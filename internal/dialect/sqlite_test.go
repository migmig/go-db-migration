package dialect

import (
	"database/sql"
	"strings"
	"testing"
)

func TestMapOracleType_SQLite(t *testing.T) {
	d := &SQLiteDialect{}
	tests := []struct {
		name      string
		oraType   string
		precision int
		want      string
	}{
		{"VARCHAR2", "VARCHAR2", 100, "TEXT"},
		{"NUMBER p<=9", "NUMBER", 5, "INTEGER"},
		{"NUMBER no precision", "NUMBER", 0, "REAL"},
		{"DATE", "DATE", 0, "TEXT"},
		{"CLOB", "CLOB", 0, "TEXT"},
		{"BLOB", "BLOB", 0, "BLOB"},
		{"FLOAT", "FLOAT", 0, "REAL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := d.MapOracleType(tt.oraType, tt.precision, 0)
			if got != tt.want {
				t.Errorf("MapOracleType(%q) = %q, want %q", tt.oraType, got, tt.want)
			}
		})
	}
}

func TestCreateTableDDL_SQLite(t *testing.T) {
	d := &SQLiteDialect{}
	cols := []ColumnDef{
		{Name: "ID", Type: "NUMBER", Precision: sql.NullInt64{Int64: 9, Valid: true}, Nullable: "N"},
	}

	t.Run("ignores schema", func(t *testing.T) {
		ddl := d.CreateTableDDL("employees", "hr", cols)
		if strings.Contains(ddl, "hr") {
			t.Error("expected schema to be ignored in SQLite DDL")
		}
		if !strings.Contains(ddl, "CREATE TABLE IF NOT EXISTS \"employees\"") {
			t.Error("expected IF NOT EXISTS")
		}
		if !strings.Contains(ddl, "\"id\" INTEGER NOT NULL") {
			t.Error("expected NOT NULL")
		}
	})
}

func TestCreateIndexDDL_SQLite(t *testing.T) {
	d := &SQLiteDialect{}

	t.Run("regular index", func(t *testing.T) {
		idx := IndexMetadata{
			Name:       "idx_emp_name",
			Uniqueness: "NONUNIQUE",
			Columns:    []IndexColumn{{Name: "name", Position: 1}},
		}
		ddl := d.CreateIndexDDL(idx, "employees", "hr")
		if !strings.Contains(ddl, "CREATE INDEX IF NOT EXISTS \"idx_emp_name\" ON \"employees\" (\"name\")") {
			t.Error("expected regular index DDL")
		}
	})

	t.Run("primary key skips", func(t *testing.T) {
		idx := IndexMetadata{
			Name:    "pk_emp",
			IsPK:    true,
			Columns: []IndexColumn{{Name: "id", Position: 1}},
		}
		ddl := d.CreateIndexDDL(idx, "employees", "")
		if ddl != "" {
			t.Error("expected empty string for PK in SQLite (handled in CreateTable)")
		}
	})
}

func TestInsertStatement_SQLite(t *testing.T) {
	d := &SQLiteDialect{}

	t.Run("batching and formatting", func(t *testing.T) {
		rows := [][]any{
			{1, "A"}, {2, "B"}, {3, "C"},
		}
		stmts := d.InsertStatement("emp", "hr", []string{"id", "name"}, rows, 2)
		if len(stmts) != 2 {
			t.Errorf("expected 2 statements, got %d", len(stmts))
		}
		// Schema should be ignored for table name in insert
		if strings.Contains(stmts[0], "hr.emp") || strings.Contains(stmts[0], "hr\".\"emp") {
			t.Errorf("schema should be ignored in sqlite inserts")
		}
	})
}