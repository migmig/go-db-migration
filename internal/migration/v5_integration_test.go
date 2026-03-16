package migration

import (
	"bufio"
	"bytes"
	"database/sql"
	"strings"
	"sync"
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/dialect"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// ── mock tracker for DDLProgressTracker tests ─────────────────────────────────

type mockDDLProgressTracker struct {
	ddlEvents []struct{ object, name, status string }
}

func (m *mockDDLProgressTracker) Init(table string, total int)       {}
func (m *mockDDLProgressTracker) Update(table string, processed int) {}
func (m *mockDDLProgressTracker) Done(table string)                  {}
func (m *mockDDLProgressTracker) Error(table string, err error)      {}
func (m *mockDDLProgressTracker) DDLProgress(object, name, status string, err error) {
	m.ddlEvents = append(m.ddlEvents, struct{ object, name, status string }{object, name, status})
}

// ── helpers ───────────────────────────────────────────────────────────────────

func newFileBuf() (*bytes.Buffer, *bufio.Writer) {
	var buf bytes.Buffer
	return &buf, bufio.NewWriter(&buf)
}

// ── 7-2: WithSequences=true → Sequence DDL이 CREATE TABLE 이전에 출력 ───────────

func TestWithSequences_DDLOutputBeforeCreateTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// 1) GetSequenceMetadata: defaultQuery — 빈 결과 반환 (패턴 기반 이름 사용)
	mock.ExpectQuery("SELECT data_default FROM all_tab_columns").
		WillReturnRows(sqlmock.NewRows([]string{"data_default"}))

	// 2) GetSequenceMetadata: seqQuery — USERS_SEQ 메타데이터 반환
	mock.ExpectQuery("all_sequences").
		WillReturnRows(sqlmock.NewRows([]string{
			"sequence_name", "min_value", "max_value",
			"increment_by", "cycle_flag", "last_number",
		}).AddRow("USERS_SEQ", int64(1), "9999999999999999999999999999", int64(1), "N", int64(100)))

	// 3) GetTableMetadata
	mock.ExpectQuery("SELECT column_name, data_type").
		WithArgs("USERS").
		WillReturnRows(sqlmock.NewRows([]string{
			"column_name", "data_type", "data_precision", "data_scale", "nullable", "data_default",
		}).AddRow("ID", "NUMBER", int64(10), nil, "N", nil))

	// 4) No primary key
	mock.ExpectQuery("SELECT c.constraint_name").
		WithArgs("OWNER", "USERS").
		WillReturnError(sql.ErrNoRows)

	// 5) SELECT * FROM USERS — 행 없음
	mock.ExpectQuery("SELECT \\* FROM \"USERS\"").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))

	buf, w := newFileBuf()
	var mu sync.Mutex
	cfg := &config.Config{
		WithDDL:       true,
		WithSequences: true,
		WithIndexes:   false,
		User:          "owner",
		BatchSize:     100,
		PerTable:      false,
	}

	dia := &dialect.PostgresDialect{}
	if _, err := MigrateTableToFile(db, dia, "USERS", w, cfg, &mu, nil, NewMigrationState("test")); err != nil {
		t.Fatalf("MigrateTableToFile: %v", err)
	}
	w.Flush()
	out := buf.String()

	seqIdx := strings.Index(out, "CREATE SEQUENCE")
	tableIdx := strings.Index(out, "CREATE TABLE")

	if seqIdx == -1 {
		t.Fatalf("Sequence DDL이 출력에 없음:\n%s", out)
	}
	if tableIdx == -1 {
		t.Fatalf("CREATE TABLE이 출력에 없음:\n%s", out)
	}
	if seqIdx > tableIdx {
		t.Errorf("Sequence DDL(pos=%d)이 CREATE TABLE(pos=%d) 이후에 출력됨", seqIdx, tableIdx)
	}
}

// ── 7-2: WithIndexes=true → Index DDL이 CREATE TABLE 이후에 출력 ─────────────────

func TestWithIndexes_DDLOutputAfterCreateTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// 1) GetTableMetadata
	mock.ExpectQuery("SELECT column_name, data_type").
		WithArgs("USERS").
		WillReturnRows(sqlmock.NewRows([]string{
			"column_name", "data_type", "data_precision", "data_scale", "nullable", "data_default",
		}).AddRow("ID", "NUMBER", int64(10), nil, "N", nil))

	// 2) No primary key
	mock.ExpectQuery("SELECT c.constraint_name").
		WithArgs("OWNER", "USERS").
		WillReturnError(sql.ErrNoRows)

	// 3) GetIndexMetadata: indexQuery — 인덱스 한 개
	mock.ExpectQuery("all_indexes").
		WillReturnRows(sqlmock.NewRows([]string{"index_name", "uniqueness", "index_type"}).
			AddRow("IDX_USERS_EMAIL", "NONUNIQUE", "NORMAL"))

	// 4) GetIndexMetadata: colQuery — 해당 인덱스 컬럼
	mock.ExpectQuery("all_ind_columns").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "column_position", "descend"}).
			AddRow("EMAIL", 1, "ASC"))

	// 5) SELECT * FROM USERS — 행 없음
	mock.ExpectQuery("SELECT \\* FROM \"USERS\"").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))

	buf, w := newFileBuf()
	var mu sync.Mutex
	cfg := &config.Config{
		WithDDL:       true,
		WithSequences: false,
		WithIndexes:   true,
		User:          "owner",
		BatchSize:     100,
		PerTable:      false,
	}

	dia := &dialect.PostgresDialect{}
	if _, err := MigrateTableToFile(db, dia, "USERS", w, cfg, &mu, nil, NewMigrationState("test")); err != nil {
		t.Fatalf("MigrateTableToFile: %v", err)
	}
	w.Flush()
	out := buf.String()

	tableIdx := strings.Index(out, "CREATE TABLE")
	idxIdx := strings.Index(out, "CREATE INDEX")

	if tableIdx == -1 {
		t.Fatalf("CREATE TABLE이 출력에 없음:\n%s", out)
	}
	if idxIdx == -1 {
		t.Fatalf("Index DDL이 출력에 없음:\n%s", out)
	}
	if idxIdx < tableIdx {
		t.Errorf("Index DDL(pos=%d)이 CREATE TABLE(pos=%d) 이전에 출력됨", idxIdx, tableIdx)
	}
}

