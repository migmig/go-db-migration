package migration

import (
	"errors"
	"testing"
)

func TestDecidePrecheckResult(t *testing.T) {
	t.Run("equal counts become skip candidate", func(t *testing.T) {
		result := DecidePrecheckResult("EMP", 10, 10, true, nil)
		if result.Decision != DecisionSkipCandidate {
			t.Fatalf("expected %q, got %q", DecisionSkipCandidate, result.Decision)
		}
		if result.TransferPlanned {
			t.Fatalf("skip candidate should not be planned for transfer")
		}
	})

	t.Run("different counts require transfer", func(t *testing.T) {
		result := DecidePrecheckResult("EMP", 15, 10, true, nil)
		if result.Decision != DecisionTransferRequired {
			t.Fatalf("expected %q, got %q", DecisionTransferRequired, result.Decision)
		}
		if !result.TransferPlanned {
			t.Fatalf("transfer required should be planned for transfer")
		}
		if result.Diff != 5 {
			t.Fatalf("expected diff 5, got %d", result.Diff)
		}
	})

	t.Run("inaccessible target is transfer required with reason", func(t *testing.T) {
		result := DecidePrecheckResult("EMP", 15, 0, false, nil)
		if result.Decision != DecisionTransferRequired {
			t.Fatalf("expected %q, got %q", DecisionTransferRequired, result.Decision)
		}
		if result.Reason == "" {
			t.Fatalf("expected reason for inaccessible target")
		}
	})

	t.Run("count failure becomes count_check_failed", func(t *testing.T) {
		err := errors.New("timeout")
		result := DecidePrecheckResult("EMP", 0, 0, true, err)
		if result.Decision != DecisionCountCheckFailed {
			t.Fatalf("expected %q, got %q", DecisionCountCheckFailed, result.Decision)
		}
		if result.Reason != "timeout" {
			t.Fatalf("expected reason timeout, got %q", result.Reason)
		}
	})
}

func TestApplyPrecheckPolicy(t *testing.T) {
	results := []PrecheckTableResult{
		{TableName: "A", Decision: DecisionTransferRequired},
		{TableName: "B", Decision: DecisionSkipCandidate},
		{TableName: "C", Decision: DecisionCountCheckFailed},
	}

	t.Run("strict blocks when failed exists", func(t *testing.T) {
		plan, err := ApplyPrecheckPolicy(results, PolicyStrict)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !plan.Blocked {
			t.Fatalf("expected blocked plan")
		}
		if len(plan.TransferTables) != 1 || plan.TransferTables[0] != "A" {
			t.Fatalf("unexpected transfer tables: %#v", plan.TransferTables)
		}
	})

	t.Run("best effort includes failed in transfer", func(t *testing.T) {
		plan, err := ApplyPrecheckPolicy(results, PolicyBestEffort)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if plan.Blocked {
			t.Fatalf("best effort should not block")
		}
		if len(plan.TransferTables) != 2 {
			t.Fatalf("expected 2 transfer tables, got %#v", plan.TransferTables)
		}
	})

	t.Run("skip equal rows also includes failed in transfer", func(t *testing.T) {
		plan, err := ApplyPrecheckPolicy(results, PolicySkipEqualRows)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if plan.Blocked {
			t.Fatalf("skip_equal_rows should not block")
		}
		if len(plan.SkipTables) != 1 || plan.SkipTables[0] != "B" {
			t.Fatalf("unexpected skip tables: %#v", plan.SkipTables)
		}
	})

	t.Run("invalid policy returns error", func(t *testing.T) {
		_, err := ApplyPrecheckPolicy(results, PrecheckPolicy("unknown"))
		if err == nil {
			t.Fatalf("expected error for invalid policy")
		}
	})
}
