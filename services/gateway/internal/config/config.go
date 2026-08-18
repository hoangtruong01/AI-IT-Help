package config

import (
	"eomp/packages/shared/pkg/config"
)

// Config represents gateway service configuration
type Config struct {
	ServiceName            string
	Port                   int
	Environment            string
	Version                string
	JWTSecret              string
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
	return &Config{
		ServiceName:            "gateway",
		Port:                   config.GetEnvInt("PORT", 8080),
		Environment:            config.GetEnv("APP_ENV", "development"),
		Version:                "0.1.0",
		JWTSecret:              config.GetEnv("JWT_SECRET", "eomp-enterprise-super-secret-jwt-key-2026"),
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
