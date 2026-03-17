package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// ── Unit Tests: TableHistoryStore ─────────────────────────────────────────────

func TestTableHistoryStore_RecordAndGetHistory(t *testing.T) {
	store := &TableHistoryStore{runs: make(map[string][]TableMigrationHistory)}
	now := time.Now()

	h1 := TableMigrationHistory{
		RunID:         "run1_USERS",
		TableName:     "USERS",
		Status:        "success",
		StartedAt:     now,
		FinishedAt:    now.Add(time.Second),
		DurationMs:    1000,
		RowsProcessed: 100,
	}
	h2 := TableMigrationHistory{
		RunID:        "run2_USERS",
		TableName:    "USERS",
		Status:       "failed",
		StartedAt:    now.Add(2 * time.Second),
		FinishedAt:   now.Add(3 * time.Second),
		DurationMs:   500,
		ErrorMessage: "duplicate key",
	}

	store.RecordTableRun(h1)
	store.RecordTableRun(h2)

	history, ok := store.GetHistory("USERS", 10)
	if !ok {
		t.Fatal("expected history to exist for USERS")
	}
	if len(history) != 2 {
		t.Fatalf("expected 2 history entries, got %d", len(history))
	}
	// Latest first
	if history[0].RunID != h2.RunID {
		t.Errorf("expected latest run first, got %s", history[0].RunID)
	}
}

func TestTableHistoryStore_GetHistory_LimitRespected(t *testing.T) {
	store := &TableHistoryStore{runs: make(map[string][]TableMigrationHistory)}
	for i := range 5 {
		store.RecordTableRun(TableMigrationHistory{
			RunID:     fmt.Sprintf("run%d", i),
			TableName: "ORDERS",
			Status:    "success",
			StartedAt: time.Now(),
		})
	}

	history, ok := store.GetHistory("ORDERS", 3)
	if !ok {
		t.Fatal("expected history for ORDERS")
	}
	if len(history) != 3 {
		t.Errorf("expected 3 entries with limit=3, got %d", len(history))
	}
}

func TestTableHistoryStore_GetHistory_NotFound(t *testing.T) {
	store := &TableHistoryStore{runs: make(map[string][]TableMigrationHistory)}
	_, ok := store.GetHistory("NONEXISTENT", 10)
	if ok {
		t.Error("expected false for non-existent table")
	}
}

func TestTableHistoryStore_MaxHistoryPerTable(t *testing.T) {
	store := &TableHistoryStore{runs: make(map[string][]TableMigrationHistory)}
	for i := range maxHistoryPerTable + 10 {
		store.RecordTableRun(TableMigrationHistory{
			RunID:     fmt.Sprintf("run%d", i),
			TableName: "BIG_TABLE",
			Status:    "success",
			StartedAt: time.Now(),
		})
	}

	history, _ := store.GetHistory("BIG_TABLE", maxHistoryPerTable+10)
	if len(history) > maxHistoryPerTable {
		t.Errorf("expected at most %d entries, got %d", maxHistoryPerTable, len(history))
	}
}

// ── Unit Tests: ListSummaries ─────────────────────────────────────────────────

func newTestStore() *TableHistoryStore {
	store := &TableHistoryStore{runs: make(map[string][]TableMigrationHistory)}
	now := time.Now()

	store.RecordTableRun(TableMigrationHistory{
		RunID:     "r1",
		TableName: "USERS",
		Status:    "success",
		StartedAt: now,
		FinishedAt: now.Add(time.Second),
		DurationMs: 1000,
	})
	store.RecordTableRun(TableMigrationHistory{
		RunID:        "r2",
		TableName:    "ORDERS",
		Status:       "failed",
		StartedAt:    now,
		FinishedAt:   now.Add(2 * time.Second),
		DurationMs:   2000,
		ErrorMessage: "constraint violation",
	})
	store.RecordTableRun(TableMigrationHistory{
		RunID:     "r3",
		TableName: "PRODUCTS",
		Status:    "success",
		StartedAt: now,
		FinishedAt: now.Add(500 * time.Millisecond),
		DurationMs: 500,
	})
	return store
}

func TestListSummaries_NoFilter(t *testing.T) {
	store := newTestStore()
	items, total := store.ListSummaries(TableSummaryFilter{PageSize: 20})
	if total != 3 {
		t.Errorf("expected total=3, got %d", total)
	}
	if len(items) != 3 {
		t.Errorf("expected 3 items, got %d", len(items))
	}
}

