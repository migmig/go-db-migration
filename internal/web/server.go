package web

import (
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"os"
	"path/filepath"
	"time"

	"dbmigrator/internal/bus"
	"dbmigrator/internal/config"
	"dbmigrator/internal/db"
	"dbmigrator/internal/dialect"
	"dbmigrator/internal/logger"
	"dbmigrator/internal/migration"
	"dbmigrator/internal/security"
	"dbmigrator/internal/web/ws"
	"dbmigrator/internal/web/ziputil"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templateFS embed.FS

var sessionManager = ws.NewSessionManager()

const authSessionCookieName = "dbm_auth_session"

type authSession struct {
	UserID     int64
	Username   string
	CreatedAt  time.Time
	LastSeenAt time.Time
}

type authSessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]authSession
	idleTTL     time.Duration
	absoluteTTL time.Duration
	metrics     *monitoringMetrics
}

func newAuthSessionManager(idleTTL, absoluteTTL time.Duration, metrics ...*monitoringMetrics) *authSessionManager {
	collector := newMonitoringMetrics()
	if len(metrics) > 0 && metrics[0] != nil {
		collector = metrics[0]
	}
	return &authSessionManager{
		sessions:    make(map[string]authSession),
		idleTTL:     idleTTL,
		absoluteTTL: absoluteTTL,
		metrics:     collector,
	}
}

func (m *authSessionManager) createSession(userID int64, username string) (string, authSession, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", authSession{}, fmt.Errorf("generate session token: %w", err)
	}

	token := hex.EncodeToString(tokenBytes)
	now := time.Now()
	s := authSession{UserID: userID, Username: username, CreatedAt: now, LastSeenAt: now}

	m.mu.Lock()
	m.sessions[token] = s
	m.mu.Unlock()

	return token, s, nil
}

func (m *authSessionManager) getSession(token string) (authSession, bool) {
	m.metrics.recordSessionCheck()

	m.mu.Lock()
	s, ok := m.sessions[token]
	if !ok {
		m.mu.Unlock()
		return authSession{}, false
	}

	now := time.Now()
	if now.Sub(s.CreatedAt) > m.absoluteTTL || now.Sub(s.LastSeenAt) > m.idleTTL {
		delete(m.sessions, token)
		m.mu.Unlock()
		m.metrics.recordSessionExpired()
		return authSession{}, false
	}

	s.LastSeenAt = now
	m.sessions[token] = s
	m.mu.Unlock()
	return s, true
}

func (m *authSessionManager) cookieMaxAge() int {
	return int(m.absoluteTTL.Seconds())
}

func (m *authSessionManager) deleteSession(token string) {
	m.mu.Lock()
	delete(m.sessions, token)
	m.mu.Unlock()
}

func RunServer(port string) {
	RunServerWithAuth(port, false)
}

