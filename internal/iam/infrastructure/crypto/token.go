package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

type SessionTokenGenerator struct{}

func NewSessionTokenGenerator() *SessionTokenGenerator {
	return &SessionTokenGenerator{}
}

func (g *SessionTokenGenerator) GenerateSessionToken() (rawToken string, tokenHash string, err error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	rawToken = hex.EncodeToString(bytes)

	hash := sha256.Sum256([]byte(rawToken))
	tokenHash = hex.EncodeToString(hash[:])

	return rawToken, tokenHash, nil
}

// HashToken computes the SHA-256 digest of a raw token.
func HashToken(rawToken string) string {
	hash := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(hash[:])
}
