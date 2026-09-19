package crypto

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestSecretBoxRoundTripAndTamperDetection(t *testing.T) {
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef"))
	box, err := NewSecretBox(key)
	if err != nil {
		t.Fatalf("NewSecretBox() error = %v", err)
	}

	encrypted, err := box.Encrypt("JBSWY3DPEHPK3PXP")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}
	if encrypted == "" || strings.Contains(encrypted, "JBSWY3DPEHPK3PXP") {
		t.Fatal("encrypted secret is empty or exposes plaintext")
	}

	decrypted, err := box.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}
	if decrypted != "JBSWY3DPEHPK3PXP" {
		t.Fatalf("Decrypt() = %q", decrypted)
	}

	tamperedBytes, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		t.Fatalf("failed to decode encrypted test value: %v", err)
	}
	tamperedBytes[len(tamperedBytes)-1] ^= 1
	tampered := base64.RawURLEncoding.EncodeToString(tamperedBytes)
	if _, err := box.Decrypt(tampered); err == nil {
		t.Fatal("Decrypt() accepted tampered ciphertext")
	}
}

func TestSecretBoxRejectsInvalidKeys(t *testing.T) {
	for _, key := range []string{"not-base64", base64.StdEncoding.EncodeToString([]byte("too short"))} {
		if _, err := NewSecretBox(key); err == nil {
			t.Fatalf("NewSecretBox(%q) unexpectedly succeeded", key)
		}
	}
}
