package web

import (
	"sort"
	"strings"
	"sync"
	"time"
)

// TableMigrationSummary는 테이블별 최신 마이그레이션 상태 요약이다.
type TableMigrationSummary struct {
	TableName      string    `json:"table_name"`
	Status         string    `json:"status"` // not_started|running|success|failed
	LastStartedAt  time.Time `json:"last_started_at,omitempty"`
	LastFinishedAt time.Time `json:"last_finished_at,omitempty"`
	DurationMs     int64     `json:"duration_ms,omitempty"`
	RunCount       int64     `json:"run_count"`
	LastError      string    `json:"last_error,omitempty"`
}

// TableMigrationHistory는 테이블 단위 단일 실행 이력이다.
type TableMigrationHistory struct {
	RunID         string    `json:"run_id"`
	TableName     string    `json:"table_name"`
	Status        string    `json:"status"` // success|failed
	StartedAt     time.Time `json:"started_at"`
	FinishedAt    time.Time `json:"finished_at,omitempty"`
	DurationMs    int64     `json:"duration_ms,omitempty"`
	RowsProcessed int64     `json:"rows_processed,omitempty"`
	ErrorMessage  string    `json:"error_message,omitempty"`
}

// TableSummaryFilter는 테이블 목록 조회 필터 옵션이다.
type TableSummaryFilter struct {
	Status         string // not_started|running|success|failed|""
	ExcludeSuccess bool
	Search         string // table name contains
	Sort           string // table_name|status|last_finished_at
	Order          string // asc|desc
	Page           int
	PageSize       int
}

var validStatuses = map[string]bool{
	"not_started": true,
	"running":     true,
	"success":     true,
	"failed":      true,
	"":            true,
}

var validSortFields = map[string]bool{
	"table_name":      true,
	"status":          true,
	"last_finished_at": true,
	"":                true,
}

// ValidateTableSummaryFilter는 필터 파라미터의 유효성을 검증한다.
func ValidateTableSummaryFilter(f *TableSummaryFilter) error {
	if !validStatuses[f.Status] {
		return errInvalidFilterParam("status", f.Status)
	}
	if !validSortFields[f.Sort] {
		return errInvalidFilterParam("sort", f.Sort)
	}
	if f.Order != "" && f.Order != "asc" && f.Order != "desc" {
		return errInvalidFilterParam("order", f.Order)
	}
	return nil
}

type filterParamError struct {
	param string
	value string
}

func (e *filterParamError) Error() string {
	return "invalid " + e.param + " value: " + e.value
}

func errInvalidFilterParam(param, value string) error {
	return &filterParamError{param: param, value: value}
}

// maxHistoryPerTable는 테이블당 보관할 최대 이력 건수이다.
const maxHistoryPerTable = 100

// TableHistoryStore는 테이블 단위 마이그레이션 이력을 관리한다.
// 인메모리 저장소로, 서버 재시작 시 초기화된다.
type TableHistoryStore struct {
	mu   sync.RWMutex
	runs map[string][]TableMigrationHistory // key: tableName, value: 최신순 정렬
}

var globalTableHistory = &TableHistoryStore{
	runs: make(map[string][]TableMigrationHistory),
}

// RecordTableRun은 테이블 마이그레이션 실행 결과를 기록한다.
func (s *TableHistoryStore) RecordTableRun(h TableMigrationHistory) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing := s.runs[h.TableName]
	// 최신순 prepend
	updated := make([]TableMigrationHistory, 0, len(existing)+1)
	updated = append(updated, h)
	updated = append(updated, existing...)
	if len(updated) > maxHistoryPerTable {
		updated = updated[:maxHistoryPerTable]
	}
	s.runs[h.TableName] = updated
}

// ListSummaries는 필터 조건에 맞는 테이블 요약 목록을 반환한다.
func (s *TableHistoryStore) ListSummaries(f TableSummaryFilter) ([]TableMigrationSummary, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]TableMigrationSummary, 0, len(s.runs))
	for tableName, history := range s.runs {
		summary := buildSummary(tableName, history)

		// 필터 적용
		if f.Status != "" && summary.Status != f.Status {
			continue
		}
		if f.ExcludeSuccess && summary.Status == "success" {
			continue
		}
		if f.Search != "" && !strings.Contains(strings.ToLower(summary.TableName), strings.ToLower(f.Search)) {
			continue
		}
		summaries = append(summaries, summary)
	}

	// 정렬
	sortSummaries(summaries, f.Sort, f.Order)

	total := len(summaries)

	// 페이징
	page := f.Page
	pageSize := f.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= total {
		return []TableMigrationSummary{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return summaries[start:end], total
}

// GetHistory는 특정 테이블의 실행 이력을 반환한다.
// 테이블 이력이 없으면 (nil, false)를 반환한다.
func (s *TableHistoryStore) GetHistory(tableName string, limit int) ([]TableMigrationHistory, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	history, ok := s.runs[tableName]
	if !ok {
		return nil, false
	}
	if limit <= 0 || limit > len(history) {
		limit = len(history)
	}
	result := make([]TableMigrationHistory, limit)
	copy(result, history[:limit])
	return result, true
}

// HasTable는 테이블 이력 존재 여부를 반환한다.
func (s *TableHistoryStore) HasTable(tableName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.runs[tableName]
	return ok
}

func buildSummary(tableName string, history []TableMigrationHistory) TableMigrationSummary {
	if len(history) == 0 {
		return TableMigrationSummary{
			TableName: tableName,
			Status:    "not_started",
		}
	}

	latest := history[0]
	summary := TableMigrationSummary{
		TableName:      tableName,
		Status:         latest.Status,
		LastStartedAt:  latest.StartedAt,
		LastFinishedAt: latest.FinishedAt,
		DurationMs:     latest.DurationMs,
		RunCount:       int64(len(history)),
		LastError:      latest.ErrorMessage,
	}
	return summary
}

func sortSummaries(summaries []TableMigrationSummary, sortField, order string) {
	if sortField == "" {
		sortField = "table_name"
	}
	if order == "" {
		order = "asc"
	}
	desc := order == "desc"

	sort.SliceStable(summaries, func(i, j int) bool {
		a, b := summaries[i], summaries[j]
		var less bool
		switch sortField {
		case "status":
			less = a.Status < b.Status
		case "last_finished_at":
			less = a.LastFinishedAt.Before(b.LastFinishedAt)
		default: // table_name
			less = a.TableName < b.TableName
		}
		if desc {
			return !less
		}
		return less
	})
}