func TestListSummaries_FilterByStatus_Failed(t *testing.T) {
	store := newTestStore()
	items, total := store.ListSummaries(TableSummaryFilter{Status: "failed", PageSize: 20})
	if total != 1 {
		t.Errorf("expected total=1 for failed filter, got %d", total)
	}
	if len(items) != 1 || items[0].TableName != "ORDERS" {
		t.Errorf("expected ORDERS in failed filter results, got %+v", items)
	}
}

func TestListSummaries_FilterByStatus_Success(t *testing.T) {
	store := newTestStore()
	items, total := store.ListSummaries(TableSummaryFilter{Status: "success", PageSize: 20})
	if total != 2 {
		t.Errorf("expected total=2 for success filter, got %d", total)
	}
	for _, item := range items {
		if item.Status != "success" {
			t.Errorf("expected success status, got %s", item.Status)
		}
	}
}

func TestListSummaries_ExcludeSuccess(t *testing.T) {
	store := newTestStore()
	items, total := store.ListSummaries(TableSummaryFilter{ExcludeSuccess: true, PageSize: 20})
	if total != 1 {
		t.Errorf("expected 1 non-success item, got %d", total)
	}
	if len(items) != 1 || items[0].Status == "success" {
		t.Errorf("expected no success items, got %+v", items)
	}
}

func TestListSummaries_SearchFilter(t *testing.T) {
	store := newTestStore()
	items, total := store.ListSummaries(TableSummaryFilter{Search: "user", PageSize: 20})
	if total != 1 {
		t.Errorf("expected 1 item for search=user, got %d", total)
	}
	if len(items) != 1 || !strings.EqualFold(items[0].TableName, "USERS") {
		t.Errorf("expected USERS in search results, got %+v", items)
	}
}

func TestListSummaries_Pagination(t *testing.T) {
	store := newTestStore()
	items, total := store.ListSummaries(TableSummaryFilter{Page: 1, PageSize: 2})
	if total != 3 {
		t.Errorf("expected total=3, got %d", total)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 items on page 1, got %d", len(items))
	}

	items2, _ := store.ListSummaries(TableSummaryFilter{Page: 2, PageSize: 2})
	if len(items2) != 1 {
		t.Errorf("expected 1 item on page 2, got %d", len(items2))
	}
}

func TestListSummaries_SortByTableName(t *testing.T) {
	store := newTestStore()
	items, _ := store.ListSummaries(TableSummaryFilter{Sort: "table_name", Order: "asc", PageSize: 20})
	if len(items) < 2 {
		t.Fatal("need at least 2 items")
	}
	for i := 1; i < len(items); i++ {
		if items[i].TableName < items[i-1].TableName {
			t.Errorf("expected ascending order, got %s before %s", items[i-1].TableName, items[i].TableName)
		}
	}
}

func TestListSummaries_SortByTableName_Desc(t *testing.T) {
	store := newTestStore()
	items, _ := store.ListSummaries(TableSummaryFilter{Sort: "table_name", Order: "desc", PageSize: 20})
	if len(items) < 2 {
		t.Fatal("need at least 2 items")
	}
	for i := 1; i < len(items); i++ {
		if items[i].TableName > items[i-1].TableName {
			t.Errorf("expected descending order, got %s before %s", items[i-1].TableName, items[i].TableName)
		}
	}
}

func TestListSummaries_PageBeyondEnd(t *testing.T) {
	store := newTestStore()
	items, total := store.ListSummaries(TableSummaryFilter{Page: 100, PageSize: 20})
	if total != 3 {
		t.Errorf("expected total=3, got %d", total)
	}
	if len(items) != 0 {
		t.Errorf("expected 0 items for out-of-range page, got %d", len(items))
	}
}

// ── Unit Tests: ValidateTableSummaryFilter ────────────────────────────────────

func TestValidateTableSummaryFilter_ValidStatus(t *testing.T) {
	for _, status := range []string{"not_started", "running", "success", "failed", ""} {
		f := &TableSummaryFilter{Status: status}
		if err := ValidateTableSummaryFilter(f); err != nil {
			t.Errorf("expected no error for status=%q, got: %v", status, err)
		}
	}
}

func TestValidateTableSummaryFilter_InvalidStatus(t *testing.T) {
	f := &TableSummaryFilter{Status: "unknown"}
	if err := ValidateTableSummaryFilter(f); err == nil {
		t.Error("expected error for invalid status")
	}
}

func TestValidateTableSummaryFilter_InvalidSort(t *testing.T) {
	f := &TableSummaryFilter{Sort: "invalid_field"}
	if err := ValidateTableSummaryFilter(f); err == nil {
		t.Error("expected error for invalid sort field")
	}
}