func RunServerWithAuth(port string, authEnabled bool) {
	r := gin.Default()
	var userStore *db.UserStore
	var authSessions *authSessionManager
	var metrics *monitoringMetrics

	if authEnabled {
		masterKey := strings.TrimSpace(os.Getenv("DBM_MASTER_KEY"))
		if masterKey == "" {
			log.Fatal("DBM_MASTER_KEY is required when auth mode is enabled")
		}

		store, err := db.OpenAuthStore(os.Getenv("DBM_AUTH_DB_PATH"), masterKey)
		if err != nil {
			log.Fatalf("Failed to open auth user store: %v", err)
		}
		userStore = store
		defer userStore.Close()
		metrics = newMonitoringMetrics()
		authSessions = newAuthSessionManager(30*time.Minute, 24*time.Hour, metrics)
	}

	tmpl := template.Must(template.ParseFS(templateFS, "templates/*"))
	r.SetHTMLTemplate(tmpl)

	r.GET("/", func(c *gin.Context) {
		sessionID := sessionManager.CreateSession()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":       "Oracle DB Migrator",
			"sessionId":   sessionID,
			"AuthEnabled": authEnabled,
		})
	})

	r.StaticFS("/static", http.FS(templateFS))
	registerV16Routes(r)

	api := r.Group("/api")
	{
		api.GET("/meta", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"authEnabled": authEnabled,
				"uiVersion":   "v16-preview",
			})
		})

		if authEnabled {
			api.POST("/auth/login", loginHandler(userStore, authSessions))
			api.POST("/auth/logout", logoutHandler(authSessions))
			api.GET("/auth/me", meHandler(authSessions))
		}

		protected := api.Group("")
		if authEnabled {
			protected.Use(requireAuth(authSessions))

			credentialRoutes := protected.Group("/credentials")
			credentialRoutes.Use(monitoringAPIErrorsMiddleware(metrics, monitoredAPICredentials))
			credentialRoutes.GET("", listCredentialsHandler(userStore))
			credentialRoutes.POST("", createCredentialHandler(userStore))
			credentialRoutes.PUT("/:id", updateCredentialHandler(userStore))
			credentialRoutes.DELETE("/:id", deleteCredentialHandler(userStore))

			historyRoutes := protected.Group("/history")
			historyRoutes.Use(monitoringAPIErrorsMiddleware(metrics, monitoredAPIHistory))
			historyRoutes.GET("", listHistoryHandler(userStore))
			historyRoutes.GET("/:id", getHistoryHandler(userStore))
			historyRoutes.POST("/:id/replay", replayHistoryHandler(userStore))

			protected.GET("/monitoring/metrics", monitoringMetricsHandler(metrics))
		}
		protected.POST("/tables", getTables)
		protected.POST("/migrate", startMigrationHandler(userStore))
		protected.POST("/migrate/retry", retryMigrationHandler(userStore))
		protected.POST("/test-target", testTargetConnection)
		protected.GET("/ws", sessionManager.HandleConnection)
		protected.GET("/download/:id", downloadZip)
		protected.GET("/report/:id", downloadReport)
	}

	log.Printf("Starting web server on port %s...", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

func registerV16Routes(r *gin.Engine) {
	if r == nil {
		return
	}

	distDir := filepath.Join("frontend", "dist")
	indexPath := filepath.Join(distDir, "index.html")

	serveMissing := func(c *gin.Context) {
		c.String(http.StatusServiceUnavailable, "v16 frontend build not found. run: cd frontend && npm install && npm run build")
	}

	if _, err := os.Stat(indexPath); err != nil {
		r.GET("/v16", serveMissing)
		r.GET("/v16/*path", serveMissing)
		return
	}

	serveIndex := func(c *gin.Context) {
		c.File(indexPath)
	}

	r.GET("/v16", serveIndex)
	r.GET("/v16/*path", func(c *gin.Context) {
		reqPath := strings.TrimPrefix(c.Param("path"), "/")
		if reqPath == "" {
			serveIndex(c)
			return
		}

		cleaned := filepath.Clean(reqPath)
		if cleaned == "." || strings.HasPrefix(cleaned, "..") {
			serveIndex(c)
			return
		}

		candidate := filepath.Join(distDir, cleaned)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			c.File(candidate)
			return
		}

		serveIndex(c)
	})
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func loginHandler(userStore *db.UserStore, sessions *authSessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessions.metrics.recordLoginAttempt()

		var req loginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			sessions.metrics.recordLoginFailure()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
			return
		}

		user, err := userStore.GetUserByUsername(strings.TrimSpace(req.Username))
		if err != nil || !security.VerifyPassword(user.PasswordHash, req.Password) {
			sessions.metrics.recordLoginFailure()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}

		token, _, err := sessions.createSession(user.ID, user.Username)
		if err != nil {
			sessions.metrics.recordLoginFailure()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
			return
		}

		setAuthCookie(c, token, sessions.cookieMaxAge())
		c.JSON(http.StatusOK, gin.H{"username": user.Username, "userId": user.ID})
	}
}

func logoutHandler(sessions *authSessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(authSessionCookieName)
		if token != "" {
			sessions.deleteSession(token)
		}
		setAuthCookie(c, "", -1)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func meHandler(sessions *authSessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(authSessionCookieName)
		if err != nil || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		s, ok := sessions.getSession(token)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"userId": s.UserID, "username": s.Username})
	}
}

