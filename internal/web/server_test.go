package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	api.GET("/download/:id", downloadZip)
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
