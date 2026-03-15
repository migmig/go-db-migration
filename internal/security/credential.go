package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

const encryptedPayloadSeparator = ":"

type CredentialCipher struct {
	gcm cipher.AEAD
}

func NewCredentialCipher(masterKey string) (*CredentialCipher, error) {
	if strings.TrimSpace(masterKey) == "" {
		return nil, fmt.Errorf("master key is required")
	}

	key := []byte(masterKey)
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("master key length must be 16, 24, or 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("create aes cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create gcm cipher: %w", err)
	}

	return &CredentialCipher{gcm: gcm}, nil
}

func (c *CredentialCipher) Encrypt(plainText string) (string, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("read nonce: %w", err)
	}

	cipherText := c.gcm.Seal(nil, nonce, []byte(plainText), nil)

	nonceEncoded := base64.StdEncoding.EncodeToString(nonce)
	cipherEncoded := base64.StdEncoding.EncodeToString(cipherText)

	return nonceEncoded + encryptedPayloadSeparator + cipherEncoded, nil
}

func (c *CredentialCipher) Decrypt(encrypted string) (string, error) {
	parts := strings.SplitN(encrypted, encryptedPayloadSeparator, 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid encrypted payload format")
	}

	nonce, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("decode nonce: %w", err)
	}

	cipherText, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("decode cipher text: %w", err)
	}

	plainText, err := c.gcm.Open(nil, nonce, cipherText, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt payload: %w", err)
	}

	return string(plainText), nil
}
