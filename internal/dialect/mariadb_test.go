package dialect

import "testing"

func TestMariaDB_Name(t *testing.T) {
	d := &MariaDBDialect{}
	if d.Name() != "mariadb" {
		t.Errorf("expected 'mariadb', got %q", d.Name())
	}
}

func TestMariaDB_InheritsMySQL(t *testing.T) {
	d := &MariaDBDialect{}
	// MariaDBDialect embeds MySQLDialect
	got := d.MapOracleType("VARCHAR2", 100, 0)
	if got != "VARCHAR(100)" {
		t.Errorf("expected VARCHAR(100), got %q", got)
	}

	got2 := d.MapOracleType("VARCHAR2", 20000, 0)
	if got2 != "LONGTEXT" {
		t.Errorf("expected LONGTEXT, got %q", got2)
	}
}

func TestCreateTableDDL_MariaDB(t *testing.T) {
	d := &MariaDBDialect{}
	cols := []ColumnDef{
		{Name: "ID", Type: "NUMBER", Nullable: "N"},
	}
	ddl := d.CreateTableDDL("test", "", cols)
	if ddl == "" {
		t.Error("expected valid DDL from embedded MySQLDialect")
	}
}

func TestInsertStatement_MariaDB(t *testing.T) {
	d := &MariaDBDialect{}
	rows := [][]any{{1}}
	stmts := d.InsertStatement("test", "", []string{"id"}, rows, 100)
	if len(stmts) == 0 {
		t.Error("expected valid InsertStatement from embedded MySQLDialect")
	}
}