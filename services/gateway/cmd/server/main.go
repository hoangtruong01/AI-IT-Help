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
	pkgRedis "eomp/packages/shared/pkg/redis"
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

	// Initialize Redis Client with Graceful Fallback (Task 6.1)
	redisClient, err := pkgRedis.NewClient(pkgRedis.Config{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}, log)
	if err != nil {
		log.Warn("gateway running with in-memory rate limiter fallback (redis offline)", slog.Any("error", err))
	} else {
		defer redisClient.Close()
	}

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
	monitoringHandler := handler.NewMonitoringHandlerWithURLs(map[string]string{
		"gateway":      fmt.Sprintf("http://127.0.0.1:%d", cfg.Port),
		"auth":         cfg.AuthServiceURL,
		"employee":     cfg.EmployeeServiceURL,
		"asset":        cfg.AssetServiceURL,
		"helpdesk":     cfg.HelpdeskServiceURL,
		"workflow":     cfg.WorkflowServiceURL,
		"notification": cfg.NotificationServiceURL,
		"knowledge":    cfg.KnowledgeServiceURL,
		"ai":           cfg.AIServiceURL,
		"audit":        cfg.AuditServiceURL,
		"reporting":    cfg.ReportingServiceURL,
	})

	// 2. Setup Routing Multiplexer
	mux := http.NewServeMux()

	// Health & Prometheus Metrics (Test Case 8.1)
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	adminManager := middleware.RequireRoles("ROLE_ADMIN", "ROLE_MANAGER")
	operator := middleware.RequireRoles("ROLE_ADMIN", "ROLE_MANAGER", "ROLE_AGENT")
	operatorWrites := middleware.RequireRolesForMethods(
		[]string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		"ROLE_ADMIN", "ROLE_MANAGER", "ROLE_AGENT",
	)
	managerWrites := middleware.RequireRolesForMethods(
		[]string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		"ROLE_ADMIN", "ROLE_MANAGER",
	)

	// Monitoring data can expose infrastructure details and is restricted.
	mux.Handle("GET /api/v1/monitoring/overview", authFilter(adminManager(http.HandlerFunc(monitoringHandler.GetOverview))))
	mux.Handle("GET /api/v1/monitoring/services", authFilter(adminManager(http.HandlerFunc(monitoringHandler.ListServices))))
	mux.Handle("POST /api/v1/monitoring/probe/{id}", authFilter(adminManager(http.HandlerFunc(monitoringHandler.ProbeService))))
	mux.Handle("GET /api/v1/monitoring/logs", authFilter(adminManager(http.HandlerFunc(monitoringHandler.GetLogs))))

	// Auth Service Routing with Strict Redis Brute-Force Rate Limiter (10 req/min/IP - Task 1.5 & Task 6.1)
	strictAuthLimiter := middleware.StrictRedisAuthRateLimiter(redisClient, 10, 1*time.Minute, cfg.TrustedProxies, log)

	// Public auth routes — no authFilter needed
	mux.Handle("POST /api/v1/auth/login", strictAuthLimiter(authProxy))
	if cfg.Environment != "production" {
		mux.Handle("POST /api/v1/auth/register", strictAuthLimiter(authProxy))
	}
	mux.Handle("POST /api/v1/auth/refresh", strictAuthLimiter(authProxy))
	mux.Handle("POST /api/v1/auth/logout", strictAuthLimiter(authProxy))

	adminOnly := middleware.RequireRole("ROLE_ADMIN")

	// Protected auth routes — require valid JWT (A-02 fix: header spoofing prevention)
	mux.Handle("GET /api/v1/auth/me", authFilter(authProxy))
	mux.Handle("GET /api/v1/auth/login-history", authFilter(authProxy))
	mux.Handle("POST /api/v1/auth/change-password", authFilter(authProxy))
	mux.Handle("/api/v1/auth/reset-password/", authFilter(adminOnly(authProxy)))

	// User Administration Routing (Gate B-02)
	mux.Handle("GET /api/v1/users", authFilter(adminManager(authProxy)))
	mux.Handle("POST /api/v1/users", authFilter(adminOnly(authProxy)))
	mux.Handle("/api/v1/users/", authFilter(adminOnly(authProxy)))

	// Employee & Department Routing
	mux.Handle("/api/v1/employees", authFilter(managerWrites(employeeProxy)))
	mux.Handle("/api/v1/employees/", authFilter(managerWrites(employeeProxy)))
	mux.Handle("/api/v1/departments", authFilter(managerWrites(employeeProxy)))
	mux.Handle("/api/v1/departments/", authFilter(managerWrites(employeeProxy)))

	// Asset & CMDB Routing
	mux.Handle("/api/v1/assets", authFilter(operator(assetProxy)))
	mux.Handle("/api/v1/assets/", authFilter(operator(assetProxy)))
	mux.Handle("/api/v1/cmdb/", authFilter(operator(assetProxy)))

	// Helpdesk & Service Catalog Routing
	// Employees may create tickets and comments; operational ticket mutations are restricted.
	mux.Handle("POST /api/v1/tickets", authFilter(helpdeskProxy))
	mux.Handle("POST /api/v1/tickets/{id}/comments", authFilter(helpdeskProxy))
	mux.Handle("/api/v1/tickets", authFilter(operatorWrites(helpdeskProxy)))
	mux.Handle("/api/v1/tickets/", authFilter(operatorWrites(helpdeskProxy)))
	mux.Handle("/api/v1/services/", authFilter(helpdeskProxy))
	mux.Handle("/api/v1/problems", authFilter(operator(helpdeskProxy)))
	mux.Handle("/api/v1/problems/", authFilter(operator(helpdeskProxy)))

	// Workflow Engine & Approvals & Changes Routing
	mux.Handle("POST /api/v1/workflows/instances", authFilter(workflowProxy))
	mux.Handle("GET /api/v1/workflows/stats", authFilter(adminManager(workflowProxy)))
	mux.Handle("/api/v1/workflows", authFilter(managerWrites(workflowProxy)))
	mux.Handle("/api/v1/workflows/", authFilter(managerWrites(workflowProxy)))
	mux.Handle("/api/v1/approvals", authFilter(managerWrites(workflowProxy)))
	mux.Handle("/api/v1/approvals/", authFilter(managerWrites(workflowProxy)))
	mux.Handle("/api/v1/changes", authFilter(adminManager(workflowProxy)))
	mux.Handle("/api/v1/changes/", authFilter(adminManager(workflowProxy)))

	// Notification Center Routing
	mux.Handle("PATCH /api/v1/notifications/{id}/read", authFilter(notificationProxy))
	mux.Handle("POST /api/v1/notifications/read-all", authFilter(notificationProxy))
	mux.Handle("/api/v1/notifications", authFilter(operatorWrites(notificationProxy)))
	mux.Handle("/api/v1/notifications/", authFilter(operatorWrites(notificationProxy)))

	// Knowledge Base & Vector Documents Routing
	mux.Handle("/api/v1/knowledge", authFilter(operatorWrites(knowledgeProxy)))
	mux.Handle("/api/v1/knowledge/", authFilter(operatorWrites(knowledgeProxy)))

	// AI Operations Copilot Routing
	mux.Handle("/api/v1/ai", authFilter(aiProxy))
	mux.Handle("/api/v1/ai/", authFilter(aiProxy))

	// Reporting & BI Analytics Routing (Phase 9)
	mux.Handle("/api/v1/reports", authFilter(adminManager(reportingProxy)))
	mux.Handle("/api/v1/reports/", authFilter(adminManager(reportingProxy)))

	// Audit Trail & Compliance Routing - Strict RBAC (Admin & Manager Only - Test Case 10.1)
	mux.Handle("/api/v1/audit", authFilter(adminManager(auditProxy)))
	mux.Handle("/api/v1/audit/", authFilter(adminManager(auditProxy)))

	// Apply Global Gateway Middleware Stack
	// StripIdentityHeaders is the OUTERMOST layer — removes all X-User-*
	// headers from client requests before any routing or auth processing.
	handlerStack := gwMiddleware.StripIdentityHeaders(
		middleware.Recoverer(log)(
			middleware.MaxBodySize(5 * 1024 * 1024)(
				middleware.RedisSlidingWindowRateLimiter(redisClient, 100, 1*time.Minute, cfg.TrustedProxies, "global", log)(
					metrics.HTTPMetricsMiddleware(cfg.ServiceName)(
						middleware.RequestLogger(log)(
							middleware.DynamicCORS(cfg.CORSAllowedOrigins)(mux),
						),
					),
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
