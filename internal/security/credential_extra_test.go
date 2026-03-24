package security

import (
	"testing"
)

func TestCredentialCipher_Encrypt_Error(t *testing.T) {
	// Key must be 32 bytes
	_, err := NewCredentialCipher("short")
	if err == nil {
		t.Error("expected error for short key")
	}
}

func TestCredentialCipher_Decrypt_Error(t *testing.T) {
	c, _ := NewCredentialCipher("12345678901234567890123456789012")
	_, err := c.Decrypt("too-short")
	if err == nil {
		t.Error("expected error for short ciphertext")
	}
	
	_, err = c.Decrypt("not-hex-at-all!!")
	if err == nil {
		t.Error("expected error for non-hex")
	}
}
