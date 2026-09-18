package services

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestValidateTOTPUsesBoundedClockSkew(t *testing.T) {
	const secret = "JBSWY3DPEHPK3PXP"
	now := time.Unix(1_800_000_000, 0)

	for _, offset := range []time.Duration{-30 * time.Second, 0, 30 * time.Second} {
		code, err := totp.GenerateCode(secret, now.Add(offset))
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}
		counter, ok := validateTOTP(code, secret, now)
		if !ok || counter != now.Add(offset).Unix()/30 {
			t.Fatalf("validateTOTP(offset=%s) = (%d, %t)", offset, counter, ok)
		}
	}

	outsideCode, err := totp.GenerateCode(secret, now.Add(60*time.Second))
	if err != nil {
		t.Fatalf("GenerateCode() error = %v", err)
	}
	if _, ok := validateTOTP(outsideCode, secret, now); ok {
		t.Fatal("validateTOTP() accepted a code outside the allowed clock skew")
	}
}

func TestGenerateRecoveryCodesProducesUniqueCodesAndHashes(t *testing.T) {
	codes, hashes, err := generateRecoveryCodes(10)
	if err != nil {
		t.Fatalf("generateRecoveryCodes() error = %v", err)
	}
	if len(codes) != 10 || len(hashes) != 10 {
		t.Fatalf("unexpected output sizes: codes=%d hashes=%d", len(codes), len(hashes))
	}
	seenCodes := make(map[string]struct{}, len(codes))
	seenHashes := make(map[string]struct{}, len(hashes))
	for i := range codes {
		if _, duplicate := seenCodes[codes[i]]; duplicate {
			t.Fatalf("duplicate recovery code %q", codes[i])
		}
		if _, duplicate := seenHashes[hashes[i]]; duplicate {
			t.Fatalf("duplicate recovery hash %q", hashes[i])
		}
		seenCodes[codes[i]] = struct{}{}
		seenHashes[hashes[i]] = struct{}{}
	}
}
