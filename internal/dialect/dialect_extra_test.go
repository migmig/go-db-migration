package dialect

import (
	"testing"
)

func TestGetDialect(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"postgres", "postgres"},
		{"mysql", "mysql"},
		{"mariadb", "mariadb"},
		{"sqlite", "sqlite"},
		{"mssql", "mssql"},
		{"unknown", ""},
	}
	for _, tt := range tests {
		got, err := GetDialect(tt.name)
		if tt.want == "" {
			if err == nil {
				t.Errorf("GetDialect(%q) expected error", tt.name)
			}
		} else {
			if err != nil {
				t.Errorf("GetDialect(%q) unexpected error: %v", tt.name, err)
			}
			if got.Name() != tt.want {
				t.Errorf("GetDialect(%q) = %q, want %q", tt.name, got.Name(), tt.want)
			}
		}
	}
}

func TestMSSQLDialect_Uncovered(t *testing.T) {
	d := &MSSQLDialect{}
	if d.Name() != "mssql" {
		t.Error("expected mssql")
	}
	if d.QuoteIdentifier("a") != "[a]" {
		t.Error("expected [a]")
	}
	if d.DriverName() != "sqlserver" {
		t.Error("expected sqlserver")
	}
	if d.NormalizeURL("url") != "url" {
		t.Error("expected url")
	}
	d.CreateConstraintDDL(ConstraintMetadata{}, "HR")
	d.CreateSequenceDDL(SequenceMetadata{}, "HR")
}

func TestMySQLDialect_Uncovered(t *testing.T) {
	d := &MySQLDialect{}
	if d.Name() != "mysql" {
		t.Error("expected mysql")
	}
	if d.QuoteIdentifier("a") != "`a`" {
		t.Error("expected `a` ")
	}
	if d.DriverName() != "mysql" {
		t.Error("expected mysql")
	}
	if d.NormalizeURL("url") != "url" {
		t.Error("expected url")
	}
	d.CreateConstraintDDL(ConstraintMetadata{}, "HR")
	d.CreateSequenceDDL(SequenceMetadata{}, "HR")
}

func TestSQLiteDialect_Uncovered(t *testing.T) {
	d := &SQLiteDialect{}
	if d.Name() != "sqlite" {
		t.Error("expected sqlite")
	}
	if d.DriverName() != "sqlite3" {
		t.Error("expected sqlite3")
	}
	if d.NormalizeURL("url") != "url" {
		t.Error("expected url")
	}
	d.CreateSequenceDDL(SequenceMetadata{}, "HR")
}
