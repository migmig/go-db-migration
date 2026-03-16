package web

import (
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
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

const testWebMasterKey = "0123456789abcdef0123456789abcdef"

func setupTestRouter() *gin.Engine {
	r := gin.New()
	registerV16Routes(r)
	api := r.Group("/api")
	api.POST("/tables", getTables)
	api.POST("/migrate", startMigration)
	api.POST("/migrate/retry", retryMigration)
	api.POST("/test-target", testTargetConnection)
	api.GET("/download/:id", downloadZip)
	return r
}

func TestRegisterV16Routes_ServesEmbeddedIndexAndFallback(t *testing.T) {
	r := gin.New()
	registerV16Routes(r)

	for _, route := range []string{"/v16", "/v16/some/deep/link"} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", route, nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 for %s, got %d", route, w.Code)
		}
		if !strings.Contains(w.Body.String(), "v16 frontend bundle") {
			t.Fatalf("expected embedded placeholder content for %s, got %q", route, w.Body.String())
		}
	}
}

func TestRootRedirectsToV16(t *testing.T) {
	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/v16")
	})
	r.HEAD("/", func(c *gin.Context) {
		c.Redirect(http.StatusTemporaryRedirect, "/v16")
	})

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(method, "/", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusTemporaryRedirect {
			t.Fatalf("expected 307 for %s, got %d", method, w.Code)
		}
		if got := w.Header().Get("Location"); got != "/v16" {
			t.Fatalf("expected redirect to /v16 for %s, got %q", method, got)
		}
	}
}

func setupAuthTestRouter(t *testing.T) (*gin.Engine, *db.UserStore) {
	r, store, _ := setupAuthTestRouterWithOptions(t, time.Hour, 24*time.Hour)
	return r, store
}

func setupAuthTestRouterWithMetrics(t *testing.T) (*gin.Engine, *db.UserStore, *monitoringMetrics) {
	return setupAuthTestRouterWithOptions(t, time.Hour, 24*time.Hour)
}

func setupAuthTestRouterWithOptions(t *testing.T, idleTTL, absoluteTTL time.Duration) (*gin.Engine, *db.UserStore, *monitoringMetrics) {
	t.Helper()
	store, err := db.OpenAuthStore(filepath.Join(t.TempDir(), "auth.db"), testWebMasterKey)
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
	if err := store.CreateUser("bob", hash, false); err != nil {
		t.Fatalf("create user: %v", err)
	}

	metrics := newMonitoringMetrics()
	sessions := newAuthSessionManager(idleTTL, absoluteTTL, metrics)

	r := gin.New()
	api := r.Group("/api")
	api.POST("/auth/login", loginHandler(store, sessions))
	api.POST("/auth/logout", logoutHandler(sessions))
	api.GET("/auth/me", meHandler(sessions))

	protected := api.Group("")
	protected.Use(requireAuth(sessions))
	credentials := protected.Group("/credentials")
	credentials.Use(monitoringAPIErrorsMiddleware(metrics, monitoredAPICredentials))
	credentials.GET("", listCredentialsHandler(store))
	credentials.POST("", createCredentialHandler(store))
	credentials.PUT("/:id", updateCredentialHandler(store))
	credentials.DELETE("/:id", deleteCredentialHandler(store))

	history := protected.Group("/history")
	history.Use(monitoringAPIErrorsMiddleware(metrics, monitoredAPIHistory))
	history.GET("", listHistoryHandler(store))
	history.GET("/:id", getHistoryHandler(store))
	history.POST("/:id/replay", replayHistoryHandler(store))

	protected.GET("/monitoring/metrics", monitoringMetricsHandler(metrics))
	protected.POST("/migrate", startMigrationHandler(store))

	return r, store, metrics
}

func loginAs(t *testing.T, r *gin.Engine, username, password string) string {
	t.Helper()
	loginReq := httptest.NewRecorder()
	body := `{"username":"` + username + `","password":"` + password + `"}`
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(loginReq, req)

	if loginReq.Code != http.StatusOK {
		t.Fatalf("expected login 200 for %s, got %d (%s)", username, loginReq.Code, loginReq.Body.String())
	}

	cookie := loginReq.Header().Get("Set-Cookie")
	if cookie == "" {
		t.Fatalf("expected session cookie for %s", username)
	}
	return cookie
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
	r, _ := setupAuthTestRouter(t)
	cookie := loginAs(t, r, "alice", "password123")

	meReq := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/auth/me", nil)
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
	r, _ := setupAuthTestRouter(t)

	w := httptest.NewRecorder()
	body := `{"oracleUrl":"h","username":"u","password":"p","tables":["T"]}`
	req, _ := http.NewRequest("POST", "/api/migrate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 without session, got %d", w.Code)
	}
}

