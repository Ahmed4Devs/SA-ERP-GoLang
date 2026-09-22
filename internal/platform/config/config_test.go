package config

import (
	"encoding/base64"
	"testing"
	"time"
)

func TestConfigLoadDefault(t *testing.T) {
	// Isolate the default-loading contract from CI and integration-test
	// environment variables inherited by the test process.
	t.Setenv("APP_ENV", "")
	t.Setenv("HTTP_PORT", "")
	t.Setenv("TEST_DATABASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error loading config, got %v", err)
	}

	if cfg.AppName != "Saudi-ERP" {
		t.Errorf("expected app name 'Saudi-ERP', got %s", cfg.AppName)
	}

	if cfg.Environment != EnvDevelopment {
		t.Errorf("expected environment 'development', got %s", cfg.Environment)
	}

	if cfg.HTTP.Port != "8080" {
		t.Errorf("expected default HTTP port '8080', got %s", cfg.HTTP.Port)
	}
}

func TestConfigValidation(t *testing.T) {
	cfg := &Config{
		AppName:     "",
		Environment: EnvDevelopment,
		HTTP:        HTTPConfig{Port: "8080"},
		Database:    DatabaseConfig{URL: "postgres://localhost/erp", MaxConns: 10},
		Storage:     StorageConfig{Provider: "local", LocalDir: "./data/storage"},
		Security:    SecurityConfig{AllowedOrigins: []string{"http://localhost:3000"}},
		Mail:        MailConfig{Driver: "mock"},
		ZATCA:       ZATCAConfig{Environment: "simulation"},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty app name, got nil")
	}

	cfg.AppName = "ERP"
	cfg.Environment = Environment("invalid_env")
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid environment, got nil")
	}

	cfg.Environment = EnvDevelopment
	cfg.HTTP.Port = ""
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty HTTP port, got nil")
	}

	cfg.HTTP.Port = "8080"
	cfg.Database.MaxConns = 0
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for maxConns <= 0, got nil")
	}

	cfg.Database.MaxConns = 5
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestGetDurationEnv(t *testing.T) {
	t.Setenv("TEST_DURATION", "30s")

	d := getDurationEnv("TEST_DURATION", 10*time.Second)
	if d != 30*time.Second {
		t.Errorf("expected 30s, got %v", d)
	}

	dFallback := getDurationEnv("NON_EXISTENT_VAR", 5*time.Minute)
	if dFallback != 5*time.Minute {
		t.Errorf("expected fallback 5m, got %v", dFallback)
	}
}

func TestProductionZATCAPolicy(t *testing.T) {
	newProductionConfig := func() *Config {
		return &Config{
			AppName:     "ERP",
			Environment: EnvProduction,
			HTTP:        HTTPConfig{Port: "8080"},
			Database:    DatabaseConfig{URL: "postgres://erp:secret@db/erp", MaxConns: 10},
			Storage:     StorageConfig{Provider: "local", LocalDir: "/app/data/storage"},
			Security: SecurityConfig{
				AllowedOrigins: []string{"https://erp.example.sa"},
				MFAKey:         base64.StdEncoding.EncodeToString(make([]byte, 32)),
				MetricsToken:   "0123456789abcdef0123456789abcdef",
			},
			Mail:      MailConfig{Driver: "smtp", Host: "smtp.example.sa", Port: 587, From: "erp@example.sa"},
			ZATCA:     ZATCAConfig{Environment: "simulation"},
			Telemetry: TelemetryConfig{SampleRatio: 0.1},
		}
	}
	zatcaKey := base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))

	t.Run("disabled integration is allowed", func(t *testing.T) {
		cfg := newProductionConfig()
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected disabled ZATCA integration to be valid in production, got %v", err)
		}
	})

	t.Run("enabled integration requires acknowledgement", func(t *testing.T) {
		cfg := newProductionConfig()
		cfg.ZATCA.Enabled = true
		cfg.ZATCA.Environment = "production"
		cfg.ZATCA.BaseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/core"
		cfg.ZATCA.CredentialsKey = zatcaKey
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected enabled ZATCA integration without acknowledgement to be rejected")
		}
	})

	t.Run("enabled integration requires HTTPS", func(t *testing.T) {
		cfg := newProductionConfig()
		cfg.ZATCA.Enabled = true
		cfg.ZATCA.Environment = "production"
		cfg.ZATCA.ProductionAcknowledged = true
		cfg.ZATCA.BaseURL = "http://zatca.example.sa"
		cfg.ZATCA.CredentialsKey = zatcaKey
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected enabled ZATCA integration with an HTTP endpoint to be rejected")
		}
	})

	t.Run("enabled acknowledged HTTPS integration is allowed", func(t *testing.T) {
		cfg := newProductionConfig()
		cfg.ZATCA.Enabled = true
		cfg.ZATCA.Environment = "production"
		cfg.ZATCA.ProductionAcknowledged = true
		cfg.ZATCA.BaseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/core"
		cfg.ZATCA.CredentialsKey = zatcaKey
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected acknowledged HTTPS ZATCA integration to be valid, got %v", err)
		}
	})

	t.Run("enabled live integration requires a credential encryption key", func(t *testing.T) {
		cfg := newProductionConfig()
		cfg.ZATCA.Enabled = true
		cfg.ZATCA.Environment = "production"
		cfg.ZATCA.ProductionAcknowledged = true
		cfg.ZATCA.BaseURL = "https://gw-fatoora.zatca.gov.sa/e-invoicing/core"
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected live ZATCA integration without credential encryption to be rejected")
		}
	})

	t.Run("configured credential encryption key is always validated", func(t *testing.T) {
		cfg := newProductionConfig()
		cfg.ZATCA.Enabled = false
		cfg.ZATCA.Environment = "simulation"
		cfg.ZATCA.CredentialsKey = "not-a-valid-key"
		if err := cfg.Validate(); err == nil {
			t.Fatal("expected an invalid optional credential encryption key to be rejected")
		}
	})
}
