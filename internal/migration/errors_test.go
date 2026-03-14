package migration

import (
	"errors"
	"fmt"
	"testing"
)

func TestMigrationError_Error(t *testing.T) {
	root := fmt.Errorf("column too long")

	cases := []struct {
		err      *MigrationError
		contains string
	}{
		{
			&MigrationError{Table: "ORDERS", Phase: "data", Category: ErrTypeMismatch, RootCause: root},
			"TYPE_MISMATCH",
		},
		{
			&MigrationError{Table: "USERS", Phase: "ddl", Category: ErrNullViolation, RootCause: root},
			"table=USERS",
		},
		{
			&MigrationError{Table: "T", Phase: "data", Category: ErrUnknown, BatchNum: 5, RowOffset: 4000, RootCause: root},
			"batch=5",
		},
		{
			&MigrationError{Table: "T", Phase: "data", Category: ErrUnknown, RowOffset: 999, Column: "DESCRIPTION", RootCause: root},
			"column=DESCRIPTION",
		},
	}

	for _, tc := range cases {
		msg := tc.err.Error()
		if msg == "" {
			t.Errorf("Error() returned empty string")
		}
		if !contains(msg, tc.contains) {
			t.Errorf("Error() = %q, want to contain %q", msg, tc.contains)
		}
	}
}

func TestMigrationError_Unwrap(t *testing.T) {
	root := fmt.Errorf("root cause")
	me := &MigrationError{Table: "T", Phase: "data", Category: ErrUnknown, RootCause: root}

	if !errors.Is(me, root) {
		t.Error("errors.Is should find root cause via Unwrap")
	}
}

func TestMigrationError_DetailedError(t *testing.T) {
	me := &MigrationError{
		Table:       "ORDERS",
		Phase:       "data",
		Category:    ErrConnectionLost,
		Suggestion:  "Check network",
		Recoverable: true,
		RootCause:   fmt.Errorf("EOF"),
	}

	var de DetailedError = me
	if de.ErrorPhase() != "data" {
		t.Errorf("ErrorPhase() = %q, want %q", de.ErrorPhase(), "data")
	}
	if de.ErrorCategory() != "CONNECTION_LOST" {
		t.Errorf("ErrorCategory() = %q, want %q", de.ErrorCategory(), "CONNECTION_LOST")
	}
	if de.ErrorSuggestion() != "Check network" {
		t.Errorf("ErrorSuggestion() = %q, want %q", de.ErrorSuggestion(), "Check network")
	}
	if !de.IsRecoverable() {
		t.Error("IsRecoverable() should be true")
	}
}

func TestClassifyError(t *testing.T) {
	cases := []struct {
		errMsg   string
		expected ErrorCategory
	}{
		{"value too large for column", ErrTypeMismatch},
		{"ORA-12899: value too large", ErrTypeMismatch},
		{"cannot insert NULL into column", ErrNullViolation},
		{"null value in column violates not-null constraint", ErrNullViolation},
		{"foreign key constraint violation", ErrFKViolation},
		{"ORA-02291 integrity constraint", ErrFKViolation},
		{"connection reset by peer", ErrConnectionLost},
		{"broken pipe", ErrConnectionLost},
		{"context deadline exceeded", ErrTimeout},
		{"ORA-01013: user requested cancel", ErrTimeout},
		{"ORA-01031: insufficient privileges", ErrPermissionDenied},
		{"permission denied for table", ErrPermissionDenied},
		{"some unknown db error xyz", ErrUnknown},
	}

	for _, tc := range cases {
		err := fmt.Errorf("%s", tc.errMsg)
		got := classifyError(err)
		if got != tc.expected {
			t.Errorf("classifyError(%q) = %q, want %q", tc.errMsg, got, tc.expected)
		}
	}

	if classifyError(nil) != ErrUnknown {
		t.Error("classifyError(nil) should return ErrUnknown")
	}
}

func TestSuggestFix(t *testing.T) {
	// 각 카테고리에 대해 빈 문자열이 아닌 제안이 반환되어야 한다
	categories := []ErrorCategory{
		ErrTypeMismatch, ErrNullViolation, ErrFKViolation,
		ErrConnectionLost, ErrTimeout, ErrPermissionDenied, ErrUnknown,
	}
	dialects := []string{"postgres", "mysql", "mssql", "sqlite"}

	for _, cat := range categories {
		for _, d := range dialects {
			s := suggestFix(cat, d)
			if s == "" {
				t.Errorf("suggestFix(%s, %s) returned empty string", cat, d)
			}
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
