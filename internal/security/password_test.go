package security

import "testing"

func TestValidatePasswordPolicy(t *testing.T) {
	t.Run("too short", func(t *testing.T) {
		err := ValidatePasswordPolicy("short")
		if err == nil {
			t.Fatal("expected error for short password")
		}
	})

	t.Run("valid length", func(t *testing.T) {
		err := ValidatePasswordPolicy("longenough")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestHashAndVerifyPassword(t *testing.T) {
	password := "very-secure-password"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected hash error: %v", err)
	}
	if hashed == password {
		t.Fatal("hashed password should not match plain password")
	}

	if !VerifyPassword(hashed, password) {
		t.Fatal("expected password verification to pass")
	}

	if VerifyPassword(hashed, "wrong-password") {
		t.Fatal("expected password verification to fail for wrong password")
	}
}

func TestHashPasswordPolicyFailure(t *testing.T) {
	_, err := HashPassword("short")
	if err == nil {
		t.Fatal("expected policy error")
	}
}

func TestVerifyPasswordEmptyInput(t *testing.T) {
	if VerifyPassword("", "abc") {
		t.Fatal("expected false for empty hash")
	}
	if VerifyPassword("hash", "") {
		t.Fatal("expected false for empty password")
	}
}
