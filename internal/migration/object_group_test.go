package migration

import (
	"bytes"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRun_AllModeWritesSequencesAfterTables(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}
	defer dbConn.Close()

	mock.ExpectQuery("SELECT data_default").
		WillReturnRows(sqlmock.NewRows([]string{"data_default"}).AddRow("SEQ_USERS.NEXTVAL"))
	mock.ExpectQuery("FROM all_sequences").
		WillReturnRows(sqlmock.NewRows([]string{
			"sequence_name", "min_value", "max_value", "increment_by", "cycle_flag", "last_number",
		}).AddRow("SEQ_USERS", 1, 999999, 1, "N", 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "USERS"`).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "USERS"`).
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

	cfg := &config.Config{
		Tables:        []string{"USERS"},
		BatchSize:     100,
		PerTable:      false,
		OutFile:       "tables.sql",
		OutputDir:     tmp,
		WithSequences: true,
		ObjectGroup:   config.ObjectGroupAll,
	}

	report, err := Run(dbConn, nil, nil, &dialect.PostgresDialect{}, cfg, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	sequenceFile := filepath.Join(tmp, "sequences.sql")
	if _, err := os.Stat(sequenceFile); err != nil {
		t.Fatalf("expected %s to be created: %v", sequenceFile, err)
	}
	if report.Stats.Tables.SuccessCount != 1 {
		t.Fatalf("table success count = %d, want 1", report.Stats.Tables.SuccessCount)
	}
	if report.Stats.Sequences.SuccessCount != 1 {
		t.Fatalf("sequence success count = %d, want 1", report.Stats.Sequences.SuccessCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRun_AllModeSkipsSequencesAfterTableFailure(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}
	defer dbConn.Close()

	mock.ExpectQuery("SELECT data_default").
		WillReturnRows(sqlmock.NewRows([]string{"data_default"}).AddRow("SEQ_USERS.NEXTVAL"))
	mock.ExpectQuery("FROM all_sequences").
		WillReturnRows(sqlmock.NewRows([]string{
			"sequence_name", "min_value", "max_value", "increment_by", "cycle_flag", "last_number",
		}).AddRow("SEQ_USERS", 1, 999999, 1, "N", 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "USERS"`).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "USERS"`).
		WillReturnError(os.ErrPermission)

	cfg := &config.Config{
		Tables:        []string{"USERS"},
		BatchSize:     100,
		PerTable:      false,
		OutFile:       "tables.sql",
		OutputDir:     tmp,
		WithSequences: true,
		ObjectGroup:   config.ObjectGroupAll,
	}

	report, err := Run(dbConn, nil, nil, &dialect.PostgresDialect{}, cfg, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	sequenceFile := filepath.Join(tmp, "sequences.sql")
	if _, err := os.Stat(sequenceFile); !os.IsNotExist(err) {
		t.Fatalf("expected %s to be skipped, stat err=%v", sequenceFile, err)
	}
	if report.Stats.Tables.ErrorCount != 1 {
		t.Fatalf("table error count = %d, want 1", report.Stats.Tables.ErrorCount)
	}
	if report.Stats.Sequences.SkippedCount != 1 {
		t.Fatalf("sequence skipped count = %d, want 1", report.Stats.Sequences.SkippedCount)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestClassifyDDLGroup(t *testing.T) {
	cases := []struct {
		name       string
		objectType string
		ddl        string
		want       string
	}{
		{name: "table object", objectType: "table", want: config.ObjectGroupTables},
		{name: "index object", objectType: "index", want: config.ObjectGroupTables},
		{name: "sequence object", objectType: "sequence", want: config.ObjectGroupSequences},
		{name: "create sequence ddl", ddl: "CREATE SEQUENCE seq_users START WITH 1", want: config.ObjectGroupSequences},
		{name: "alter table ddl", ddl: "ALTER TABLE users ADD CONSTRAINT pk_users PRIMARY KEY (id)", want: config.ObjectGroupTables},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := classifyDDLGroup(tc.objectType, tc.ddl); got != tc.want {
				t.Fatalf("classifyDDLGroup(%q, %q) = %q, want %q", tc.objectType, tc.ddl, got, tc.want)
			}
		})
	}
}

func TestRun_TablesOnlySkipsSequenceArtifacts(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}
	defer dbConn.Close()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "USERS"`).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "USERS"`).
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

	cfg := &config.Config{
		Tables:        []string{"USERS"},
		BatchSize:     100,
		PerTable:      false,
		OutFile:       "tables.sql",
		OutputDir:     tmp,
		WithSequences: true,
		ObjectGroup:   config.ObjectGroupTables,
	}

	report, err := Run(dbConn, nil, nil, &dialect.PostgresDialect{}, cfg, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(tmp, "sequences.sql")); !os.IsNotExist(err) {
		t.Fatalf("expected no sequences.sql for tables-only mode, err=%v", err)
	}
	if report.Stats.Sequences.TotalItems != 0 {
		t.Fatalf("expected no sequence stats in tables-only mode, got %+v", report.Stats.Sequences)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRun_SequencesOnlyWritesOnlySequenceDDL(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}
	defer dbConn.Close()

	mock.ExpectQuery("SELECT data_default").
		WillReturnRows(sqlmock.NewRows([]string{"data_default"}).AddRow("SEQ_USERS.NEXTVAL"))
	mock.ExpectQuery("FROM all_sequences").
		WillReturnRows(sqlmock.NewRows([]string{
			"sequence_name", "min_value", "max_value", "increment_by", "cycle_flag", "last_number",
		}).AddRow("SEQ_USERS", 1, 999999, 1, "N", 1))

	cfg := &config.Config{
		Tables:      []string{"USERS"},
		PerTable:    false,
		OutFile:     "migration.sql",
		OutputDir:   tmp,
		ObjectGroup: config.ObjectGroupSequences,
	}

	if _, err := Run(dbConn, nil, nil, &dialect.PostgresDialect{}, cfg, nil); err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmp, "migration.sql"))
	if err != nil {
		t.Fatalf("read sequences-only output: %v", err)
	}
	out := string(data)
	if !strings.Contains(out, "CREATE SEQUENCE") {
		t.Fatalf("expected sequence DDL in sequences-only output:\n%s", out)
	}
	if strings.Contains(out, "CREATE TABLE") || strings.Contains(out, "INSERT INTO") {
		t.Fatalf("sequences-only output should not contain table SQL:\n%s", out)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestRun_DefaultObjectGroupBehavesAsAll(t *testing.T) {
	tmp := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir temp dir: %v", err)
	}
	defer os.Chdir(origDir)

	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}
	defer dbConn.Close()

	mock.ExpectQuery("SELECT data_default").
		WillReturnRows(sqlmock.NewRows([]string{"data_default"}).AddRow("SEQ_USERS.NEXTVAL"))
	mock.ExpectQuery("FROM all_sequences").
		WillReturnRows(sqlmock.NewRows([]string{
			"sequence_name", "min_value", "max_value", "increment_by", "cycle_flag", "last_number",
		}).AddRow("SEQ_USERS", 1, 999999, 1, "N", 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "USERS"`).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT"}).AddRow(1))
	mock.ExpectQuery(`SELECT \* FROM "USERS"`).
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

	cfg := &config.Config{
		Tables:        []string{"USERS"},
		BatchSize:     100,
		PerTable:      false,
		OutFile:       "tables.sql",
		OutputDir:     tmp,
		WithSequences: true,
	}

	report, err := Run(dbConn, nil, nil, &dialect.PostgresDialect{}, cfg, nil)
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if report.ObjectGroup != config.ObjectGroupAll {
		t.Fatalf("ObjectGroup = %q, want %q", report.ObjectGroup, config.ObjectGroupAll)
	}
	if _, err := os.Stat(filepath.Join(tmp, "sequences.sql")); err != nil {
		t.Fatalf("expected sequences.sql in default all mode: %v", err)
	}
}

func TestRun_DryRunPrintsGroupedSections(t *testing.T) {
	dbConn, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error: %v", err)
	}
	defer dbConn.Close()

	mock.ExpectQuery("SELECT data_default").
		WillReturnRows(sqlmock.NewRows([]string{"data_default"}).AddRow("SEQ_USERS.NEXTVAL"))
	mock.ExpectQuery("FROM all_sequences").
		WillReturnRows(sqlmock.NewRows([]string{
			"sequence_name", "min_value", "max_value", "increment_by", "cycle_flag", "last_number",
		}).AddRow("SEQ_USERS", 1, 999999, 1, "N", 1))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "USERS"`).
		WillReturnRows(sqlmock.NewRows([]string{"COUNT"}).AddRow(42))

	cfg := &config.Config{
		Tables:        []string{"USERS"},
		DryRun:        true,
		WithDDL:       true,
		WithSequences: true,
		ObjectGroup:   config.ObjectGroupAll,
	}

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error: %v", err)
	}
	os.Stdout = w

	_, runErr := Run(dbConn, nil, nil, &dialect.PostgresDialect{}, cfg, nil)
	_ = w.Close()
	os.Stdout = origStdout
	if runErr != nil {
		t.Fatalf("Run() error: %v", runErr)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy stdout: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "TABLES SQL") || !strings.Contains(out, "SEQUENCES SQL") {
		t.Fatalf("expected grouped dry-run sections, got:\n%s", out)
	}
	if !strings.Contains(out, "estimated_rows: 42") {
		t.Fatalf("expected estimated row count in dry-run output, got:\n%s", out)
	}
}

func TestLogStatementFailedIncludesObjectGroup(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	defer slog.SetDefault(prev)

	logStatementFailed(config.ObjectGroupSequences, "sequence", "SEQ_USERS", "", "ddl", os.ErrPermission)

	out := buf.String()
	if !strings.Contains(out, "migration.statement.failed") {
		t.Fatalf("expected migration.statement.failed log, got %q", out)
	}
	if !strings.Contains(out, "object_group=sequences") {
		t.Fatalf("expected object_group field in log, got %q", out)
	}
}
