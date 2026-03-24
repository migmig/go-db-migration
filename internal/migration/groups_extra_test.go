package migration

import (
	"dbmigrator/internal/dialect"
	"testing"
)

func TestSequenceNames(t *testing.T) {
	gm := &GroupedMetadata{
		Sequences: []dialect.SequenceMetadata{
			{Name: "S1"},
			{Name: "S2"},
		},
	}
	names := gm.SequenceNames()
	if len(names) != 2 {
		t.Errorf("expected 2 names, got %d", len(names))
	}
	if names[0] != "S1" || names[1] != "S2" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestPublishSequenceError_NilSafe(t *testing.T) {
	// Should not panic
	publishSequenceError(nil, nil, false, "S1", nil)
}
