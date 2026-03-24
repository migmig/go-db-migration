package web

import (
	"net/http"
	"sync/atomic"
	"time"

	"dbmigrator/internal/config"
	"dbmigrator/internal/migration"

	"github.com/gin-gonic/gin"
)

type monitoredAPI string

const (
	monitoredAPICredentials monitoredAPI = "credentials"
	monitoredAPIHistory     monitoredAPI = "history"
)

type monitoringMetrics struct {
	startedAt time.Time

	loginAttempts uint64
	loginFailures uint64

	sessionChecks   uint64
	sessionExpired  uint64
	sessionCleanups uint64
	sessionEvicted  uint64

	credentialsRequests uint64
	credentialsErrors   uint64

	historyRequests uint64
	historyErrors   uint64

	migrationAllRuns                 uint64
	migrationTablesRuns              uint64
	migrationSequencesRuns           uint64
	migrationAllFailures             uint64
	migrationTablesFailures          uint64
	migrationSequencesFailures       uint64
	migrationAllRetryAttempts        uint64
	migrationTablesRetryAttempts     uint64
	migrationSequencesRetryAttempts  uint64
	migrationAllRetrySuccesses       uint64
	migrationTablesRetrySuccesses    uint64
	migrationSequencesRetrySuccesses uint64

	// v18: 테이블 필터/재시도/상태 메트릭
	tableFilterUsageAll            uint64
	tableFilterUsageStatus         uint64
	tableFilterUsageExcludeSuccess uint64
	tableFilterUsageSearch         uint64
	tableRetryTotal                uint64
	tableStatusSuccess             uint64
	tableStatusFailed              uint64
	tableStatusRunning             uint64
	tableStatusNotStarted          uint64
	tableStatusPartialSuccess      uint64

	// v20 metrics
	migrationPartialSuccessTotal uint64
	migrationSkippedBatchesTotal uint64
	migrationRetryTotal          uint64

	// v19: pre-check 메트릭
	precheckRunTotal              uint64
	precheckTablesTotal           uint64
	precheckTransferRequiredTotal uint64
	precheckSkipCandidateTotal    uint64
	precheckCountCheckFailedTotal uint64
}

type loginMetricsSnapshot struct {
	Attempts       uint64  `json:"attempts"`
	Failures       uint64  `json:"failures"`
	FailureRatePct float64 `json:"failureRatePct"`
}

type sessionMetricsSnapshot struct {
	Checks            uint64  `json:"checks"`
	Expired           uint64  `json:"expired"`
	ExpirationRatePct float64 `json:"expirationRatePct"`
	Cleanups          uint64  `json:"cleanups"`
	Evicted           uint64  `json:"evicted"`
}

type apiErrorMetricsSnapshot struct {
	Requests     uint64  `json:"requests"`
	Errors       uint64  `json:"errors"`
	ErrorRatePct float64 `json:"errorRatePct"`
}

type precheckMetricsSnapshot struct {
	RunTotal              uint64 `json:"runTotal"`
	TablesTotal           uint64 `json:"tablesTotal"`
	TransferRequiredTotal uint64 `json:"transferRequiredTotal"`
	SkipCandidateTotal    uint64 `json:"skipCandidateTotal"`
	CountCheckFailedTotal uint64 `json:"countCheckFailedTotal"`
}

type monitoringSnapshot struct {
	UptimeSeconds int64                       `json:"uptimeSeconds"`
	Login         loginMetricsSnapshot        `json:"login"`
	Session       sessionMetricsSnapshot      `json:"session"`
	Credentials   apiErrorMetricsSnapshot     `json:"credentialsApi"`
	History       apiErrorMetricsSnapshot     `json:"historyApi"`
	Migrations    migrationMetricsSnapshot    `json:"migrations"`
	TableHistory  tableHistoryMetricsSnapshot `json:"tableHistory"`
	Precheck      precheckMetricsSnapshot     `json:"precheck"`
}

type tableHistoryMetricsSnapshot struct {
	FilterUsage struct {
		Total          uint64 `json:"total"`
		Status         uint64 `json:"status"`
		ExcludeSuccess uint64 `json:"excludeSuccess"`
		Search         uint64 `json:"search"`
	} `json:"filterUsage"`
	RetryTotal  uint64 `json:"retryTotal"`
	StatusTotal struct {
		Success        uint64 `json:"success"`
		Failed         uint64 `json:"failed"`
		Running        uint64 `json:"running"`
		NotStarted     uint64 `json:"notStarted"`
		PartialSuccess uint64 `json:"partialSuccess"`
	} `json:"statusTotal"`
}

