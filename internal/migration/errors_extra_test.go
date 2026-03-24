package migration

import (
	"errors"
	"testing"
)

func TestMigrationError_Error_Extra(t *testing.T) {
	err := &MigrationError{
		Table:     "T",
		Phase:     "P",
		Category:  ErrConnectionLost,
		BatchNum:  1,
		RowOffset: 0,
		RootCause: errors.New("root"),
	}
	s := err.Error()
	if s == "" {
		t.Error("expected non-empty error string")
	}
}

func TestClassifyError_Nil(t *testing.T) {
	if classifyError(nil) != ErrUnknown {
		t.Error("expected ErrUnknown")
	}
}

func TestSuggestFix_Unknown(t *testing.T) {
	expected := "Check logs for details. Use --log-json for structured error output."
	if suggestFix(ErrUnknown, "postgres") != expected {
		t.Errorf("expected %q", expected)
	}
}
