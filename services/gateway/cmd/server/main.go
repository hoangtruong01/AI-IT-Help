package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"eomp/packages/shared/pkg/auth"
	"eomp/packages/shared/pkg/logger"
	"eomp/packages/shared/pkg/metrics"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/gateway/internal/config"
	"eomp/services/gateway/internal/handler"
	gwMiddleware "eomp/services/gateway/internal/middleware"
	"eomp/services/gateway/internal/proxy"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

	// Fail-fast configuration validation
	if err := cfg.Validate(); err != nil {
		log.Error("gateway configuration validation failed (fail-fast)", slog.Any("error", err))
		os.Exit(1)
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, 60*time.Minute, 7*24*time.Hour)
	authFilter := gwMiddleware.GatewayAuth(jwtManager)

	// 1. Initialize Reverse Proxies
	authProxy, err := proxy.NewReverseProxy(cfg.AuthServiceURL, log)
	if err != nil {
		log.Error("failed to initialize auth proxy", slog.Any("error", err))
		os.Exit(1)
	}

	employeeProxy, err := proxy.NewReverseProxy(cfg.EmployeeServiceURL, log)
	if err != nil {
		log.Error("failed to initialize employee proxy", slog.Any("error", err))
		os.Exit(1)
	}

	assetProxy, err := proxy.NewReverseProxy(cfg.AssetServiceURL, log)
	if err != nil {
		log.Error("failed to initialize asset proxy", slog.Any("error", err))
		os.Exit(1)
	}

	helpdeskProxy, err := proxy.NewReverseProxy(cfg.HelpdeskServiceURL, log)
	if err != nil {
		log.Error("failed to initialize helpdesk proxy", slog.Any("error", err))
		os.Exit(1)
	}

	workflowProxy, err := proxy.NewReverseProxy(cfg.WorkflowServiceURL, log)
	if err != nil {
		log.Error("failed to initialize workflow proxy", slog.Any("error", err))
		os.Exit(1)
	}

	notificationProxy, err := proxy.NewReverseProxy(cfg.NotificationServiceURL, log)
	if err != nil {
		log.Error("failed to initialize notification proxy", slog.Any("error", err))
		os.Exit(1)
	}

	knowledgeProxy, err := proxy.NewReverseProxy(cfg.KnowledgeServiceURL, log)
	if err != nil {
		log.Error("failed to initialize knowledge proxy", slog.Any("error", err))
		os.Exit(1)
	}

	aiProxy, err := proxy.NewReverseProxy(cfg.AIServiceURL, log)
	if err != nil {
		log.Error("failed to initialize ai proxy", slog.Any("error", err))
		os.Exit(1)
	}

	reportingProxy, err := proxy.NewReverseProxy(cfg.ReportingServiceURL, log)
	if err != nil {
		log.Error("failed to initialize reporting proxy", slog.Any("error", err))
		os.Exit(1)
	}

	auditProxy, err := proxy.NewReverseProxy(cfg.AuditServiceURL, log)
	if err != nil {
		log.Error("failed to initialize audit proxy", slog.Any("error", err))
		os.Exit(1)
	}

	healthHandler := handler.NewHealthHandler(cfg)
	monitoringHandler := handler.NewMonitoringHandler()

	// 2. Setup Routing Multiplexer
	mux := http.NewServeMux()

	// Health & Prometheus Metrics (Test Case 8.1)
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	// Enterprise Observability & SRE Monitoring APIs (Phase 8)
	mux.HandleFunc("GET /api/v1/monitoring/overview", monitoringHandler.GetOverview)
	mux.HandleFunc("GET /api/v1/monitoring/services", monitoringHandler.ListServices)
	mux.HandleFunc("POST /api/v1/monitoring/probe/{id}", monitoringHandler.ProbeService)
	mux.HandleFunc("GET /api/v1/monitoring/logs", monitoringHandler.GetLogs)

	// Auth Service Routing with Strict Brute-Force Rate Limiter (10 req/min/IP - Task 1.5)
	strictAuthLimiter := middleware.IPRateLimiterWithProxies(10, 1*time.Minute, cfg.TrustedProxies)
	mux.Handle("/api/v1/auth/", strictAuthLimiter(authProxy))

	// Employee & Department Routing
	mux.Handle("/api/v1/employees", authFilter(employeeProxy))
	mux.Handle("/api/v1/employees/", authFilter(employeeProxy))
	mux.Handle("/api/v1/departments", authFilter(employeeProxy))
	mux.Handle("/api/v1/departments/", authFilter(employeeProxy))

	// Asset & CMDB Routing
	mux.Handle("/api/v1/assets", authFilter(assetProxy))
	mux.Handle("/api/v1/assets/", authFilter(assetProxy))
	mux.Handle("/api/v1/cmdb/", authFilter(assetProxy))

	// Helpdesk & Service Catalog Routing
	mux.Handle("/api/v1/tickets", authFilter(helpdeskProxy))
	mux.Handle("/api/v1/tickets/", authFilter(helpdeskProxy))
	mux.Handle("/api/v1/services/", authFilter(helpdeskProxy))
	mux.Handle("/api/v1/problems", authFilter(helpdeskProxy))
	mux.Handle("/api/v1/problems/", authFilter(helpdeskProxy))

	// Workflow Engine & Approvals & Changes Routing
	mux.Handle("/api/v1/workflows", authFilter(workflowProxy))
	mux.Handle("/api/v1/workflows/", authFilter(workflowProxy))
	mux.Handle("/api/v1/approvals", authFilter(workflowProxy))
	mux.Handle("/api/v1/approvals/", authFilter(workflowProxy))
	mux.Handle("/api/v1/changes", authFilter(workflowProxy))
	mux.Handle("/api/v1/changes/", authFilter(workflowProxy))

	// Notification Center Routing
	mux.Handle("/api/v1/notifications", authFilter(notificationProxy))
	mux.Handle("/api/v1/notifications/", authFilter(notificationProxy))

	// Knowledge Base & Vector Documents Routing
	mux.Handle("/api/v1/knowledge", authFilter(knowledgeProxy))
	mux.Handle("/api/v1/knowledge/", authFilter(knowledgeProxy))

	// AI Operations Copilot Routing
	mux.Handle("/api/v1/ai", authFilter(aiProxy))
	mux.Handle("/api/v1/ai/", authFilter(aiProxy))

	// Reporting & BI Analytics Routing (Phase 9)
	mux.Handle("/api/v1/reports", authFilter(reportingProxy))
	mux.Handle("/api/v1/reports/", authFilter(reportingProxy))

	// Audit Trail & Compliance Routing - Strict RBAC (Admin & Manager Only - Test Case 10.1)
	adminRoleFilter := middleware.RequireRoles("ROLE_ADMIN", "ROLE_MANAGER")
	mux.Handle("/api/v1/audit", authFilter(adminRoleFilter(auditProxy)))
	mux.Handle("/api/v1/audit/", authFilter(adminRoleFilter(auditProxy)))

	// Apply Global Gateway Middleware Stack with Dynamic CORS, Anti-Spoofing Limiter (100 req/min), and RED Metrics
	handlerStack := middleware.Recoverer(log)(
		middleware.IPRateLimiterWithProxies(100, 1*time.Minute, cfg.TrustedProxies)(
			metrics.HTTPMetricsMiddleware(cfg.ServiceName)(
				middleware.RequestLogger(log)(
					middleware.DynamicCORS(cfg.CORSAllowedOrigins)(mux),
				),
			),
		),
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handlerStack,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("gateway server starting",
			slog.Int("port", cfg.Port),
			slog.String("version", cfg.Version),
			slog.String("auth_service", cfg.AuthServiceURL),
			slog.String("employee_service", cfg.EmployeeServiceURL),
			slog.String("asset_service", cfg.AssetServiceURL),
			slog.String("helpdesk_service", cfg.HelpdeskServiceURL),
			slog.String("workflow_service", cfg.WorkflowServiceURL),
			slog.String("notification_service", cfg.NotificationServiceURL),
			slog.String("knowledge_service", cfg.KnowledgeServiceURL),
			slog.String("ai_service", cfg.AIServiceURL),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("gateway failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-shutdownChan
	log.Info("shutting down gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("gateway forced shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("gateway stopped gracefully")
}
