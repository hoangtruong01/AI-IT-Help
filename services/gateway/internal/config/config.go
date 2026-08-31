package config

import "eomp/packages/shared/pkg/config"

// Config represents gateway service configuration.
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
	RedisHost              string
	RedisPort              int
	RedisPassword          string
	RedisDB                int
}

// Load reads gateway configuration from the environment.
func Load() *Config {
	defaultCORS := []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:8080"}
	defaultProxies := []string{"127.0.0.1", "::1"}

	return &Config{
		ServiceName:            "gateway",
		Port:                   config.GetEnvInt("PORT", 8080),
		Environment:            config.GetEnv("APP_ENV", "development"),
		Version:                "0.1.0",
		JWTSecret:              config.GetEnv("JWT_SECRET", ""),
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
		RedisHost:              config.GetEnv("REDIS_HOST", "localhost"),
		RedisPort:              config.GetEnvInt("REDIS_PORT", 6379),
		RedisPassword:          config.GetEnv("REDIS_PASSWORD", ""),
		RedisDB:                config.GetEnvInt("REDIS_DB", 0),
	}
}

// Validate performs fail-fast verification in every runtime environment.
func (c *Config) Validate() error {
	return config.ValidateRequiredSecret(
		"JWT_SECRET",
		c.JWTSecret,
		32,
		"eomp-enterprise-super-secret-jwt-key-2026",
	)
}
