package dialect

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"
)

func TestMapOracleType_MSSQL(t *testing.T) {
	d := &MSSQLDialect{}
	tests := []struct {
		name      string
		oraType   string
		precision int
		scale     int
		want      string
	}{
		{"VARCHAR2 precision<=4000", "VARCHAR2", 500, 0, "NVARCHAR(500)"},
		{"VARCHAR2 precision>4000", "VARCHAR2", 8000, 0, "NVARCHAR(MAX)"},
		{"VARCHAR2 no precision", "VARCHAR2", 0, 0, "NVARCHAR(MAX)"},
		{"CHAR precision<=4000", "CHAR", 500, 0, "NCHAR(500)"},
		{"CHAR precision>4000", "CHAR", 8000, 0, "NCHAR(4000)"},
		{"CHAR no precision", "CHAR", 0, 0, "NCHAR(255)"},
		{"NUMBER with scale", "NUMBER", 10, 2, "DECIMAL(10, 2)"},
		{"NUMBER p<=4", "NUMBER", 3, 0, "SMALLINT"},
		{"NUMBER p<=9", "NUMBER", 8, 0, "INT"},
		{"NUMBER p>9", "NUMBER", 15, 0, "BIGINT"},
		{"NUMBER no precision", "NUMBER", 0, 0, "NUMERIC"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.MapOracleType(tt.oraType, tt.precision, tt.scale)
			if got != tt.want {
				t.Errorf("MapOracleType(%q, %d, %d) = %q, want %q", tt.oraType, tt.precision, tt.scale, got, tt.want)
			}
		})
	}
}

func TestCreateTableDDL_MSSQL(t *testing.T) {
	d := &MSSQLDialect{}
	cols := []ColumnDef{
		{Name: "ID", Type: "NUMBER", Precision: sql.NullInt64{Int64: 9, Valid: true}, Nullable: "N"},
		{Name: "NAME", Type: "VARCHAR2", Precision: sql.NullInt64{Int64: 100, Valid: true}, Nullable: "Y"},
	}

	t.Run("with schema", func(t *testing.T) {
		ddl := d.CreateTableDDL("employees", "hr", cols)
		if !strings.Contains(ddl, "TABLE_SCHEMA = 'hr'") {
			t.Error("expected TABLE_SCHEMA condition in DDL")
		}
		if !strings.Contains(ddl, "TABLE_NAME = 'employees'") {
			t.Error("expected TABLE_NAME condition in DDL")
		}
	})

	t.Run("without schema uses dbo", func(t *testing.T) {
		ddl := d.CreateTableDDL("employees", "", cols)
		if !strings.Contains(ddl, "TABLE_SCHEMA = 'dbo'") {
			t.Error("expected TABLE_SCHEMA = 'dbo' as default")
		}
	})
}

func TestCreateIndexDDL_MSSQL(t *testing.T) {
	d := &MSSQLDialect{}

	t.Run("regular index with object_id filter", func(t *testing.T) {
		idx := IndexMetadata{
			Name:       "idx_emp_name",
			Uniqueness: "NONUNIQUE",
			Columns:    []IndexColumn{{Name: "name", Position: 1}},
		}
		ddl := d.CreateIndexDDL(idx, "employees", "hr")
		if !strings.Contains(ddl, "sys.objects o ON i.object_id = o.object_id") {
			t.Error("expected object_id join in DDL")
		}
		if !strings.Contains(ddl, "o.name = 'employees'") {
			t.Error("expected table name filter in DDL")
		}
		if !strings.Contains(ddl, "SCHEMA_NAME(o.schema_id) = 'hr'") {
			t.Error("expected schema filter in DDL")
		}
	})

	t.Run("primary key", func(t *testing.T) {
		idx := IndexMetadata{
			Name:    "pk_emp",
			IsPK:    true,
			Columns: []IndexColumn{{Name: "id", Position: 1}},
		}
		ddl := d.CreateIndexDDL(idx, "employees", "hr")
		if !strings.Contains(ddl, "ALTER TABLE") {
			t.Error("PK should use ALTER TABLE ADD PRIMARY KEY")
		}
	})
}

func TestInsertStatement_MSSQL(t *testing.T) {
	d := &MSSQLDialect{}

	t.Run("batch limit 1000", func(t *testing.T) {
		rows := make([][]any, 1500)
		for i := range rows {
			rows[i] = []any{i, fmt.Sprintf("name_%d", i)}
		}

		stmts := d.InsertStatement("emp", "", []string{"id", "name"}, rows, 2000)
		if len(stmts) != 2 {
			t.Errorf("expected 2 statements, got %d", len(stmts))
		}
	})

	t.Run("formatting", func(t *testing.T) {
		rows := [][]any{
			{"hello", []byte{0xAB, 0xCD}},
		}
		stmts := d.InsertStatement("emp", "", []string{"str", "bin"}, rows, 100)
		stmt := stmts[0]
		if !strings.Contains(stmt, "N'hello'") {
			t.Errorf("expected N'' prefix for string")
		}
		if !strings.Contains(stmt, "0xabcd") {
			t.Errorf("expected 0x prefix for binary")
		}
	})
}