package db

import (
	"errors"
	"path/filepath"
	"testing"
)

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