func requireAuth(sessions *authSessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(authSessionCookieName)
		if err != nil || token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		s, ok := sessions.getSession(token)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		c.Set("user_id", s.UserID)
		c.Set("username", s.Username)
		c.Next()
	}
}

type credentialRequest struct {
	Alias        string `json:"alias" binding:"required"`
	DBType       string `json:"dbType" binding:"required"`
	Host         string `json:"host" binding:"required"`
	Port         *int   `json:"port"`
	DatabaseName string `json:"databaseName"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

func setAuthCookie(c *gin.Context, value string, maxAge int) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authSessionCookieName, value, maxAge, "/", "", isSecureRequest(c.Request), true)
}

func isSecureRequest(r *http.Request) bool {
	if r == nil {
		return false
	}
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func currentUserID(c *gin.Context) int64 {
	value, ok := c.Get("user_id")
	if !ok {
		return 0
	}

	switch typed := value.(type) {
	case int64:
		return typed
	case int:
		return int64(typed)
	default:
		return 0
	}
}

func parseInt64Param(c *gin.Context, name string) (int64, bool) {
	value, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || value <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return 0, false
	}
	return value, true
}

func listCredentialsHandler(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		credentials, err := store.ListCredentialsByUser(currentUserID(c))
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, db.ErrCredentialCipherUnavailable) {
				status = http.StatusServiceUnavailable
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"items": credentials})
	}
}

func createCredentialHandler(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req credentialRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
			return
		}

		credential, err := store.CreateCredential(currentUserID(c), db.Credential{
			Alias:        strings.TrimSpace(req.Alias),
			DBType:       strings.TrimSpace(req.DBType),
			Host:         strings.TrimSpace(req.Host),
			Port:         req.Port,
			DatabaseName: strings.TrimSpace(req.DatabaseName),
			Username:     strings.TrimSpace(req.Username),
			Password:     req.Password,
		})
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, db.ErrCredentialCipherUnavailable) {
				status = http.StatusServiceUnavailable
			}
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, credential)
	}
}

func updateCredentialHandler(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		credentialID, ok := parseInt64Param(c, "id")
		if !ok {
			return
		}

		var req credentialRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
			return
		}

		credential, err := store.UpdateCredential(currentUserID(c), credentialID, db.Credential{
			Alias:        strings.TrimSpace(req.Alias),
			DBType:       strings.TrimSpace(req.DBType),
			Host:         strings.TrimSpace(req.Host),
			Port:         req.Port,
			DatabaseName: strings.TrimSpace(req.DatabaseName),
			Username:     strings.TrimSpace(req.Username),
			Password:     req.Password,
		})
		if err != nil {
			switch {
			case errors.Is(err, db.ErrCredentialNotFound):
				c.JSON(http.StatusNotFound, gin.H{"error": "Credential not found"})
			case errors.Is(err, db.ErrCredentialCipherUnavailable):
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			default:
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}

		c.JSON(http.StatusOK, credential)
	}
}

func deleteCredentialHandler(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		credentialID, ok := parseInt64Param(c, "id")
		if !ok {
			return
		}

		if err := store.DeleteCredential(currentUserID(c), credentialID); err != nil {
			if errors.Is(err, db.ErrCredentialNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "Credential not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

func listHistoryHandler(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

		items, total, err := store.ListHistoryByUser(currentUserID(c), page, pageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if page <= 0 {
			page = 1
		}
		if pageSize <= 0 {
			pageSize = 20
		}

		c.JSON(http.StatusOK, gin.H{
			"items":    items,
			"page":     page,
			"pageSize": pageSize,
			"total":    total,
		})
	}
}

func getHistoryHandler(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		historyID, ok := parseInt64Param(c, "id")
		if !ok {
			return
		}

		entry, err := store.GetHistoryByID(currentUserID(c), historyID)
		if err != nil {
			if errors.Is(err, db.ErrHistoryNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "History not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, entry)
	}
}

func replayHistoryHandler(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		historyID, ok := parseInt64Param(c, "id")
		if !ok {
			return
		}

		entry, err := store.GetHistoryByID(currentUserID(c), historyID)
		if err != nil {
			if errors.Is(err, db.ErrHistoryNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "History not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		payload := map[string]any{}
		if err := json.Unmarshal([]byte(entry.OptionsJSON), &payload); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode history payload"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"history": entry,
			"payload": payload,
		})
	}
}

type getTablesRequest struct {
	OracleURL string `json:"oracleUrl" binding:"required"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Like      string `json:"like"`
}

