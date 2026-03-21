package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dbmigrator/internal/migration"

	"github.com/gin-gonic/gin"
)

func setupPrecheckRouter() *gin.Engine {
	r := gin.New()
	api := r.Group("/api")
	api.POST("/migrations/precheck", precheckHandler(nil))
	api.GET("/migrations/precheck/results", precheckResultsHandler())
	return r
}

func seedPrecheckStore(t *testing.T, results []migration.PrecheckTableResult) {
	t.Helper()
	s := buildPrecheckSummaryFromResults(results)
	globalPrecheckStore.set(results, s)
}

func buildPrecheckSummaryFromResults(results []migration.PrecheckTableResult) migration.PrecheckSummary {
	s := migration.PrecheckSummary{TotalTables: len(results)}
	for _, r := range results {
		switch r.Decision {
		case migration.DecisionTransferRequired:
			s.TransferRequiredCount++
		case migration.DecisionSkipCandidate:
			s.SkipCandidateCount++
		case migration.DecisionCountCheckFailed:
			s.CountCheckFailedCount++
		}
	}
	return s
}

// --- POST /api/migrations/precheck 검증 오류 테스트 ---

func TestPrecheckHandler_InvalidPolicy(t *testing.T) {
	r := setupPrecheckRouter()
	body := `{"oracleUrl":"oracle://x","username":"u","password":"p","tables":["EMP"],"policy":"invalid"}`
	req, _ := http.NewRequest("POST", "/api/migrations/precheck", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if !strings.Contains(resp["error"], "invalid policy") {
		t.Errorf("expected error about invalid policy, got %q", resp["error"])
	}
}

func TestPrecheckHandler_MissingRequiredFields(t *testing.T) {
	r := setupPrecheckRouter()
	body := `{"oracleUrl":"oracle://x","username":"u"}` // password and tables missing
	req, _ := http.NewRequest("POST", "/api/migrations/precheck", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing required fields, got %d", w.Code)
	}
}

func TestPrecheckHandler_EmptyTables(t *testing.T) {
	r := setupPrecheckRouter()
	body := `{"oracleUrl":"oracle://x","username":"u","password":"p","tables":[]}`
	req, _ := http.NewRequest("POST", "/api/migrations/precheck", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty tables, got %d", w.Code)
	}
}

// --- GET /api/migrations/precheck/results 필터링/페이지네이션 테스트 ---

func TestPrecheckResultsHandler_Empty(t *testing.T) {
	globalPrecheckStore.set(nil, migration.PrecheckSummary{})

	r := setupPrecheckRouter()
	req, _ := http.NewRequest("GET", "/api/migrations/precheck/results", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	total, _ := resp["total"].(float64)
	if total != 0 {
		t.Errorf("expected total 0, got %v", total)
	}
}

func TestPrecheckResultsHandler_DecisionFilter(t *testing.T) {
	fixtures := []migration.PrecheckTableResult{
		{TableName: "EMP", Decision: migration.DecisionTransferRequired, CheckedAt: time.Now()},
		{TableName: "DEPT", Decision: migration.DecisionSkipCandidate, CheckedAt: time.Now()},
		{TableName: "PROJ", Decision: migration.DecisionCountCheckFailed, Reason: "timeout", CheckedAt: time.Now()},
		{TableName: "SAL", Decision: migration.DecisionTransferRequired, CheckedAt: time.Now()},
	}
	seedPrecheckStore(t, fixtures)

	r := setupPrecheckRouter()

	tests := []struct {
		decision string
		want     int
	}{
		{"all", 4},
		{"", 4},
		{"transfer_required", 2},
		{"skip_candidate", 1},
		{"count_check_failed", 1},
	}

	for _, tc := range tests {
		url := "/api/migrations/precheck/results"
		if tc.decision != "" {
			url += "?decision=" + tc.decision
		}
		req, _ := http.NewRequest("GET", url, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("decision=%q: expected 200, got %d", tc.decision, w.Code)
			continue
		}
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		total, _ := resp["total"].(float64)
		if int(total) != tc.want {
			t.Errorf("decision=%q: expected total %d, got %v", tc.decision, tc.want, total)
		}
	}
}

func TestPrecheckResultsHandler_InvalidDecisionFilter(t *testing.T) {
	r := setupPrecheckRouter()
	req, _ := http.NewRequest("GET", "/api/migrations/precheck/results?decision=bad_value", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid decision, got %d", w.Code)
	}
}

func TestPrecheckResultsHandler_SearchFilter(t *testing.T) {
	fixtures := []migration.PrecheckTableResult{
		{TableName: "EMP_DETAIL", Decision: migration.DecisionTransferRequired},
		{TableName: "DEPT", Decision: migration.DecisionSkipCandidate},
		{TableName: "EMP_SUMMARY", Decision: migration.DecisionTransferRequired},
	}
	seedPrecheckStore(t, fixtures)

	r := setupPrecheckRouter()
	req, _ := http.NewRequest("GET", "/api/migrations/precheck/results?search=EMP", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	total, _ := resp["total"].(float64)
	if int(total) != 2 {
		t.Errorf("expected 2 results matching 'EMP', got %v", total)
	}
}

func TestPrecheckResultsHandler_Pagination(t *testing.T) {
	fixtures := make([]migration.PrecheckTableResult, 15)
	for i := range fixtures {
		fixtures[i] = migration.PrecheckTableResult{
			TableName: fmt.Sprintf("TABLE_%02d", i),
			Decision:  migration.DecisionTransferRequired,
		}
	}
	seedPrecheckStore(t, fixtures)

	r := setupPrecheckRouter()

	// page 1, page_size 10 → 10 items, total 15
	req, _ := http.NewRequest("GET", "/api/migrations/precheck/results?page=1&page_size=10", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	total, _ := resp["total"].(float64)
	if int(total) != 15 {
		t.Errorf("expected total 15, got %v", total)
	}
	items, _ := resp["items"].([]interface{})
	if len(items) != 10 {
		t.Errorf("expected 10 items on page 1, got %d", len(items))
	}

	// page 2, page_size 10 → 5 items
	req2, _ := http.NewRequest("GET", "/api/migrations/precheck/results?page=2&page_size=10", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var resp2 map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &resp2)
	items2, _ := resp2["items"].([]interface{})
	if len(items2) != 5 {
		t.Errorf("expected 5 items on page 2, got %d", len(items2))
	}

	// page beyond end → empty items
	req3, _ := http.NewRequest("GET", "/api/migrations/precheck/results?page=99&page_size=10", nil)
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	var resp3 map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &resp3)
	items3 := resp3["items"]
	if items3 != nil {
		if arr, ok := items3.([]interface{}); ok && len(arr) > 0 {
			t.Errorf("expected empty items for out-of-range page, got %d items", len(arr))
		}
	}
}

func TestPrecheckResultsHandler_SummaryReflected(t *testing.T) {
	fixtures := []migration.PrecheckTableResult{
		{TableName: "A", Decision: migration.DecisionTransferRequired},
		{TableName: "B", Decision: migration.DecisionTransferRequired},
		{TableName: "C", Decision: migration.DecisionSkipCandidate},
	}
	seedPrecheckStore(t, fixtures)

	r := setupPrecheckRouter()
	req, _ := http.NewRequest("GET", "/api/migrations/precheck/results", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	summary, _ := resp["summary"].(map[string]interface{})
	if summary == nil {
		t.Fatal("expected summary in response")
	}
	if v, _ := summary["transfer_required_count"].(float64); int(v) != 2 {
		t.Errorf("expected transfer_required_count 2, got %v", v)
	}
	if v, _ := summary["skip_candidate_count"].(float64); int(v) != 1 {
		t.Errorf("expected skip_candidate_count 1, got %v", v)
	}
}

// --- precheckEnabled 테스트 ---

func TestPrecheckEnabled_DefaultTrue(t *testing.T) {
	t.Setenv("DBM_PRECHECK_ENABLED", "")
	if !precheckEnabled() {
		t.Error("expected precheck enabled by default")
	}
}

func TestPrecheckEnabled_DisabledByEnv(t *testing.T) {
	t.Setenv("DBM_PRECHECK_ENABLED", "false")
	if precheckEnabled() {
		t.Error("expected precheck disabled when DBM_PRECHECK_ENABLED=false")
	}
}

func TestPrecheckEnabled_InvalidEnv(t *testing.T) {
	t.Setenv("DBM_PRECHECK_ENABLED", "not_a_bool")
	if !precheckEnabled() {
		t.Error("expected default true on invalid env value")
	}
}

// --- validatePrecheckPolicy / validatePrecheckDecision 테스트 ---

func TestValidatePrecheckPolicy(t *testing.T) {
	valid := []string{"strict", "best_effort", "skip_equal_rows", ""}
	for _, p := range valid {
		if !validatePrecheckPolicy(p) {
			t.Errorf("expected valid policy %q", p)
		}
	}
	invalid := []string{"unknown", "STRICT", "fail_closed"}
	for _, p := range invalid {
		if validatePrecheckPolicy(p) {
			t.Errorf("expected invalid policy %q", p)
		}
	}
}

func TestValidatePrecheckDecision(t *testing.T) {
	valid := []string{"all", "", "transfer_required", "skip_candidate", "count_check_failed"}
	for _, d := range valid {
		if !validatePrecheckDecision(d) {
			t.Errorf("expected valid decision %q", d)
		}
	}
	invalid := []string{"bad", "TRANSFER_REQUIRED", "unknown"}
	for _, d := range invalid {
		if validatePrecheckDecision(d) {
			t.Errorf("expected invalid decision %q", d)
		}
	}
}
