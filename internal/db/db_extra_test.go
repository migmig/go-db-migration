package db

import (
	"context"
	"fmt"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pashagolub/pgxmock/v3"
)

func TestConnectOracle_Invalid(t *testing.T) {
	_, err := ConnectOracle("invalid_url", "user", "pass")
	if err == nil {
		t.Error("expected error")
	}
}

func TestFetchTargetTables_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	mock.ExpectQuery("SELECT table_name FROM information_schema.tables").
		WithArgs("public").
		WillReturnRows(pgxmock.NewRows([]string{"table_name"}).AddRow("T1"))
	
	tables, err := FetchTargetTables(context.Background(), mock, "public")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tables) != 1 || tables[0] != "T1" {
		t.Errorf("unexpected tables: %v", tables)
	}
}

func TestFetchColumnTypes_Mock(t *testing.T) {
	mock, _ := pgxmock.NewPool()
	mock.ExpectQuery("SELECT column_name, data_type FROM information_schema.columns").
		WithArgs("public", "t1").
		WillReturnRows(pgxmock.NewRows([]string{"column_name", "data_type"}).AddRow("id", "integer"))
	
	types, err := FetchColumnTypes(context.Background(), mock, "public", "t1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if types["id"] != "integer" {
		t.Errorf("expected integer, got %s", types["id"])
	}
}

func TestSQLDBCountFn_Error(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	
	mock.ExpectQuery("SELECT COUNT").WillReturnError(fmt.Errorf("count error"))
	
	fn := SQLDBCountFn(db, nil)
	_, err := fn(context.Background(), "T1")
	if err == nil {
		t.Error("expected error")
	}
}

func TestConnectPostgres_Invalid(t *testing.T) {
	_, err := ConnectPostgres("invalid://url", 1, 1, 10)
	if err == nil {
		t.Error("expected error")
	}
}

func TestConnectTargetDB_Invalid(t *testing.T) {
	_, err := ConnectTargetDB("mysql", "invalid_dsn")
	if err == nil {
		t.Error("expected error")
	}
}
