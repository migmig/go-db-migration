package web

import (
	"context"
	"crypto/rand"
	"database/sql"
	"embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"path"
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
	"google.golang.org/api/idtoken"
)

//go:embed templates/* assets/frontend
var templateFS embed.FS

var sessionManager = ws.NewSessionManager()

const authSessionCookieName = "dbm_auth_session"

type authSession struct {
	UserID     int64
	Username   string
	CreatedAt  time.Time
	LastSeenAt time.Time
	ExpiresAt  time.Time
}

type authSessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]authSession
	idleTTL     time.Duration
	absoluteTTL time.Duration
	maxSessions int
	stopCleanup chan struct{}
	metrics     *monitoringMetrics
}

func newAuthSessionManager(idleTTL, absoluteTTL time.Duration, maxSessions int, cleanupInterval time.Duration, metrics ...*monitoringMetrics) *authSessionManager {
	collector := newMonitoringMetrics()
	if len(metrics) > 0 && metrics[0] != nil {
		collector = metrics[0]
	}
	manager := &authSessionManager{
		sessions:    make(map[string]authSession),
		idleTTL:     idleTTL,
		absoluteTTL: absoluteTTL,
		maxSessions: maxSessions,
		stopCleanup: make(chan struct{}),
		metrics:     collector,
	}
	go manager.startCleanupLoop(cleanupInterval)
	return manager
}

func (m *authSessionManager) createSession(userID int64, username string) (string, authSession, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", authSession{}, fmt.Errorf("generate session token: %w", err)
	}

	token := hex.EncodeToString(tokenBytes)
	now := time.Now()
	s := authSession{UserID: userID, Username: username, CreatedAt: now, LastSeenAt: now, ExpiresAt: now.Add(m.absoluteTTL)}

	m.mu.Lock()
	if m.maxSessions > 0 && len(m.sessions) >= m.maxSessions {
		m.evictOldest()
	}
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
	if now.After(s.ExpiresAt) || now.Sub(s.LastSeenAt) > m.idleTTL {
		delete(m.sessions, token)
		m.mu.Unlock()
		m.metrics.recordSessionExpired()
		return authSession{}, false
	}

	s.LastSeenAt = now
	s.ExpiresAt = s.CreatedAt.Add(m.absoluteTTL)
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

func (m *authSessionManager) close() {
	if m == nil || m.stopCleanup == nil {
		return
	}
	close(m.stopCleanup)
}

func (m *authSessionManager) startCleanupLoop(interval time.Duration) {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			m.purgeExpired()
		case <-m.stopCleanup:
			return
		}
	}
}

func (m *authSessionManager) purgeExpired() {
	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()
	cleaned := 0
	for token, s := range m.sessions {
		if now.After(s.ExpiresAt) || now.Sub(s.LastSeenAt) > m.idleTTL {
			delete(m.sessions, token)
			m.metrics.recordSessionExpired()
			m.metrics.recordSessionCleanup()
			cleaned++
		}
	}
	if cleaned > 0 {
		slog.Info("session cleanup", "cleaned_count", cleaned)
	}
}

func (m *authSessionManager) evictOldest() {
	var (
		oldestToken string
		oldestTime  time.Time
		found       bool
	)
	for token, s := range m.sessions {
		if !found || s.CreatedAt.Before(oldestTime) {
			found = true
			oldestToken = token
			oldestTime = s.CreatedAt
		}
	}
	if found {
		prefix := oldestToken
		if len(prefix) > 8 {
			prefix = prefix[:8]
		}
		delete(m.sessions, oldestToken)
		m.metrics.recordSessionEvicted()
		slog.Warn("session evicted due to capacity", "evicted_token_prefix", prefix)
	}
}

func RunServer(port string) {
	RunServerWithAuth(port, false)
}

