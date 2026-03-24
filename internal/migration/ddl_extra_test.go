package migration

import (
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
)

func TestGetConstraintMetadata(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	// Mocking the query for constraints
	rows := sqlmock.NewRows([]string{"CONSTRAINT_NAME", "CONSTRAINT_TYPE", "SEARCH_CONDITION", "R_CONSTRAINT_NAME", "DELETE_RULE"}).
		AddRow("SYS_C001", "C", "SALARY > 0", nil, nil).
		AddRow("SYS_R001", "R", nil, "SYS_PK_DEPT", "CASCADE")

	mock.ExpectQuery("SELECT c\\.constraint_name.*").
		WithArgs("HR", "EMP").
		WillReturnRows(rows)

	// 1. Local columns for SYS_C001 (Check constraint)
	colRowsC := sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("SALARY")
	mock.ExpectQuery("SELECT column_name FROM all_cons_columns").WithArgs("HR", "SYS_C001").WillReturnRows(colRowsC)

	// 2. Ref columns for SYS_R001 (FK constraint)
	colRowsRRef := sqlmock.NewRows([]string{"TABLE_NAME", "COLUMN_NAME"}).AddRow("DEPT", "ID")
	mock.ExpectQuery("SELECT table_name, column_name FROM all_cons_columns").WithArgs("HR", "SYS_PK_DEPT").WillReturnRows(colRowsRRef)

	// 3. Local columns for SYS_R001 (FK constraint)
	colRowsRLoc := sqlmock.NewRows([]string{"COLUMN_NAME"}).AddRow("DEPT_ID")
	mock.ExpectQuery("SELECT column_name FROM all_cons_columns").WithArgs("HR", "SYS_R001").WillReturnRows(colRowsRLoc)

	constraints, err := GetConstraintMetadata(db, "EMP", "HR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(constraints) != 2 {
		t.Fatalf("expected 2 constraints, got %d", len(constraints))
	}
	if constraints[0].Type != "C" {
		t.Errorf("expected C, got %s", constraints[0].Type)
	}
	if constraints[1].Type != "R" {
		t.Errorf("expected R, got %s", constraints[1].Type)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestMigrateTablePgUpsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to open mock pg connection: %v", err)
	}
	defer pgMock.Close()

	cfg := &config.Config{
		Schema:    "public",
		BatchSize: 100,
		CopyBatch: 100,
	}

	pgMock.ExpectExec("CREATE TEMP TABLE IF NOT EXISTS").WillReturnResult(pgxmock.NewResult("CREATE", 0))

	// 1 row in Oracle
	mock.ExpectQuery("SELECT \\* FROM \"MOCK_TABLE\" OFFSET 0 ROWS FETCH NEXT 100 ROWS ONLY").
		WillReturnRows(sqlmock.NewRows([]string{"id", "val"}).AddRow(1, "A"))

	// pg transaction
	pgMock.ExpectBegin()
	pgMock.ExpectCopyFrom(pgx.Identifier{"_dbm_stage_mock_table"}, []string{"id", "val"}).WillReturnResult(1)
	pgMock.ExpectExec("INSERT INTO \"public\".\"mock_table\" SELECT \\* FROM \"_dbm_stage_mock_table\" ON CONFLICT \\(\"id\"\\) DO NOTHING").WillReturnResult(pgxmock.NewResult("INSERT", 1))
	pgMock.ExpectCommit()

	state := NewMigrationState("t")

	count, err := migrateTablePgUpsert(db, pgMock, "MOCK_TABLE", []string{"id", "val"}, []string{"id"}, cfg, nil, state)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Oracle expectations unfulfilled: %s", err)
	}
	if err := pgMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Postgres expectations unfulfilled: %s", err)
	}
}

func TestRunValidation(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to open mock pg connection: %v", err)
	}
	defer pgMock.Close()

	cfg := &config.Config{
		Schema: "public",
		Tables: []string{"T1"},
	}
	dia := &dialect.PostgresDialect{}
	report := NewMigrationReport("job1", "oracle", "postgres", "pgurl", "all")

	mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"T1\"").WillReturnRows(sqlmock.NewRows([]string{"C"}).AddRow(10))
	pgMock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM \"public\".\"t1\"").WillReturnRows(pgxmock.NewRows([]string{"c"}).AddRow(10))

	// It logs results and returns a struct, just need to see if it executes without error.
	runValidation(db, nil, pgMock, dia, cfg, nil, report)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Oracle expectations unfulfilled: %s", err)
	}
	if err := pgMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Postgres expectations unfulfilled: %s", err)
	}
}

func TestGetPrimaryKeyMetadata_NoPK(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	// Mocking empty result for PK search in all_constraints
	mock.ExpectQuery("SELECT c\\.constraint_name FROM all_constraints c.*AND c\\.constraint_type = 'P'").
		WithArgs("HR", "EMP").
		WillReturnRows(sqlmock.NewRows([]string{"CONSTRAINT_NAME"}))

	pk, err := GetPrimaryKeyMetadata(db, "EMP", "HR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pk != nil {
		t.Error("expected nil pk")
	}
}

func TestGetIndexMetadata_None(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open mock db: %v", err)
	}
	defer db.Close()

	// Mocking empty result for indexes search in all_indexes
	mock.ExpectQuery("SELECT i\\.index_name.*").
		WithArgs("HR", "EMP").
		WillReturnRows(sqlmock.NewRows([]string{"INDEX_NAME", "UNIQUENESS"}))

	indexes, err := GetIndexMetadata(db, "EMP", "HR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(indexes) != 0 {
		t.Error("expected 0 indexes")
	}
}
