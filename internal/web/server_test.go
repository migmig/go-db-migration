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
