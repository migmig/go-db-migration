package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"dbmigrator/internal/db"
	"dbmigrator/internal/security"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestRouter() *gin.Engine {
	r := gin.New()
	api := r.Group("/api")
	api.POST("/tables", getTables)
	api.POST("/migrate", startMigration)
	api.POST("/migrate/retry", retryMigration)
	api.POST("/test-target", testTargetConnection)
	api.GET("/download/:id", downloadZip)
	return r
}

func setupAuthTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	store, err := db.OpenUserStore(filepath.Join(t.TempDir(), "auth.db"))
	if err != nil {
		t.Fatalf("open user store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	hash, err := security.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	if err := store.CreateUser("alice", hash, false); err != nil {
		t.Fatalf("create user: %v", err)
	}

	sessions := newAuthSessionManager(time.Hour)

	r := gin.New()
	api := r.Group("/api")
	api.POST("/auth/login", loginHandler(store, sessions))
	api.POST("/auth/logout", logoutHandler(sessions))
	api.GET("/auth/me", meHandler(sessions))

	protected := api.Group("")
	protected.Use(requireAuth(sessions))
	protected.POST("/migrate", startMigration)

	return r
}

// ── validateMigrationRequest ─────────────────────────────────────────────────

func TestValidateMigrationRequest_ValidInput(t *testing.T) {
	req := &startMigrationRequest{
		OutFile:   "output.sql",
		Schema:    "myschema",
		BatchSize: 1000,
		Workers:   4,
	}
	if err := validateMigrationRequest(req); err != nil {
		t.Errorf("expected no error for valid input, got: %v", err)
	}
}

func TestValidateMigrationRequest_PathTraversal(t *testing.T) {
	cases := []string{"../etc/passwd", "sub/file.sql", `C:\Windows\file.sql`}
	for _, outFile := range cases {
		req := &startMigrationRequest{OutFile: outFile}
		if err := validateMigrationRequest(req); err == nil {
			t.Errorf("expected error for outFile=%q containing path separator, got nil", outFile)
		}
	}
}

func TestValidateMigrationRequest_InvalidSchema(t *testing.T) {
	cases := []string{"my schema", "schema-name", "schema;DROP", "123bad"}
	for _, schema := range cases {
		req := &startMigrationRequest{Schema: schema}
		if err := validateMigrationRequest(req); err == nil {
			t.Errorf("expected error for schema=%q, got nil", schema)
		}
	}
}

func TestValidateMigrationRequest_NegativeBatchSize(t *testing.T) {
	req := &startMigrationRequest{BatchSize: -1}
	if err := validateMigrationRequest(req); err == nil {
		t.Error("expected error for negative batchSize, got nil")
	}
}

func TestValidateMigrationRequest_NegativeWorkers(t *testing.T) {
	req := &startMigrationRequest{Workers: -1}
	if err := validateMigrationRequest(req); err == nil {
		t.Error("expected error for negative workers, got nil")
	}
}

func TestValidateMigrationRequest_NegativeDBPoolSettings(t *testing.T) {
	validReq := startMigrationRequest{
		OracleURL: "host/svc",
		Username:  "u",
		Password:  "p",
		Tables:    []string{"T1"},
	}

	req1 := validReq
	req1.DBMaxOpen = -1
	req2 := validReq
	req2.DBMaxIdle = -1
	req3 := validReq
	req3.DBMaxLife = -1

	cases := []startMigrationRequest{req1, req2, req3}

	for _, req := range cases {
		if err := validateMigrationRequest(&req); err == nil {
			t.Errorf("expected error for negative DB pool setting, got nil for %+v", req)
		}
	}
}

// ── /api/migrate validation ───────────────────────────────────────────────────

func TestStartMigration_PathTraversal_Returns400(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	body := `{"oracleUrl":"h","username":"u","password":"p","tables":["T"],"outFile":"../evil.sql"}`
	req, _ := http.NewRequest("POST", "/api/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for path traversal in outFile, got %d", w.Code)
	}
}

