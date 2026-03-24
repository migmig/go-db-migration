package migration

import (
	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v3"
)

func TestMigrateTableDirect(t *testing.T) {
	// Mock Oracle
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql connection: %v", err)
	}
	defer db.Close()

	tableName := "MOCK_TABLE"
	rows := sqlmock.NewRows([]string{"ID", "NAME"}).
		AddRow(1, "Alice").
		AddRow(2, "Bob")
	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\"").WillReturnRows(rows)

	// Mock Postgres
	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to open mock pg connection: %v", err)
	}
	defer pgMock.Close()

	// 1. TableExists check (happens before Begin)
	pgMock.ExpectQuery("SELECT EXISTS").
		WithArgs("public", "MOCK_TABLE").
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	// 2. Migration transaction
	pgMock.ExpectBegin()
	pgMock.ExpectCopyFrom(
		pgx.Identifier{"public", "mock_table"},
		[]string{"id", "name"},
	).WillReturnResult(2)
	pgMock.ExpectCommit()

	cfg := &config.Config{
		Schema:  "public",
		WithDDL: false,
	}
	dia := &dialect.PostgresDialect{}

	_, err = MigrateTableDirect(db, nil, pgMock, dia, tableName, cfg, nil, NewMigrationState("test"))
	if err != nil {
		t.Errorf("MigrateTableDirect returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Oracle expectations unfulfilled: %s", err)
	}
	if err := pgMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Postgres expectations unfulfilled: %s", err)
	}
}

func TestMigrateTableDirect_WithDDL(t *testing.T) {
	// Mock Oracle
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql connection: %v", err)
	}
	defer db.Close()

	tableName := "MOCK_TABLE"

	// Mock GetTableMetadata
	metaRows := sqlmock.NewRows([]string{"column_name", "data_type", "data_precision", "data_scale", "nullable", "data_default"}).
		AddRow("ID", "NUMBER", 10, 0, "N", nil).
		AddRow("NAME", "VARCHAR2", nil, nil, "Y", nil)
	mock.ExpectQuery("SELECT column_name, data_type").WillReturnRows(metaRows)

	// Mock PK metadata
	mock.ExpectQuery("FROM all_constraints c").
		WillReturnRows(sqlmock.NewRows([]string{"constraint_name"}).AddRow("PK_MOCK_TABLE"))
	mock.ExpectQuery("FROM all_cons_columns").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "position"}).AddRow("ID", 1))

	// Mock Data Select
	rows := sqlmock.NewRows([]string{"ID", "NAME"}).AddRow(1, "Alice")
	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\"").WillReturnRows(rows)

	// Mock Postgres
	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to open mock pg connection: %v", err)
	}
	defer pgMock.Close()

	// 1. DDL Exec
	pgMock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(pgxmock.NewResult("CREATE", 0))
	pgMock.ExpectExec("ALTER TABLE .* ADD PRIMARY KEY").WillReturnResult(pgxmock.NewResult("ALTER TABLE", 0))

	// 2. Migration transaction
	pgMock.ExpectBegin()
	pgMock.ExpectCopyFrom(
		pgx.Identifier{"public", "mock_table"},
		[]string{"id", "name"},
	).WillReturnResult(1)
	pgMock.ExpectCommit()

	cfg := &config.Config{
		Schema:  "public",
		WithDDL: true,
	}
	dia := &dialect.PostgresDialect{}

	_, err = MigrateTableDirect(db, nil, pgMock, dia, tableName, cfg, nil, NewMigrationState("test"))
	if err != nil {
		t.Errorf("MigrateTableDirect returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Oracle expectations unfulfilled: %s", err)
	}
	if err := pgMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Postgres expectations unfulfilled: %s", err)
	}
}

func TestMigrateTableDirect_WithDDLAndIndexes_BatchedCopy(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql connection: %v", err)
	}
	defer db.Close()

	tableName := "SAMPLE_DATA2"

	metaRows := sqlmock.NewRows([]string{"column_name", "data_type", "data_precision", "data_scale", "nullable", "data_default"}).
		AddRow("ID", "NUMBER", 10, 0, "N", nil).
		AddRow("NAME", "VARCHAR2", nil, nil, "Y", nil)
	mock.ExpectQuery("SELECT column_name, data_type").WillReturnRows(metaRows)

	mock.ExpectQuery("FROM all_constraints c").
		WillReturnRows(sqlmock.NewRows([]string{"constraint_name"}).AddRow("SAMPLE_DATA2_PK"))
	mock.ExpectQuery("FROM all_cons_columns").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "position"}).AddRow("ID", 1))

	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\"").
		WillReturnRows(sqlmock.NewRows([]string{"ID", "NAME"}))

	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\" OFFSET 0 ROWS FETCH NEXT 10 ROWS ONLY").
		WillReturnRows(sqlmock.NewRows([]string{"ID", "NAME"}).AddRow(1, "Alice"))

	mock.ExpectQuery("all_indexes").
		WillReturnRows(sqlmock.NewRows([]string{"index_name", "uniqueness", "index_type"}).
			AddRow("SAMPLE_DATA2_NAME_INDEX", "NONUNIQUE", "NORMAL"))
	mock.ExpectQuery("all_ind_columns").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "column_position", "descend"}).
			AddRow("NAME", 1, "ASC"))

	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to open mock pg connection: %v", err)
	}
	defer pgMock.Close()

	pgMock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(pgxmock.NewResult("CREATE", 0))
	pgMock.ExpectExec("ALTER TABLE .* ADD PRIMARY KEY").WillReturnResult(pgxmock.NewResult("ALTER TABLE", 0))
	pgMock.ExpectBegin()
	pgMock.ExpectCopyFrom(
		pgx.Identifier{"sample_data2"},
		[]string{"id", "name"},
	).WillReturnResult(1)
	pgMock.ExpectCommit()
	pgMock.ExpectExec("CREATE INDEX IF NOT EXISTS .*sample_data2_name_index.*").WillReturnResult(pgxmock.NewResult("CREATE INDEX", 0))

	cfg := &config.Config{
		WithDDL:     true,
		WithIndexes: true,
		CopyBatch:   10,
	}
	dia := &dialect.PostgresDialect{}

	_, err = MigrateTableDirect(db, nil, pgMock, dia, tableName, cfg, nil, NewMigrationState("test"))
	if err != nil {
		t.Errorf("MigrateTableDirect returned error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Oracle expectations unfulfilled: %s", err)
	}
	if err := pgMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Postgres expectations unfulfilled: %s", err)
	}
}

