package web

import (
	"net/http"
	"sync/atomic"
	"time"

	"dbmigrator/internal/config"

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

	sessionChecks  uint64
	sessionExpired uint64

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
}

type apiErrorMetricsSnapshot struct {
	Requests     uint64  `json:"requests"`
	Errors       uint64  `json:"errors"`
	ErrorRatePct float64 `json:"errorRatePct"`
}

type monitoringSnapshot struct {
	UptimeSeconds int64                    `json:"uptimeSeconds"`
	Login         loginMetricsSnapshot     `json:"login"`
	Session       sessionMetricsSnapshot   `json:"session"`
	Credentials   apiErrorMetricsSnapshot  `json:"credentialsApi"`
	History       apiErrorMetricsSnapshot  `json:"historyApi"`
	Migrations    migrationMetricsSnapshot `json:"migrations"`
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
	All       migrationGroupMetricsSnapshot `json:"all"`
	Tables    migrationGroupMetricsSnapshot `json:"tables"`
	Sequences migrationGroupMetricsSnapshot `json:"sequences"`
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
