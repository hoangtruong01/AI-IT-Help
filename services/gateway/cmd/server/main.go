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
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/gateway/internal/config"
	"eomp/services/gateway/internal/handler"
	gwMiddleware "eomp/services/gateway/internal/middleware"
	"eomp/services/gateway/internal/proxy"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

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

	helpdeskProxy, err := proxy.NewReverseProxy(cfg.HelpdeskServiceURL, log)
	if err != nil {
		log.Error("failed to initialize helpdesk proxy", slog.Any("error", err))
		os.Exit(1)
	}

	aiProxy, err := proxy.NewReverseProxy(cfg.AIServiceURL, log)
	if err != nil {
		log.Error("failed to initialize ai proxy", slog.Any("error", err))
		os.Exit(1)
	}

	healthHandler := handler.NewHealthHandler(cfg)

	// 2. Setup Routing Multiplexer
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)

	// Auth Service Routing (Public routes for register, login, refresh; me requires token)
	mux.Handle("/api/v1/auth/", authProxy)

	// Employee & Department Routing
	mux.Handle("/api/v1/employees", authFilter(employeeProxy))
	mux.Handle("/api/v1/employees/", authFilter(employeeProxy))
	mux.Handle("/api/v1/departments", authFilter(employeeProxy))
	mux.Handle("/api/v1/departments/", authFilter(employeeProxy))

	// Helpdesk & Service Catalog Routing
	mux.Handle("/api/v1/tickets", authFilter(helpdeskProxy))
	mux.Handle("/api/v1/tickets/", authFilter(helpdeskProxy))
	mux.Handle("/api/v1/services/", authFilter(helpdeskProxy))

	// AI Operations Copilot Routing
	mux.Handle("/api/v1/ai/", authFilter(aiProxy))

	// Apply Global Gateway Middleware Stack
	handlerStack := middleware.Recoverer(log)(
		middleware.RequestLogger(log)(
			middleware.CORS(mux),
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
			slog.String("helpdesk_service", cfg.HelpdeskServiceURL),
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
