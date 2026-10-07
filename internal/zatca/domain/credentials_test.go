package domain

import (
	"encoding/base64"
	"testing"
)

func TestLiveCredentialsRejectSimulation(t *testing.T) {
	fake := &CSID{Status: "ACTIVE", CSIDType: CSIDTypeProduction, Secret: "simulated-secret", BinarySecurityToken: base64.StdEncoding.EncodeToString([]byte("simulated-token"))}
	if err := ValidateLiveCSID(fake, EnvironmentSimulation); err != nil {
		t.Fatal(err)
	}
	for _, env := range []Environment{EnvironmentSandbox, EnvironmentProduction} {
		if err := ValidateLiveCSID(fake, env); err == nil {
			t.Fatalf("simulator credentials accepted in %s", env)
		}
	}
}
