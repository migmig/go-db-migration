package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pashagolub/pgxmock/v3"
)

// ── Count Functions ───────────────────────────────────────────────────────────

func TestSQLDBCountFn_UsesQuotedIdentifier(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM "Users"`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(7))

	countFn := SQLDBCountFn(db, func(name string) string {
		return `"` + name + `"`
	})
	got, err := countFn(context.Background(), "Users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 7 {
		t.Fatalf("expected count 7, got %d", got)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPGPoolCountFn_UsesQuotedIdentifier(t *testing.T) {
	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	defer pgMock.Close()

	pgMock.ExpectQuery(`SELECT COUNT\(\*\) FROM "Orders"`).
		WillReturnRows(pgxmock.NewRows([]string{"count"}).AddRow(3))

	countFn := PGPoolCountFn(pgMock, func(name string) string {
		return `"` + name + `"`
	})
	got, err := countFn(context.Background(), "Orders")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3 {
		t.Fatalf("expected count 3, got %d", got)
	}
	if err := pgMock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

// ── FetchTables ───────────────────────────────────────────────────────────────

func TestFetchTables_NoFilter(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT table_name FROM user_tables").
		WillReturnRows(sqlmock.NewRows([]string{"table_name"}).
			AddRow("ORDERS").
			AddRow("USERS"))

	tables, err := FetchTables(db, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tables) != 2 {
		t.Fatalf("expected 2 tables, got %d: %v", len(tables), tables)
	}
	if tables[0] != "ORDERS" || tables[1] != "USERS" {
		t.Errorf("unexpected tables: %v", tables)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestFetchTables_WithFilter_IncludesLIKEClause(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// The filter argument must be passed to the query
	mock.ExpectQuery("SELECT table_name FROM user_tables WHERE table_name LIKE").
		WithArgs("USER%").
		WillReturnRows(sqlmock.NewRows([]string{"table_name"}).AddRow("USERS"))

	tables, err := FetchTables(db, "USER%")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tables) != 1 || tables[0] != "USERS" {
		t.Errorf("unexpected tables: %v", tables)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations (LIKE clause not used): %v", err)
	}
}

func TestFetchTables_EmptyResultSet(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT table_name FROM user_tables").
		WillReturnRows(sqlmock.NewRows([]string{"table_name"}))

	tables, err := FetchTables(db, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tables) != 0 {
		t.Errorf("expected 0 tables, got %d", len(tables))
	}
}

func TestFetchTables_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT table_name FROM user_tables").
		WillReturnError(fmt.Errorf("oracle: connection refused"))

	_, err = FetchTables(db, "")
	if err == nil {
		t.Error("expected error on query failure, got nil")
	}
}

// ── TableExists ───────────────────────────────────────────────────────────────

func TestTableExists_ReturnsTrue(t *testing.T) {
	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	defer pgMock.Close()

	pgMock.ExpectQuery("SELECT EXISTS").
		WithArgs("public", "USERS").
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	exists, err := TableExists(context.Background(), pgMock, "public", "USERS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected exists=true, got false")
	}
	if err := pgMock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestTableExists_ReturnsFalse(t *testing.T) {
	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	defer pgMock.Close()

	pgMock.ExpectQuery("SELECT EXISTS").
		WithArgs("myschema", "NONEXISTENT").
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(false))

	exists, err := TableExists(context.Background(), pgMock, "myschema", "NONEXISTENT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected exists=false, got true")
	}
}

func TestTableExists_QueryError(t *testing.T) {
	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	defer pgMock.Close()

	pgMock.ExpectQuery("SELECT EXISTS").
		WillReturnError(fmt.Errorf("pg: connection closed"))

	_, err = TableExists(context.Background(), pgMock, "public", "T")
	if err == nil {
		t.Error("expected error on query failure, got nil")
	}
}

func TestTableExists_PassesSchemaAndTableAsArgs(t *testing.T) {
	pgMock, err := pgxmock.NewPool()
	if err != nil {
		t.Fatalf("failed to create pgxmock: %v", err)
	}
	defer pgMock.Close()

	// The args must match exactly — this verifies the function passes them in the right order
	pgMock.ExpectQuery("SELECT EXISTS").
		WithArgs("billing", "INVOICES").
		WillReturnRows(pgxmock.NewRows([]string{"exists"}).AddRow(true))

	_, err = TableExists(context.Background(), pgMock, "billing", "INVOICES")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := pgMock.ExpectationsWereMet(); err != nil {
		t.Errorf("schema/table args not passed correctly: %v", err)
	}
}
