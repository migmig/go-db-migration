package migration

import (
	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"
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
	mock.ExpectQuery("SELECT \\* FROM " + tableName).WillReturnRows(rows)

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
		pgx.Identifier{"public", "MOCK_TABLE"},
		[]string{"ID", "NAME"},
	).WillReturnResult(2)
	pgMock.ExpectCommit()

	cfg := &config.Config{
		Schema:  "public",
		WithDDL: false,
	}
	dia := &dialect.PostgresDialect{}

	err = MigrateTableDirect(db, nil, pgMock, dia, tableName, cfg, nil)
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
	metaRows := sqlmock.NewRows([]string{"column_name", "data_type", "data_precision", "data_scale", "nullable"}).
		AddRow("ID", "NUMBER", 10, 0, "N").
		AddRow("NAME", "VARCHAR2", nil, nil, "Y")
	mock.ExpectQuery("SELECT column_name, data_type").WillReturnRows(metaRows)

	// Mock Data Select
	rows := sqlmock.NewRows([]string{"ID", "NAME"}).AddRow(1, "Alice")
	mock.ExpectQuery("SELECT \\* FROM " + tableName).WillReturnRows(rows)

	// Mock Postgres
	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("Failed to open mock pg connection: %v", err)
	}
	defer pgMock.Close()

	// 1. DDL Exec
	pgMock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnResult(pgxmock.NewResult("CREATE", 0))

	// 2. Migration transaction
	pgMock.ExpectBegin()
	pgMock.ExpectCopyFrom(
		pgx.Identifier{"public", "MOCK_TABLE"},
		[]string{"ID", "NAME"},
	).WillReturnResult(1)
	pgMock.ExpectCommit()

	cfg := &config.Config{
		Schema:  "public",
		WithDDL: true,
	}
	dia := &dialect.PostgresDialect{}

	err = MigrateTableDirect(db, nil, pgMock, dia, tableName, cfg, nil)
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
