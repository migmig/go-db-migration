package web

import (
	"net/http"
	"sync/atomic"
	"time"

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
	UptimeSeconds int64                   `json:"uptimeSeconds"`
	Login         loginMetricsSnapshot    `json:"login"`
	Session       sessionMetricsSnapshot  `json:"session"`
	Credentials   apiErrorMetricsSnapshot `json:"credentialsApi"`
	History       apiErrorMetricsSnapshot `json:"historyApi"`
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
