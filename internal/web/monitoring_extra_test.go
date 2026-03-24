package web

import (
	"dbmigrator/internal/migration"
	"testing"
)

func TestMonitoringMetrics_All(t *testing.T) {
	m := newMonitoringMetrics()
	m.recordLoginAttempt()
	m.recordLoginFailure()
	m.recordSessionCheck()
	m.recordSessionExpired()
	m.recordSessionCleanup()
	m.recordSessionEvicted()
	m.recordAPIResponse(monitoredAPICredentials, 200)
	m.recordAPIResponse(monitoredAPIHistory, 400)
	m.recordTableStatus("pass")
	m.recordMigrationPartialSuccess(5)
	m.recordMigrationRetry()
	m.recordMigrationStart("all", true)
	m.recordMigrationFinish("all", true, true)
	m.recordPrecheckRun(migration.PrecheckSummary{})
	m.recordTableFilterUsage(&TableSummaryFilter{})
	m.recordTableRetry()

	s := m.snapshot()
	_ = s
}

func TestMonitoringMetrics_NilSafe_All(t *testing.T) {
	var m *monitoringMetrics
	m.recordLoginAttempt()
	m.recordLoginFailure()
	m.recordSessionCheck()
	m.recordSessionExpired()
	m.recordSessionCleanup()
	m.recordSessionEvicted()
	m.recordAPIResponse(monitoredAPICredentials, 200)
	m.recordTableStatus("pass")
}