func TestMigrateTableDirect_SkipBatch_ReturnsPartialSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql connection: %v", err)
	}
	defer db.Close()

	tableName := "MOCK_TABLE"
	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\"").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))

	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\" OFFSET 0 ROWS FETCH NEXT 1 ROWS ONLY").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))
	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\" OFFSET 1 ROWS FETCH NEXT 1 ROWS ONLY").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(2))
	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\" OFFSET 2 ROWS FETCH NEXT 1 ROWS ONLY").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))

	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to open mock pg connection: %v", err)
	}
	defer pgMock.Close()

	pgMock.ExpectQuery("SELECT EXISTS").
		WithArgs("public", "MOCK_TABLE").
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	pgMock.ExpectBegin()
	pgMock.ExpectCopyFrom(pgx.Identifier{"public", "mock_table"}, []string{"id"}).WillReturnError(fmt.Errorf("copy error"))
	pgMock.ExpectRollback()

	pgMock.ExpectBegin()
	pgMock.ExpectCopyFrom(pgx.Identifier{"public", "mock_table"}, []string{"id"}).WillReturnResult(1)
	pgMock.ExpectCommit()

	pgMock.ExpectBegin()
	pgMock.ExpectCopyFrom(pgx.Identifier{"public", "mock_table"}, []string{"id"}).WillReturnResult(0)
	pgMock.ExpectCommit()

	cfg := &config.Config{
		Schema:     "public",
		CopyBatch:  1,
		OnError:    "skip_batch",
		MaxRetries: 0,
	}
	dia := &dialect.PostgresDialect{}

	_, err = MigrateTableDirect(db, nil, pgMock, dia, tableName, cfg, nil, NewMigrationState("test"))
	if err == nil {
		t.Errorf("expected PartialBatchError, got nil")
	} else if pbe, ok := err.(*PartialBatchError); !ok {
		t.Errorf("expected PartialBatchError, got %T: %v", err, err)
	} else {
		if pbe.SkippedBatches != 1 {
			t.Errorf("expected skipped batches 1, got %d", pbe.SkippedBatches)
		}
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Oracle expectations unfulfilled: %s", err)
	}
	if err := pgMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Postgres expectations unfulfilled: %s", err)
	}
}

func TestMigrateTableDirect_FailFast_ReturnsError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to open mock sql connection: %v", err)
	}
	defer db.Close()

	tableName := "MOCK_TABLE"
	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\"").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))

	mock.ExpectQuery("SELECT \\* FROM \"" + tableName + "\" OFFSET 0 ROWS FETCH NEXT 1 ROWS ONLY").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to open mock pg connection: %v", err)
	}
	defer pgMock.Close()

	pgMock.ExpectQuery("SELECT EXISTS").
		WithArgs("public", "MOCK_TABLE").
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	pgMock.ExpectBegin()
	pgMock.ExpectCopyFrom(pgx.Identifier{"public", "mock_table"}, []string{"id"}).WillReturnError(fmt.Errorf("copy error"))
	pgMock.ExpectRollback()

	cfg := &config.Config{
		Schema:     "public",
		CopyBatch:  1,
		OnError:    "fail_fast",
		MaxRetries: 0,
	}
	dia := &dialect.PostgresDialect{}

	_, err = MigrateTableDirect(db, nil, pgMock, dia, tableName, cfg, nil, NewMigrationState("test"))
	if err == nil {
		t.Errorf("expected error, got nil")
	} else if _, ok := err.(*PartialBatchError); ok {
		t.Errorf("expected normal error, got PartialBatchError")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Oracle expectations unfulfilled: %s", err)
	}
	if err := pgMock.ExpectationsWereMet(); err != nil {
		t.Errorf("Postgres expectations unfulfilled: %s", err)
	}
}