func RunServerWithAuth(port string, authEnabled bool) {
	r := gin.Default()
	var userStore *db.UserStore
	var authSessions *authSessionManager
	var metrics *monitoringMetrics
	var googleClientID string

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
		maxSessions, cleanupInterval := loadAuthSessionEnv()
		authSessions = newAuthSessionManager(30*time.Minute, 24*time.Hour, maxSessions, cleanupInterval, metrics)
		defer authSessions.close()

		googleClientID, err = config.DecryptEnvValue(os.Getenv("GOOGLE_CLIENT_ID"), masterKey)
		if err != nil {
			log.Fatalf("Failed to decrypt GOOGLE_CLIENT_ID: %v", err)
		}
	}

	tmpl := template.Must(template.ParseFS(templateFS, "templates/*"))
	r.SetHTMLTemplate(tmpl)

	serveLegacyIndex := func(c *gin.Context) {
		sessionID := sessionManager.CreateSession()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title":       "Oracle DB Migrator",
			"sessionId":   sessionID,
			"AuthEnabled": authEnabled,
		})
	}

	frontendReady := registerFrontendRoutes(r)

	r.GET("/", func(c *gin.Context) {
		if frontendReady {
			c.Redirect(http.StatusTemporaryRedirect, "/app")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, "/legacy")
	})
	r.HEAD("/", func(c *gin.Context) {
		if frontendReady {
			c.Redirect(http.StatusTemporaryRedirect, "/app")
			return
		}
		c.Redirect(http.StatusTemporaryRedirect, "/legacy")
	})
	r.GET("/legacy", serveLegacyIndex)
	r.GET("/v15", serveLegacyIndex)

	r.StaticFS("/static", http.FS(templateFS))

	isTableHistoryEnabled := tableHistoryEnabled()

	uiVersion := "current"
	if !isTableHistoryEnabled {
		uiVersion = "preview"
	}

	api := r.Group("/api")
	{
		api.GET("/meta", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"authEnabled":    authEnabled,
				"uiVersion":      uiVersion,
				"googleClientId": googleClientID,
				"features": gin.H{
					"objectGroupMode":  objectGroupModeEnabled(),
					"tableHistory":     isTableHistoryEnabled,
					"precheckRowCount": precheckEnabled(),
				},
			})
		})

		if authEnabled {
			api.POST("/auth/login", loginHandler(userStore, authSessions))
			api.POST("/auth/google", googleLoginHandler(userStore, authSessions, googleClientID))
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
		protected.POST("/tables", getTables(userStore))
		protected.POST("/migrate", startMigrationHandler(userStore, metrics))
		protected.POST("/migrate/retry", retryMigrationHandler(userStore, metrics))
		if precheckEnabled() {
			protected.POST("/migrations/precheck", precheckHandler(metrics))
			protected.GET("/migrations/precheck/results", precheckResultsHandler())
		}
		protected.POST("/test-target", testTargetConnection(userStore))
		protected.POST("/target-tables", targetTablesHandler(userStore))
		protected.GET("/ws", sessionManager.HandleConnection)
		protected.GET("/download/:id", downloadZip)
		protected.GET("/report/:id", downloadReport)
		if isTableHistoryEnabled {
			protected.GET("/migrations/tables", listTableSummariesHandler(globalTableHistory, metrics))
			protected.GET("/migrations/tables/:tableName/history", getTableHistoryHandler(globalTableHistory))
		}
	}

	log.Printf("Starting web server on port %s...", port)
	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatalf("Failed to start web server: %v", err)
	}
}

func loadAuthSessionEnv() (int, time.Duration) {
	maxSessions := 100
	if raw := strings.TrimSpace(os.Getenv("DBM_MAX_SESSIONS")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			log.Fatalf("invalid DBM_MAX_SESSIONS=%q: must be an integer >= 0", raw)
		}
		maxSessions = parsed
	}

	cleanupInterval := 5 * time.Minute
	if raw := strings.TrimSpace(os.Getenv("DBM_SESSION_CLEANUP_INTERVAL")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			log.Fatalf("invalid DBM_SESSION_CLEANUP_INTERVAL=%q: must be a positive duration", raw)
		}
		cleanupInterval = parsed
	}

	return maxSessions, cleanupInterval
}

func tableHistoryEnabled() bool {
	raw, ok := os.LookupEnv("DBM_TABLE_HISTORY_ENABLED")
	if !ok || strings.TrimSpace(raw) == "" {
		return true // 기본값: 활성화
	}
	enabled, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return true
	}
	return enabled
}

func objectGroupModeEnabled() bool {
	raw, ok := os.LookupEnv("DBM_OBJECT_GROUP_UI_ENABLED")
	if !ok || strings.TrimSpace(raw) == "" {
		return true
	}

	enabled, err := strconv.ParseBool(strings.TrimSpace(raw))
	if err != nil {
		return true
	}
	return enabled
}

