package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// SecretBox encrypts application secrets with AES-256-GCM. Ciphertexts include
// a fresh nonce and are authenticated, so tampering causes decryption to fail.
type SecretBox struct {
	aead cipher.AEAD
}

func NewSecretBox(base64Key string) (*SecretBox, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil || len(key) != 32 {
		return nil, fmt.Errorf("encryption key must contain 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SecretBox{aead: aead}, nil
}

func (b *SecretBox) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := b.aead.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

func (b *SecretBox) Decrypt(encoded string) (string, error) {
	sealed, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(sealed) < b.aead.NonceSize() || base64.RawURLEncoding.EncodeToString(sealed) != encoded {
		return "", fmt.Errorf("invalid encrypted secret")
	}
	nonce, ciphertext := sealed[:b.aead.NonceSize()], sealed[b.aead.NonceSize():]
	plaintext, err := b.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret: %w", err)
	}
	return string(plaintext), nil
}
