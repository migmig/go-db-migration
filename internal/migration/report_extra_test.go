package migration

import (
	"testing"
)

func TestMigrationReport_SkipGroup(t *testing.T) {
	r := NewMigrationReport("job", "s", "t", "tu", "all")
	r.SkipGroup("tables", 5)
	// Just check if it executes without panic
}

func TestMigrationReport_StartTable_Duplicate(t *testing.T) {
	r := NewMigrationReport("job", "s", "t", "tu", "all")
	r.StartTable("T1", false)
	r.StartTable("T1", true) // Re-starting as retry should be fine
}