func registerFrontendRoutes(r *gin.Engine) bool {
	if r == nil {
		return false
	}

	distFS, ok := resolveFrontendAssetsFS()
	if !ok {
		redirectLegacy := func(c *gin.Context) {
			c.Redirect(http.StatusTemporaryRedirect, "/legacy")
		}
		r.GET("/app", redirectLegacy)
		r.GET("/v16", redirectLegacy)
		r.HEAD("/app", redirectLegacy)
		r.HEAD("/v16", redirectLegacy)
		r.GET("/app/*path", redirectLegacy)
		r.GET("/v16/*path", redirectLegacy)
		r.HEAD("/app/*path", redirectLegacy)
		r.HEAD("/v16/*path", redirectLegacy)
		return false
	}

	serveIndex := func(c *gin.Context) {
		indexHTML, readErr := fs.ReadFile(distFS, "index.html")
		if readErr != nil {
			c.String(http.StatusServiceUnavailable, "frontend assets are unavailable in this binary")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	}

	servePath := func(c *gin.Context) {
		reqPath := strings.TrimPrefix(c.Param("path"), "/")
		if reqPath == "" {
			serveIndex(c)
			return
		}

		cleaned := path.Clean(reqPath)
		if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") {
			serveIndex(c)
			return
		}

		if info, err := fs.Stat(distFS, cleaned); err == nil && !info.IsDir() {
			c.FileFromFS(cleaned, http.FS(distFS))
			return
		}

		serveIndex(c)
	}

	for _, route := range []string{"/app", "/v16"} {
		r.GET(route, serveIndex)
		r.HEAD(route, serveIndex)
		r.GET(route+"/*path", servePath)
		r.HEAD(route+"/*path", servePath)
	}
	return true
}

func resolveFrontendAssetsFS() (fs.FS, bool) {
	for _, candidate := range candidateFrontendDistDirs() {
		if info, err := os.Stat(filepath.Join(candidate, "index.html")); err == nil && !info.IsDir() {
			return os.DirFS(candidate), true
		}
	}

	distFS, err := fs.Sub(templateFS, "assets/frontend")
	if err != nil {
		return nil, false
	}
	if embeddedFrontendIsPlaceholder(distFS) {
		return nil, false
	}
	return distFS, true
}

func candidateFrontendDistDirs() []string {
	candidates := []string{}

	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, "frontend", "dist"))
	}
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates, filepath.Join(exeDir, "frontend", "dist"))
	}

	seen := make(map[string]struct{}, len(candidates))
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		out = append(out, candidate)
	}
	return out
}

