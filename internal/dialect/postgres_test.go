package dialect

import (
	"database/sql"
	"strings"
	"testing"
	"time"
)

func TestPostgresDialect_Name(t *testing.T) {
	d := &PostgresDialect{}
	if n := d.Name(); n != "postgres" {
		t.Errorf("expected postgres, got %s", n)
	}
}

func TestPostgresDialect_QuoteIdentifier(t *testing.T) {
	d := &PostgresDialect{}
	if q := d.QuoteIdentifier("user"); q != "\"user\"" {
		t.Errorf("expected \"user\", got %s", q)
	}
}

func TestPostgresDialect_DriverName(t *testing.T) {
	d := &PostgresDialect{}
	if n := d.DriverName(); n != "pgx" {
		t.Errorf("expected pgx, got %s", n)
	}
}

func TestPostgresDialect_NormalizeURL(t *testing.T) {
	d := &PostgresDialect{}
	if u := d.NormalizeURL("url"); u != "url" {
		t.Errorf("expected url, got %s", u)
	}
}

func TestPostgresDialect_MapOracleType(t *testing.T) {
	d := &PostgresDialect{}
	tests := []struct {
		name      string
		oraType   string
		precision int
		scale     int
		want      string
	}{
		{"VARCHAR2", "VARCHAR2", 100, 0, "text"},
		{"CHAR", "CHAR", 10, 0, "text"},
		{"NUMBER no prec", "NUMBER", 0, 0, "numeric"},
		{"NUMBER p<=4", "NUMBER", 4, 0, "smallint"},
		{"NUMBER p<=9", "NUMBER", 9, 0, "integer"},
		{"NUMBER p>9", "NUMBER", 10, 0, "bigint"},
		{"NUMBER with scale", "NUMBER", 10, 2, "numeric(10, 2)"},
		{"DATE", "DATE", 0, 0, "timestamp"},
		{"TIMESTAMP", "TIMESTAMP", 0, 0, "timestamp"},
		{"CLOB", "CLOB", 0, 0, "text"},
		{"BLOB", "BLOB", 0, 0, "bytea"},
		{"RAW", "RAW", 0, 0, "bytea"},
		{"FLOAT", "FLOAT", 0, 0, "double precision"},
		{"UNKNOWN", "UNKNOWN", 0, 0, "text"},
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

func TestNormalizeOracleDefaultForPostgres(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{" SYSDATE ", "CURRENT_TIMESTAMP"},
		{"SYSTIMESTAMP", "CURRENT_TIMESTAMP"},
		{"seq.nextval", "nextval('seq')"},
		{"\"schema\".\"seq\".nextval", "nextval('\"schema\".\"seq\"')"},
		{"'A'", "'A'"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeOracleDefaultForPostgres(tt.input)
			if got != tt.want {
				t.Errorf("normalizeOracleDefaultForPostgres(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCreateTableDDL_Postgres(t *testing.T) {
	d := &PostgresDialect{}
	cols := []ColumnDef{
		{Name: "ID", Type: "NUMBER", Precision: sql.NullInt64{Int64: 9, Valid: true}, Nullable: "N"},
		{Name: "NAME", Type: "VARCHAR2", Nullable: "Y", DefaultValue: sql.NullString{String: "'A'", Valid: true}},
	}

	t.Run("with schema", func(t *testing.T) {
		ddl := d.CreateTableDDL("employees", "hr", cols)
		if !strings.Contains(ddl, "\"hr\".\"employees\"") {
			t.Error("expected schema in DDL")
		}
		if !strings.Contains(ddl, "\"id\" integer NOT NULL") {
			t.Error("expected integer NOT NULL")
		}
		if !strings.Contains(ddl, "DEFAULT 'A'") {
			t.Error("expected DEFAULT 'A'")
		}
	})
	t.Run("no schema", func(t *testing.T) {
		ddl := d.CreateTableDDL("employees", "", cols)
		if !strings.Contains(ddl, "CREATE TABLE IF NOT EXISTS \"employees\"") {
			t.Error("expected CREATE TABLE IF NOT EXISTS")
		}
	})
}

func TestCreateConstraintDDL_Postgres(t *testing.T) {
	d := &PostgresDialect{}
	t.Run("foreign key", func(t *testing.T) {
		c := ConstraintMetadata{
			Name:         "fk_emp_dept",
			Type:         "R",
			TableName:    "EMP",
			Columns:      []string{"DEPT_ID"},
			RefTableName: "DEPT",
			RefColumns:   []string{"ID"},
			DeleteRule:   "CASCADE",
		}
		ddl := d.CreateConstraintDDL(c, "HR")
		if !strings.Contains(ddl, "ALTER TABLE \"hr\".\"emp\" ADD CONSTRAINT \"fk_emp_dept\" FOREIGN KEY (\"dept_id\") REFERENCES \"hr\".\"dept\" (\"id\") ON DELETE CASCADE") {
			t.Errorf("unexpected ddl: %s", ddl)
		}
	})
	t.Run("check constraint", func(t *testing.T) {
		c := ConstraintMetadata{
			Name:            "chk_emp_sal",
			Type:            "C",
			TableName:       "EMP",
			SearchCondition: "SAL > 0",
		}
		ddl := d.CreateConstraintDDL(c, "")
		if !strings.Contains(ddl, "CHECK (SAL > 0)") {
			t.Errorf("unexpected ddl: %s", ddl)
		}
	})
}

func TestCreateSequenceDDL_Postgres(t *testing.T) {
	d := &PostgresDialect{}
	seq := SequenceMetadata{
		Name:        "SEQ_EMP",
		LastNumber:  1,
		IncrementBy: 1,
		MinValue:    1,
		MaxValue:    "9999999999999999999999999999",
		CycleFlag:   "Y",
	}
	ddl, ok := d.CreateSequenceDDL(seq, "HR")
	if !ok {
		t.Error("expected true")
	}
	if !strings.Contains(ddl, "\"hr\".\"seq_emp\"") {
		t.Errorf("unexpected ddl: %s", ddl)
	}
	if !strings.Contains(ddl, "CYCLE") {
		t.Errorf("unexpected ddl: %s", ddl)
	}
}

func TestCreateIndexDDL_Postgres(t *testing.T) {
	d := &PostgresDialect{}
	t.Run("primary key", func(t *testing.T) {
		idx := IndexMetadata{
			Name:    "pk_emp",
			IsPK:    true,
			Columns: []IndexColumn{{Name: "id", Position: 1}},
		}
		ddl := d.CreateIndexDDL(idx, "emp", "hr")
		if !strings.Contains(ddl, "ALTER TABLE \"hr\".\"emp\" ADD PRIMARY KEY (\"id\")") {
			t.Errorf("unexpected ddl: %s", ddl)
		}
	})
	t.Run("unique index desc", func(t *testing.T) {
		idx := IndexMetadata{
			Name:       "idx_emp_name",
			Uniqueness: "UNIQUE",
			Columns:    []IndexColumn{{Name: "name", Position: 1, Descend: "DESC"}},
		}
		ddl := d.CreateIndexDDL(idx, "emp", "")
		if !strings.Contains(ddl, "CREATE UNIQUE INDEX IF NOT EXISTS \"idx_emp_name\" ON \"emp\" (\"name\" DESC)") {
			t.Errorf("unexpected ddl: %s", ddl)
		}
	})
	t.Run("regular index", func(t *testing.T) {
		idx := IndexMetadata{
			Name:       "idx_emp_name",
			Uniqueness: "NONUNIQUE",
			Columns:    []IndexColumn{{Name: "name", Position: 1}},
		}
		ddl := d.CreateIndexDDL(idx, "emp", "")
		if !strings.Contains(ddl, "CREATE INDEX IF NOT EXISTS \"idx_emp_name\" ON \"emp\" (\"name\")") {
			t.Errorf("unexpected ddl: %s", ddl)
		}
	})
}

func TestInsertStatement_Postgres(t *testing.T) {
	d := &PostgresDialect{}
	cols := []string{"id", "name"}
	rows := [][]any{
		{1, "A"}, {2, "B"}, {3, "C"},
	}
	stmts := d.InsertStatement("emp", "hr", cols, rows, 2)
	if len(stmts) != 2 {
		t.Errorf("expected 2 statements, got %d", len(stmts))
	}
	if !strings.Contains(stmts[0], "INSERT INTO \"hr\".\"emp\"") {
		t.Errorf("unexpected ddl: %s", stmts[0])
	}
	if !strings.Contains(stmts[0], "(1, 'A')") {
		t.Errorf("unexpected ddl: %s", stmts[0])
	}
}

func TestFormatValue_Postgres(t *testing.T) {
	d := &PostgresDialect{}
	if d.formatValue(nil) != "NULL" {
		t.Error("expected NULL")
	}
	if d.formatValue("It's") != "'It''s'" {
		t.Error("expected 'It''s'")
	}
	if d.formatValue([]byte{0x01, 0x02}) != "'\\x0102'" {
		t.Error("expected '\\x0102'")
	}
	tm := time.Date(2020, 1, 1, 10, 0, 0, 0, time.UTC)
	if !strings.Contains(d.formatValue(tm), "2020-01-01") {
		t.Error("expected date string")
	}
	if d.formatValue(10) != "10" {
		t.Error("expected 10")
	}
	if d.formatValue(true) != "TRUE" {
		t.Error("expected TRUE")
	}
	if d.formatValue(false) != "FALSE" {
		t.Error("expected FALSE")
	}
	if d.formatValue(struct{}{}) != "'{}'" {
		t.Error("expected '{}'")
	}
}