type migrationGroupMetricsSnapshot struct {
	Runs                uint64  `json:"runs"`
	Failures            uint64  `json:"failures"`
	FailureRatePct      float64 `json:"failureRatePct"`
	RetryAttempts       uint64  `json:"retryAttempts"`
	RetrySuccesses      uint64  `json:"retrySuccesses"`
	RetrySuccessRatePct float64 `json:"retrySuccessRatePct"`
}

type migrationMetricsSnapshot struct {
	All                 migrationGroupMetricsSnapshot `json:"all"`
	Tables              migrationGroupMetricsSnapshot `json:"tables"`
	Sequences           migrationGroupMetricsSnapshot `json:"sequences"`
	PartialSuccessTotal uint64                        `json:"partialSuccessTotal"`
	SkippedBatchesTotal uint64                        `json:"skippedBatchesTotal"`
	RetryTotal          uint64                        `json:"retryTotal"`
}

func newMonitoringMetrics() *monitoringMetrics {
	return &monitoringMetrics{startedAt: time.Now()}
}

func (m *monitoringMetrics) recordLoginAttempt() {
	if m == nil {
		return
	}
	atomic.AddUint64(&m.loginAttempts, 1)
}

func (m *monitoringMetrics) recordLoginFailure() {
	if m == nil {
		return
	}
	atomic.AddUint64(&m.loginFailures, 1)
}

func (m *monitoringMetrics) recordSessionCheck() {
	if m == nil {
		return
	}
	atomic.AddUint64(&m.sessionChecks, 1)
}

func (m *monitoringMetrics) recordSessionExpired() {
	if m == nil {
		return
	}
	atomic.AddUint64(&m.sessionExpired, 1)
}

func (m *monitoringMetrics) recordSessionCleanup() {
	if m == nil {
		return
	}
	atomic.AddUint64(&m.sessionCleanups, 1)
}

func (m *monitoringMetrics) recordSessionEvicted() {
	if m == nil {
		return
	}
	atomic.AddUint64(&m.sessionEvicted, 1)
}

func (m *monitoringMetrics) recordAPIResponse(api monitoredAPI, status int) {
	if m == nil {
		return
	}

	switch api {
	case monitoredAPICredentials:
		atomic.AddUint64(&m.credentialsRequests, 1)
		if status >= http.StatusBadRequest {
			atomic.AddUint64(&m.credentialsErrors, 1)
		}
	case monitoredAPIHistory:
		atomic.AddUint64(&m.historyRequests, 1)
		if status >= http.StatusBadRequest {
			atomic.AddUint64(&m.historyErrors, 1)
		}
	}
}

func (m *monitoringMetrics) recordTableFilterUsage(f *TableSummaryFilter) {
	if m == nil || f == nil {
		return
	}
	atomic.AddUint64(&m.tableFilterUsageAll, 1)
	if f.Status != "" {
		atomic.AddUint64(&m.tableFilterUsageStatus, 1)
	}
	if f.ExcludeSuccess {
		atomic.AddUint64(&m.tableFilterUsageExcludeSuccess, 1)
	}
	if f.Search != "" {
		atomic.AddUint64(&m.tableFilterUsageSearch, 1)
	}
}

func (m *monitoringMetrics) recordTableRetry() {
	if m == nil {
		return
	}
	atomic.AddUint64(&m.tableRetryTotal, 1)
}

func (m *monitoringMetrics) recordPrecheckRun(summary migration.PrecheckSummary) {
	if m == nil {
		return
	}
	atomic.AddUint64(&m.precheckRunTotal, 1)
	atomic.AddUint64(&m.precheckTablesTotal, uint64(summary.TotalTables))
	atomic.AddUint64(&m.precheckTransferRequiredTotal, uint64(summary.TransferRequiredCount))
	atomic.AddUint64(&m.precheckSkipCandidateTotal, uint64(summary.SkipCandidateCount))
	atomic.AddUint64(&m.precheckCountCheckFailedTotal, uint64(summary.CountCheckFailedCount))
}

func (m *monitoringMetrics) recordTableStatus(status string) {
	if m == nil {
		return
	}
	switch status {
	case "success":
		atomic.AddUint64(&m.tableStatusSuccess, 1)
	case "failed":
		atomic.AddUint64(&m.tableStatusFailed, 1)
	case "running":
		atomic.AddUint64(&m.tableStatusRunning, 1)
	case "not_started":
		atomic.AddUint64(&m.tableStatusNotStarted, 1)
	case "partial_success":
		atomic.AddUint64(&m.tableStatusPartialSuccess, 1)
	}
}

func (m *monitoringMetrics) recordMigrationPartialSuccess(skippedBatches int) {
	if m == nil {
		return
	}
	atomic.AddUint64(&m.migrationPartialSuccessTotal, 1)
	atomic.AddUint64(&m.migrationSkippedBatchesTotal, uint64(skippedBatches))
}

