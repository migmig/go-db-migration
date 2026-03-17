package web

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"dbmigrator/internal/db"
	"dbmigrator/internal/migration"

	"github.com/gin-gonic/gin"
)

// v19PrecheckEnabled는 DBM_V19_PRECHECK 환경변수로 pre-check 기능을 켜거나 끈다. 기본값은 활성화.
func v19PrecheckEnabled() bool {
	raw, ok := os.LookupEnv("DBM_V19_PRECHECK")
	if !ok || strings.TrimSpace(raw) == "" {
		return true
	}
	enabled, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return true
	}
	return enabled
}

// precheckRequest는 POST /api/migrations/precheck 요청 바디다.
type precheckRequest struct {
	OracleURL   string   `json:"oracleUrl"  binding:"required"`
	Username    string   `json:"username"   binding:"required"`
	Password    string   `json:"password"   binding:"required"`
	Tables      []string `json:"tables"     binding:"required"`
	TargetDB    string   `json:"targetDb"`
	TargetURL   string   `json:"targetUrl"`
	Policy      string   `json:"policy"`
	Concurrency int      `json:"concurrency"`
	TimeoutMs   int      `json:"timeoutMs"`
}

func validatePrecheckPolicy(p string) bool {
	switch migration.PrecheckPolicy(p) {
	case migration.PolicyStrict, migration.PolicyBestEffort, migration.PolicySkipEqualRows, "":
		return true
	}
	return false
}

func validatePrecheckDecision(d string) bool {
	switch d {
	case "all", "", string(migration.DecisionTransferRequired),
		string(migration.DecisionSkipCandidate), string(migration.DecisionCountCheckFailed):
		return true
	}
	return false
}

// precheckHandler는 POST /api/migrations/precheck 핸들러다.
func precheckHandler(metrics *monitoringMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req precheckRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
			return
		}

		if !validatePrecheckPolicy(req.Policy) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid policy: must be strict, best_effort, or skip_equal_rows"})
			return
		}

		if len(req.Tables) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "tables must not be empty"})
			return
		}

		policy := migration.PrecheckPolicy(req.Policy)
		if policy == "" {
			policy = migration.PolicyStrict
		}

		oracleDB, err := db.ConnectOracle(req.OracleURL, req.Username, req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Oracle: " + err.Error()})
			return
		}
		defer oracleDB.Close()

		sourceCountFn := db.SQLDBCountFn(oracleDB)
		var targetCountFn migration.RowCountFn

		if req.TargetURL != "" {
			targetDBName := req.TargetDB
			if targetDBName == "" {
				targetDBName = "postgres"
			}

			if targetDBName == "postgres" {
				pgPool, pgErr := db.ConnectPostgres(req.TargetURL, 4, 2, 30)
				if pgErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to target DB: " + pgErr.Error()})
					return
				}
				defer pgPool.Close()
				targetCountFn = db.PGPoolCountFn(pgPool)
			} else {
				targetConn, tErr := db.ConnectTargetDB(targetDBName, req.TargetURL)
				if tErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to target DB: " + tErr.Error()})
					return
				}
				defer targetConn.Close()
				targetCountFn = db.SQLDBCountFn(targetConn)
			}
		}

		cfg := migration.PrecheckEngineConfig{
			Concurrency: req.Concurrency,
			TimeoutMs:   req.TimeoutMs,
			Policy:      policy,
		}

		slog.Info("starting precheck",
			"tables", len(req.Tables),
			"policy", policy,
			"concurrency", cfg.Concurrency,
		)

		results, summary := migration.RunPrecheckRowCount(c.Request.Context(), req.Tables, sourceCountFn, targetCountFn, cfg)
		globalPrecheckStore.set(results, summary)

		if metrics != nil {
			metrics.recordPrecheckRun(summary)
		}

		c.JSON(http.StatusOK, gin.H{
			"summary": summary,
			"items":   results,
		})
	}
}

// precheckResultsHandler는 GET /api/migrations/precheck/results 핸들러다.
func precheckResultsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		decision := c.Query("decision")
		if !validatePrecheckDecision(decision) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid decision filter: must be all, transfer_required, skip_candidate, or count_check_failed"})
			return
		}

		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 || pageSize > 500 {
			pageSize = 50
		}

		allResults, summary := globalPrecheckStore.getAll()
		filtered := migration.FilterPrecheckResults(allResults, decision)

		// 테이블명 검색 필터
		if search := strings.TrimSpace(c.Query("search")); search != "" {
			search = strings.ToUpper(search)
			out := filtered[:0]
			for _, r := range filtered {
				if strings.Contains(strings.ToUpper(r.TableName), search) {
					out = append(out, r)
				}
			}
			filtered = out
		}

		total := len(filtered)

		// 페이지네이션
		start := (page - 1) * pageSize
		end := start + pageSize
		if start >= total {
			filtered = nil
		} else {
			if end > total {
				end = total
			}
			filtered = filtered[start:end]
		}

		c.JSON(http.StatusOK, gin.H{
			"summary":   summary,
			"items":     filtered,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	}
}