func TestStartMigration_InvalidSchema_Returns400(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	body := `{"oracleUrl":"h","username":"u","password":"p","tables":["T"],"schema":"bad schema!"}`
	req, _ := http.NewRequest("POST", "/api/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid schema, got %d", w.Code)
	}
}

func TestStartMigration_InvalidTableName_Returns400(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	// 악의적인 테이블명 포함
	body := `{"oracleUrl":"h","username":"u","password":"p","tables":["USERS; DROP TABLE USERS --"]}`
	req, _ := http.NewRequest("POST", "/api/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid table name, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid table name") {
		t.Errorf("expected error message to contain 'invalid table name', got: %s", w.Body.String())
	}
}

func TestStartMigration_InvalidOracleOwner_Returns400(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	// 유효하지 않은 oracleOwner
	body := `{"oracleUrl":"h","username":"u","password":"p","tables":["USERS"],"oracleOwner":"HR;--"}`
	req, _ := http.NewRequest("POST", "/api/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid oracle owner, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "invalid oracle owner") {
		t.Errorf("expected error message to contain 'invalid oracle owner', got: %s", w.Body.String())
	}
}

// ── Retry Migration ──────────────────────────────────────────────────────────

func TestRetryMigration_ValidRequest_Returns200(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	body := `{"oracleUrl":"host/svc","username":"u","password":"p","tables":["T1"]}`
	req, _ := http.NewRequest("POST", "/api/migrate/retry", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Migration started") {
		t.Errorf("expected 'Migration started' in response body, got: %s", w.Body.String())
	}
}

func TestRetryMigration_InvalidJSON(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/migrate/retry", strings.NewReader("{bad json}"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ── Test Target Connection ───────────────────────────────────────────────────

func TestTestTargetConnection_MissingFields(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/test-target", strings.NewReader(`{"targetDb":"postgres"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing targetUrl, got %d", w.Code)
	}
}

func TestTestTargetConnection_UnsupportedDB(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	body := `{"targetDb":"unsupported","targetUrl":"some_url"}`
	req, _ := http.NewRequest("POST", "/api/test-target", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unsupported DB, got %d", w.Code)
	}
}

// ── 6-2 하위 호환성: 새 필드 미포함 요청 정상 동작 ────────────────────────────────

func TestStartMigration_BackwardCompat_NoNewFields(t *testing.T) {
	r := setupTestRouter()

	// 기존 형식: v4 새 필드(outFile, perTable, schema, dryRun, logJson) 없이 요청
	w := httptest.NewRecorder()
	body := `{"oracleUrl":"host/svc","username":"u","password":"p","tables":["T1"],"batchSize":500,"workers":2}`
	req, _ := http.NewRequest("POST", "/api/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// 400이 아닌 200을 반환해야 함 (validation 통과, 백그라운드 고루틴에서 DB 연결 시도)
	if w.Code != http.StatusOK {
		t.Errorf("backward-compat request should return 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

// ── /api/tables ──────────────────────────────────────────────────────────────

func TestGetTables_InvalidJSON(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/tables", strings.NewReader("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestGetTables_MissingRequiredFields(t *testing.T) {
	r := setupTestRouter()

	// oracleUrl present but username and password missing
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/tables", strings.NewReader(`{"oracleUrl":"host/svc"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing required fields, got %d", w.Code)
	}
}

func TestGetTables_EmptyBody(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/tables", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty body, got %d", w.Code)
	}
}

func TestAuth_LoginMeLogoutFlow(t *testing.T) {
	r := setupAuthTestRouter(t)

	loginReq := httptest.NewRecorder()
	body := `{"username":"alice","password":"password123"}`
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(loginReq, req)

	if loginReq.Code != http.StatusOK {
		t.Fatalf("expected login 200, got %d (%s)", loginReq.Code, loginReq.Body.String())
	}

	cookie := loginReq.Header().Get("Set-Cookie")
	if cookie == "" {
		t.Fatal("expected session cookie on login")
	}

	meReq := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(meReq, req)

	if meReq.Code != http.StatusOK {
		t.Fatalf("expected me 200, got %d (%s)", meReq.Code, meReq.Body.String())
	}

	var me map[string]any
	if err := json.Unmarshal(meReq.Body.Bytes(), &me); err != nil {
		t.Fatalf("decode me response: %v", err)
	}
	if me["username"] != "alice" {
		t.Fatalf("expected username alice, got %v", me["username"])
	}

	logoutReq := httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/auth/logout", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(logoutReq, req)

	if logoutReq.Code != http.StatusOK {
		t.Fatalf("expected logout 200, got %d", logoutReq.Code)
	}

	meAfter := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(meAfter, req)

	if meAfter.Code != http.StatusUnauthorized {
		t.Fatalf("expected me after logout 401, got %d", meAfter.Code)
	}
}

func TestAuth_ProtectedEndpointRequiresSession(t *testing.T) {
	r := setupAuthTestRouter(t)

	w := httptest.NewRecorder()
	body := `{"oracleUrl":"h","username":"u","password":"p","tables":["T"]}`
	req, _ := http.NewRequest("POST", "/api/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session, got %d", w.Code)
	}
}

