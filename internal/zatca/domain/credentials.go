package domain

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"strings"
	"time"
)

// ValidateLiveCSID rejects simulator placeholders and expired certificates before
// any external call. This is a local preflight, not authority-chain certification.
func ValidateLiveCSID(csid *CSID, environment Environment) error {
	if environment == EnvironmentSimulation {
		return nil
	}
	if csid == nil || csid.Status != "ACTIVE" || strings.TrimSpace(csid.Secret) == "" {
		return fmt.Errorf("active live CSID credentials are required")
	}
	if environment == EnvironmentProduction && csid.CSIDType != CSIDTypeProduction {
		return fmt.Errorf("production CSID is required for production submissions")
	}
	encoded, err := base64.StdEncoding.DecodeString(csid.BinarySecurityToken)
	if err != nil {
		return fmt.Errorf("live CSID security token must contain a certificate")
	}
	if block, _ := pem.Decode(encoded); block != nil {
		encoded = block.Bytes
	}
	cert, err := x509.ParseCertificate(encoded)
	if err != nil {
		return fmt.Errorf("live CSID security token is not an X.509 certificate; simulator credentials are forbidden")
	}
	now := time.Now()
	if now.Before(cert.NotBefore) || !now.Before(cert.NotAfter) || (csid.ExpiresAt != nil && !now.Before(*csid.ExpiresAt)) {
		return fmt.Errorf("live CSID certificate is not currently valid")
	}
	return nil
}