func TestValidateTableSummaryFilter_InvalidOrder(t *testing.T) {
	f := &TableSummaryFilter{Order: "random"}
	if err := ValidateTableSummaryFilter(f); err == nil {
		t.Error("expected error for invalid order value")
	}
}

func TestValidateTableSummaryFilter_ValidSortAndOrder(t *testing.T) {
	for _, sort := range []string{"table_name", "status", "last_finished_at", ""} {
		for _, order := range []string{"asc", "desc", ""} {
			f := &TableSummaryFilter{Sort: sort, Order: order}
			if err := ValidateTableSummaryFilter(f); err != nil {
				t.Errorf("expected no error for sort=%q order=%q, got: %v", sort, order, err)
			}
		}
	}
}

// ── Integration Tests: HTTP endpoints ─────────────────────────────────────────

func setupTableHistoryRouter(store *TableHistoryStore) *gin.Engine {
	r := gin.New()
	api := r.Group("/api")
	api.GET("/migrations/tables", listTableSummariesHandler(store, nil))
	api.GET("/migrations/tables/:tableName/history", getTableHistoryHandler(store))
	return r
}

func TestListTableSummaries_EmptyStore(t *testing.T) {
	store := &TableHistoryStore{runs: make(map[string][]TableMigrationHistory)}
	r := setupTableHistoryRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/migrations/tables", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Items []TableMigrationSummary `json:"items"`
		Total int                     `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 0 || len(resp.Items) != 0 {
		t.Errorf("expected empty results, got total=%d items=%d", resp.Total, len(resp.Items))
	}
}

func TestListTableSummaries_WithData(t *testing.T) {
	store := newTestStore()
	r := setupTableHistoryRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/migrations/tables", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	var resp struct {
		Items []TableMigrationSummary `json:"items"`
		Total int                     `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 3 {
		t.Errorf("expected total=3, got %d", resp.Total)
	}
}

func TestListTableSummaries_FilterByStatus(t *testing.T) {
	store := newTestStore()
	r := setupTableHistoryRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/migrations/tables?status=failed", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Items []TableMigrationSummary `json:"items"`
		Total int                     `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total != 1 {
		t.Errorf("expected 1 failed item, got %d", resp.Total)
	}
	if len(resp.Items) != 1 || resp.Items[0].Status != "failed" {
		t.Errorf("expected failed status item, got %+v", resp.Items)
	}
}

func TestListTableSummaries_ExcludeSuccess(t *testing.T) {
	store := newTestStore()
	r := setupTableHistoryRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/migrations/tables?exclude_success=true", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Items []TableMigrationSummary `json:"items"`
		Total int                     `json:"total"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, item := range resp.Items {
		if item.Status == "success" {
			t.Errorf("unexpected success item after exclude_success=true: %s", item.TableName)
		}
	}
}

func TestListTableSummaries_InvalidStatus_Returns400(t *testing.T) {
	store := newTestStore()
	r := setupTableHistoryRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/migrations/tables?status=invalid", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid status, got %d", w.Code)
	}
}

func TestListTableSummaries_InvalidSort_Returns400(t *testing.T) {
	store := newTestStore()
	r := setupTableHistoryRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/migrations/tables?sort=bad_field", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid sort, got %d", w.Code)
	}
}

func TestGetTableHistory_Exists(t *testing.T) {
	store := newTestStore()
	r := setupTableHistoryRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/migrations/tables/USERS/history", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", w.Code, w.Body.String())
	}
	var resp struct {
		TableName string                  `json:"table_name"`
		Items     []TableMigrationHistory `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.TableName != "USERS" {
		t.Errorf("expected table_name=USERS, got %s", resp.TableName)
	}
	if len(resp.Items) == 0 {
		t.Error("expected at least 1 history item")
	}
}

func TestGetTableHistory_NotFound_Returns404(t *testing.T) {
	store := newTestStore()
	r := setupTableHistoryRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/migrations/tables/NONEXISTENT/history", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent table, got %d", w.Code)
	}
}

func TestGetTableHistory_LimitParam(t *testing.T) {
	store := &TableHistoryStore{runs: make(map[string][]TableMigrationHistory)}
	for i := range 10 {
		store.RecordTableRun(TableMigrationHistory{
			RunID:     fmt.Sprintf("run%d", i),
			TableName: "T1",
			Status:    "success",
			StartedAt: time.Now(),
		})
	}
	r := setupTableHistoryRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/migrations/tables/T1/history?limit=3", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Items []TableMigrationHistory `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Items) != 3 {
		t.Errorf("expected 3 items with limit=3, got %d", len(resp.Items))
	}
}