func embeddedFrontendIsPlaceholder(distFS fs.FS) bool {
	indexHTML, err := fs.ReadFile(distFS, "index.html")
	if err != nil {
		return true
	}
	return strings.Contains(string(indexHTML), "frontend bundle is not embedded in this binary")
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

func googleLoginHandler(userStore *db.UserStore, sessions *authSessionManager, clientID string) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessions.metrics.recordLoginAttempt()

		var req struct {
			Credential string `json:"credential"` // Google ID Token
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			sessions.metrics.recordLoginFailure()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		payload, err := idtoken.Validate(context.Background(), req.Credential, clientID)
		if err != nil {
			// If clientID is not set, idtoken.Validate might fail.
			// In some development environments, we might want to skip validation if clientID is empty,
			// but for security we should expect it.
			sessions.metrics.recordLoginFailure()
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Google token: " + err.Error()})
			return
		}

		googleID := payload.Subject
		email := payload.Claims["email"].(string)

		user, err := userStore.GetUserByGoogleID(googleID)
		if err != nil {
			if errors.Is(err, db.ErrUserNotFound) {
				// Auto-register google user
				userID, err := userStore.CreateGoogleUser(email, googleID)
				if err != nil {
					sessions.metrics.recordLoginFailure()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
					return
				}
				user = &db.User{ID: userID, Username: email}
			} else {
				sessions.metrics.recordLoginFailure()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				return
			}
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
	OracleURL      string `json:"oracleUrl" binding:"required"`
	Username       string `json:"username" binding:"required"`
	Password       string `json:"password" binding:"required"`
	Like           string `json:"like"`
	SaveCredential bool   `json:"saveCredential"`
	Alias          string `json:"alias"`
}

func getTables(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		if req.SaveCredential && store != nil {
			uid := currentUserID(c)
			if uid > 0 {
				alias := req.Alias
				if alias == "" {
					alias = "Source: " + req.Username + "@" + req.OracleURL
				}
				_, _ = store.CreateCredential(uid, db.Credential{
					Alias:    alias,
					DBType:   "oracle",
					Host:     req.OracleURL,
					Username: req.Username,
					Password: req.Password,
				})
			}
		}

		tables, err := db.FetchTables(oracleDB, req.Like)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tables: " + err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"tables": tables})
	}
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
	// v17
	ObjectGroup string `json:"objectGroup"`
	// v18
	Truncate bool   `json:"truncate"`
	Upsert   bool   `json:"upsert"`
	OnError  string `json:"onError"`
	// v19
	UsePrecheckResults bool   `json:"usePrecheckResults"`
	PrecheckPolicy     string `json:"precheckPolicy"`
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
	if req.OnError != "" && req.OnError != "fail_fast" && req.OnError != "skip_batch" {
		return fmt.Errorf("onError must be one of fail_fast, skip_batch")
	}
	group := strings.ToLower(strings.TrimSpace(req.ObjectGroup))
	if group != "" && group != config.ObjectGroupAll && group != config.ObjectGroupTables && group != config.ObjectGroupSequences {
		return fmt.Errorf("objectGroup must be one of all, tables, sequences")
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
	handleMigration(c, false, nil, nil)
}

func startMigrationHandler(store *db.UserStore, metrics *monitoringMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		handleMigration(c, false, store, metrics)
	}
}

func retryMigration(c *gin.Context) {
	handleMigration(c, true, nil, nil)
}

func retryMigrationHandler(store *db.UserStore, metrics *monitoringMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		handleMigration(c, true, store, metrics)
	}
}

