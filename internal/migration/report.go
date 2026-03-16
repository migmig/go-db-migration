package migration

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"dbmigrator/internal/config"
)

// TableReport는 단일 테이블 마이그레이션 결과를 담는다.
type TableReport struct {
	Name          string   `json:"name"`
	RowCount      int      `json:"row_count"`
	DurationNs    int64    `json:"duration_ns"`
	DurationHuman string   `json:"duration"`
	RowsPerSec    float64  `json:"rows_per_sec"`
	DDLExecuted   bool     `json:"ddl_executed"`
	Status        string   `json:"status"` // "ok", "error"
	Errors        []string `json:"errors,omitempty"`
}

type GroupStats struct {
	TotalItems   int `json:"total_items"`
	SuccessCount int `json:"success_count"`
	ErrorCount   int `json:"error_count"`
	SkippedCount int `json:"skipped_count"`
	TotalRows    int `json:"total_rows,omitempty"`
}

type GroupedStats struct {
	Tables    GroupStats `json:"tables"`
	Sequences GroupStats `json:"sequences"`
}

type ReportSummary struct {
	TotalRows    int          `json:"total_rows"`
	SuccessCount int          `json:"success_count"`
	ErrorCount   int          `json:"error_count"`
	Duration     string       `json:"duration"`
	ReportID     string       `json:"report_id"`
	ObjectGroup  string       `json:"object_group"`
	Stats        GroupedStats `json:"stats"`
}

// MigrationReport는 마이그레이션 전체 실행 결과 감사 로그이다.
type MigrationReport struct {
	JobID         string        `json:"job_id"`
	UserID        int64         `json:"user_id,omitempty"`
	StartedAt     time.Time     `json:"started_at"`
	FinishedAt    time.Time     `json:"finished_at,omitempty"`
	DurationHuman string        `json:"duration,omitempty"`
	SourceURL     string        `json:"source_url"`
	TargetDB      string        `json:"target_db"`
	TargetURL     string        `json:"target_url"`
	ObjectGroup   string        `json:"object_group"`
	Tables        []TableReport `json:"tables"`
	Stats         GroupedStats  `json:"stats"`
	TotalRows     int           `json:"total_rows"`
	SuccessCount  int           `json:"success_count"`
	ErrorCount    int           `json:"error_count"`
	mu            sync.Mutex
}

// NewMigrationReport는 새 MigrationReport를 생성한다. URL의 비밀번호는 마스킹된다.
func NewMigrationReport(jobID, sourceURL, targetDB, targetURL, objectGroup string) *MigrationReport {
	if objectGroup == "" {
		objectGroup = config.ObjectGroupAll
	}
	return &MigrationReport{
		JobID:       jobID,
		StartedAt:   time.Now(),
		SourceURL:   maskPassword(sourceURL),
		TargetDB:    targetDB,
		TargetURL:   maskPassword(targetURL),
		ObjectGroup: objectGroup,
	}
}

// StartTable은 테이블 마이그레이션 시작 시간을 기록하고 완료 콜백을 반환한다.
// 콜백은 worker 함수에서 MigrateTable 완료 직후 호출한다.
func (r *MigrationReport) StartTable(name string, withDDL bool) func(rowCount int, err error) {
	start := time.Now()
	return func(rowCount int, err error) {
		elapsed := time.Since(start)
		tr := TableReport{
			Name:          name,
			RowCount:      rowCount,
			DurationNs:    elapsed.Nanoseconds(),
			DurationHuman: formatDuration(elapsed),
			DDLExecuted:   withDDL,
		}
		if elapsed.Seconds() > 0 && rowCount > 0 {
			tr.RowsPerSec = math.Round(float64(rowCount)/elapsed.Seconds()*10) / 10
		}
		if err != nil {
			tr.Status = "error"
			tr.Errors = append(tr.Errors, err.Error())
		} else {
			tr.Status = "ok"
		}

		r.mu.Lock()
		r.Tables = append(r.Tables, tr)
		r.TotalRows += rowCount
		if err != nil {
			r.ErrorCount++
		} else {
			r.SuccessCount++
		}
		r.recordGroupResultLocked(config.ObjectGroupTables, rowCount, err)
		r.mu.Unlock()
	}
}

func (r *MigrationReport) RecordSequenceResult(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recordGroupResultLocked(config.ObjectGroupSequences, 0, err)
}