func getTables(c *gin.Context) {
	var req getTablesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	oracleDB, err := db.ConnectOracle(req.OracleURL, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Oracle DB: " + err.Error()})
		return
	}
	defer oracleDB.Close()

	tables, err := db.FetchTables(oracleDB, req.Like)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tables: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tables": tables})
}

type startMigrationRequest struct {
	SessionID string   `json:"sessionId"`
	OracleURL string   `json:"oracleUrl" binding:"required"`
	Username  string   `json:"username" binding:"required"`
	Password  string   `json:"password" binding:"required"`
	Tables    []string `json:"tables" binding:"required"`
	Direct    bool     `json:"direct"`
	PGURL     string   `json:"pgUrl"`
	WithDDL   bool     `json:"withDdl"`
	BatchSize int      `json:"batchSize"`
	Workers   int      `json:"workers"`
	// v4 추가 필드
	OutFile  string `json:"outFile"`
	PerTable bool   `json:"perTable"`
	Schema   string `json:"schema"`
	DryRun   bool   `json:"dryRun"`
	LogJSON  bool   `json:"logJson"`
	// v5 추가 필드
	WithSequences bool   `json:"withSequences"`
	WithIndexes   bool   `json:"withIndexes"`
	OracleOwner   string `json:"oracleOwner"`
	// v6 추가 필드
	TargetDB  string `json:"targetDb"`
	TargetURL string `json:"targetUrl"`
	// v8 추가 필드
	WithConstraints bool `json:"withConstraints"`
	DBMaxOpen       int  `json:"dbMaxOpen"`
	DBMaxIdle       int  `json:"dbMaxIdle"`
	DBMaxLife       int  `json:"dbMaxLife"`
	// v9 추가 필드
	Validate  bool `json:"validate"`
	CopyBatch int  `json:"copyBatch"`
}

var schemaPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