func TestWithDDL_WritesPrimaryKeyWithoutWithIndexes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT column_name, data_type").
		WithArgs("USERS").
		WillReturnRows(sqlmock.NewRows([]string{
			"column_name", "data_type", "data_precision", "data_scale", "nullable", "data_default",
		}).AddRow("ID", "NUMBER", int64(10), nil, "N", nil))

	mock.ExpectQuery("FROM all_constraints c").
		WillReturnRows(sqlmock.NewRows([]string{"constraint_name"}).AddRow("PK_USERS"))
	mock.ExpectQuery("FROM all_cons_columns").
		WillReturnRows(sqlmock.NewRows([]string{"column_name", "position"}).AddRow("ID", 1))

	mock.ExpectQuery("SELECT \\* FROM \"USERS\"").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}))

	buf, w := newFileBuf()
	var mu sync.Mutex
	cfg := &config.Config{
		WithDDL:     true,
		WithIndexes: false,
		User:        "owner",
		BatchSize:   100,
		PerTable:    false,
	}

	dia := &dialect.PostgresDialect{}
	if _, err := MigrateTableToFile(db, dia, "USERS", w, cfg, &mu, nil, NewMigrationState("test")); err != nil {
		t.Fatalf("MigrateTableToFile: %v", err)
	}
	w.Flush()
	out := buf.String()

	if !strings.Contains(out, "ADD PRIMARY KEY") {
		t.Fatalf("PK DDL이 출력에 없음:\n%s", out)
	}
	if strings.Contains(out, "CREATE INDEX") {
		t.Fatalf("WithIndexes=false 인데 일반 인덱스 DDL이 출력됨:\n%s", out)
	}
}

// ── 7-2: WithSequences=false, WithIndexes=false → 기존 동작 완전 유지 ─────────────

func TestWithoutSequencesAndIndexes_NoExtraDDL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	// WithDDL=false → sequence/index 쿼리 없이 바로 데이터 쿼리만
	mock.ExpectQuery("SELECT \\* FROM \"ITEMS\"").
		WillReturnRows(sqlmock.NewRows([]string{"ID"}).AddRow(1))

	buf, w := newFileBuf()
	var mu sync.Mutex
	cfg := &config.Config{
		WithDDL:       false,
		WithSequences: false,
		WithIndexes:   false,
		BatchSize:     100,
		PerTable:      false,
	}

	dia := &dialect.PostgresDialect{}
	if _, err := MigrateTableToFile(db, dia, "ITEMS", w, cfg, &mu, nil, NewMigrationState("test")); err != nil {
		t.Fatalf("MigrateTableToFile: %v", err)
	}
	w.Flush()
	out := buf.String()

	if strings.Contains(out, "CREATE SEQUENCE") {
		t.Errorf("WithSequences=false 인데 Sequence DDL이 출력됨:\n%s", out)
	}
	if strings.Contains(out, "CREATE INDEX") || strings.Contains(out, "ADD PRIMARY KEY") {
		t.Errorf("WithIndexes=false 인데 Index DDL이 출력됨:\n%s", out)
	}
	if strings.Contains(out, "CREATE TABLE") {
		t.Errorf("WithDDL=false 인데 CREATE TABLE이 출력됨:\n%s", out)
	}
	if !strings.Contains(out, "INSERT INTO") {
		t.Errorf("INSERT 구문이 출력에 없음:\n%s", out)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %v", err)
	}
}

// ── 7-2: OracleOwner 기본값 → 빈 문자열 시 User 값 대문자로 대체 ──────────────────

func TestResolveOwner_EmptyOracleOwner_UsesUserUppercased(t *testing.T) {
	cfg := &config.Config{User: "myuser", OracleOwner: ""}
	owner := resolveOwner(cfg)
	if owner != "MYUSER" {
		t.Errorf("resolveOwner = %q, want %q", owner, "MYUSER")
	}
}

func TestResolveOwner_WithOracleOwner_UsesOracleOwnerUppercased(t *testing.T) {
	cfg := &config.Config{User: "myuser", OracleOwner: "myowner"}
	owner := resolveOwner(cfg)
	if owner != "MYOWNER" {
		t.Errorf("resolveOwner = %q, want %q", owner, "MYOWNER")
	}
}
