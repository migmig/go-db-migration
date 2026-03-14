package migration

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"dbmigrator/internal/bus"
	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/dialect"
)

// ValidationTracker는 검증 진행 상황을 외부로 전달하는 인터페이스이다.
// Deprecated: Use EventBus instead
type ValidationTracker interface {
	ValidationStart(table string)
	ValidationResult(table string, sourceCount, targetCount int, status string, detail string)
}

// ValidationResult는 단일 테이블 검증 결과를 담는다.
type ValidationResult struct {
	Table       string `json:"table"`
	SourceCount int    `json:"source_count"`
	TargetCount int    `json:"target_count"`
	Status      string `json:"status"` // "pass", "mismatch", "error"
	Detail      string `json:"detail,omitempty"`
}

// runValidation은 직접 마이그레이션 후 소스-타겟 행 수를 비교 검증한다.
func runValidation(
	dbConn *sql.DB,
	targetDB *sql.DB,
	pgPool db.PGPool,
	dia dialect.Dialect,
	cfg *config.Config,
	tracker ProgressTracker,
	report *MigrationReport,
) {
	valTracker, hasValTracker := tracker.(ValidationTracker)

	for _, tableName := range cfg.Tables {
		if tracker != nil && tracker.EventBus() != nil {
			tracker.EventBus().Publish(bus.Event{Type: bus.EventValidationStart, Table: tableName})
		} else if hasValTracker {
			valTracker.ValidationStart(tableName)
		}

		result := validateTable(dbConn, targetDB, pgPool, dia, tableName, cfg)

		if tracker != nil && tracker.EventBus() != nil {
			tracker.EventBus().Publish(bus.Event{
				Type:    bus.EventValidationResult,
				Table:   tableName,
				Total:   result.SourceCount,
				Count:   result.TargetCount,
				Status:  result.Status,
				Message: result.Detail,
			})
		} else if hasValTracker {
			valTracker.ValidationResult(
				tableName, result.SourceCount, result.TargetCount,
				result.Status, result.Detail,
			)
		}

		slog.Info("validation result",
			"table", tableName,
			"source_count", result.SourceCount,
			"target_count", result.TargetCount,
			"status", result.Status,
		)
	}
}

func validateTable(
	dbConn *sql.DB,
	targetDB *sql.DB,
	pgPool db.PGPool,
	dia dialect.Dialect,
	tableName string,
	cfg *config.Config,
) ValidationResult {
	result := ValidationResult{Table: tableName}
	quotedSrc := dialect.QuoteOracleIdentifier(tableName)

	// 1. 소스 행 수 조회
	if err := dbConn.QueryRow(
		fmt.Sprintf("SELECT COUNT(*) FROM %s", quotedSrc),
	).Scan(&result.SourceCount); err != nil {
		result.Status = "error"
		result.Detail = "source count query failed: " + err.Error()
		return result
	}

	// 2. 타겟 행 수 조회
	targetTable := dia.QuoteIdentifier(strings.ToLower(tableName))
	if cfg.Schema != "" {
		targetTable = dia.QuoteIdentifier(strings.ToLower(cfg.Schema)) + "." + targetTable
	}

	var targetErr error
	if pgPool != nil {
		targetErr = pgPool.QueryRow(
			context.Background(),
			fmt.Sprintf("SELECT COUNT(*) FROM %s", targetTable),
		).Scan(&result.TargetCount)
	} else if targetDB != nil {
		targetErr = targetDB.QueryRow(
			fmt.Sprintf("SELECT COUNT(*) FROM %s", targetTable),
		).Scan(&result.TargetCount)
	}
	if targetErr != nil {
		result.Status = "error"
		result.Detail = "target count query failed: " + targetErr.Error()
		return result
	}

	// 3. 비교
	if result.SourceCount != result.TargetCount {
		result.Status = "mismatch"
		diff := result.SourceCount - result.TargetCount
		result.Detail = fmt.Sprintf("%d rows difference", diff)
	} else {
		result.Status = "pass"
	}

	return result
}
