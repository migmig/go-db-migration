package migration

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"dbmigrator/internal/config"

	"github.com/DATA-DOG/go-sqlmock"
)

// ── ProcessRow — extended type coverage ──────────────────────────────────────

func TestProcessRow_TimeValue(t *testing.T) {
	ts := time.Date(2024, 3, 15, 9, 5, 3, 0, time.UTC)
	result := ProcessRow([]interface{}{ts}, []string{"DATE"})
	want := "'2024-03-15 09:05:03'"
	if result != want {
		t.Errorf("ProcessRow(time.Time) = %q, want %q", result, want)
	}
}

func TestProcessRow_BlobBytesHexEncoded(t *testing.T) {
	result := ProcessRow([]interface{}{[]byte{0x01, 0x02, 0xFF}}, []string{"BLOB"})
	want := "'\\x0102ff'"
	if result != want {
		t.Errorf("ProcessRow(BLOB) = %q, want %q", result, want)
	}
}

func TestProcessRow_RawBytesHexEncoded(t *testing.T) {
	result := ProcessRow([]interface{}{[]byte{0xAB, 0xCD}}, []string{"RAW"})
	want := "'\\xabcd'"
	if result != want {
		t.Errorf("ProcessRow(RAW) = %q, want %q", result, want)
	}
}

func TestProcessRow_BytesAsText(t *testing.T) {
	// []byte column that is NOT BLOB/RAW should be treated as a string
	result := ProcessRow([]interface{}{[]byte("hello world")}, []string{"VARCHAR2"})
	want := "'hello world'"
	if result != want {
		t.Errorf("ProcessRow([]byte VARCHAR2) = %q, want %q", result, want)
	}
}

func TestProcessRow_BytesAsText_QuoteEscaping(t *testing.T) {
	result := ProcessRow([]interface{}{[]byte("it's fine")}, []string{"CLOB"})
	want := "'it''s fine'"
	if result != want {
		t.Errorf("ProcessRow([]byte quote escape) = %q, want %q", result, want)
	}
}

func TestProcessRow_BoolTrue(t *testing.T) {
	result := ProcessRow([]interface{}{true}, []string{"BOOLEAN"})
	if result != "TRUE" {
		t.Errorf("ProcessRow(true) = %q, want %q", result, "TRUE")
	}
}

func TestProcessRow_BoolFalse(t *testing.T) {
	result := ProcessRow([]interface{}{false}, []string{"BOOLEAN"})
	if result != "FALSE" {
		t.Errorf("ProcessRow(false) = %q, want %q", result, "FALSE")
	}
}

func TestProcessRow_Int64(t *testing.T) {
	result := ProcessRow([]interface{}{int64(99999)}, []string{"NUMBER"})
	if result != "99999" {
		t.Errorf("ProcessRow(int64) = %q, want %q", result, "99999")
	}
}

func TestProcessRow_Float64(t *testing.T) {
	result := ProcessRow([]interface{}{float64(3.14)}, []string{"FLOAT"})
	if result != "3.14" {
		t.Errorf("ProcessRow(float64) = %q, want %q", result, "3.14")
	}
}

func TestProcessRow_DefaultFallback(t *testing.T) {
	// An unrecognised type should be cast to string and quoted
	type custom struct{ v int }
	result := ProcessRow([]interface{}{custom{42}}, []string{"UNKNOWN"})
	if !strings.HasPrefix(result, "'") || !strings.HasSuffix(result, "'") {
		t.Errorf("ProcessRow(default) should be quoted, got %q", result)
	}
}

func TestProcessRow_MultipleColumns(t *testing.T) {
	ts := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	result := ProcessRow(
		[]interface{}{int64(1), "Alice", nil, ts},
		[]string{"NUMBER", "VARCHAR2", "VARCHAR2", "DATE"},
	)
	want := "1, 'Alice', NULL, '2023-01-01 00:00:00'"
	if result != want {
		t.Errorf("ProcessRow(multi) = %q, want %q", result, want)
	}
}

// ── Run — dry-run mode ────────────────────────────────────────────────────────

func TestRun_DryRun_QueriesRowCounts(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM USERS").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(100))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM ORDERS").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(50))

	cfg := &config.Config{
		Tables: []string{"USERS", "ORDERS"},
		DryRun: true,
	}

	if err := Run(db, nil, cfg, nil); err != nil {
		t.Errorf("Run(DryRun) returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestRun_DryRun_NoFilesCreated(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM T").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

	outDir := t.TempDir()
	cfg := &config.Config{
		Tables:    []string{"T"},
		DryRun:    true,
		OutputDir: outDir,
	}

	if err := Run(db, nil, cfg, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No SQL files should have been written
	entries, _ := readDir(outDir)
	if len(entries) != 0 {
		t.Errorf("expected no files written in dry-run mode, found: %v", entries)
	}
}

// ── MigrateTableToFile — WithDDL flag ─────────────────────────────────────────

func TestMigrateTableToFile_WithDDL_PrependsDDL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// GetTableMetadata query
	mock.ExpectQuery("SELECT column_name, data_type").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "data_precision", "data_scale", "nullable"}).
			AddRow("ID", "NUMBER", 10, 0, "N"))

	// Data query
	mock.ExpectQuery("SELECT \\* FROM USERS").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	var mu sync.Mutex
	cfg := &config.Config{
		BatchSize: 1000,
		WithDDL:   true,
		PerTable:  false,
	}

	if err := MigrateTableToFile(db, "USERS", w, cfg, &mu, nil); err != nil {
		t.Fatalf("MigrateTableToFile returned error: %v", err)
	}
	w.Flush()

	out := buf.String()
	if !strings.Contains(out, "CREATE TABLE IF NOT EXISTS USERS") {
		t.Errorf("expected DDL in output, got:\n%s", out)
	}
	if !strings.Contains(out, "INSERT INTO USERS") {
		t.Errorf("expected INSERT in output, got:\n%s", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMigrateTableToFile_WithDDL_SchemaPrefix(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT column_name, data_type").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "data_precision", "data_scale", "nullable"}).
			AddRow("ID", "NUMBER", 5, 0, "Y"))

	mock.ExpectQuery("SELECT \\* FROM T").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	var mu sync.Mutex
	cfg := &config.Config{
		BatchSize: 1000,
		WithDDL:   true,
		Schema:    "myschema",
		PerTable:  false,
	}

	if err := MigrateTableToFile(db, "T", w, cfg, &mu, nil); err != nil {
		t.Fatalf("MigrateTableToFile returned error: %v", err)
	}
	w.Flush()

	out := buf.String()
	if !strings.Contains(out, "CREATE TABLE IF NOT EXISTS myschema.T") {
		t.Errorf("expected schema-qualified DDL, got:\n%s", out)
	}
}

// ── oracleCopySource ──────────────────────────────────────────────────────────

func TestOracleCopySource_Err_WithNoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT \\* FROM T").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))

	rows, err := db.Query("SELECT * FROM T")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rows.Close()

	src := &oracleCopySource{rows: rows, cols: []string{"ID"}}

	// No rows, so Next() should return false immediately
	if src.Next() {
		t.Error("expected Next() to return false for empty result set")
	}
	if src.Err() != nil {
		t.Errorf("expected Err() to be nil, got: %v", src.Err())
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// readDir returns file names in a directory (used to check no files were written).
func readDir(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names, nil
}
