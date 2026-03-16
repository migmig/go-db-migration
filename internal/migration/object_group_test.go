package migration

import (
	"os"
	"path/filepath"
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