func (r *MigrationReport) SkipGroup(group string, count int) {
	if count <= 0 {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	stats := r.groupStatsLocked(group)
	stats.TotalItems += count
	stats.SkippedCount += count
}

func (r *MigrationReport) recordGroupResultLocked(group string, rowCount int, err error) {
	stats := r.groupStatsLocked(group)
	stats.TotalItems++
	stats.TotalRows += rowCount
	if err != nil {
		stats.ErrorCount++
		return
	}
	stats.SuccessCount++
}

func (r *MigrationReport) groupStatsLocked(group string) *GroupStats {
	switch group {
	case config.ObjectGroupSequences:
		return &r.Stats.Sequences
	default:
		return &r.Stats.Tables
	}
}

// Finalize는 리포트를 마무리하고 .migration_state/{job_id}_report.json에 저장한다.
func (r *MigrationReport) Finalize() error {
	r.mu.Lock()
	r.FinishedAt = time.Now()
	r.DurationHuman = formatDuration(r.FinishedAt.Sub(r.StartedAt))
	r.mu.Unlock()

	dir := ".migration_state"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	path := filepath.Join(dir, fmt.Sprintf("%s_report.json", r.JobID))
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// PrintSummary는 마이그레이션 결과 요약 테이블을 표준 출력에 출력한다.
func (r *MigrationReport) PrintSummary() {
	r.mu.Lock()
	tables := make([]TableReport, len(r.Tables))
	copy(tables, r.Tables)
	totalRows := r.TotalRows
	successCount := r.SuccessCount
	errorCount := r.ErrorCount
	duration := r.DurationHuman
	r.mu.Unlock()

	maxName := 5
	for _, t := range tables {
		if len(t.Name) > maxName {
			maxName = len(t.Name)
		}
	}

	line := func(left, mid, right, sep string) string {
		return left + strings.Repeat(sep, maxName+2) + mid +
			strings.Repeat(sep, 11) + mid +
			strings.Repeat(sep, 10) + mid +
			strings.Repeat(sep, 9) + right
	}

	fmt.Println(line("┌", "┬", "┐", "─"))
	fmt.Printf("│ %-*s │ %-9s │ %-8s │ %-7s │\n", maxName, "Table", "Rows", "Duration", "Status")
	fmt.Println(line("├", "┼", "┤", "─"))
	for _, t := range tables {
		status := "OK"
		if t.Status != "ok" {
			status = "ERROR"
		}
		fmt.Printf("│ %-*s │ %9s │ %8s │ %-7s │\n",
			maxName, t.Name, formatCount(t.RowCount), t.DurationHuman, status)
	}
	fmt.Println(line("└", "┴", "┘", "─"))
	fmt.Printf("Total: %s rows, %d ok, %d errors, %s elapsed\n",
		formatCount(totalRows), successCount, errorCount, duration)
}

// ToSummary는 ws.ReportSummary DTO에 들어갈 값을 반환한다.
// server.go에서 AllDone 호출 시 사용한다.
func (r *MigrationReport) ToSummary() ReportSummary {
	r.mu.Lock()
	defer r.mu.Unlock()
	return ReportSummary{
		TotalRows:    r.TotalRows,
		SuccessCount: r.SuccessCount,
		ErrorCount:   r.ErrorCount,
		Duration:     r.DurationHuman,
		ReportID:     r.JobID,
		ObjectGroup:  r.ObjectGroup,
		Stats:        r.Stats,
	}
}

// maskPassword는 URL에서 비밀번호를 "***"로 치환한다.
// "scheme://user:password@host" 형태를 처리한다.
var passwordPattern = regexp.MustCompile(`(://[^:@]+):(.+)@`)

func maskPassword(url string) string {
	return passwordPattern.ReplaceAllString(url, "$1:***@")
}

// formatDuration은 time.Duration을 사람이 읽기 쉬운 문자열로 변환한다.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%.0fms", float64(d.Milliseconds()))
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	min := int(d.Minutes())
	sec := int(d.Seconds()) % 60
	return fmt.Sprintf("%dm%ds", min, sec)
}

// formatCount는 정수를 읽기 쉬운 문자열로 변환한다 (1,234 / 12K / 2.1M).
func formatCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 10_000:
		return fmt.Sprintf("%dK", n/1000)
	case n >= 1000:
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	default:
		return fmt.Sprintf("%d", n)
	}
}