func handleMigration(c *gin.Context, isRetry bool, store *db.UserStore, metrics *monitoringMetrics) {
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

	if !requirePostgres(c, req.TargetDB) {
		return
	}
	req.ObjectGroup = strings.ToLower(strings.TrimSpace(req.ObjectGroup))
	if req.ObjectGroup == "" {
		req.ObjectGroup = config.ObjectGroupAll
	}
	req.OnError = strings.ToLower(strings.TrimSpace(req.OnError))
	if req.OnError == "" {
		req.OnError = "fail_fast"
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
		runGroup := req.ObjectGroup
		if runGroup == "" {
			runGroup = config.ObjectGroupAll
		}
		metrics.recordMigrationStart(runGroup, isRetry)
		success := false
		defer func() {
			metrics.recordMigrationFinish(runGroup, isRetry, success)
		}()

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
			pgPool, err = db.ConnectPostgres(targetURL, req.DBMaxOpen, req.DBMaxIdle, req.DBMaxLife)
			if err != nil {
				log.Printf("Failed to connect to Postgres: %v", err)
				if !isRetry {
					tracker.AllDone("", nil)
				}
				return
			}
			defer pgPool.Close()
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

		// v19: use_precheck 연계 - precheck 결과 기반 전송 대상 테이블 필터링
		tables := req.Tables
		if req.UsePrecheckResults && precheckEnabled() {
			precheckResults, _ := globalPrecheckStore.getAll()
			if len(precheckResults) > 0 {
				policy := migration.PrecheckPolicy(req.PrecheckPolicy)
				if policy == "" {
					policy = migration.PolicyStrict
				}
				plan, planErr := migration.ApplyPrecheckPolicy(precheckResults, policy)
				if planErr == nil && !plan.Blocked {
					tables = plan.TransferTables
					slog.Info("precheck filtering applied",
						"original_tables", len(req.Tables),
						"filtered_tables", len(tables),
						"skip_tables", len(plan.SkipTables),
						"policy", policy,
					)
				} else if planErr != nil {
					slog.Warn("precheck policy error, using original table list", "error", planErr)
				} else if plan.Blocked {
					slog.Warn("precheck plan blocked by strict policy, using original table list",
						"block_reason", plan.BlockReason,
					)
				}
			}
		}

		cfg := &config.Config{
			UserID:          authUserID,
			Tables:          tables,
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
			ObjectGroup:     strings.ToLower(strings.TrimSpace(req.ObjectGroup)),
			Truncate:        req.Truncate,
			Upsert:          req.Upsert,
			OnError:         req.OnError,
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
		recordTableHistory(globalTableHistory, jobID, report)
		if isRetry {
			metrics.recordTableRetry()
		}
		saveHistoryForRequest(store, authUserID, req, targetDBName, targetURL, report, err)
		if err == nil {
			success = true
		}

		buildSummary := func() *ws.ReportSummary {
			if report == nil {
				return nil
			}
			summary := report.ToSummary()
			return &ws.ReportSummary{
				TotalRows:    summary.TotalRows,
				SuccessCount: summary.SuccessCount,
				ErrorCount:   summary.ErrorCount,
				Duration:     summary.Duration,
				ReportID:     summary.ReportID,
				ObjectGroup:  summary.ObjectGroup,
				Stats: ws.GroupedStats{
					Tables: ws.GroupStats{
						TotalItems:   summary.Stats.Tables.TotalItems,
						SuccessCount: summary.Stats.Tables.SuccessCount,
						ErrorCount:   summary.Stats.Tables.ErrorCount,
						SkippedCount: summary.Stats.Tables.SkippedCount,
						TotalRows:    summary.Stats.Tables.TotalRows,
					},
					Sequences: ws.GroupStats{
						TotalItems:   summary.Stats.Sequences.TotalItems,
						SuccessCount: summary.Stats.Sequences.SuccessCount,
						ErrorCount:   summary.Stats.Sequences.ErrorCount,
						SkippedCount: summary.Stats.Sequences.SkippedCount,
						TotalRows:    summary.Stats.Sequences.TotalRows,
					},
				},
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

func recordTableHistory(histStore *TableHistoryStore, jobID string, report *migration.MigrationReport) {
	if histStore == nil || report == nil {
		return
	}
	for _, tr := range report.Tables {
		status := "success"
		var errMsg string
		if tr.Status == "partial_success" || tr.Status == migration.StatusPartialSuccess {
			status = "partial_success"
			if len(tr.Errors) > 0 {
				errMsg = tr.Errors[0]
			}
		} else if tr.Status != "ok" {
			status = "failed"
			if len(tr.Errors) > 0 {
				errMsg = tr.Errors[0]
			}
		}
		durationMs := tr.DurationNs / int64(time.Millisecond)
		finishedAt := report.StartedAt.Add(time.Duration(tr.DurationNs))
		h := TableMigrationHistory{
			RunID:                jobID + "_" + tr.Name,
			TableName:            tr.Name,
			Status:               status,
			StartedAt:            report.StartedAt,
			FinishedAt:           finishedAt,
			DurationMs:           durationMs,
			RowsProcessed:        int64(tr.RowCount),
			ErrorMessage:         errMsg,
			SkippedBatches:       tr.SkippedBatches,
			EstimatedSkippedRows: tr.EstimatedSkippedRows,
		}
		histStore.RecordTableRun(h)
	}
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
		summary := report.ToSummary()
		logSummary = fmt.Sprintf(
			"report=%s rows=%d success=%d error=%d duration=%s object_group=%s tables_ok=%d tables_error=%d sequences_ok=%d sequences_error=%d sequences_skipped=%d",
			summary.ReportID,
			summary.TotalRows,
			summary.SuccessCount,
			summary.ErrorCount,
			summary.Duration,
			summary.ObjectGroup,
			summary.Stats.Tables.SuccessCount,
			summary.Stats.Tables.ErrorCount,
			summary.Stats.Sequences.SuccessCount,
			summary.Stats.Sequences.ErrorCount,
			summary.Stats.Sequences.SkippedCount,
		)
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
	objectGroup := strings.ToLower(strings.TrimSpace(req.ObjectGroup))
	if objectGroup == "" {
		objectGroup = config.ObjectGroupAll
	}
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
		"objectGroup":     objectGroup,
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

func listTableSummariesHandler(store *TableHistoryStore, metrics *monitoringMetrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		f := TableSummaryFilter{
			Status: c.Query("status"),
			Search: c.Query("search"),
			Sort:   c.Query("sort"),
			Order:  c.Query("order"),
		}
		if v := c.Query("exclude_success"); v == "true" || v == "1" {
			f.ExcludeSuccess = true
		}
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
		f.Page = page
		f.PageSize = pageSize

		if err := ValidateTableSummaryFilter(&f); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		metrics.recordTableFilterUsage(&f)

		items, total := store.ListSummaries(f)
		for _, item := range items {
			metrics.recordTableStatus(item.Status)
		}
		c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
	}
}

func getTableHistoryHandler(store *TableHistoryStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		tableName := c.Param("tableName")
		if tableName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing table name"})
			return
		}

		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if limit <= 0 {
			limit = 20
		}

		history, ok := store.GetHistory(tableName, limit)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "Table not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"table_name": tableName, "items": history})
	}
}

// requirePostgres는 targetDb 값이 postgres가 아니면 400을 반환하고 false를 돌려준다.
func requirePostgres(c *gin.Context, targetDB string) bool {
	if targetDB != "" && targetDB != "postgres" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "v22 이후 타겟 DB는 PostgreSQL만 지원합니다 (입력값: " + targetDB + ")",
		})
		return false
	}
	return true
}

