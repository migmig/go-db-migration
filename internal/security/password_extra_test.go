package security

import (
	"strings"
	"testing"
)

func TestHashPassword_TooLong(t *testing.T) {
	// Bcrypt max length is 72. But HashPassword calls ValidatePasswordPolicy first.
	// Let's check ValidatePasswordPolicy.
	longPass := strings.Repeat("a", 100)
	_, err := HashPassword(longPass)
	if err == nil {
		t.Error("expected error for too long password")
	}
}

func TestVerifyPassword_Empty(t *testing.T) {
	if VerifyPassword("", "p") {
		t.Error("expected false")
	}
	if VerifyPassword("h", "") {
		t.Error("expected false")
	}
}
