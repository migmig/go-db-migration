package security

import (
	"strings"
	"testing"
)

func TestNewCredentialCipherRequiresKey(t *testing.T) {
	_, err := NewCredentialCipher("")
	if err == nil {
		t.Fatal("expected error for empty key")
	}
}

func TestNewCredentialCipherValidatesKeyLength(t *testing.T) {
	_, err := NewCredentialCipher("short-key")
	if err == nil {
		t.Fatal("expected error for invalid key length")
	}
}

func TestCredentialCipherEncryptDecrypt(t *testing.T) {
	c, err := NewCredentialCipher("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	enc, err := c.Encrypt("my-db-password")
	if err != nil {
		t.Fatalf("unexpected encrypt error: %v", err)
	}
	if enc == "my-db-password" {
		t.Fatal("encrypted value must differ from plain text")
	}

	dec, err := c.Decrypt(enc)
	if err != nil {
		t.Fatalf("unexpected decrypt error: %v", err)
	}
	if dec != "my-db-password" {
		t.Fatalf("decrypt result = %q, want %q", dec, "my-db-password")
	}
}

func TestCredentialCipherDecryptInvalidPayload(t *testing.T) {
	c, err := NewCredentialCipher("0123456789abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = c.Decrypt("invalid-payload")
	if err == nil {
		t.Fatal("expected format error")
	}
}

func TestCredentialCipherDecryptTamperedPayload(t *testing.T) {
	c, err := NewCredentialCipher("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	enc, err := c.Encrypt("top-secret")
	if err != nil {
		t.Fatalf("unexpected encrypt error: %v", err)
	}

	tampered := enc[:len(enc)-2] + "xx"
	_, err = c.Decrypt(tampered)
	if err == nil {
		t.Fatal("expected decrypt error for tampered payload")
	}
}

func TestCredentialCipherPayloadContainsSeparator(t *testing.T) {
	c, err := NewCredentialCipher("0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	enc, err := c.Encrypt("abc")
	if err != nil {
		t.Fatalf("unexpected encrypt error: %v", err)
	}
	if !strings.Contains(enc, ":") {
		t.Fatalf("expected encoded payload format nonce:ciphertext, got %q", enc)
	}
}
