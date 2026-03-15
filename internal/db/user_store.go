package db

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const DefaultAuthDBPath = ".migration_state/auth.db"

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	IsAdmin      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserStore struct {
	db *sql.DB
}

func OpenUserStore(path string) (*UserStore, error) {
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

	store := &UserStore{db: dbConn}
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
		is_admin INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`
	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("create users table: %w", err)
	}
	return nil
}

func (s *UserStore) CreateUser(username, passwordHash string, isAdmin bool) error {
	now := time.Now().UTC()
	_, err := s.db.Exec(`
		INSERT INTO users (username, password_hash, is_admin, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, username, passwordHash, boolToInt(isAdmin), now, now)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (s *UserStore) GetUserByUsername(username string) (*User, error) {
	row := s.db.QueryRow(`
		SELECT id, username, password_hash, is_admin, created_at, updated_at
		FROM users WHERE username = ?
	`, username)

	var u User
	var isAdmin int
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &isAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	u.IsAdmin = isAdmin == 1
	return &u, nil
}

func (s *UserStore) ListUsers() ([]User, error) {
	rows, err := s.db.Query(`
		SELECT id, username, password_hash, is_admin, created_at, updated_at
		FROM users ORDER BY id ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]User, 0)
	for rows.Next() {
		var u User
		var isAdmin int
		if err := rows.Scan(&u.ID, &u.Username, &u.PasswordHash, &isAdmin, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		u.IsAdmin = isAdmin == 1
		users = append(users, u)
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

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
