package migration

import (
	"database/sql"
	"fmt"
	"strings"
	"testing"

	"dbmigrator/internal/dialect"
	"github.com/DATA-DOG/go-sqlmock"
)

func TestMapOracleToPostgres(t *testing.T) {
	nullInt := func(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }
	dia := &dialect.PostgresDialect{}

	tests := []struct {
		name     string
		col      dialect.ColumnDef
		expected string
	}{
		{"VARCHAR2", dialect.ColumnDef{Type: "VARCHAR2"}, "text"},
		{"CHAR", dialect.ColumnDef{Type: "CHAR"}, "text"},
		{"NCHAR", dialect.ColumnDef{Type: "NCHAR"}, "text"},
		{"NVARCHAR2", dialect.ColumnDef{Type: "NVARCHAR2"}, "text"},
		{"NUMBER no precision", dialect.ColumnDef{Type: "NUMBER"}, "numeric"},
		{"NUMBER precision=9 → integer", dialect.ColumnDef{Type: "NUMBER", Precision: nullInt(9)}, "integer"},
		{"NUMBER precision=10 → bigint", dialect.ColumnDef{Type: "NUMBER", Precision: nullInt(10)}, "bigint"},
		{"NUMBER precision=1 → smallint", dialect.ColumnDef{Type: "NUMBER", Precision: nullInt(1)}, "smallint"},
		{"NUMBER with scale → numeric(p,s)", dialect.ColumnDef{Type: "NUMBER", Precision: nullInt(10), Scale: nullInt(2)}, "numeric(10, 2)"},
		{"NUMBER scale=0 not numeric", dialect.ColumnDef{Type: "NUMBER", Precision: nullInt(5), Scale: nullInt(0)}, "integer"},
		{"DATE", dialect.ColumnDef{Type: "DATE"}, "timestamp"},
		{"TIMESTAMP", dialect.ColumnDef{Type: "TIMESTAMP(6)"}, "timestamp"},
		{"TIMESTAMP WITH TZ", dialect.ColumnDef{Type: "TIMESTAMP WITH TIME ZONE"}, "timestamp"},
		{"CLOB", dialect.ColumnDef{Type: "CLOB"}, "text"},
		{"BLOB", dialect.ColumnDef{Type: "BLOB"}, "bytea"},
		{"RAW", dialect.ColumnDef{Type: "RAW"}, "bytea"},
		{"FLOAT", dialect.ColumnDef{Type: "FLOAT"}, "double precision"},
		{"lowercase float", dialect.ColumnDef{Type: "float"}, "double precision"},
		{"unknown type falls back to text", dialect.ColumnDef{Type: "XMLTYPE"}, "text"},
		{"unknown type INTERVAL", dialect.ColumnDef{Type: "INTERVAL YEAR TO MONTH"}, "text"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prec := 0
			if tt.col.Precision.Valid {
				prec = int(tt.col.Precision.Int64)
			}
			s := 0
			if tt.col.Scale.Valid {
				s = int(tt.col.Scale.Int64)
			}
			result := dia.MapOracleType(tt.col.Type, prec, s)
			if result != tt.expected {
				t.Errorf("MapOracleType(%+v) = %q; want %q", tt.col, result, tt.expected)
			}
		})
	}
}

