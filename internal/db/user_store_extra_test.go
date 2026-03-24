package db

import (
	"testing"
)

func TestUserStore_GoogleUser(t *testing.T) {
	store := newTestStore(t)

	_, err := store.CreateGoogleUser("alice@gmail.com", "g1")
	if err != nil {
		t.Fatalf("create google user: %v", err)
	}

	user, err := store.GetUserByGoogleID("g1")
	if err != nil {
		t.Fatalf("get google user: %v", err)
	}
	if user.Username != "alice@gmail.com" {
		t.Errorf("expected alice@gmail.com, got %s", user.Username)
	}

	_, err = store.GetUserByGoogleID("unknown")
	if err == nil {
		t.Error("expected error for unknown google ID")
	}
}

func TestUserStore_EnsureLegacyUser(t *testing.T) {
	store := newTestStore(t)

	// This should create the legacy user
	_, err := store.ensureLegacyUser()
	if err != nil {
		t.Fatalf("ensure legacy user: %v", err)
	}

	user, err := store.GetUserByUsername("legacy")
	if err != nil {
		t.Fatalf("get legacy user: %v", err)
	}
	if user.Username != "legacy" {
		t.Errorf("expected legacy, got %s", user.Username)
	}

	// Running again should be fine
	_, err = store.ensureLegacyUser()
	if err != nil {
		t.Fatalf("ensure legacy user again: %v", err)
	}
}

func TestUserStore_CredentialErrors(t *testing.T) {
	store := newTestAuthStore(t)
	userID := createUserAndGetID(t, store, "u")

	// Get non-existent
	_, err := store.getCredentialByID(userID, 999)
	if err == nil {
		t.Error("expected error for non-existent cred")
	}
}
