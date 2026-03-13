package dialect

import (
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestMapOracleType_MySQL(t *testing.T) {
	d := &MySQLDialect{}
	tests := []struct {
		name      string
		oraType   string
		precision int
		scale     int
		want      string
	}{
		{"VARCHAR2 with precision", "VARCHAR2", 100, 0, "VARCHAR(100)"},
		{"VARCHAR2 over limit", "VARCHAR2", 20000, 0, "LONGTEXT"},
		{"VARCHAR2 no precision", "VARCHAR2", 0, 0, "VARCHAR(255)"},
		{"CHAR with precision", "CHAR", 10, 0, "CHAR(10)"},
		{"CHAR no precision", "CHAR", 0, 0, "CHAR(255)"},
		{"NUMBER with scale", "NUMBER", 10, 2, "DECIMAL(10, 2)"},
		{"NUMBER p<=4", "NUMBER", 3, 0, "SMALLINT"},
		{"NUMBER p<=9", "NUMBER", 8, 0, "INT"},
		{"NUMBER p>9", "NUMBER", 15, 0, "BIGINT"},
		{"NUMBER no precision", "NUMBER", 0, 0, "DOUBLE"},
		{"DATE", "DATE", 0, 0, "DATETIME"},
		{"CLOB", "CLOB", 0, 0, "LONGTEXT"},
		{"BLOB", "BLOB", 0, 0, "LONGBLOB"},
		{"FLOAT", "FLOAT", 0, 0, "DOUBLE"},
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

func TestCreateTableDDL_MySQL(t *testing.T) {
	d := &MySQLDialect{}
	cols := []ColumnDef{
		{Name: "ID", Type: "NUMBER", Precision: sql.NullInt64{Int64: 9, Valid: true}, Nullable: "N"},
		{Name: "NAME", Type: "VARCHAR2", Precision: sql.NullInt64{Int64: 100, Valid: true}, Nullable: "Y"},
	}

	t.Run("with schema", func(t *testing.T) {
		ddl := d.CreateTableDDL("employees", "hr", cols)
		if !strings.Contains(ddl, "CREATE TABLE IF NOT EXISTS `hr`.`employees`") {
			t.Error("expected schema in DDL")
		}
		if !strings.Contains(ddl, "`id` INT NOT NULL") {
			t.Error("expected INT NOT NULL")
		}
	})

	t.Run("without schema", func(t *testing.T) {
		ddl := d.CreateTableDDL("employees", "", cols)
		if !strings.Contains(ddl, "CREATE TABLE IF NOT EXISTS `employees`") {
			t.Error("expected table without schema in DDL")
		}
	})
}

func TestCreateIndexDDL_MySQL(t *testing.T) {
	d := &MySQLDialect{}

	t.Run("regular index", func(t *testing.T) {
		idx := IndexMetadata{
			Name:       "idx_emp_name",
			Uniqueness: "NONUNIQUE",
			Columns:    []IndexColumn{{Name: "name", Position: 1}},
		}
		ddl := d.CreateIndexDDL(idx, "employees", "hr")
		if !strings.Contains(ddl, "CREATE INDEX `idx_emp_name` ON `hr`.`employees` (`name`)") {
			t.Error("expected regular index DDL")
		}
	})

	t.Run("unique index", func(t *testing.T) {
		idx := IndexMetadata{
			Name:       "idx_emp_email",
			Uniqueness: "UNIQUE",
			Columns:    []IndexColumn{{Name: "email", Position: 1}},
		}
		ddl := d.CreateIndexDDL(idx, "employees", "hr")
		if !strings.Contains(ddl, "CREATE UNIQUE INDEX `idx_emp_email`") {
			t.Error("expected unique index DDL")
		}
	})

	t.Run("primary key", func(t *testing.T) {
		idx := IndexMetadata{
			Name:    "pk_emp",
			IsPK:    true,
			Columns: []IndexColumn{{Name: "id", Position: 1}},
		}
		ddl := d.CreateIndexDDL(idx, "employees", "hr")
		if !strings.Contains(ddl, "ALTER TABLE `hr`.`employees` ADD PRIMARY KEY (`id`)") {
			t.Error("expected PK DDL")
		}
	})
}

func TestInsertStatement_MySQL(t *testing.T) {
	d := &MySQLDialect{}

	t.Run("formatting", func(t *testing.T) {
		tm := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
		rows := [][]any{
			{1, "O'Reilly", tm, nil, []byte{0xDE, 0xAD}},
		}
		stmts := d.InsertStatement("emp", "", []string{"id", "name", "dt", "nullval", "bin"}, rows, 100)
		if len(stmts) != 1 {
			t.Errorf("expected 1 statement")
		}
		stmt := stmts[0]
		if !strings.Contains(stmt, "1") || !strings.Contains(stmt, "'O''Reilly'") || !strings.Contains(stmt, "'2025-01-01 12:00:00'") || !strings.Contains(stmt, "NULL") || !strings.Contains(stmt, "X'dead'") {
			t.Errorf("unexpected formatting: %s", stmt)
		}
	})

	t.Run("batching", func(t *testing.T) {
		rows := make([][]any, 5)
		for i := 0; i < 5; i++ {
			rows[i] = []any{i}
		}
		stmts := d.InsertStatement("emp", "", []string{"id"}, rows, 2)
		if len(stmts) != 3 {
			t.Errorf("expected 3 statements, got %d", len(stmts))
		}
	})
}