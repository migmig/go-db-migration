package migration

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockCountFn은 테스트용 COUNT 함수 모킹 헬퍼다.
func mockCountFn(counts map[string]int, errs map[string]error) RowCountFn {
	return func(ctx context.Context, tableName string) (int, error) {
		if errs != nil {
			if err, ok := errs[tableName]; ok {
				return 0, err
			}
		}
		if counts != nil {
			if cnt, ok := counts[tableName]; ok {
				return cnt, nil
			}
		}
		return 0, nil
	}
}

func TestRunPrecheckRowCount_BasicDecisions(t *testing.T) {
	sourceCounts := map[string]int{
		"EMP":  100,
		"DEPT": 50,
		"PROJ": 30,
	}
	targetCounts := map[string]int{
		"EMP":  100, // equal → skip_candidate
		"DEPT": 40, // different → transfer_required
		// PROJ: missing → transfer_required
	}
	targetErrs := map[string]error{
		"PROJ": errors.New("table not found"),
	}

	tables := []string{"EMP", "DEPT", "PROJ"}
	results, summary := RunPrecheckRowCount(
		context.Background(),
		tables,
		mockCountFn(sourceCounts, nil),
		mockCountFn(targetCounts, targetErrs),
		PrecheckEngineConfig{Policy: PolicyStrict},
	)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	byTable := make(map[string]PrecheckTableResult)
	for _, r := range results {
		byTable[r.TableName] = r
	}

	if byTable["EMP"].Decision != DecisionSkipCandidate {
		t.Errorf("EMP: expected skip_candidate, got %q", byTable["EMP"].Decision)
	}
	if byTable["DEPT"].Decision != DecisionTransferRequired {
		t.Errorf("DEPT: expected transfer_required, got %q", byTable["DEPT"].Decision)
	}
	if byTable["DEPT"].Diff != 10 {
		t.Errorf("DEPT: expected diff 10, got %d", byTable["DEPT"].Diff)
	}
	if byTable["PROJ"].Decision != DecisionTransferRequired {
		t.Errorf("PROJ: expected transfer_required (target missing), got %q", byTable["PROJ"].Decision)
	}

	if summary.TotalTables != 3 {
		t.Errorf("expected total 3, got %d", summary.TotalTables)
	}
	if summary.SkipCandidateCount != 1 {
		t.Errorf("expected skip_candidate 1, got %d", summary.SkipCandidateCount)
	}
	if summary.TransferRequiredCount != 2 {
		t.Errorf("expected transfer_required 2, got %d", summary.TransferRequiredCount)
	}
}

func TestRunPrecheckRowCount_SourceCountFailure(t *testing.T) {
	sourceErrs := map[string]error{
		"EMP": errors.New("timeout"),
	}

	results, summary := RunPrecheckRowCount(
		context.Background(),
		[]string{"EMP"},
		mockCountFn(nil, sourceErrs),
		mockCountFn(map[string]int{"EMP": 10}, nil),
		PrecheckEngineConfig{Policy: PolicyBestEffort},
	)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Decision != DecisionCountCheckFailed {
		t.Errorf("expected count_check_failed, got %q", results[0].Decision)
	}
	if summary.CountCheckFailedCount != 1 {
		t.Errorf("expected count_check_failed count 1, got %d", summary.CountCheckFailedCount)
	}
}

func TestRunPrecheckRowCount_NilTargetFn(t *testing.T) {
	// targetCountFn이 nil이면 target inaccessible → transfer_required
	results, _ := RunPrecheckRowCount(
		context.Background(),
		[]string{"EMP"},
		mockCountFn(map[string]int{"EMP": 50}, nil),
		nil, // no target
		PrecheckEngineConfig{Policy: PolicyStrict},
	)

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Decision != DecisionTransferRequired {
		t.Errorf("expected transfer_required for nil target, got %q", results[0].Decision)
	}
}

func TestRunPrecheckRowCount_PolicyAndCheckedAt(t *testing.T) {
	before := time.Now()
	results, _ := RunPrecheckRowCount(
		context.Background(),
		[]string{"T1"},
		mockCountFn(map[string]int{"T1": 5}, nil),
		mockCountFn(map[string]int{"T1": 5}, nil),
		PrecheckEngineConfig{Policy: PolicySkipEqualRows},
	)
	after := time.Now()

	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	if results[0].Policy != string(PolicySkipEqualRows) {
		t.Errorf("expected policy skip_equal_rows, got %q", results[0].Policy)
	}
	if results[0].CheckedAt.Before(before) || results[0].CheckedAt.After(after) {
		t.Errorf("checked_at out of expected range: %v", results[0].CheckedAt)
	}
}

