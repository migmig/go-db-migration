package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dbmigrator/internal/security"

	_ "github.com/mattn/go-sqlite3"
)

const DefaultAuthDBPath = ".migration_state/auth.db"

var ErrUserNotFound = errors.New("user not found")
var ErrCredentialNotFound = errors.New("credential not found")
var ErrHistoryNotFound = errors.New("history not found")
var ErrCredentialCipherUnavailable = errors.New("credential cipher unavailable")

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	GoogleID     string
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Credential struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"userId"`
	Alias        string    `json:"alias"`
	DBType       string    `json:"dbType"`
	Host         string    `json:"host"`
	Port         *int      `json:"port,omitempty"`
	DatabaseName string    `json:"databaseName,omitempty"`
	Username     string    `json:"username"`
	Password     string    `json:"password,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type HistoryEntry struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"userId"`
	Status        string    `json:"status"`
	SourceSummary string    `json:"sourceSummary"`
	TargetSummary string    `json:"targetSummary"`
	OptionsJSON   string    `json:"optionsJson"`
	LogSummary    string    `json:"logSummary,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type rowScanner interface {
	Scan(dest ...any) error
}

type UserStore struct {
	db               *sql.DB
	credentialCipher *security.CredentialCipher
}

func OpenAuthStore(path, masterKey string) (*UserStore, error) {
	var credentialCipher *security.CredentialCipher
	var err error
	if strings.TrimSpace(masterKey) != "" {
		credentialCipher, err = security.NewCredentialCipher(masterKey)
		if err != nil {
			return nil, fmt.Errorf("init credential cipher: %w", err)
		}
	}
	return openUserStore(path, credentialCipher)
}

func OpenUserStore(path string) (*UserStore, error) {
	return openUserStore(path, nil)
}

func openUserStore(path string, credentialCipher *security.CredentialCipher) (*UserStore, error) {
	if path == "" {
		path = DefaultAuthDBPath
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("create auth db directory: %w", err)
	}

	dbConn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open auth db: %w", err)
	}

	if _, err := dbConn.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = dbConn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	store := &UserStore{db: dbConn, credentialCipher: credentialCipher}
	if err := store.ensureSchema(); err != nil {
		_ = dbConn.Close()
		return nil, err
	}

	return store, nil
}

func (s *UserStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *UserStore) ensureSchema() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		google_id TEXT NULL UNIQUE,
		is_admin INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS db_credentials (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		alias TEXT NOT NULL,
		db_type TEXT NOT NULL,
		host TEXT NOT NULL,
		port INTEGER NULL,
		database_name TEXT NULL,
		username TEXT NOT NULL,
		password_enc TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS migration_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		status TEXT NOT NULL,
		source_summary TEXT NOT NULL,
		target_summary TEXT NOT NULL,
		options_json TEXT NOT NULL,
		log_summary TEXT,
		created_at DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS ix_db_credentials_user_id ON db_credentials (user_id);
	CREATE UNIQUE INDEX IF NOT EXISTS ux_db_credentials_user_alias ON db_credentials (user_id, alias);
	CREATE INDEX IF NOT EXISTS ix_migration_history_user_created_at ON migration_history (user_id, created_at DESC);
	`
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("create auth schema: %w", err)
	}
	if err := s.ensureLegacyHistoryUserID(); err != nil {
		return err
	}
	return nil
}

func (s *UserStore) ensureLegacyHistoryUserID() error {
	exists, err := s.tableExists("migration_history")
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	hasUserID, err := s.columnExists("migration_history", "user_id")
	if err != nil {
		return err
	}
	if hasUserID {
		return nil
	}

	if _, err := s.db.Exec(`ALTER TABLE migration_history ADD COLUMN user_id INTEGER`); err != nil {
		return fmt.Errorf("add migration_history.user_id: %w", err)
	}

	legacyUserID, err := s.ensureLegacyUser()
	if err != nil {
		return err
	}

	if _, err := s.db.Exec(`UPDATE migration_history SET user_id = ? WHERE user_id IS NULL`, legacyUserID); err != nil {
		return fmt.Errorf("backfill migration_history.user_id: %w", err)
	}
	return nil
}

