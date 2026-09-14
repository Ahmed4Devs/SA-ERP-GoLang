package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigLoadDefault(t *testing.T) {
	// Clear env vars that might interfere
	os.Unsetenv("APP_ENV")
	os.Unsetenv("HTTP_PORT")

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
		Database:    DatabaseConfig{MaxConns: 10},
	}

	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty app name, got nil")
	}

	cfg.AppName = "ERP"
	cfg.Environment = Environment("invalid_env")
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid environment, got nil")
	}

	cfg.Environment = EnvProduction
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
	os.Setenv("TEST_DURATION", "30s")
	defer os.Unsetenv("TEST_DURATION")

	d := getDurationEnv("TEST_DURATION", 10*time.Second)
	if d != 30*time.Second {
		t.Errorf("expected 30s, got %v", d)
	}

	dFallback := getDurationEnv("NON_EXISTENT_VAR", 5*time.Minute)
	if dFallback != 5*time.Minute {
		t.Errorf("expected fallback 5m, got %v", dFallback)
	}
}