func (m *monitoringMetrics) recordMigrationRetry() {
	if m == nil {
		return
	}
	atomic.AddUint64(&m.migrationRetryTotal, 1)
}

func (m *monitoringMetrics) recordMigrationStart(group string, isRetry bool) {
	if m == nil {
		return
	}

	runCounter, retryCounter := m.migrationCounters(group)
	atomic.AddUint64(runCounter, 1)
	if isRetry {
		atomic.AddUint64(retryCounter, 1)
	}
}

func (m *monitoringMetrics) recordMigrationFinish(group string, isRetry, success bool) {
	if m == nil {
		return
	}

	_, _, failureCounter, retrySuccessCounter := m.migrationResultCounters(group)
	if !success {
		atomic.AddUint64(failureCounter, 1)
		return
	}
	if isRetry {
		atomic.AddUint64(retrySuccessCounter, 1)
	}
}

func (m *monitoringMetrics) migrationCounters(group string) (*uint64, *uint64) {
	switch group {
	case config.ObjectGroupTables:
		return &m.migrationTablesRuns, &m.migrationTablesRetryAttempts
	case config.ObjectGroupSequences:
		return &m.migrationSequencesRuns, &m.migrationSequencesRetryAttempts
	default:
		return &m.migrationAllRuns, &m.migrationAllRetryAttempts
	}
}

func (m *monitoringMetrics) migrationResultCounters(group string) (*uint64, *uint64, *uint64, *uint64) {
	switch group {
	case config.ObjectGroupTables:
		return &m.migrationTablesRuns, &m.migrationTablesRetryAttempts, &m.migrationTablesFailures, &m.migrationTablesRetrySuccesses
	case config.ObjectGroupSequences:
		return &m.migrationSequencesRuns, &m.migrationSequencesRetryAttempts, &m.migrationSequencesFailures, &m.migrationSequencesRetrySuccesses
	default:
		return &m.migrationAllRuns, &m.migrationAllRetryAttempts, &m.migrationAllFailures, &m.migrationAllRetrySuccesses
	}
}