// isPermissionError는 PostgreSQL permission denied 오류 여부를 판별한다.
func isPermissionError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "permission denied") ||
		strings.Contains(msg, "42501") ||
		strings.Contains(msg, "insufficient privileges") ||
		strings.Contains(msg, "ora-01031") ||
		strings.Contains(msg, "access denied")
}

type targetTablesRequest struct {
	TargetURL      string `json:"targetUrl" binding:"required"`
	Schema         string `json:"schema"    binding:"required"`
	SaveCredential bool   `json:"saveCredential"`
	Alias          string `json:"alias"`
}

type targetTablesResponse struct {
	Tables    []string `json:"tables"`
	FetchedAt string   `json:"fetchedAt"`
}

func targetTablesHandler(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req targetTablesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
		defer cancel()

		pool, err := db.ConnectPostgres(req.TargetURL, 1, 1, 30)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "타겟 DB 연결 실패: " + err.Error()})
			return
		}
		defer pool.Close()

		if req.SaveCredential && store != nil {
			uid := currentUserID(c)
			if uid > 0 {
				alias := req.Alias
				if alias == "" {
					alias = "Target: " + req.TargetURL
				}
				_, _ = store.CreateCredential(uid, db.Credential{
					Alias:  alias,
					DBType: "postgres",
					Host:   req.TargetURL,
					// Password/Username are usually in TargetURL but we can store the full URL as Host
					// or try to parse it. For now, Host = URL is safe as it's encrypted.
				})
			}
		}

		tables, err := db.FetchTargetTables(ctx, pool, req.Schema)
		if err != nil {
			if isPermissionError(err) {
				c.JSON(http.StatusForbidden, gin.H{"error": "테이블 목록 조회 권한이 없습니다: " + err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "테이블 목록 조회 실패: " + err.Error()})
			return
		}

		if tables == nil {
			tables = []string{}
		}

		c.JSON(http.StatusOK, targetTablesResponse{
			Tables:    tables,
			FetchedAt: time.Now().UTC().Format(time.RFC3339),
		})
	}
}

type testTargetRequest struct {
	TargetDB       string `json:"targetDb"`
	TargetURL      string `json:"targetUrl" binding:"required"`
	SaveCredential bool   `json:"saveCredential"`
	Alias          string `json:"alias"`
}

func testTargetConnection(store *db.UserStore) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req testTargetRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request parameters"})
			return
		}

		if !requirePostgres(c, req.TargetDB) {
			return
		}

		pgPool, err := db.ConnectPostgres(req.TargetURL, 1, 1, 10)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Target DB: " + err.Error()})
			return
		}
		defer pgPool.Close()
		if err := pgPool.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to ping Target DB: " + err.Error()})
			return
		}

		if req.SaveCredential && store != nil {
			uid := currentUserID(c)
			if uid > 0 {
				alias := req.Alias
				if alias == "" {
					alias = "Target: " + req.TargetURL
				}
				_, _ = store.CreateCredential(uid, db.Credential{
					Alias:  alias,
					DBType: "postgres",
					Host:   req.TargetURL,
				})
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "Connection successful"})
	}
}