func validateMigrationRequest(req *startMigrationRequest) error {
	if strings.ContainsAny(req.OutFile, `/\`) {
		return fmt.Errorf("outFile must not contain path separators")
	}
	if req.Schema != "" && !schemaPattern.MatchString(req.Schema) {
		return fmt.Errorf("schema name contains invalid characters")
	}
	if req.BatchSize < 0 {
		return fmt.Errorf("batchSize must be non-negative")
	}
	if req.Workers < 0 {
		return fmt.Errorf("workers must be non-negative")
	}
	if req.DBMaxOpen < 0 {
		return fmt.Errorf("dbMaxOpen must be non-negative")
	}
	if req.DBMaxIdle < 0 {
		return fmt.Errorf("dbMaxIdle must be non-negative")
	}
	if req.DBMaxLife < 0 {
		return fmt.Errorf("dbMaxLife must be non-negative")
	}
	// v9: 테이블명 및 Oracle 소유자 식별자 검증 (SQL Injection 방어)
	for _, table := range req.Tables {
		if err := dialect.ValidateOracleIdentifier(table); err != nil {
			return fmt.Errorf("invalid table name %q: %w", table, err)
		}
	}
	if req.OracleOwner != "" {
		if err := dialect.ValidateOracleIdentifier(req.OracleOwner); err != nil {
			return fmt.Errorf("invalid oracle owner %q: %w", req.OracleOwner, err)
		}
	}
	return nil
}

func startMigration(c *gin.Context) {
	handleMigration(c, false, nil)
}

func startMigrationHandler(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		handleMigration(c, false, store)
	}
}

func retryMigration(c *gin.Context) {
	handleMigration(c, true, nil)
}

func retryMigrationHandler(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		handleMigration(c, true, store)
	}
}

func handleMigration(c *gin.Context, isRetry bool, store *db.UserStore) {
	var req startMigrationRequest
	// set defaults for db max idle to be safe
	req.DBMaxIdle = 2

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	if err := validateMigrationRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tracker := sessionManager.GetTracker(req.SessionID)
	if tracker == nil && req.SessionID != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired session ID"})
		return
	}
	if tracker == nil {
		tracker = ws.NewWebSocketTracker()
	}

	authUserID := currentUserID(c)

	go func() {
		if req.LogJSON {
			logger.SetJSONMode(true)
			defer logger.SetJSONMode(false)
		}

		// Start migration process in background
		oracleDB, err := db.ConnectOracle(req.OracleURL, req.Username, req.Password)
		if err != nil {
			log.Printf("Failed to connect to Oracle: %v", err)
			if !isRetry {
				tracker.AllDone("", nil)
			}
			return
		}
		defer oracleDB.Close()

		if req.DBMaxOpen > 0 {
			oracleDB.SetMaxOpenConns(req.DBMaxOpen)
		}
		if req.DBMaxIdle > 0 {
			oracleDB.SetMaxIdleConns(req.DBMaxIdle)
		}
		if req.DBMaxLife > 0 {
			oracleDB.SetConnMaxLifetime(time.Duration(req.DBMaxLife) * time.Second)
		}

		targetDBName := req.TargetDB
		if targetDBName == "" {
			targetDBName = "postgres"
		}

		dia, err := dialect.GetDialect(targetDBName)
		if err != nil {
			log.Printf("Failed to get dialect: %v", err)
			if !isRetry {
				tracker.AllDone("", nil)
			}
			return
		}

		var pgPool db.PGPool
		var targetDB *sql.DB

		targetURL := req.TargetURL
		if targetURL == "" {
			targetURL = req.PGURL // backward compat
		}

		if req.Direct && targetURL != "" {
			if targetDBName == "postgres" {
				// Wait, pgxpool config needs DBMaxOpen etc.
				// The db.ConnectPostgres doesn't take these parameters yet, so I will update ConnectPostgres in db.go as well.
				pgPool, err = db.ConnectPostgres(targetURL, req.DBMaxOpen, req.DBMaxIdle, req.DBMaxLife)
				if err != nil {
					log.Printf("Failed to connect to Postgres: %v", err)
					if !isRetry {
						tracker.AllDone("", nil)
					}
					return
				}
				defer pgPool.Close()
			} else {
				targetDB, err = db.ConnectTargetDB(dia.DriverName(), dia.NormalizeURL(targetURL))
				if err != nil {
					log.Printf("Failed to connect to Target DB: %v", err)
					if !isRetry {
						tracker.AllDone("", nil)
					}
					return
				}
				defer targetDB.Close()

				if req.DBMaxOpen > 0 {
					targetDB.SetMaxOpenConns(req.DBMaxOpen)
				}
				if req.DBMaxIdle > 0 {
					targetDB.SetMaxIdleConns(req.DBMaxIdle)
				}
				if req.DBMaxLife > 0 {
					targetDB.SetConnMaxLifetime(time.Duration(req.DBMaxLife) * time.Second)
				}
			}
		}

		workers := req.Workers
		if workers <= 0 {
			workers = 4
		}
		batchSize := req.BatchSize
		if batchSize <= 0 {
			batchSize = 1000
		}
		outFile := req.OutFile
		if outFile == "" {
			outFile = "migration.sql"
		}

		jobID := time.Now().Format("20060102150405")
		outDir := filepath.Join(os.TempDir(), "dbmigrator_"+jobID)
		if !req.Direct && !req.DryRun {
			if err := os.MkdirAll(outDir, 0755); err != nil {
				log.Printf("Failed to create temp directory: %v", err)
				return
			}
		}

		cfg := &config.Config{
			UserID:          authUserID,
			Tables:          req.Tables,
			Parallel:        true,
			Workers:         workers,
			BatchSize:       batchSize,
			PerTable:        req.PerTable,
			OutFile:         outFile,
			Schema:          req.Schema,
			DryRun:          req.DryRun,
			OutputDir:       outDir,
			TargetDB:        targetDBName,
			TargetURL:       targetURL,
			WithDDL:         req.WithDDL,
			WithSequences:   req.WithSequences,
			WithIndexes:     req.WithIndexes,
			WithConstraints: req.WithConstraints,
			OracleOwner:     req.OracleOwner,
			DBMaxOpen:       req.DBMaxOpen,
			DBMaxIdle:       req.DBMaxIdle,
			DBMaxLife:       req.DBMaxLife,
			Validate:        req.Validate,
			CopyBatch:       req.CopyBatch,
		}

		// Start background metrics collection
		doneMetrics := make(chan bool)
		defer close(doneMetrics)
		go func() {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					var m runtime.MemStats
					runtime.ReadMemStats(&m)
					memUsageMB := float64(m.Alloc) / 1024 / 1024
					// Dummy CPU usage & IOPS, in real scenario we can calculate diffs or use OS specific calls
					// Here we just mock CPU usage based on goroutines
					cpuUsagePct := float64(runtime.NumGoroutine()) * 2.5

					metricsData := map[string]interface{}{
						"mem_usage_mb":  fmt.Sprintf("%.2f", memUsageMB),
						"cpu_usage_pct": fmt.Sprintf("%.1f", cpuUsagePct),
						// IOPS and network can be sent from Tracker logic
					}
					metricsJSON, _ := json.Marshal(metricsData)

					tracker.EventBus().Publish(bus.Event{
						Type:    bus.EventMetrics,
						Message: string(metricsJSON),
					})
				case <-doneMetrics:
					return
				}
			}
		}()

		report, err := migration.Run(oracleDB, targetDB, pgPool, dia, cfg, tracker)
		saveHistoryForRequest(store, authUserID, req, targetDBName, targetURL, report, err)

		buildSummary := func() *ws.ReportSummary {
			if report == nil {
				return nil
			}
			totalRows, successCount, errorCount, duration, reportID := report.ToSummary()
			return &ws.ReportSummary{
				TotalRows:    totalRows,
				SuccessCount: successCount,
				ErrorCount:   errorCount,
				Duration:     duration,
				ReportID:     reportID,
			}
		}

		if err != nil {
			log.Printf("Migration failed: %v", err)
			if !isRetry {
				tracker.AllDone("", buildSummary())
			} else {
				log.Printf("Retry migration finished with error")
			}
		} else if req.DryRun {
			if !isRetry {
				tracker.AllDone("", nil)
			}
		} else if !req.Direct {
			// Create ZIP
			zipFilePath := filepath.Join(os.TempDir(), "migration_"+jobID+".zip")
			if err := ziputil.ZipDirectory(outDir, zipFilePath); err != nil {
				log.Printf("Failed to create zip: %v", err)
				if !isRetry {
					tracker.AllDone("", buildSummary())
				}
			} else {
				if !isRetry {
					tracker.AllDone("migration_"+jobID+".zip", buildSummary())
				}
			}
		} else {
			if !isRetry {
				tracker.AllDone("", buildSummary())
			} else {
				log.Printf("Retry migration finished successfully")
			}
		}

		// Clean up the temporary SQL files folder (keep zip)
		if !req.Direct && !req.DryRun {
			os.RemoveAll(outDir)
		}
	}()

	c.JSON(http.StatusOK, gin.H{"message": "Migration started"})
}

func saveHistoryForRequest(store *db.UserStore, userID int64, req startMigrationRequest, targetDBName, targetURL string, report *migration.MigrationReport, runErr error) {
	if store == nil || userID <= 0 {
		return
	}

	payload, err := json.Marshal(buildReplayPayload(req))
	if err != nil {
		log.Printf("Failed to encode migration history payload: %v", err)
		return
	}

	status := "success"
	logSummary := "completed"
	if runErr != nil {
		status = "failed"
		logSummary = runErr.Error()
	} else if report != nil {
		totalRows, successCount, errorCount, duration, reportID := report.ToSummary()
		logSummary = fmt.Sprintf("report=%s rows=%d success=%d error=%d duration=%s", reportID, totalRows, successCount, errorCount, duration)
	}

	if _, err := store.InsertHistory(userID, db.HistoryEntry{
		Status:        status,
		SourceSummary: buildSourceSummary(req),
		TargetSummary: buildTargetSummary(req, targetDBName, targetURL),
		OptionsJSON:   string(payload),
		LogSummary:    logSummary,
	}); err != nil {
		log.Printf("Failed to persist migration history: %v", err)
	}
}

func buildReplayPayload(req startMigrationRequest) map[string]any {
	targetURL := maskedURL(req.TargetURL)
	if targetURL == "" {
		targetURL = maskedURL(req.PGURL)
	}

	return map[string]any{
		"oracleUrl":       req.OracleURL,
		"username":        req.Username,
		"tables":          req.Tables,
		"direct":          req.Direct,
		"pgUrl":           maskedURL(req.PGURL),
		"targetDb":        req.TargetDB,
		"targetUrl":       targetURL,
		"withDdl":         req.WithDDL,
		"batchSize":       req.BatchSize,
		"workers":         req.Workers,
		"outFile":         req.OutFile,
		"perTable":        req.PerTable,
		"schema":          req.Schema,
		"dryRun":          req.DryRun,
		"logJson":         req.LogJSON,
		"withSequences":   req.WithSequences,
		"withIndexes":     req.WithIndexes,
		"oracleOwner":     req.OracleOwner,
		"withConstraints": req.WithConstraints,
		"dbMaxOpen":       req.DBMaxOpen,
		"dbMaxIdle":       req.DBMaxIdle,
		"dbMaxLife":       req.DBMaxLife,
		"validate":        req.Validate,
		"copyBatch":       req.CopyBatch,
	}
}

func buildSourceSummary(req startMigrationRequest) string {
	return fmt.Sprintf("%s@%s", strings.TrimSpace(req.Username), strings.TrimSpace(req.OracleURL))
}

func buildTargetSummary(req startMigrationRequest, targetDBName, targetURL string) string {
	if req.Direct {
		return fmt.Sprintf("%s:%s", targetDBName, maskedURL(targetURL))
	}
	outFile := req.OutFile
	if outFile == "" {
		outFile = "migration.sql"
	}
	return fmt.Sprintf("file:%s", outFile)
}

func maskedURL(raw string) string {
	if raw == "" {
		return ""
	}

	replacer := regexp.MustCompile(`(://[^:@]+):([^@]+)@`)
	return replacer.ReplaceAllString(raw, "$1:***@")
}

func downloadReport(c *gin.Context) {
	id := filepath.Base(c.Param("id"))
	if id == "" || id == "." || id == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing report ID"})
		return
	}

	reportPath := filepath.Join(".migration_state", id+"_report.json")
	if _, err := os.Stat(reportPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Report not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+id+"_report.json")
	c.Header("Content-Type", "application/json")
	c.File(reportPath)
}

func downloadZip(c *gin.Context) {
	id := filepath.Base(c.Param("id")) // prevent path traversal
	if id == "" || id == "." || id == "/" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing zip file ID"})
		return
	}

	zipPath := filepath.Join(os.TempDir(), id)
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+id)
	c.Header("Content-Type", "application/zip")
	c.File(zipPath)

	// Clean up zip after download
	// Wait a moment before deleting to ensure transfer completes
	go func() {
		time.Sleep(5 * time.Minute)
		os.Remove(zipPath)
	}()
}

type testTargetRequest struct {
	TargetDB  string `json:"targetDb" binding:"required"`
	TargetURL string `json:"targetUrl" binding:"required"`
}

func testTargetConnection(c *gin.Context) {
	var req testTargetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
		return
	}

	dia, err := dialect.GetDialect(req.TargetDB)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported target DB: " + req.TargetDB})
		return
	}

	if req.TargetDB == "postgres" {
		pgPool, err := db.ConnectPostgres(req.TargetURL, 1, 1, 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Target DB: " + err.Error()})
			return
		}
		defer pgPool.Close()
		// Ping to ensure connection is valid
		if err := pgPool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ping Target DB: " + err.Error()})
			return
		}
	} else {
		targetDB, err := db.ConnectTargetDB(dia.DriverName(), dia.NormalizeURL(req.TargetURL))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Target DB: " + err.Error()})
			return
		}
		defer targetDB.Close()
		if err := targetDB.Ping(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ping Target DB: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
}
