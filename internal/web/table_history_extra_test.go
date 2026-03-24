package web

import (
	"testing"
)

func TestTableHistory_HasTable(t *testing.T) {
	th := &TableHistoryStore{
		runs: make(map[string][]TableMigrationHistory),
	}
	th.RecordTableRun(TableMigrationHistory{
		TableName: "T1",
		Status:    "pass",
	})
	
	if !th.HasTable("T1") {
		t.Error("expected T1 to exist")
	}
	if th.HasTable("T2") {
		t.Error("expected T2 not to exist")
	}
}