func TestAuth_CredentialsAreScopedPerUser(t *testing.T) {
	r, store := setupAuthTestRouter(t)
	aliceCookie := loginAs(t, r, "alice", "password123")
	bobCookie := loginAs(t, r, "bob", "password123")

	bob := mustGetUser(t, store, "bob")
	_, err := store.CreateCredential(bob.ID, db.Credential{
		Alias:        "bob-main",
		DBType:       "postgres",
		Host:         "bob-db",
		Username:     "bob",
		Password:     "bob-secret",
		DatabaseName: "bobdb",
	})
	if err != nil {
		t.Fatalf("seed bob credential: %v", err)
	}

	createReq := httptest.NewRecorder()
	body := `{"alias":"alice-main","dbType":"postgres","host":"localhost","port":5432,"databaseName":"appdb","username":"alice","password":"secret"}`
	req, _ := http.NewRequest("POST", "/api/credentials", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", aliceCookie)
	r.ServeHTTP(createReq, req)

	if createReq.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d (%s)", createReq.Code, createReq.Body.String())
	}

	var created db.Credential
	if err := json.Unmarshal(createReq.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create credential response: %v", err)
	}
	if created.Password != "secret" {
		t.Fatalf("expected decrypted password in response, got %q", created.Password)
	}

	listAlice := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/credentials", nil)
	req.Header.Set("Cookie", aliceCookie)
	r.ServeHTTP(listAlice, req)

	if listAlice.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listAlice.Code)
	}

	var aliceResp struct {
		Items []db.Credential `json:"items"`
	}
	if err := json.Unmarshal(listAlice.Body.Bytes(), &aliceResp); err != nil {
		t.Fatalf("decode alice credential list: %v", err)
	}
	if len(aliceResp.Items) != 1 || aliceResp.Items[0].Alias != "alice-main" {
		t.Fatalf("unexpected alice credentials: %+v", aliceResp.Items)
	}

	listBob := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/credentials", nil)
	req.Header.Set("Cookie", bobCookie)
	r.ServeHTTP(listBob, req)

	if listBob.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listBob.Code)
	}

	var bobResp struct {
		Items []db.Credential `json:"items"`
	}
	if err := json.Unmarshal(listBob.Body.Bytes(), &bobResp); err != nil {
		t.Fatalf("decode bob credential list: %v", err)
	}
	if len(bobResp.Items) != 1 || bobResp.Items[0].Alias != "bob-main" {
		t.Fatalf("unexpected bob credentials: %+v", bobResp.Items)
	}

	deleteReq := httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/credentials/"+strconv.FormatInt(created.ID, 10), nil)
	req.Header.Set("Cookie", bobCookie)
	r.ServeHTTP(deleteReq, req)

	if deleteReq.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-user delete, got %d", deleteReq.Code)
	}

}

func TestAuth_HistoryPaginationAndReplayAreScopedPerUser(t *testing.T) {
	r, store := setupAuthTestRouter(t)
	aliceCookie := loginAs(t, r, "alice", "password123")
	bobCookie := loginAs(t, r, "bob", "password123")

	alice := mustGetUser(t, store, "alice")
	bob := mustGetUser(t, store, "bob")

	firstID, err := store.InsertHistory(alice.ID, db.HistoryEntry{
		Status:        "success",
		SourceSummary: "alice@oracle",
		TargetSummary: "postgres://local",
		OptionsJSON:   `{"tables":["USERS"],"schema":"public"}`,
		LogSummary:    "rows=10",
	})
	if err != nil {
		t.Fatalf("seed alice history 1: %v", err)
	}
	_, err = store.InsertHistory(alice.ID, db.HistoryEntry{
		Status:        "failed",
		SourceSummary: "alice@oracle",
		TargetSummary: "postgres://local",
		OptionsJSON:   `{"tables":["ORDERS"],"schema":"audit"}`,
		LogSummary:    "rows=0",
	})
	if err != nil {
		t.Fatalf("seed alice history 2: %v", err)
	}
	bobHistoryID, err := store.InsertHistory(bob.ID, db.HistoryEntry{
		Status:        "success",
		SourceSummary: "bob@oracle",
		TargetSummary: "mysql://target",
		OptionsJSON:   `{"tables":["PAYMENTS"]}`,
	})
	if err != nil {
		t.Fatalf("seed bob history: %v", err)
	}

	listReq := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/history?page=1&pageSize=1", nil)
	req.Header.Set("Cookie", aliceCookie)
	r.ServeHTTP(listReq, req)

	if listReq.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", listReq.Code, listReq.Body.String())
	}

	var listResp struct {
		Items    []db.HistoryEntry `json:"items"`
		Page     int               `json:"page"`
		PageSize int               `json:"pageSize"`
		Total    int               `json:"total"`
	}
	if err := json.Unmarshal(listReq.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode history list: %v", err)
	}
	if listResp.Total != 2 || len(listResp.Items) != 1 {
		t.Fatalf("unexpected history pagination response: %+v", listResp)
	}

	replayReq := httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/history/"+strconv.FormatInt(firstID, 10)+"/replay", nil)
	req.Header.Set("Cookie", aliceCookie)
	r.ServeHTTP(replayReq, req)

	if replayReq.Code != http.StatusOK {
		t.Fatalf("expected 200 replay, got %d (%s)", replayReq.Code, replayReq.Body.String())
	}

	var replayResp struct {
		History db.HistoryEntry `json:"history"`
		Payload map[string]any  `json:"payload"`
	}
	if err := json.Unmarshal(replayReq.Body.Bytes(), &replayResp); err != nil {
		t.Fatalf("decode replay response: %v", err)
	}
	if replayResp.Payload["schema"] != "public" {
		t.Fatalf("expected replay schema public, got %v", replayResp.Payload["schema"])
	}

	crossUserReq := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/history/"+strconv.FormatInt(bobHistoryID, 10), nil)
	req.Header.Set("Cookie", aliceCookie)
	r.ServeHTTP(crossUserReq, req)

	if crossUserReq.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-user history access, got %d", crossUserReq.Code)
	}

	bobList := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/history", nil)
	req.Header.Set("Cookie", bobCookie)
	r.ServeHTTP(bobList, req)

	if bobList.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", bobList.Code)
	}
}