func TestGenerateCreateTableDDL_WithSchema(t *testing.T) {
	nullInt := func(v int64) sql.NullInt64 { return sql.NullInt64{Int64: v, Valid: true} }
	cols := []dialect.ColumnDef{
		{Name: "ID", Type: "NUMBER", Precision: nullInt(10), Nullable: "N"},
		{Name: "NAME", Type: "VARCHAR2", Nullable: "Y"},
		{Name: "SCORE", Type: "NUMBER", Precision: nullInt(8), Scale: nullInt(2), Nullable: "Y"},
	}
	dia := &dialect.PostgresDialect{}

	ddl := GenerateCreateTableDDL("USERS", "public", cols, dia)

	if !strings.Contains(ddl, "CREATE TABLE IF NOT EXISTS \"public\".\"users\"") {
		t.Errorf("expected schema.table in DDL, got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "\"id\" bigint NOT NULL") {
		t.Errorf("expected 'id bigint NOT NULL', got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "\"name\" text") {
		t.Errorf("expected 'name text', got:\n%s", ddl)
	}
	if strings.Contains(ddl, "name text NOT NULL") {
		t.Errorf("nullable column should not have NOT NULL, got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "\"score\" numeric(8, 2)") {
		t.Errorf("expected 'score numeric(8, 2)', got:\n%s", ddl)
	}
}

func TestGenerateCreateTableDDL_WithoutSchema(t *testing.T) {
	cols := []dialect.ColumnDef{
		{Name: "ID", Type: "NUMBER", Nullable: "N"},
	}
	dia := &dialect.PostgresDialect{}

	ddl := GenerateCreateTableDDL("ORDERS", "", cols, dia)

	if !strings.Contains(ddl, "CREATE TABLE IF NOT EXISTS \"orders\"") {
		t.Errorf("expected table without schema prefix, got:\n%s", ddl)
	}
	if strings.Contains(ddl, ".\"orders\"") {
		t.Errorf("schema separator should not appear when schema is empty, got:\n%s", ddl)
	}
}

func TestGenerateCreateTableDDL_EndsWithSemicolon(t *testing.T) {
	cols := []dialect.ColumnDef{{Name: "X", Type: "VARCHAR2", Nullable: "Y"}}
	dia := &dialect.PostgresDialect{}
	ddl := GenerateCreateTableDDL("T", "", cols, dia)
	trimmed := strings.TrimSpace(ddl)
	if !strings.HasSuffix(trimmed, ";") {
		t.Errorf("DDL should end with ';', got:\n%s", ddl)
	}
}

func TestGenerateCreateTableDDL_PostgresDefaultNextvalConverted(t *testing.T) {
	cols := []dialect.ColumnDef{
		{Name: "ID", Type: "NUMBER", Nullable: "N", DefaultValue: sql.NullString{String: "USERS_SEQ.NEXTVAL", Valid: true}},
	}
	dia := &dialect.PostgresDialect{}

	ddl := GenerateCreateTableDDL("USERS", "public", cols, dia)
	if !strings.Contains(ddl, "DEFAULT nextval('users_seq')") {
		t.Fatalf("expected Postgres nextval default conversion, got:\n%s", ddl)
	}
}

func TestGenerateCreateTableDDL_PostgresDefaultSysdateConverted(t *testing.T) {
	cols := []dialect.ColumnDef{
		{Name: "CREATED_AT", Type: "DATE", Nullable: "N", DefaultValue: sql.NullString{String: "SYSDATE", Valid: true}},
	}
	dia := &dialect.PostgresDialect{}

	ddl := GenerateCreateTableDDL("AUDIT_LOG", "", cols, dia)
	if !strings.Contains(ddl, "DEFAULT CURRENT_TIMESTAMP") {
		t.Fatalf("expected SYSDATE to become CURRENT_TIMESTAMP, got:\n%s", ddl)
	}
}

func TestGenerateCreateTableDDL_ColumnNamesLowercased(t *testing.T) {
	cols := []dialect.ColumnDef{
		{Name: "MY_COL", Type: "VARCHAR2", Nullable: "Y"},
	}
	dia := &dialect.PostgresDialect{}
	ddl := GenerateCreateTableDDL("T", "", cols, dia)
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
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "data_precision", "data_scale", "nullable", "data_default"}).
			AddRow("ID", "NUMBER", 10, 0, "N", nil).
			AddRow("EMAIL", "VARCHAR2", nil, nil, "Y", nil))

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
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "data_type", "data_precision", "data_scale", "nullable", "data_default"}))

	_, err = GetTableMetadata(db, "users")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations (table name not uppercased): %v", err)
	}
}

// ── GenerateSequenceDDL ────────────────────────────────────────────────────────