func (s *UserStore) ensureLegacyUser() (int64, error) {
	user, err := s.GetUserByUsername("legacy")
	if err == nil {
		return user.ID, nil
	}
	if !errors.Is(err, ErrUserNotFound) {
		return 0, err
	}

	now := time.Now().UTC()
	result, err := s.db.Exec(`
		INSERT INTO users (username, password_hash, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, "legacy", "!", 0, now, now)
	if err != nil {
		return 0, fmt.Errorf("create legacy user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get legacy user id: %w", err)
	}
	return id, nil
}

func (s *UserStore) tableExists(table string) (bool, error) {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&count); err != nil {
		return false, fmt.Errorf("check table %s: %w", table, err)
	}
	return count > 0, nil
}

func (s *UserStore) columnExists(table, column string) (bool, error) {
	rows, err := s.db.Query(fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return false, fmt.Errorf("check column %s.%s: %w", table, column, err)
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name string
		var columnType string
		var notNull int
		var defaultValue sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return false, fmt.Errorf("scan pragma table_info for %s: %w", table, err)
		}
		if name == column {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, fmt.Errorf("iterate pragma table_info for %s: %w", table, err)
	}
	return false, nil
}

func (s *UserStore) CreateUser(username, passwordHash string, isAdmin bool) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(`
		INSERT INTO users (username, password_hash, google_id, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, username, passwordHash, nil, boolToInt(isAdmin), now, now)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (s *UserStore) CreateGoogleUser(username, googleID string) (int64, error) {
	now := time.Now().UTC()
	result, err := s.db.Exec(`
		INSERT INTO users (username, password_hash, google_id, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, username, "", googleID, 0, now, now)
	if err != nil {
		return 0, fmt.Errorf("create google user: %w", err)
	}
	return result.LastInsertId()
}

func (s *UserStore) GetUserByUsername(username string) (*User, error) {
	row := s.db.QueryRow(`
		SELECT id, username, password_hash, google_id, is_admin, created_at, updated_at
		FROM users WHERE username = ?
	`, username)

	var user User
	var isAdmin int
	var googleID sql.NullString
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &googleID, &isAdmin, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	user.IsAdmin = isAdmin == 1
	user.GoogleID = googleID.String
	return &user, nil
}

func (s *UserStore) GetUserByGoogleID(googleID string) (*User, error) {
	row := s.db.QueryRow(`
		SELECT id, username, password_hash, google_id, is_admin, created_at, updated_at
		FROM users WHERE google_id = ?
	`, googleID)

	var user User
	var isAdmin int
	var gID sql.NullString
	if err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &gID, &isAdmin, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by google id: %w", err)
	}
	user.IsAdmin = isAdmin == 1
	user.GoogleID = gID.String
	return &user, nil
}

func (s *UserStore) ListUsers() ([]User, error) {
	rows, err := s.db.Query(`
		SELECT id, username, password_hash, google_id, is_admin, created_at, updated_at
		FROM users ORDER BY id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var user User
		var isAdmin int
		var googleID sql.NullString
		if err := rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &googleID, &isAdmin, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		user.IsAdmin = isAdmin == 1
		user.GoogleID = googleID.String
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate users: %w", err)
	}
	return users, nil
}

func (s *UserStore) DeleteUser(username string) error {
	result, err := s.db.Exec(`DELETE FROM users WHERE username = ?`, username)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *UserStore) ResetPassword(username, passwordHash string) error {
	result, err := s.db.Exec(`
		UPDATE users SET password_hash = ?, updated_at = ? WHERE username = ?
	`, passwordHash, time.Now().UTC(), username)
	if err != nil {
		return fmt.Errorf("reset password: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrUserNotFound
	}
	return nil
}

func (s *UserStore) CreateCredential(userID int64, credential Credential) (*Credential, error) {
	encryptedPassword, err := s.encryptCredentialPassword(credential.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	result, err := s.db.Exec(`
		INSERT INTO db_credentials (
			user_id, alias, db_type, host, port, database_name, username, password_enc, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		userID,
		credential.Alias,
		credential.DBType,
		credential.Host,
		nullableInt64(credential.Port),
		nullableString(credential.DatabaseName),
		credential.Username,
		encryptedPassword,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("create credential: %w", err)
	}

	credentialID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get created credential id: %w", err)
	}

	return s.getCredentialByID(userID, credentialID)
}

func (s *UserStore) ListCredentialsByUser(userID int64) ([]Credential, error) {
	rows, err := s.db.Query(`
		SELECT id, user_id, alias, db_type, host, port, database_name, username, password_enc, created_at, updated_at
		FROM db_credentials
		WHERE user_id = ?
		ORDER BY alias ASC, id ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list credentials: %w", err)
	}
	defer rows.Close()

	credentials := make([]Credential, 0)
	for rows.Next() {
		credential, err := s.scanCredential(rows)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, *credential)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate credentials: %w", err)
	}
	return credentials, nil
}

func (s *UserStore) UpdateCredential(userID, credentialID int64, credential Credential) (*Credential, error) {
	encryptedPassword, err := s.encryptCredentialPassword(credential.Password)
	if err != nil {
		return nil, err
	}

	result, err := s.db.Exec(`
		UPDATE db_credentials
		SET alias = ?, db_type = ?, host = ?, port = ?, database_name = ?, username = ?, password_enc = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`,
		credential.Alias,
		credential.DBType,
		credential.Host,
		nullableInt64(credential.Port),
		nullableString(credential.DatabaseName),
		credential.Username,
		encryptedPassword,
		time.Now().UTC(),
		credentialID,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("update credential: %w", err)
	}

	affected, _ := result.RowsAffected()
	if affected == 0 {
		return nil, ErrCredentialNotFound
	}

	return s.getCredentialByID(userID, credentialID)
}

func (s *UserStore) DeleteCredential(userID, credentialID int64) error {
	result, err := s.db.Exec(`DELETE FROM db_credentials WHERE id = ? AND user_id = ?`, credentialID, userID)
	if err != nil {
		return fmt.Errorf("delete credential: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrCredentialNotFound
	}
	return nil
}

func (s *UserStore) InsertHistory(userID int64, entry HistoryEntry) (int64, error) {
	optionsJSON := entry.OptionsJSON
	if optionsJSON == "" {
		optionsJSON = "{}"
	}

	result, err := s.db.Exec(`
		INSERT INTO migration_history (user_id, status, source_summary, target_summary, options_json, log_summary, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		userID,
		entry.Status,
		entry.SourceSummary,
		entry.TargetSummary,
		optionsJSON,
		entry.LogSummary,
		time.Now().UTC(),
	)
	if err != nil {
		return 0, fmt.Errorf("insert history: %w", err)
	}

	historyID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get created history id: %w", err)
	}
	return historyID, nil
}

func (s *UserStore) ListHistoryByUser(userID int64, page, pageSize int) ([]HistoryEntry, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM migration_history WHERE user_id = ?`, userID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count history: %w", err)
	}

	rows, err := s.db.Query(`
		SELECT id, user_id, status, source_summary, target_summary, options_json, log_summary, created_at
		FROM migration_history
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?
	`, userID, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list history: %w", err)
	}
	defer rows.Close()

	entries := make([]HistoryEntry, 0)
	for rows.Next() {
		entry, err := scanHistoryEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		entries = append(entries, *entry)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate history: %w", err)
	}
	return entries, total, nil
}

func (s *UserStore) GetHistoryByID(userID, historyID int64) (*HistoryEntry, error) {
	row := s.db.QueryRow(`
		SELECT id, user_id, status, source_summary, target_summary, options_json, log_summary, created_at
		FROM migration_history
		WHERE id = ? AND user_id = ?
	`, historyID, userID)

	entry, err := scanHistoryEntry(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrHistoryNotFound
		}
		return nil, err
	}
	return entry, nil
}

func (s *UserStore) getCredentialByID(userID, credentialID int64) (*Credential, error) {
	row := s.db.QueryRow(`
		SELECT id, user_id, alias, db_type, host, port, database_name, username, password_enc, created_at, updated_at
		FROM db_credentials
		WHERE id = ? AND user_id = ?
	`, credentialID, userID)

	credential, err := s.scanCredential(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCredentialNotFound
		}
		return nil, err
	}
	return credential, nil
}

func (s *UserStore) scanCredential(scanner rowScanner) (*Credential, error) {
	var credential Credential
	var port sql.NullInt64
	var databaseName sql.NullString
	var encryptedPassword string
	if err := scanner.Scan(
		&credential.ID,
		&credential.UserID,
		&credential.Alias,
		&credential.DBType,
		&credential.Host,
		&port,
		&databaseName,
		&credential.Username,
		&encryptedPassword,
		&credential.CreatedAt,
		&credential.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan credential: %w", err)
	}

	if port.Valid {
		portValue := int(port.Int64)
		credential.Port = &portValue
	}
	if databaseName.Valid {
		credential.DatabaseName = databaseName.String
	}

	password, err := s.decryptCredentialPassword(encryptedPassword)
	if err != nil {
		return nil, err
	}
	credential.Password = password
	return &credential, nil
}

func scanHistoryEntry(scanner rowScanner) (*HistoryEntry, error) {
	var entry HistoryEntry
	var logSummary sql.NullString
	if err := scanner.Scan(
		&entry.ID,
		&entry.UserID,
		&entry.Status,
		&entry.SourceSummary,
		&entry.TargetSummary,
		&entry.OptionsJSON,
		&logSummary,
		&entry.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("scan history: %w", err)
	}
	if logSummary.Valid {
		entry.LogSummary = logSummary.String
	}
	return &entry, nil
}

func (s *UserStore) encryptCredentialPassword(password string) (string, error) {
	if s.credentialCipher == nil {
		return "", ErrCredentialCipherUnavailable
	}
	encryptedPassword, err := s.credentialCipher.Encrypt(password)
	if err != nil {
		return "", fmt.Errorf("encrypt credential password: %w", err)
	}
	return encryptedPassword, nil
}

func (s *UserStore) decryptCredentialPassword(passwordEnc string) (string, error) {
	if s.credentialCipher == nil {
		return "", ErrCredentialCipherUnavailable
	}
	password, err := s.credentialCipher.Decrypt(passwordEnc)
	if err != nil {
		return "", fmt.Errorf("decrypt credential password: %w", err)
	}
	return password, nil
}

func nullableString(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableInt64(value *int) any {
	if value == nil {
		return nil
	}
	return int64(*value)
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