func TestMonitoring_LoginFailureAndSessionExpiration(t *testing.T) {
	r, _, metrics := setupAuthTestRouterWithOptions(t, 2*time.Millisecond, time.Hour)

	loginFail := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(`{"username":"alice","password":"wrong"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(loginFail, req)
	if loginFail.Code != http.StatusUnauthorized {
		t.Fatalf("expected login failure 401, got %d", loginFail.Code)
	}

	cookie := loginAs(t, r, "alice", "password123")
	time.Sleep(5 * time.Millisecond)

	expiredReq := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/auth/me", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(expiredReq, req)
	if expiredReq.Code != http.StatusUnauthorized {
		t.Fatalf("expected expired session 401, got %d", expiredReq.Code)
	}

	snapshot := metrics.snapshot()
	if snapshot.Login.Attempts != 2 || snapshot.Login.Failures != 1 {
		t.Fatalf("unexpected login metrics: %+v", snapshot.Login)
	}
	if !nearlyEqual(snapshot.Login.FailureRatePct, 50.0) {
		t.Fatalf("expected login failure rate 50.0, got %.2f", snapshot.Login.FailureRatePct)
	}
	if snapshot.Session.Checks != 1 || snapshot.Session.Expired != 1 {
		t.Fatalf("unexpected session metrics: %+v", snapshot.Session)
	}
	if !nearlyEqual(snapshot.Session.ExpirationRatePct, 100.0) {
		t.Fatalf("expected session expiration rate 100.0, got %.2f", snapshot.Session.ExpirationRatePct)
	}
}

func TestMonitoring_CredentialsAndHistoryAPIErrorRates(t *testing.T) {
	r, _, _ := setupAuthTestRouterWithMetrics(t)
	cookie := loginAs(t, r, "alice", "password123")

	createReq := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/credentials", strings.NewReader(`{"alias":"main","dbType":"postgres","host":"localhost","port":5432,"databaseName":"appdb","username":"alice","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(createReq, req)
	if createReq.Code != http.StatusCreated {
		t.Fatalf("expected credential create 201, got %d (%s)", createReq.Code, createReq.Body.String())
	}

	credentialErr := httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/credentials/999999", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(credentialErr, req)
	if credentialErr.Code != http.StatusNotFound {
		t.Fatalf("expected credential delete miss 404, got %d", credentialErr.Code)
	}

	historyErr := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/history/999999", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(historyErr, req)
	if historyErr.Code != http.StatusNotFound {
		t.Fatalf("expected history miss 404, got %d", historyErr.Code)
	}

	historyOK := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/history?page=1&pageSize=20", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(historyOK, req)
	if historyOK.Code != http.StatusOK {
		t.Fatalf("expected history list 200, got %d", historyOK.Code)
	}

	metricsRes := httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/monitoring/metrics", nil)
	req.Header.Set("Cookie", cookie)
	r.ServeHTTP(metricsRes, req)
	if metricsRes.Code != http.StatusOK {
		t.Fatalf("expected monitoring endpoint 200, got %d (%s)", metricsRes.Code, metricsRes.Body.String())
	}

	var snapshot monitoringSnapshot
	if err := json.Unmarshal(metricsRes.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("decode monitoring snapshot: %v", err)
	}

	if snapshot.Credentials.Requests != 2 || snapshot.Credentials.Errors != 1 {
		t.Fatalf("unexpected credential API metrics: %+v", snapshot.Credentials)
	}
	if !nearlyEqual(snapshot.Credentials.ErrorRatePct, 50.0) {
		t.Fatalf("expected credential API error rate 50.0, got %.2f", snapshot.Credentials.ErrorRatePct)
	}
	if snapshot.History.Requests != 2 || snapshot.History.Errors != 1 {
		t.Fatalf("unexpected history API metrics: %+v", snapshot.History)
	}
	if !nearlyEqual(snapshot.History.ErrorRatePct, 50.0) {
		t.Fatalf("expected history API error rate 50.0, got %.2f", snapshot.History.ErrorRatePct)
	}
}

func mustGetUser(t *testing.T, store *db.UserStore, username string) *db.User {
	t.Helper()
	user, err := store.GetUserByUsername(username)
	if err != nil {
		t.Fatalf("get user %s: %v", username, err)
	}
	return user
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

func nearlyEqual(got, want float64) bool {
	return math.Abs(got-want) <= 0.0001
}
