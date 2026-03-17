package web

import (
	"testing"

	"dbmigrator/internal/config"
	"dbmigrator/internal/migration"
)

func TestMonitoringMigrationMetricsByObjectGroup(t *testing.T) {
	metrics := newMonitoringMetrics()

	metrics.recordMigrationStart(config.ObjectGroupAll, false)
	metrics.recordMigrationFinish(config.ObjectGroupAll, false, true)

	metrics.recordMigrationStart(config.ObjectGroupTables, false)
	metrics.recordMigrationFinish(config.ObjectGroupTables, false, false)

	metrics.recordMigrationStart(config.ObjectGroupSequences, true)
	metrics.recordMigrationFinish(config.ObjectGroupSequences, true, true)

	snapshot := metrics.snapshot()

	if snapshot.Migrations.All.Runs != 1 || snapshot.Migrations.All.Failures != 0 {
		t.Fatalf("unexpected all metrics: %+v", snapshot.Migrations.All)
	}
	if snapshot.Migrations.Tables.Runs != 1 || snapshot.Migrations.Tables.Failures != 1 {
		t.Fatalf("unexpected tables metrics: %+v", snapshot.Migrations.Tables)
	}
	if !nearlyEqualFloat(snapshot.Migrations.Tables.FailureRatePct, 100.0) {
		t.Fatalf("expected tables failure rate 100.0, got %.2f", snapshot.Migrations.Tables.FailureRatePct)
	}
	if snapshot.Migrations.Sequences.RetryAttempts != 1 || snapshot.Migrations.Sequences.RetrySuccesses != 1 {
		t.Fatalf("unexpected sequences retry metrics: %+v", snapshot.Migrations.Sequences)
	}
	if !nearlyEqualFloat(snapshot.Migrations.Sequences.RetrySuccessRatePct, 100.0) {
		t.Fatalf("expected sequences retry success rate 100.0, got %.2f", snapshot.Migrations.Sequences.RetrySuccessRatePct)
	}
}

func TestMonitoringTableHistoryMetrics(t *testing.T) {
	metrics := newMonitoringMetrics()

	// Simulate filter usage
	metrics.recordTableFilterUsage(&TableSummaryFilter{Status: "failed"})
	metrics.recordTableFilterUsage(&TableSummaryFilter{ExcludeSuccess: true, Search: "users"})
	metrics.recordTableFilterUsage(&TableSummaryFilter{})

	// Simulate retry
	metrics.recordTableRetry()
	metrics.recordTableRetry()

	// Simulate status recording
	metrics.recordTableStatus("success")
	metrics.recordTableStatus("success")
	metrics.recordTableStatus("failed")
	metrics.recordTableStatus("running")
	metrics.recordTableStatus("not_started")

	snap := metrics.snapshot()

	if snap.TableHistory.FilterUsage.Total != 3 {
		t.Fatalf("expected filter total 3, got %d", snap.TableHistory.FilterUsage.Total)
	}
	if snap.TableHistory.FilterUsage.Status != 1 {
		t.Fatalf("expected filter status 1, got %d", snap.TableHistory.FilterUsage.Status)
	}
	if snap.TableHistory.FilterUsage.ExcludeSuccess != 1 {
		t.Fatalf("expected filter excludeSuccess 1, got %d", snap.TableHistory.FilterUsage.ExcludeSuccess)
	}
	if snap.TableHistory.FilterUsage.Search != 1 {
		t.Fatalf("expected filter search 1, got %d", snap.TableHistory.FilterUsage.Search)
	}
	if snap.TableHistory.RetryTotal != 2 {
		t.Fatalf("expected retry total 2, got %d", snap.TableHistory.RetryTotal)
	}
	if snap.TableHistory.StatusTotal.Success != 2 {
		t.Fatalf("expected status success 2, got %d", snap.TableHistory.StatusTotal.Success)
	}
	if snap.TableHistory.StatusTotal.Failed != 1 {
		t.Fatalf("expected status failed 1, got %d", snap.TableHistory.StatusTotal.Failed)
	}
	if snap.TableHistory.StatusTotal.Running != 1 {
		t.Fatalf("expected status running 1, got %d", snap.TableHistory.StatusTotal.Running)
	}
	if snap.TableHistory.StatusTotal.NotStarted != 1 {
		t.Fatalf("expected status not_started 1, got %d", snap.TableHistory.StatusTotal.NotStarted)
	}
}

func TestMonitoringTableHistoryMetrics_NilSafe(t *testing.T) {
	var metrics *monitoringMetrics
	// Should not panic
	metrics.recordTableFilterUsage(&TableSummaryFilter{Status: "failed"})
	metrics.recordTableRetry()
	metrics.recordTableStatus("success")
}

func TestMonitoringPrecheckMetrics(t *testing.T) {
	metrics := newMonitoringMetrics()

	// precheck 실행 1회: 5개 테이블, transfer_required 2, skip_candidate 2, count_check_failed 1
	metrics.recordPrecheckRun(migration.PrecheckSummary{
		TotalTables:           5,
		TransferRequiredCount: 2,
		SkipCandidateCount:    2,
		CountCheckFailedCount: 1,
	})

	// precheck 실행 2회: 3개 테이블, transfer_required 3
	metrics.recordPrecheckRun(migration.PrecheckSummary{
		TotalTables:           3,
		TransferRequiredCount: 3,
	})

	snap := metrics.snapshot()

	if snap.Precheck.RunTotal != 2 {
		t.Errorf("expected run total 2, got %d", snap.Precheck.RunTotal)
	}
	if snap.Precheck.TablesTotal != 8 {
		t.Errorf("expected tables total 8, got %d", snap.Precheck.TablesTotal)
	}
	if snap.Precheck.TransferRequiredTotal != 5 {
		t.Errorf("expected transfer_required total 5, got %d", snap.Precheck.TransferRequiredTotal)
	}
	if snap.Precheck.SkipCandidateTotal != 2 {
		t.Errorf("expected skip_candidate total 2, got %d", snap.Precheck.SkipCandidateTotal)
	}
	if snap.Precheck.CountCheckFailedTotal != 1 {
		t.Errorf("expected count_check_failed total 1, got %d", snap.Precheck.CountCheckFailedTotal)
	}
}

func TestMonitoringPrecheckMetrics_NilSafe(t *testing.T) {
	var metrics *monitoringMetrics
	// Should not panic
	metrics.recordPrecheckRun(migration.PrecheckSummary{TotalTables: 5})
}

func nearlyEqualFloat(a, b float64) bool {
	const eps = 0.0001
	if a > b {
		return a-b < eps
	}
	return b-a < eps
}