// ── /api/migrate ─────────────────────────────────────────────────────────────

func TestStartMigration_InvalidJSON(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/migrate", strings.NewReader("{not valid"))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestStartMigration_MissingRequiredFields(t *testing.T) {
	r := setupTestRouter()

	// tables field is missing
	w := httptest.NewRecorder()
	body := `{"oracleUrl":"host/svc","username":"u","password":"p"}`
	req, _ := http.NewRequest("POST", "/api/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing tables, got %d", w.Code)
	}
}

func TestStartMigration_ValidRequest_Returns200(t *testing.T) {
	r := setupTestRouter()

	// All required fields present — handler responds 200 immediately and the
	// background goroutine will fail to connect (no real DB), which is fine.
	w := httptest.NewRecorder()
	body := `{"oracleUrl":"host/svc","username":"u","password":"p","tables":["T1"]}`
	req, _ := http.NewRequest("POST", "/api/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Migration started") {
		t.Errorf("expected 'Migration started' in response body, got: %s", w.Body.String())
	}
}

// ── /api/download/:id ────────────────────────────────────────────────────────

func TestDownloadZip_NotFound(t *testing.T) {
	r := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/download/migration_99999999999999.zip", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDownloadZip_PathTraversal_IsNeutralised(t *testing.T) {
	r := setupTestRouter()

	// filepath.Base inside downloadZip strips path components
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/download/../../../etc/passwd", nil)
	r.ServeHTTP(w, req)

	// Must NOT return 200 (file contents served from an arbitrary path)
	if w.Code == http.StatusOK {
		t.Errorf("path traversal must not return 200, got %d", w.Code)
	}
}

func TestDownloadZip_ExistingFile(t *testing.T) {
	// Write a temp file into os.TempDir() so the handler can find it
	f, err := os.CreateTemp(os.TempDir(), "migration_test_*.zip")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_, _ = f.WriteString("PK fake zip content")
	f.Close()
	defer os.Remove(f.Name())

	r := setupTestRouter()

	filename := filepath.Base(f.Name())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/download/"+filename, nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for existing file, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Disposition"); !strings.Contains(ct, filename) {
		t.Errorf("expected Content-Disposition to contain filename, got: %s", ct)
	}
}

func TestDownloadZip_EmptyID(t *testing.T) {
	// Gin's router requires a non-empty param; hitting /api/download/ (no id)
	// won't match the route and returns 301/404 from the router, which is fine.
	// This test focuses on the "." or "/" id branch inside the handler.
	r := setupTestRouter()

	// The param ":id" will be "." after the route matches, but Gin won't match
	// an empty segment — test with a dot-only name that survives URL parsing.
	// filepath.Base(".") == "." so the handler should return 400.
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/download/.", nil)
	r.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		t.Errorf("dot-only id should not return 200, got %d", w.Code)
	}
}
