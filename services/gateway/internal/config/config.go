package config

import (
	"errors"
	"fmt"

	"eomp/packages/shared/pkg/config"
)

// Config represents gateway service configuration
type Config struct {
	ServiceName            string
	Port                   int
	Environment            string
	Version                string
	JWTSecret              string
	CORSAllowedOrigins     []string
	TrustedProxies         []string
	AuthServiceURL         string
	EmployeeServiceURL     string
	AssetServiceURL        string
	HelpdeskServiceURL     string
	WorkflowServiceURL     string
	NotificationServiceURL string
	KnowledgeServiceURL    string
	AIServiceURL           string
	AuditServiceURL        string
	ReportingServiceURL    string
}

// Load reads gateway configuration from environment
func Load() *Config {
	defaultCORS := []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:8080"}
	defaultProxies := []string{"127.0.0.1", "::1", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}

	return &Config{
		ServiceName:            "gateway",
		Port:                   config.GetEnvInt("PORT", 8080),
		Environment:            config.GetEnv("APP_ENV", "development"),
		Version:                "0.1.0",
		JWTSecret:              config.GetEnv("JWT_SECRET", "eomp-enterprise-super-secret-jwt-key-2026"),
		CORSAllowedOrigins:     config.GetEnvSlice("CORS_ALLOWED_ORIGINS", defaultCORS),
		TrustedProxies:         config.GetEnvSlice("TRUSTED_PROXIES", defaultProxies),
		AuthServiceURL:         config.GetEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		EmployeeServiceURL:     config.GetEnv("EMPLOYEE_SERVICE_URL", "http://localhost:8082"),
		AssetServiceURL:        config.GetEnv("ASSET_SERVICE_URL", "http://localhost:8083"),
		HelpdeskServiceURL:     config.GetEnv("HELPDESK_SERVICE_URL", "http://localhost:8084"),
		WorkflowServiceURL:     config.GetEnv("WORKFLOW_SERVICE_URL", "http://localhost:8085"),
		NotificationServiceURL: config.GetEnv("NOTIFICATION_SERVICE_URL", "http://localhost:8086"),
		KnowledgeServiceURL:    config.GetEnv("KNOWLEDGE_SERVICE_URL", "http://localhost:8087"),
		AIServiceURL:           config.GetEnv("AI_SERVICE_URL", "http://localhost:8088"),
		AuditServiceURL:        config.GetEnv("AUDIT_SERVICE_URL", "http://localhost:8089"),
		ReportingServiceURL:    config.GetEnv("REPORTING_SERVICE_URL", "http://localhost:8090"),
	}
}

// Validate performs fail-fast verification on gateway configuration.
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return errors.New("security violation: JWT_SECRET must not be empty in gateway")
	}
	if len(c.JWTSecret) < 16 {
		return fmt.Errorf("security violation: JWT_SECRET must be at least 16 characters long, got %d", len(c.JWTSecret))
	}
	if c.Environment == "production" {
		if c.JWTSecret == "eomp-enterprise-super-secret-jwt-key-2026" {
			return errors.New("security violation: default dev JWT_SECRET is strictly prohibited in production")
		}
	}
	return nil
}
