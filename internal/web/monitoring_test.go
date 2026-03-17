package web

import (
	"testing"

	"dbmigrator/internal/config"
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

func nearlyEqualFloat(a, b float64) bool {
	const eps = 0.0001
	if a > b {
		return a-b < eps
	}
	return b-a < eps
}