func TestGenerateSequenceDDL_Basic(t *testing.T) {
	seq := dialect.SequenceMetadata{
		Name:        "USERS_SEQ",
		MinValue:    1,
		MaxValue:    "100000",
		IncrementBy: 1,
		CycleFlag:   "N",
		LastNumber:  42,
	}
	dia := &dialect.PostgresDialect{}
	ddl, _ := GenerateSequenceDDL(seq, "", dia)

	if !strings.Contains(ddl, "CREATE SEQUENCE IF NOT EXISTS \"users_seq\"") {
		t.Errorf("expected sequence name in DDL, got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "START WITH 42") {
		t.Errorf("expected START WITH 42, got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "INCREMENT BY 1") {
		t.Errorf("expected INCREMENT BY 1, got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "MINVALUE 1") {
		t.Errorf("expected MINVALUE 1, got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "MAXVALUE 100000") {
		t.Errorf("expected MAXVALUE 100000, got:\n%s", ddl)
	}
	if !strings.Contains(ddl, "NO CYCLE") {
		t.Errorf("expected NO CYCLE, got:\n%s", ddl)
	}
}

func TestGenerateSequenceDDL_MaxValueOmit(t *testing.T) {
	seq := dialect.SequenceMetadata{
		Name:        "MY_SEQ",
		MinValue:    1,
		MaxValue:    "9999999999999999999999999999", // Oracle 기본값
		IncrementBy: 1,
		CycleFlag:   "N",
		LastNumber:  1,
	}
	dia := &dialect.PostgresDialect{}
	ddl, _ := GenerateSequenceDDL(seq, "", dia)

	if strings.Contains(ddl, "MAXVALUE") {
		t.Errorf("Oracle 기본 MAXVALUE는 생략되어야 하는데 포함됨:\n%s", ddl)
	}
}

func TestGenerateSequenceDDL_Cycle(t *testing.T) {
	seq := dialect.SequenceMetadata{
		Name:        "CYCLE_SEQ",
		MinValue:    1,
		MaxValue:    "1000",
		IncrementBy: 1,
		CycleFlag:   "Y",
		LastNumber:  1,
	}
	dia := &dialect.PostgresDialect{}
	ddl, _ := GenerateSequenceDDL(seq, "", dia)

	if strings.Contains(ddl, "NO CYCLE") {
		t.Errorf("CycleFlag=Y 이면 NO CYCLE이 아닌 CYCLE이어야 함:\n%s", ddl)
	}
	if !strings.Contains(ddl, "CYCLE") {
		t.Errorf("CycleFlag=Y 이면 CYCLE이 포함되어야 함:\n%s", ddl)
	}
}

func TestGenerateSequenceDDL_WithSchema(t *testing.T) {
	seq := dialect.SequenceMetadata{
		Name:        "ORDER_SEQ",
		MinValue:    1,
		MaxValue:    "9999999999999999999999999999",
		IncrementBy: 1,
		CycleFlag:   "N",
		LastNumber:  100,
	}
	dia := &dialect.PostgresDialect{}
	ddl, _ := GenerateSequenceDDL(seq, "myschema", dia)

	if !strings.Contains(ddl, "CREATE SEQUENCE IF NOT EXISTS \"myschema\".\"order_seq\"") {
		t.Errorf("스키마 접두사가 포함되어야 함, got:\n%s", ddl)
	}
}

// ── GenerateIndexDDL ───────────────────────────────────────────────────────────

func TestGenerateIndexDDL_Normal(t *testing.T) {
	idx := dialect.IndexMetadata{
		Name:       "IDX_USERS_EMAIL",
		Uniqueness: "NONUNIQUE",
		IndexType:  "NORMAL",
		IsPK:       false,
		Columns:    []dialect.IndexColumn{{Name: "EMAIL", Position: 1, Descend: "ASC"}},
	}
	dia := &dialect.PostgresDialect{}
	ddl := GenerateIndexDDL(idx, "USERS", "", dia)

	if !strings.Contains(ddl, "CREATE INDEX IF NOT EXISTS \"idx_users_email\" ON \"users\" (\"email\")") {
		t.Errorf("일반 인덱스 DDL 불일치:\n%s", ddl)
	}
}

func TestGenerateIndexDDL_Unique(t *testing.T) {
	idx := dialect.IndexMetadata{
		Name:       "UQ_USERS_EMAIL",
		Uniqueness: "UNIQUE",
		IndexType:  "NORMAL",
		IsPK:       false,
		Columns:    []dialect.IndexColumn{{Name: "EMAIL", Position: 1, Descend: "ASC"}},
	}
	dia := &dialect.PostgresDialect{}
	ddl := GenerateIndexDDL(idx, "USERS", "", dia)

	if !strings.Contains(ddl, "CREATE UNIQUE INDEX IF NOT EXISTS") {
		t.Errorf("UNIQUE INDEX 키워드가 없음:\n%s", ddl)
	}
	if !strings.Contains(ddl, "uq_users_email") {
		t.Errorf("인덱스 이름(소문자)이 없음:\n%s", ddl)
	}
}

func TestGenerateIndexDDL_PrimaryKey(t *testing.T) {
	idx := dialect.IndexMetadata{
		Name:       "SYS_C001234",
		Uniqueness: "UNIQUE",
		IndexType:  "NORMAL",
		IsPK:       true,
		Columns:    []dialect.IndexColumn{{Name: "ID", Position: 1, Descend: "ASC"}},
	}
	dia := &dialect.PostgresDialect{}
	ddl := GenerateIndexDDL(idx, "USERS", "", dia)

	if !strings.Contains(ddl, "ALTER TABLE \"users\" ADD PRIMARY KEY (\"id\")") {
		t.Errorf("PK는 ALTER TABLE ... ADD PRIMARY KEY 형태여야 함:\n%s", ddl)
	}
}

func TestGenerateIndexDDL_Descend(t *testing.T) {
	idx := dialect.IndexMetadata{
		Name:       "IDX_USERS_CREATED",
		Uniqueness: "NONUNIQUE",
		IndexType:  "NORMAL",
		IsPK:       false,
		Columns:    []dialect.IndexColumn{{Name: "CREATED_AT", Position: 1, Descend: "DESC"}},
	}
	dia := &dialect.PostgresDialect{}
	ddl := GenerateIndexDDL(idx, "USERS", "", dia)

	if !strings.Contains(ddl, "\"created_at\" DESC") {
		t.Errorf("DESC 컬럼 표현식이 없음:\n%s", ddl)
	}
}
