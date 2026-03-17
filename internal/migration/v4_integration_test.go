package migration

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"

	"dbmigrator/internal/bus"
	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// ── 6-2 Integration: SQL File + PerTable=false + outFile 이름 반영 ─────────────

func TestRun_PerTableFalse_SingleFileWithOutFileName(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.MatchExpectationsInOrder(false)
	for _, table := range []string{"T1", "T2"} {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"" + table + "\"").
			WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(1))
		mock.ExpectQuery("SELECT \\* FROM \"" + table + "\"").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))
	}

	outDir := t.TempDir()
	outFile := "custom_output.sql"
	cfg := &config.Config{
		Tables:    []string{"T1", "T2"},
		Parallel:  false,
		Workers:   1,
		BatchSize: 100,
		PerTable:  false,
		OutFile:   outFile,
		OutputDir: outDir,
	}

	dia := &dialect.PostgresDialect{}
	if _, err := Run(db, nil, nil, dia, cfg, nil); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	// custom_output.sql should exist in outDir
	path := outDir + "/" + outFile
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected output file %q to be created", path)
	}
	if _, err := os.Stat(outDir + "/tables.sql"); os.IsNotExist(err) {
		t.Errorf("expected grouped tables artifact %q to be created", outDir+"/tables.sql")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ── 6-2 Integration: SQL File + PerTable=true → 테이블별 파일 생성 ──────────────

func TestRun_PerTableTrue_CreatesPerTableFiles(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.MatchExpectationsInOrder(false)
	tables := []string{"ALPHA", "BETA"}
	for _, tbl := range tables {
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"" + tbl + "\"").
			WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(5))
		mock.ExpectQuery("SELECT \\* FROM \"" + tbl + "\"").
			WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1).AddRow(2))
	}

	outDir := t.TempDir()
	cfg := &config.Config{
		Tables:    tables,
		Parallel:  false,
		Workers:   1,
		BatchSize: 100,
		PerTable:  true,
		OutputDir: outDir,
	}

	dia := &dialect.PostgresDialect{}
	if _, err := Run(db, nil, nil, dia, cfg, nil); err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	for _, tbl := range tables {
		path := outDir + "/" + tbl + ".sql"
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected per-table file %q to be created", path)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ── 6-2 Integration: SQL File + Schema 지정 → INSERT에 스키마 접두사 포함 ────────

func TestMigrateTableToFile_Schema_InsertsHaveSchemaPrefix(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT \\* FROM \"ITEMS\"").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(42))

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)
	var mu sync.Mutex
	cfg := &config.Config{
		BatchSize: 100,
		Schema:    "myschema",
		PerTable:  false,
	}

	dia := &dialect.PostgresDialect{}
	if _, err := MigrateTableToFile(db, dia, "ITEMS", w, cfg, &mu, nil, NewMigrationState("test")); err != nil {
		t.Fatalf("MigrateTableToFile: %v", err)
	}
	w.Flush()

	out := buf.String()
	if !strings.Contains(out, "INSERT INTO \"myschema\".\"items\"") {
		t.Errorf("expected INSERT to contain schema prefix, got:\n%s", out)
	}
}

// ── 6-2 Integration: Dry-Run → DryRunTracker 호출 검증 ──────────────────────

type mockDryRunTracker struct {
	results []struct {
		table        string
		totalRows    int
		connectionOk bool
	}
	errors []string
}

func (m *mockDryRunTracker) Init(table string, totalRows int)       {}
func (m *mockDryRunTracker) Update(table string, processedRows int) {}
func (m *mockDryRunTracker) Done(table string)                      {}
func (m *mockDryRunTracker) Error(table string, err error)          { m.errors = append(m.errors, table) }
func (m *mockDryRunTracker) EventBus() bus.EventBus                 { return nil }
func (m *mockDryRunTracker) DryRunResult(table string, totalRows int, connectionOk bool) {
	m.results = append(m.results, struct {
		table        string
		totalRows    int
		connectionOk bool
	}{table, totalRows, connectionOk})
}

func TestRun_DryRun_CallsDryRunTracker(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"USERS\"").
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(150))
	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"ORDERS\"").
		WillReturnRows(sqlmock.NewRows([]string{"c"}).AddRow(300))

	tracker := &mockDryRunTracker{}
	cfg := &config.Config{
		Tables: []string{"USERS", "ORDERS"},
		DryRun: true,
	}

	dia := &dialect.PostgresDialect{}
	if _, err := Run(db, nil, nil, dia, cfg, tracker); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(tracker.results) != 2 {
		t.Fatalf("expected 2 DryRunResult calls, got %d", len(tracker.results))
	}
	byTable := map[string]int{}
	for _, r := range tracker.results {
		byTable[r.table] = r.totalRows
		if !r.connectionOk {
			t.Errorf("DryRunResult for %s: connectionOk should be true", r.table)
		}
	}
	if byTable["USERS"] != 150 {
		t.Errorf("USERS totalRows = %d, want 150", byTable["USERS"])
	}
	if byTable["ORDERS"] != 300 {
		t.Errorf("ORDERS totalRows = %d, want 300", byTable["ORDERS"])
	}
}

func TestRun_DryRun_ErrorCallsTrackerError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"BAD\"").
		WillReturnError(os.ErrNotExist)

	tracker := &mockDryRunTracker{}
	cfg := &config.Config{
		Tables: []string{"BAD"},
		DryRun: true,
	}

	dia := &dialect.PostgresDialect{}
	if _, err := Run(db, nil, nil, dia, cfg, tracker); err != nil {
		t.Fatalf("Run should not return error on per-table count failure: %v", err)
	}

	if len(tracker.errors) != 1 || tracker.errors[0] != "BAD" {
		t.Errorf("expected Error('BAD') to be called, got: %v", tracker.errors)
	}
}
