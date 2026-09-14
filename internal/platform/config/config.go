package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvStaging     Environment = "staging"
	EnvProduction  Environment = "production"
	EnvTesting     Environment = "testing"
)

type Config struct {
	AppName     string
	AppVersion  string
	Environment Environment
	HTTP        HTTPConfig
	Database    DatabaseConfig
	LogLevel    string
}

type HTTPConfig struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	URL      string
	MaxConns int32
	MinConns int32
}

func Load() (*Config, error) {
	envStr := getEnv("APP_ENV", string(EnvDevelopment))
	env := Environment(envStr)

	cfg := &Config{
		AppName:     getEnv("APP_NAME", "Saudi-ERP"),
		AppVersion:  getEnv("APP_VERSION", "0.1.0"),
		Environment: env,
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		HTTP: HTTPConfig{
			Port:         getEnv("HTTP_PORT", "8080"),
			ReadTimeout:  getDurationEnv("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", "postgres://postgres:toor@localhost:5432/erp_dev?sslmode=disable"),
			MaxConns: int32(getIntEnv("DB_MAX_CONNS", 25)),
			MinConns: int32(getIntEnv("DB_MIN_CONNS", 5)),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.AppName == "" {
		return errors.New("app name cannot be empty")
	}

	switch c.Environment {
	case EnvDevelopment, EnvStaging, EnvProduction, EnvTesting:
	default:
		return fmt.Errorf("invalid environment: %s", c.Environment)
	}

	if c.HTTP.Port == "" {
		return errors.New("http port cannot be empty")
	}

	if c.Database.MaxConns <= 0 {
		return errors.New("database max connections must be greater than 0")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return fallback
}