func (m *monitoringMetrics) snapshot() monitoringSnapshot {
	if m == nil {
		return monitoringSnapshot{}
	}

	loginAttempts := atomic.LoadUint64(&m.loginAttempts)
	loginFailures := atomic.LoadUint64(&m.loginFailures)
	sessionChecks := atomic.LoadUint64(&m.sessionChecks)
	sessionExpired := atomic.LoadUint64(&m.sessionExpired)
	credentialRequests := atomic.LoadUint64(&m.credentialsRequests)
	credentialErrors := atomic.LoadUint64(&m.credentialsErrors)
	historyRequests := atomic.LoadUint64(&m.historyRequests)
	historyErrors := atomic.LoadUint64(&m.historyErrors)
	allRuns := atomic.LoadUint64(&m.migrationAllRuns)
	tablesRuns := atomic.LoadUint64(&m.migrationTablesRuns)
	sequencesRuns := atomic.LoadUint64(&m.migrationSequencesRuns)
	allFailures := atomic.LoadUint64(&m.migrationAllFailures)
	tablesFailures := atomic.LoadUint64(&m.migrationTablesFailures)
	sequencesFailures := atomic.LoadUint64(&m.migrationSequencesFailures)
	allRetryAttempts := atomic.LoadUint64(&m.migrationAllRetryAttempts)
	tablesRetryAttempts := atomic.LoadUint64(&m.migrationTablesRetryAttempts)
	sequencesRetryAttempts := atomic.LoadUint64(&m.migrationSequencesRetryAttempts)
	allRetrySuccesses := atomic.LoadUint64(&m.migrationAllRetrySuccesses)
	tablesRetrySuccesses := atomic.LoadUint64(&m.migrationTablesRetrySuccesses)
	sequencesRetrySuccesses := atomic.LoadUint64(&m.migrationSequencesRetrySuccesses)

	filterAll := atomic.LoadUint64(&m.tableFilterUsageAll)
	filterStatus := atomic.LoadUint64(&m.tableFilterUsageStatus)
	filterExclude := atomic.LoadUint64(&m.tableFilterUsageExcludeSuccess)
	filterSearch := atomic.LoadUint64(&m.tableFilterUsageSearch)
	retryTotal := atomic.LoadUint64(&m.tableRetryTotal)
	statusSuccess := atomic.LoadUint64(&m.tableStatusSuccess)
	statusFailed := atomic.LoadUint64(&m.tableStatusFailed)
	statusRunning := atomic.LoadUint64(&m.tableStatusRunning)
	statusNotStarted := atomic.LoadUint64(&m.tableStatusNotStarted)

	var tableHistorySnap tableHistoryMetricsSnapshot
	tableHistorySnap.FilterUsage.Total = filterAll
	tableHistorySnap.FilterUsage.Status = filterStatus
	tableHistorySnap.FilterUsage.ExcludeSuccess = filterExclude
	tableHistorySnap.FilterUsage.Search = filterSearch
	tableHistorySnap.RetryTotal = retryTotal
	tableHistorySnap.StatusTotal.Success = statusSuccess
	tableHistorySnap.StatusTotal.Failed = statusFailed
	tableHistorySnap.StatusTotal.Running = statusRunning
	tableHistorySnap.StatusTotal.NotStarted = statusNotStarted
	tableHistorySnap.StatusTotal.PartialSuccess = atomic.LoadUint64(&m.tableStatusPartialSuccess)

	precheckRunTotal := atomic.LoadUint64(&m.precheckRunTotal)
	precheckTablesTotal := atomic.LoadUint64(&m.precheckTablesTotal)
	precheckTransferRequired := atomic.LoadUint64(&m.precheckTransferRequiredTotal)
	precheckSkipCandidate := atomic.LoadUint64(&m.precheckSkipCandidateTotal)
	precheckCountCheckFailed := atomic.LoadUint64(&m.precheckCountCheckFailedTotal)

	return monitoringSnapshot{
		UptimeSeconds: int64(time.Since(m.startedAt).Seconds()),
		Login: loginMetricsSnapshot{
			Attempts:       loginAttempts,
			Failures:       loginFailures,
			FailureRatePct: percentage(loginFailures, loginAttempts),
		},
		Session: sessionMetricsSnapshot{
			Checks:            sessionChecks,
			Expired:           sessionExpired,
			ExpirationRatePct: percentage(sessionExpired, sessionChecks),
			Cleanups:          atomic.LoadUint64(&m.sessionCleanups),
			Evicted:           atomic.LoadUint64(&m.sessionEvicted),
		},
		Credentials: apiErrorMetricsSnapshot{
			Requests:     credentialRequests,
			Errors:       credentialErrors,
			ErrorRatePct: percentage(credentialErrors, credentialRequests),
		},
		History: apiErrorMetricsSnapshot{
			Requests:     historyRequests,
			Errors:       historyErrors,
			ErrorRatePct: percentage(historyErrors, historyRequests),
		},
		Migrations: migrationMetricsSnapshot{
			All: migrationGroupMetricsSnapshot{
				Runs:                allRuns,
				Failures:            allFailures,
				FailureRatePct:      percentage(allFailures, allRuns),
				RetryAttempts:       allRetryAttempts,
				RetrySuccesses:      allRetrySuccesses,
				RetrySuccessRatePct: percentage(allRetrySuccesses, allRetryAttempts),
			},
			Tables: migrationGroupMetricsSnapshot{
				Runs:                tablesRuns,
				Failures:            tablesFailures,
				FailureRatePct:      percentage(tablesFailures, tablesRuns),
				RetryAttempts:       tablesRetryAttempts,
				RetrySuccesses:      tablesRetrySuccesses,
				RetrySuccessRatePct: percentage(tablesRetrySuccesses, tablesRetryAttempts),
			},
			Sequences: migrationGroupMetricsSnapshot{
				Runs:                sequencesRuns,
				Failures:            sequencesFailures,
				FailureRatePct:      percentage(sequencesFailures, sequencesRuns),
				RetryAttempts:       sequencesRetryAttempts,
				RetrySuccesses:      sequencesRetrySuccesses,
				RetrySuccessRatePct: percentage(sequencesRetrySuccesses, sequencesRetryAttempts),
			},
			PartialSuccessTotal: atomic.LoadUint64(&m.migrationPartialSuccessTotal),
			SkippedBatchesTotal: atomic.LoadUint64(&m.migrationSkippedBatchesTotal),
			RetryTotal:          atomic.LoadUint64(&m.migrationRetryTotal),
		},
		TableHistory: tableHistorySnap,
		Precheck: precheckMetricsSnapshot{
			RunTotal:              precheckRunTotal,
			TablesTotal:           precheckTablesTotal,
			TransferRequiredTotal: precheckTransferRequired,
			SkipCandidateTotal:    precheckSkipCandidate,
			CountCheckFailedTotal: precheckCountCheckFailed,
		},
	}
}

func percentage(numerator, denominator uint64) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) * 100 / float64(denominator)
}

func monitoringAPIErrorsMiddleware(metrics *monitoringMetrics, api monitoredAPI) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		metrics.recordAPIResponse(api, c.Writer.Status())
	}
}

func monitoringMetricsHandler(metrics *monitoringMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, metrics.snapshot())
	}
}
