package migration

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestMapOracleToPostgres(t *testing.T) {
	nullInt := func(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }

	tests := []struct {
		name     string
		col      ColumnMetadata
		expected string
	}{
		{"VARCHAR2", ColumnMetadata{Type: "VARCHAR2"}, "text"},
		{"CHAR", ColumnMetadata{Type: "CHAR"}, "text"},
		{"NCHAR", ColumnMetadata{Type: "NCHAR"}, "text"},
		{"NVARCHAR2", ColumnMetadata{Type: "NVARCHAR2"}, "text"},
		{"NUMBER no precision", ColumnMetadata{Type: "NUMBER"}, "numeric"},
		{"NUMBER precision=9 → integer", ColumnMetadata{Type: "NUMBER", Precision: nullInt(9)}, "integer"},
		{"NUMBER precision=10 → bigint", ColumnMetadata{Type: "NUMBER", Precision: nullInt(10)}, "bigint"},
		{"NUMBER precision=1 → integer", ColumnMetadata{Type: "NUMBER", Precision: nullInt(1)}, "integer"},
		{"NUMBER with scale → numeric(p,s)", ColumnMetadata{Type: "NUMBER", Precision: nullInt(10), Scale: nullInt(2)}, "numeric(10, 2)"},
		{"NUMBER scale=0 not numeric", ColumnMetadata{Type: "NUMBER", Precision: nullInt(5), Scale: nullInt(0)}, "integer"},
		{"DATE", ColumnMetadata{Type: "DATE"}, "timestamp"},
		{"TIMESTAMP", ColumnMetadata{Type: "TIMESTAMP(6)"}, "timestamp"},
		{"TIMESTAMP WITH TZ", ColumnMetadata{Type: "TIMESTAMP WITH TIME ZONE"}, "timestamp"},
		{"CLOB", ColumnMetadata{Type: "CLOB"}, "text"},
		{"BLOB", ColumnMetadata{Type: "BLOB"}, "bytea"},
		{"RAW", ColumnMetadata{Type: "RAW"}, "bytea"},
		{"FLOAT", ColumnMetadata{Type: "FLOAT"}, "double precision"},
		{"lowercase float", ColumnMetadata{Type: "float"}, "double precision"},
		{"unknown type falls back to text", ColumnMetadata{Type: "XMLTYPE"}, "text"},
		{"unknown type INTERVAL", ColumnMetadata{Type: "INTERVAL YEAR TO MONTH"}, "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapOracleToPostgres(tt.col)
			if result != tt.expected {
				t.Errorf("MapOracleToPostgres(%+v) = %q; want %q", tt.col, result, tt.expected)
			}
		})
	}
}

func TestGenerateCreateTableDDL_WithSchema(t *testing.T) {
	nullInt := func(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }
	cols := []ColumnMetadata{
		{Name: "ID", Type: "NUMBER", Precision: nullInt(10), Nullable: "N"},
		{Name: "NAME", Type: "VARCHAR2", Nullable: "Y"},
		{Name: "SCORE", Type: "NUMBER", Precision: nullInt(8), Scale: nullInt(2), Nullable: "Y"},
	}

	ddl := GenerateCreateTableDDL("USERS", "public", cols)

	if !strings.Contains(ddl, "CREATE TABLE IF NOT EXISTS public.USERS") {
		t.Errorf("expected schema.table in DDL, got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "id bigint NOT NULL") {
		t.Errorf("expected 'id bigint NOT NULL', got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "name text") {
		t.Errorf("expected 'name text', got:\n%s", ddl)
	}
	if strings.Contains(ddl, "name text NOT NULL") {
		t.Errorf("nullable column should not have NOT NULL, got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "score numeric(8, 2)") {
		t.Errorf("expected 'score numeric(8, 2)', got:\n%s", ddl)
	}
}

func TestGenerateCreateTableDDL_WithoutSchema(t *testing.T) {
	cols := []ColumnMetadata{
		{Name: "ID", Type: "NUMBER", Nullable: "N"},
	}

	ddl := GenerateCreateTableDDL("ORDERS", "", cols)

	if !strings.Contains(ddl, "CREATE TABLE IF NOT EXISTS ORDERS") {
		t.Errorf("expected table without schema prefix, got:\n%s", ddl)
	}
	if strings.Contains(ddl, ".ORDERS") {
		t.Errorf("schema separator should not appear when schema is empty, got:\n%s", ddl)
	}
}

func TestGenerateCreateTableDDL_EndsWithSemicolon(t *testing.T) {
	cols := []ColumnMetadata{{Name: "X", Type: "VARCHAR2", Nullable: "Y"}}
	ddl := GenerateCreateTableDDL("T", "", cols)
	trimmed := strings.TrimSpace(ddl)
	if !strings.HasSuffix(trimmed, ";") {
		t.Errorf("DDL should end with ';', got:\n%s", ddl)
	}
}

func TestGenerateCreateTableDDL_ColumnNamesLowercased(t *testing.T) {
	cols := []ColumnMetadata{
		{Name: "MY_COL", Type: "VARCHAR2", Nullable: "Y"},
	}
	ddl := GenerateCreateTableDDL("T", "", cols)
	if !strings.Contains(ddl, "my_col") {
		t.Errorf("expected column name to be lowercased, got:\n%s", ddl)
	}
}

func TestGetTableMetadata_ReturnsColumns(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT column_name, data_type").
		WithArgs("USERS").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "data_precision", "data_scale", "nullable"}).
			AddRow("ID", "NUMBER", 10, 0, "N").
			AddRow("EMAIL", "VARCHAR2", nil, nil, "Y"))

	cols, err := GetTableMetadata(db, "USERS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cols) != 2 {
		t.Fatalf("expected 2 columns, got %d", len(cols))
	}
	if cols[0].Name != "ID" || cols[0].Type != "NUMBER" || cols[0].Nullable != "N" {
		t.Errorf("unexpected first column: %+v", cols[0])
	}
	if cols[1].Name != "EMAIL" || cols[1].Type != "VARCHAR2" || cols[1].Nullable != "Y" {
		t.Errorf("unexpected second column: %+v", cols[1])
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

func TestGetTableMetadata_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT column_name, data_type").
		WillReturnError(fmt.Errorf("oracle connection lost"))

	_, err = GetTableMetadata(db, "USERS")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestGetTableMetadata_TableNameUppercased(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	// Expect the uppercase version of "users"
	mock.ExpectQuery("SELECT column_name, data_type").
		WithArgs("USERS").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "data_precision", "data_scale", "nullable"}))

	_, err = GetTableMetadata(db, "users")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations (table name not uppercased): %v", err)
	}
}
