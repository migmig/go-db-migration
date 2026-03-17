package migration

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// RowCountFn은 특정 테이블의 행 수를 반환하는 함수 타입이다.
// DB 종류(Oracle, PostgreSQL, MySQL 등)에 따라 적합한 구현체를 주입한다.
type RowCountFn func(ctx context.Context, tableName string) (int, error)

// PrecheckEngineConfig는 pre-check 엔진의 실행 설정을 담는다.
type PrecheckEngineConfig struct {
	Concurrency int            // 병렬 처리 수 (기본 4)
	TimeoutMs   int            // 테이블별 타임아웃(밀리초) (기본 5000)
	Policy      PrecheckPolicy // 정책 (기본 strict)
}

func (cfg *PrecheckEngineConfig) applyDefaults() {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 4
	}
	if cfg.TimeoutMs <= 0 {
		cfg.TimeoutMs = 5000
	}
	if cfg.Policy == "" {
		cfg.Policy = PolicyStrict
	}
}

// RunPrecheckRowCount는 테이블 목록에 대해 source/target COUNT를 병렬로 수행하고 결과를 반환한다.
// sourceCountFn 또는 targetCountFn이 nil이면 해당 DB에 접근 불가로 처리한다.
func RunPrecheckRowCount(ctx context.Context, tables []string, sourceCountFn RowCountFn, targetCountFn RowCountFn, cfg PrecheckEngineConfig) ([]PrecheckTableResult, PrecheckSummary) {
	cfg.applyDefaults()
	if ctx == nil {
		ctx = context.Background()
	}

	results := make([]PrecheckTableResult, len(tables))
	sem := make(chan struct{}, cfg.Concurrency)
	var wg sync.WaitGroup

	for i, table := range tables {
		wg.Add(1)
		go func(idx int, tbl string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			tableCtx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TimeoutMs)*time.Millisecond)
			defer cancel()

			result := fetchAndDecide(tableCtx, tbl, sourceCountFn, targetCountFn, cfg.Policy)
			results[idx] = result

			slog.Info("precheck result",
				"table_name", result.TableName,
				"source_row_count", result.SourceRowCount,
				"target_row_count", result.TargetRowCount,
				"decision", result.Decision,
				"policy", result.Policy,
				"reason", result.Reason,
			)
		}(i, table)
	}

	wg.Wait()
	summary := buildPrecheckSummary(results)
	return results, summary
}

func fetchAndDecide(ctx context.Context, tableName string, sourceCountFn, targetCountFn RowCountFn, policy PrecheckPolicy) PrecheckTableResult {
	if sourceCountFn == nil {
		result := DecidePrecheckResult(tableName, 0, 0, false, fmt.Errorf("source db unavailable"))
		result.Policy = string(policy)
		result.CheckedAt = time.Now()
		return result
	}

	sourceCount, sourceErr := sourceCountFn(ctx, tableName)
	if sourceErr != nil {
		result := DecidePrecheckResult(tableName, 0, 0, false, fmt.Errorf("source count failed: %w", sourceErr))
		result.Policy = string(policy)
		result.CheckedAt = time.Now()
		return result
	}

	var targetCount int
	var targetAccessible bool
	if targetCountFn != nil {
		count, err := targetCountFn(ctx, tableName)
		if err == nil {
			targetCount = count
			targetAccessible = true
		}
	}

	result := DecidePrecheckResult(tableName, sourceCount, targetCount, targetAccessible, nil)
	result.Policy = string(policy)
	result.CheckedAt = time.Now()
	return result
}

func buildPrecheckSummary(results []PrecheckTableResult) PrecheckSummary {
	s := PrecheckSummary{TotalTables: len(results)}
	for _, r := range results {
		switch r.Decision {
		case DecisionTransferRequired:
			s.TransferRequiredCount++
		case DecisionSkipCandidate:
			s.SkipCandidateCount++
		case DecisionCountCheckFailed:
			s.CountCheckFailedCount++
		}
	}
	return s
}

// FilterPrecheckResults는 decision 필터 기준으로 결과를 필터링한다.
// filter가 "" 또는 "all"이면 전체를 반환한다.
func FilterPrecheckResults(results []PrecheckTableResult, decision string) []PrecheckTableResult {
	if decision == "" || decision == "all" {
		return results
	}
	out := make([]PrecheckTableResult, 0, len(results))
	for _, r := range results {
		if string(r.Decision) == decision {
			out = append(out, r)
		}
	}
	return out
}
