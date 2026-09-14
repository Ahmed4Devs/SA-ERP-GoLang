package crypto_test

import (
	"testing"

	"erp/internal/iam/infrastructure/crypto"
)

func TestArgon2idHasher(t *testing.T) {
	hasher := crypto.NewArgon2idHasher()
	password := "SecretPassword@123"

	// 1. Hash password
	encoded, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if len(encoded) == 0 {
		t.Fatal("expected non-empty hash")
	}

	// 2. Verify with correct password
	match, err := hasher.Verify(password, encoded)
	if err != nil {
		t.Fatalf("unexpected verification error: %v", err)
	}
	if !match {
		t.Error("expected password to match hash")
	}

	// 3. Verify with wrong password
	matchWrong, err := hasher.Verify("WrongPassword!999", encoded)
	if err != nil {
		t.Fatalf("unexpected verification error: %v", err)
	}
	if matchWrong {
		t.Error("expected wrong password to fail verification")
	}
}
