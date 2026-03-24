package web

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGetTables_ConnectOracleError(t *testing.T) {
	r := gin.New()
	r.POST("/tables", getTables(nil))

	body := []byte(`{"oracleUrl":"invalid-url","username":"user","password":"pwd"}`)
	req, _ := http.NewRequest("POST", "/tables", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestTestTargetConnection_ConnectPostgresError(t *testing.T) {
	r := gin.New()
	r.POST("/test-target", testTargetConnection(nil))

	body := []byte(`{"targetDb":"postgres","targetUrl":"postgres://invalid-url"}`)
	req, _ := http.NewRequest("POST", "/test-target", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestStartMigrationHandler_MissingSession(t *testing.T) {
	r := gin.New()
	r.POST("/migrate", startMigration)

	// Direct = false, so it tries to use WebSocket, which needs SessionID
	body := []byte(`{"oracleUrl":"url","username":"user","password":"pwd","tables":["T1"],"direct":false}`)
	req, _ := http.NewRequest("POST", "/migrate", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for missing session, got %d", w.Code)
	}
}

func TestTableHistoryEnabled(t *testing.T) {
	t.Setenv("DBM_TABLE_HISTORY_ENABLED", "false")
	if tableHistoryEnabled() {
		t.Error("expected false")
	}
	t.Setenv("DBM_TABLE_HISTORY_ENABLED", "true")
	if !tableHistoryEnabled() {
		t.Error("expected true")
	}
}

func TestIsPermissionError(t *testing.T) {
	if !isPermissionError(errors.New("permission denied")) {
		t.Error("should be permission error")
	}
	if !isPermissionError(errors.New("ORA-01031")) {
		t.Error("should be ORA-01031")
	}
	if !isPermissionError(errors.New("insufficient privileges")) {
		t.Error("should be insufficient")
	}
	if isPermissionError(errors.New("other")) {
		t.Error("should not be permission error")
	}
}

func TestMaskedURL(t *testing.T) {
	if maskedURL("") != "" {
		t.Error("expected empty")
	}
	u := "oracle://user:pass@host:1521/sid"
	m := maskedURL(u)
	if strings.Contains(m, "pass") {
		t.Errorf("password not masked: %s", m)
	}
}

func TestDownloadReport_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/report/:id", downloadReport)

	req, _ := http.NewRequest("GET", "/report/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestIsSecureRequest(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	if isSecureRequest(req) {
		t.Error("expected false")
	}
	req.Header.Set("X-Forwarded-Proto", "https")
	if !isSecureRequest(req) {
		t.Error("expected true")
	}
}

func TestParseInt64Param_Errors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Params = gin.Params{gin.Param{Key: "id", Value: "abc"}}
	val, ok := parseInt64Param(c, "id")
	if ok || val != 0 {
		t.Error("expected failure for non-int")
	}
	if w.Code != http.StatusBadRequest {
		t.Error("expected 400")
	}
	
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "-1"}}
	val, ok = parseInt64Param(c, "id")
	if ok {
		t.Error("expected failure for negative")
	}
}

func TestListTableSummariesHandler(t *testing.T) {
	defer func() { recover() }()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	th := &TableHistoryStore{runs: make(map[string][]TableMigrationHistory)}
	r.GET("/history", func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Next()
	}, listTableSummariesHandler(th, nil))

	req, _ := http.NewRequest("GET", "/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
}

func TestGetTableHistoryHandler(t *testing.T) {
	defer func() { recover() }()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	th := &TableHistoryStore{runs: make(map[string][]TableMigrationHistory)}
	r.GET("/history/:table", func(c *gin.Context) {
		c.Set("user_id", int64(1))
		c.Next()
	}, getTableHistoryHandler(th))

	req, _ := http.NewRequest("GET", "/history/T1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
}

func TestHandleMigration_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	c.Request, _ = http.NewRequest("POST", "/migrate", strings.NewReader(`{invalid`))
	handleMigration(c, false, nil, nil)
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestTargetTablesHandler_ConnectError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/target-tables", targetTablesHandler(nil))

	body := []byte(`{"targetUrl":"postgres://invalid","schema":"public"}`)
	req, _ := http.NewRequest("POST", "/target-tables", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
