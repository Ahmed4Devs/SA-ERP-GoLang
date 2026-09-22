package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
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
	Storage     StorageConfig
	Security    SecurityConfig
	Bootstrap   BootstrapConfig
	Mail        MailConfig
	ZATCA       ZATCAConfig
	Telemetry   TelemetryConfig
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

type StorageConfig struct {
	Provider string
	LocalDir string
}

type SecurityConfig struct {
	AllowedOrigins []string
	TrustedProxies []string
	MFAKey         string
	MetricsToken   string
}

type BootstrapConfig struct {
	AdminEnabled    bool
	DemoDataEnabled bool
	AdminUsername   string
	AdminEmail      string
	AdminName       string
	AdminPassword   string
}

type MailConfig struct {
	Driver   string
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type ZATCAConfig struct {
	Enabled                bool
	Environment            string
	BaseURL                string
	Timeout                time.Duration
	ProductionAcknowledged bool
	CredentialsKey         string
}

type TelemetryConfig struct {
	OTLPEndpoint string
	Insecure     bool
	SampleRatio  float64
}

func Load() (*Config, error) {
	testDatabaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	_, appEnvironmentSet := os.LookupEnv("APP_ENV")
	envStr := getEnv("APP_ENV", string(EnvDevelopment))
	if testDatabaseURL != "" && !appEnvironmentSet {
		envStr = string(EnvTesting)
	}
	env := Environment(envStr)
	databaseURL := getEnv("DATABASE_URL", "postgres://postgres:toor@localhost:5432/erp_dev?sslmode=disable")
	if env == EnvTesting && testDatabaseURL != "" {
		databaseURL = testDatabaseURL
	}
	defaultMaxConns := 25
	defaultMinConns := 5
	if env == EnvTesting {
		defaultMaxConns = 5
		defaultMinConns = 0
	}

	cfg := &Config{
		AppName:     getEnv("APP_NAME", "Saudi-ERP"),
		AppVersion:  getEnv("APP_VERSION", "1.0.0"),
		Environment: env,
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		HTTP: HTTPConfig{
			Port:         getEnv("HTTP_PORT", "8080"),
			ReadTimeout:  getDurationEnv("HTTP_READ_TIMEOUT", 15*time.Second),
			WriteTimeout: getDurationEnv("HTTP_WRITE_TIMEOUT", 15*time.Second),
			IdleTimeout:  getDurationEnv("HTTP_IDLE_TIMEOUT", 60*time.Second),
		},
		Database: DatabaseConfig{
			URL:      databaseURL,
			MaxConns: int32(getIntEnv("DB_MAX_CONNS", defaultMaxConns)),
			MinConns: int32(getIntEnv("DB_MIN_CONNS", defaultMinConns)),
		},
		Storage: StorageConfig{
			Provider: getEnv("STORAGE_PROVIDER", "local"),
			LocalDir: getEnv("STORAGE_LOCAL_DIR", "./data/storage"),
		},
		Security: SecurityConfig{
			AllowedOrigins: getCSVEnv("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://127.0.0.1:3000"}),
			TrustedProxies: getCSVEnv("TRUSTED_PROXIES", []string{"127.0.0.1", "::1"}),
			MFAKey:         strings.TrimSpace(getEnv("MFA_ENCRYPTION_KEY", "")),
			MetricsToken:   strings.TrimSpace(getEnv("METRICS_BEARER_TOKEN", "")),
		},
		Bootstrap: BootstrapConfig{
			AdminEnabled:    getBoolEnv("BOOTSTRAP_ADMIN_ENABLED", env != EnvProduction),
			DemoDataEnabled: getBoolEnv("BOOTSTRAP_DEMO_DATA_ENABLED", env != EnvProduction),
			AdminUsername:   getEnv("BOOTSTRAP_ADMIN_USERNAME", "admin"),
			AdminEmail:      getEnv("BOOTSTRAP_ADMIN_EMAIL", "admin@erp.sa"),
			AdminName:       getEnv("BOOTSTRAP_ADMIN_NAME", "مدير النظام الشامل"),
			AdminPassword:   getEnv("BOOTSTRAP_ADMIN_PASSWORD", "Admin@ERP123!"),
		},
		Mail: MailConfig{
			Driver:   strings.ToLower(getEnv("MAIL_DRIVER", "mock")),
			Host:     getEnv("SMTP_HOST", ""),
			Port:     getIntEnv("SMTP_PORT", 587),
			Username: getEnv("SMTP_USERNAME", ""),
			Password: getEnv("SMTP_PASSWORD", ""),
			From:     getEnv("SMTP_FROM", ""),
		},
		ZATCA: ZATCAConfig{
			Enabled:                getBoolEnv("ZATCA_ENABLED", env != EnvProduction),
			Environment:            strings.ToLower(getEnv("ZATCA_ENVIRONMENT", "simulation")),
			BaseURL:                getEnv("ZATCA_BASE_URL", ""),
			Timeout:                getDurationEnv("ZATCA_HTTP_TIMEOUT", 30*time.Second),
			ProductionAcknowledged: getBoolEnv("ZATCA_PRODUCTION_ACKNOWLEDGED", false),
			CredentialsKey:         strings.TrimSpace(getEnv("ZATCA_CREDENTIALS_ENCRYPTION_KEY", "")),
		},
		Telemetry: TelemetryConfig{
			OTLPEndpoint: strings.TrimSpace(getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")),
			Insecure:     getBoolEnv("OTEL_EXPORTER_OTLP_INSECURE", false),
			SampleRatio:  getFloatEnv("OTEL_TRACES_SAMPLER_RATIO", 0.1),
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
	if strings.TrimSpace(c.Database.URL) == "" {
		return errors.New("database URL cannot be empty")
	}
	if c.Database.MinConns < 0 || c.Database.MinConns > c.Database.MaxConns {
		return errors.New("database min connections must be between 0 and max connections")
	}
	if c.Storage.Provider != "local" {
		return fmt.Errorf("unsupported storage provider: %q", c.Storage.Provider)
	}
	if strings.TrimSpace(c.Storage.LocalDir) == "" {
		return errors.New("STORAGE_LOCAL_DIR cannot be empty")
	}
	if len(c.Security.AllowedOrigins) == 0 {
		return errors.New("at least one CORS allowed origin is required")
	}
	for _, origin := range c.Security.AllowedOrigins {
		u, err := url.Parse(origin)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("invalid CORS allowed origin: %q", origin)
		}
		if c.Environment == EnvProduction && u.Scheme != "https" {
			return fmt.Errorf("production CORS origin must use HTTPS: %q", origin)
		}
	}
	if c.Mail.Driver != "mock" && c.Mail.Driver != "smtp" {
		return fmt.Errorf("invalid mail driver: %q", c.Mail.Driver)
	}
	if c.Mail.Driver == "smtp" && (c.Mail.Host == "" || c.Mail.Port <= 0 || c.Mail.From == "") {
		return errors.New("SMTP host, port, and from address are required when MAIL_DRIVER=smtp")
	}
	if c.ZATCA.Environment != "sandbox" && c.ZATCA.Environment != "simulation" && c.ZATCA.Environment != "production" {
		return fmt.Errorf("invalid ZATCA environment: %q", c.ZATCA.Environment)
	}
	if c.ZATCA.CredentialsKey != "" {
		decoded, err := base64.StdEncoding.DecodeString(c.ZATCA.CredentialsKey)
		if err != nil || len(decoded) != 32 {
			return errors.New("ZATCA_CREDENTIALS_ENCRYPTION_KEY must be a base64-encoded 32-byte key")
		}
	}
	if c.ZATCA.Enabled && c.ZATCA.Environment != "simulation" && c.ZATCA.CredentialsKey == "" {
		return errors.New("ZATCA_CREDENTIALS_ENCRYPTION_KEY is required when live ZATCA is enabled")
	}
	if c.Telemetry.SampleRatio < 0 || c.Telemetry.SampleRatio > 1 {
		return errors.New("OTEL_TRACES_SAMPLER_RATIO must be between 0 and 1")
	}
	if c.Security.MFAKey != "" {
		decoded, err := base64.StdEncoding.DecodeString(c.Security.MFAKey)
		if err != nil || len(decoded) != 32 {
			return errors.New("MFA_ENCRYPTION_KEY must be a base64-encoded 32-byte key")
		}
	}
	if c.Environment == EnvProduction {
		if c.Security.MFAKey == "" {
			return errors.New("MFA_ENCRYPTION_KEY is required in production")
		}
		if c.Mail.Driver != "smtp" {
			return errors.New("MAIL_DRIVER must be smtp in production")
		}
		if len(c.Security.MetricsToken) < 32 {
			return errors.New("METRICS_BEARER_TOKEN of at least 32 characters is required in production")
		}
		if c.ZATCA.Enabled {
			if c.ZATCA.Environment != "production" {
				return errors.New("ZATCA_ENVIRONMENT must be production when ZATCA is enabled in production")
			}
			if !c.ZATCA.ProductionAcknowledged {
				return errors.New("ZATCA_PRODUCTION_ACKNOWLEDGED=true is required to enable the uncertified ZATCA integration")
			}
			endpoint, err := url.Parse(c.ZATCA.BaseURL)
			if err != nil || endpoint.Scheme != "https" || endpoint.Host == "" {
				return errors.New("ZATCA_BASE_URL must be a valid HTTPS URL in production")
			}
		}
		if c.Bootstrap.AdminEnabled {
			if len(c.Bootstrap.AdminPassword) < 16 || c.Bootstrap.AdminPassword == "Admin@ERP123!" || strings.Contains(strings.ToLower(c.Bootstrap.AdminPassword), "changeme") {
				return errors.New("a unique bootstrap admin password of at least 16 characters is required in production")
			}
		}
		if c.Bootstrap.DemoDataEnabled {
			return errors.New("BOOTSTRAP_DEMO_DATA_ENABLED must be false in production")
		}
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

func getFloatEnv(key string, fallback float64) float64 {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if parsed, err := strconv.ParseBool(val); err == nil {
			return parsed
		}
	}
	return fallback
}

func getCSVEnv(key string, fallback []string) []string {
	value, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(value) == "" {
		return append([]string(nil), fallback...)
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
