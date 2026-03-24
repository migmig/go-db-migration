package migration

import (
	"fmt"
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pashagolub/pgxmock/v3"
)

func TestMigrateTableDirect_DDLFail(t *testing.T) {
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
		WithDDL: true,
		Schema:  "public",
	}
	dia := &dialect.PostgresDialect{}

	// Mocking metadata
	mock.ExpectQuery("SELECT column_name, data_type").WillReturnRows(
		sqlmock.NewRows([]string{"COLUMN_NAME", "DATA_TYPE"}).AddRow("ID", "NUMBER"),
	)

	// Mocking DDL failure
	pgMock.ExpectExec("CREATE TABLE IF NOT EXISTS").WillReturnError(fmt.Errorf("ddl error"))

	state := NewMigrationState("t")

	_, err = MigrateTableDirect(db, nil, pgMock, dia, "EMP", cfg, nil, state)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMigrateTableDirect_TableNotExists(t *testing.T) {
	db, _, err := sqlmock.New()
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
		WithDDL: false,
		Schema:  "public",
	}
	dia := &dialect.PostgresDialect{}

	// Mocking TableExists returning false
	pgMock.ExpectQuery("SELECT EXISTS").WithArgs("public", "EMP").WillReturnRows(
		pgxmock.NewRows([]string{"exists"}).AddRow(false),
	)

	state := NewMigrationState("t")

	_, err = MigrateTableDirect(db, nil, pgMock, dia, "EMP", cfg, nil, state)
	if err == nil {
		t.Fatal("expected error")
	}
}
