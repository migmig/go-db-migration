package db

import (
	"errors"
	"path/filepath"
	"testing"
)

const testMasterKey = "0123456789abcdef0123456789abcdef"

func newTestStore(t *testing.T) *UserStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "auth.db")
	store, err := OpenUserStore(path)
	if err != nil {
		t.Fatalf("open user store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func newTestAuthStore(t *testing.T) *UserStore {
	t.Helper()
	path := filepath.Join(t.TempDir(), "auth.db")
	store, err := OpenAuthStore(path, testMasterKey)
	if err != nil {
		t.Fatalf("open auth store: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func createUserAndGetID(t *testing.T, store *UserStore, username string) int64 {
	t.Helper()
	if err := store.CreateUser(username, "hash-"+username, false); err != nil {
		t.Fatalf("create user %s: %v", username, err)
	}
	user, err := store.GetUserByUsername(username)
	if err != nil {
		t.Fatalf("get user %s: %v", username, err)
	}
	return user.ID
}

func TestUserStore_CreateGetList(t *testing.T) {
	store := newTestStore(t)

	if err := store.CreateUser("alice", "hash1", true); err != nil {
		t.Fatalf("create user alice: %v", err)
	}
	if err := store.CreateUser("bob", "hash2", false); err != nil {
		t.Fatalf("create user bob: %v", err)
	}

	alice, err := store.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("get user alice: %v", err)
	}
	if !alice.IsAdmin {
		t.Fatalf("expected alice admin=true")
	}

	users, err := store.ListUsers()
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
}

func TestUserStore_ResetPasswordAndDelete(t *testing.T) {
	store := newTestStore(t)

	if err := store.CreateUser("alice", "hash1", false); err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := store.ResetPassword("alice", "hash-new"); err != nil {
		t.Fatalf("reset password: %v", err)
	}

	alice, err := store.GetUserByUsername("alice")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if alice.PasswordHash != "hash-new" {
		t.Fatalf("unexpected password hash: %s", alice.PasswordHash)
	}

	if err := store.DeleteUser("alice"); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	_, err = store.GetUserByUsername("alice")
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestUserStore_NotFound(t *testing.T) {
	store := newTestStore(t)

	if err := store.DeleteUser("unknown"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound on delete, got %v", err)
	}
	if err := store.ResetPassword("unknown", "hash"); !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound on reset, got %v", err)
	}
}

func TestUserStore_CredentialsAreScopedPerUser(t *testing.T) {
	store := newTestAuthStore(t)
	aliceID := createUserAndGetID(t, store, "alice")
	bobID := createUserAndGetID(t, store, "bob")

	port := 5432
	created, err := store.CreateCredential(aliceID, Credential{
		Alias:        "pg-main",
		DBType:       "postgres",
		Host:         "localhost",
		Port:         &port,
		DatabaseName: "appdb",
		Username:     "pguser",
		Password:     "secret",
	})
	if err != nil {
		t.Fatalf("create credential: %v", err)
	}

	credentials, err := store.ListCredentialsByUser(aliceID)
	if err != nil {
		t.Fatalf("list alice credentials: %v", err)
	}
	if len(credentials) != 1 {
		t.Fatalf("expected 1 alice credential, got %d", len(credentials))
	}
	if credentials[0].Password != "secret" {
		t.Fatalf("expected decrypted password, got %q", credentials[0].Password)
	}

	_, err = store.UpdateCredential(bobID, created.ID, Credential{
		Alias:        "pg-bob",
		DBType:       "postgres",
		Host:         "bob-host",
		Username:     "bob",
		Password:     "bob-secret",
		DatabaseName: "bobdb",
	})
	if !errors.Is(err, ErrCredentialNotFound) {
		t.Fatalf("expected ErrCredentialNotFound for cross-user update, got %v", err)
	}

	if err := store.DeleteCredential(bobID, created.ID); !errors.Is(err, ErrCredentialNotFound) {
		t.Fatalf("expected ErrCredentialNotFound for cross-user delete, got %v", err)
	}

	bobCredentials, err := store.ListCredentialsByUser(bobID)
	if err != nil {
		t.Fatalf("list bob credentials: %v", err)
	}
	if len(bobCredentials) != 0 {
		t.Fatalf("expected 0 bob credentials, got %d", len(bobCredentials))
	}
}

func TestUserStore_HistoryIsScopedAndPaginated(t *testing.T) {
	store := newTestAuthStore(t)
	aliceID := createUserAndGetID(t, store, "alice")
	bobID := createUserAndGetID(t, store, "bob")

	firstID, err := store.InsertHistory(aliceID, HistoryEntry{
		Status:        "success",
		SourceSummary: "alice@oracle",
		TargetSummary: "postgres://target",
		OptionsJSON:   `{"tables":["USERS"]}`,
		LogSummary:    "rows=10",
	})
	if err != nil {
		t.Fatalf("insert history 1: %v", err)
	}
	_, err = store.InsertHistory(aliceID, HistoryEntry{
		Status:        "failed",
		SourceSummary: "alice@oracle",
		TargetSummary: "postgres://target",
		OptionsJSON:   `{"tables":["ORDERS"]}`,
		LogSummary:    "rows=0",
	})
	if err != nil {
		t.Fatalf("insert history 2: %v", err)
	}
	bobHistoryID, err := store.InsertHistory(bobID, HistoryEntry{
		Status:        "success",
		SourceSummary: "bob@oracle",
		TargetSummary: "mysql://target",
		OptionsJSON:   `{"tables":["PAYMENTS"]}`,
	})
	if err != nil {
		t.Fatalf("insert bob history: %v", err)
	}

	pageOne, total, err := store.ListHistoryByUser(aliceID, 1, 1)
	if err != nil {
		t.Fatalf("list alice history page 1: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if len(pageOne) != 1 {
		t.Fatalf("expected 1 history item on page 1, got %d", len(pageOne))
	}

	entry, err := store.GetHistoryByID(aliceID, firstID)
	if err != nil {
		t.Fatalf("get alice history: %v", err)
	}
	if entry.SourceSummary != "alice@oracle" {
		t.Fatalf("unexpected source summary: %s", entry.SourceSummary)
	}

	_, err = store.GetHistoryByID(aliceID, bobHistoryID)
	if !errors.Is(err, ErrHistoryNotFound) {
		t.Fatalf("expected ErrHistoryNotFound for cross-user access, got %v", err)
	}
}
