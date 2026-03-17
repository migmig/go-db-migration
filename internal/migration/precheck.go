package migration

import (
	"fmt"
	"time"
)

// PrecheckDecision은 사전 점검 결과를 바탕으로 테이블별 전송 필요 여부를 나타낸다.
type PrecheckDecision string

const (
	DecisionTransferRequired PrecheckDecision = "transfer_required"
	DecisionSkipCandidate    PrecheckDecision = "skip_candidate"
	DecisionCountCheckFailed PrecheckDecision = "count_check_failed"
)

// PrecheckPolicy는 pre-check 결과를 실제 실행 계획에 적용하는 방식을 정의한다.
type PrecheckPolicy string

const (
	PolicyStrict        PrecheckPolicy = "strict"
	PolicyBestEffort    PrecheckPolicy = "best_effort"
	PolicySkipEqualRows PrecheckPolicy = "skip_equal_rows"
)

// PrecheckTableResult는 테이블 단위 pre-check 판정 결과를 담는다.
type PrecheckTableResult struct {
	TableName       string           `json:"table_name"`
	SourceRowCount  int              `json:"source_row_count"`
	TargetRowCount  int              `json:"target_row_count"`
	Diff            int              `json:"diff"`
	Decision        PrecheckDecision `json:"decision"`
	Policy          string           `json:"policy,omitempty"`
	Reason          string           `json:"reason,omitempty"`
	TransferPlanned bool             `json:"transfer_planned"`
	CheckedAt       time.Time        `json:"checked_at"`
}

// PrecheckSummary는 pre-check 실행 요약 정보를 담는다.
type PrecheckSummary struct {
	TotalTables           int `json:"total_tables"`
	TransferRequiredCount int `json:"transfer_required_count"`
	SkipCandidateCount    int `json:"skip_candidate_count"`
	CountCheckFailedCount int `json:"count_check_failed_count"`
}

// PrecheckExecutionPlan은 policy 적용 이후 실제 전송/제외/차단 상태를 요약한다.
type PrecheckExecutionPlan struct {
	TransferTables []string `json:"transfer_tables"`
	SkipTables     []string `json:"skip_tables"`
	FailedTables   []string `json:"failed_tables"`
	Blocked        bool     `json:"blocked"`
	BlockReason    string   `json:"block_reason,omitempty"`
}

// DecidePrecheckResult는 source/target count 상태를 기준으로 decision을 산출한다.
func DecidePrecheckResult(tableName string, sourceCount, targetCount int, targetAccessible bool, countErr error) PrecheckTableResult {
	result := PrecheckTableResult{
		TableName:      tableName,
		SourceRowCount: sourceCount,
		TargetRowCount: targetCount,
		Diff:           sourceCount - targetCount,
	}

	if countErr != nil {
		result.Decision = DecisionCountCheckFailed
		result.Reason = countErr.Error()
		return result
	}

	if !targetAccessible {
		result.Decision = DecisionTransferRequired
		result.Reason = "target table missing or inaccessible"
		result.TransferPlanned = true
		return result
	}

	if sourceCount == targetCount {
		result.Decision = DecisionSkipCandidate
		return result
	}

	result.Decision = DecisionTransferRequired
	result.TransferPlanned = true
	return result
}

// ApplyPrecheckPolicy는 판정 결과를 policy에 맞춰 실행 계획으로 변환한다.
func ApplyPrecheckPolicy(results []PrecheckTableResult, policy PrecheckPolicy) (PrecheckExecutionPlan, error) {
	plan := PrecheckExecutionPlan{}

	switch policy {
	case PolicyStrict, PolicyBestEffort, PolicySkipEqualRows:
	default:
		return plan, fmt.Errorf("invalid precheck policy: %s", policy)
	}

	for _, result := range results {
		switch result.Decision {
		case DecisionSkipCandidate:
			plan.SkipTables = append(plan.SkipTables, result.TableName)
		case DecisionTransferRequired:
			plan.TransferTables = append(plan.TransferTables, result.TableName)
		case DecisionCountCheckFailed:
			plan.FailedTables = append(plan.FailedTables, result.TableName)
			if policy == PolicyStrict {
				plan.Blocked = true
				if plan.BlockReason == "" {
					plan.BlockReason = "count_check_failed exists under strict policy"
				}
			} else {
				plan.TransferTables = append(plan.TransferTables, result.TableName)
			}
		}
	}

	return plan, nil
}