func TestRunPrecheckRowCount_Concurrency(t *testing.T) {
	// 병렬 처리 시 모든 테이블이 정확히 처리되는지 확인
	tables := make([]string, 20)
	counts := make(map[string]int, 20)
	for i := range tables {
		name := "TABLE_" + string(rune('A'+i%26)) + string(rune('0'+i/26))
		tables[i] = name
		counts[name] = i * 10
	}

	results, summary := RunPrecheckRowCount(
		context.Background(),
		tables,
		mockCountFn(counts, nil),
		mockCountFn(counts, nil), // equal → all skip_candidate
		PrecheckEngineConfig{Concurrency: 5, Policy: PolicyStrict},
	)

	if len(results) != 20 {
		t.Fatalf("expected 20 results, got %d", len(results))
	}
	if summary.SkipCandidateCount != 20 {
		t.Errorf("expected all 20 as skip_candidate, got %d", summary.SkipCandidateCount)
	}
}

func TestRunPrecheckRowCount_ContextTimeout(t *testing.T) {
	slowFn := RowCountFn(func(ctx context.Context, tableName string) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		case <-time.After(5 * time.Second):
			return 100, nil
		}
	})

	results, _ := RunPrecheckRowCount(
		context.Background(),
		[]string{"SLOW"},
		slowFn,
		slowFn,
		PrecheckEngineConfig{TimeoutMs: 50, Policy: PolicyStrict},
	)

	if len(results) != 1 {
		t.Fatalf("expected 1 result")
	}
	// timeout → source count failed → count_check_failed
	if results[0].Decision != DecisionCountCheckFailed {
		t.Errorf("expected count_check_failed on timeout, got %q", results[0].Decision)
	}
}

func TestFilterPrecheckResults(t *testing.T) {
	items := []PrecheckTableResult{
		{TableName: "A", Decision: DecisionTransferRequired},
		{TableName: "B", Decision: DecisionSkipCandidate},
		{TableName: "C", Decision: DecisionCountCheckFailed},
		{TableName: "D", Decision: DecisionTransferRequired},
	}

	all := FilterPrecheckResults(items, "all")
	if len(all) != 4 {
		t.Errorf("expected 4 for all, got %d", len(all))
	}

	empty := FilterPrecheckResults(items, "")
	if len(empty) != 4 {
		t.Errorf("expected 4 for empty filter, got %d", len(empty))
	}

	tr := FilterPrecheckResults(items, "transfer_required")
	if len(tr) != 2 {
		t.Errorf("expected 2 transfer_required, got %d", len(tr))
	}

	sc := FilterPrecheckResults(items, "skip_candidate")
	if len(sc) != 1 || sc[0].TableName != "B" {
		t.Errorf("unexpected skip_candidate results: %v", sc)
	}

	ccf := FilterPrecheckResults(items, "count_check_failed")
	if len(ccf) != 1 || ccf[0].TableName != "C" {
		t.Errorf("unexpected count_check_failed results: %v", ccf)
	}
}

func TestPrecheckExecutionPlan_Integration(t *testing.T) {
	sourceCounts := map[string]int{
		"A": 100, // equal → skip
		"B": 200, // different → transfer
		"C": 50,  // count fail
	}
	targetCounts := map[string]int{
		"A": 100,
		"B": 150,
	}
	targetErrs := map[string]error{
		"C": errors.New("permission denied"),
	}

	tables := []string{"A", "B", "C"}

	t.Run("strict policy blocks on count_check_failed", func(t *testing.T) {
		results, _ := RunPrecheckRowCount(context.Background(), tables,
			mockCountFn(sourceCounts, nil),
			mockCountFn(targetCounts, targetErrs),
			PrecheckEngineConfig{Policy: PolicyStrict},
		)
		plan, err := ApplyPrecheckPolicy(results, PolicyStrict)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !plan.Blocked {
			t.Error("expected plan to be blocked under strict policy")
		}
	})

	t.Run("best_effort includes failed in transfer", func(t *testing.T) {
		results, _ := RunPrecheckRowCount(context.Background(), tables,
			mockCountFn(sourceCounts, nil),
			mockCountFn(targetCounts, targetErrs),
			PrecheckEngineConfig{Policy: PolicyBestEffort},
		)
		plan, err := ApplyPrecheckPolicy(results, PolicyBestEffort)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if plan.Blocked {
			t.Error("expected plan not blocked under best_effort")
		}
		if len(plan.TransferTables) != 2 { // B + C
			t.Errorf("expected 2 transfer tables, got %v", plan.TransferTables)
		}
		if len(plan.SkipTables) != 1 || plan.SkipTables[0] != "A" {
			t.Errorf("expected A as skip table, got %v", plan.SkipTables)
		}
	})

	t.Run("skip_equal_rows excludes skip candidates", func(t *testing.T) {
		results, _ := RunPrecheckRowCount(context.Background(), tables,
			mockCountFn(sourceCounts, nil),
			mockCountFn(targetCounts, targetErrs),
			PrecheckEngineConfig{Policy: PolicySkipEqualRows},
		)
		plan, err := ApplyPrecheckPolicy(results, PolicySkipEqualRows)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if plan.Blocked {
			t.Error("expected not blocked under skip_equal_rows")
		}
		// skip A, transfer B and C (failed goes to transfer in non-strict)
		if len(plan.SkipTables) != 1 {
			t.Errorf("expected 1 skip table, got %v", plan.SkipTables)
		}
	})
}
