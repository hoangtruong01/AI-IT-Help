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
	"eomp/packages/shared/pkg/database"
	"eomp/packages/shared/pkg/logger"
	"eomp/packages/shared/pkg/metrics"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/auth/internal/config"
	"eomp/services/auth/internal/handler"
	"eomp/services/auth/internal/repository"
	"eomp/services/auth/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

	if err := cfg.Validate(); err != nil {
		log.Error("configuration validation failed (fail-fast)", slog.Any("error", err))
		os.Exit(1)
	}

	// 1. Initialize PostgreSQL Connection
	dbCfg := database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	db, err := database.Connect(dbCfg)
	if err != nil {
		log.Error("could not connect to PostgreSQL", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("connected to PostgreSQL successfully", slog.String("db", cfg.DBName))
	if err := database.RunMigrations(db, cfg.MigrationsPath); err != nil {
		log.Error("failed to run database migrations", slog.Any("error", err))
		os.Exit(1)
	}
	log.Info("database migrations applied successfully")

	// 2. Instantiate Dependencies
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	userRepo := repository.NewUserRepository(db)
	authSvc := service.NewAuthService(userRepo, jwtManager)
	if err := authSvc.BootstrapAdmin(context.Background(), cfg.BootstrapAdminEmail, cfg.BootstrapAdminPassword, cfg.BootstrapAdminName); err != nil {
		log.Error("failed to bootstrap initial administrator", slog.Any("error", err))
		os.Exit(1)
	}
	authHandler := handler.NewAuthHandler(authSvc)
	healthHandler := handler.NewHealthHandler(cfg)

	// 3. Register HTTP Routes
	mux := http.NewServeMux()

	// Health & Metrics
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)
	mux.HandleFunc("GET /ready", database.ReadinessHandler(db))
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	// Public auth routes
	mux.HandleFunc("POST /api/v1/auth/register", authHandler.Register)
	mux.HandleFunc("POST /api/v1/auth/login", authHandler.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", authHandler.RefreshToken)
	mux.HandleFunc("POST /api/v1/auth/logout", authHandler.Logout)

	// Protected routes — always require Bearer token validation.
	authMiddleware := middleware.Authenticate(jwtManager)
	adminOnly := middleware.RequireRole("ROLE_ADMIN")
	adminOrManager := middleware.RequireRole("ROLE_ADMIN", "ROLE_MANAGER")

	mux.Handle("GET /api/v1/auth/me", authMiddleware(http.HandlerFunc(authHandler.GetMe)))
	mux.Handle("GET /api/v1/auth/login-history", authMiddleware(http.HandlerFunc(authHandler.GetLoginHistory)))
	mux.Handle("POST /api/v1/auth/change-password", authMiddleware(http.HandlerFunc(authHandler.ChangePassword)))
	mux.Handle("POST /api/v1/auth/reset-password/{id}", authMiddleware(adminOnly(http.HandlerFunc(authHandler.AdminResetPassword))))

	// User provisioning and administration (Gate B-02)
	mux.Handle("GET /api/v1/users", authMiddleware(adminOrManager(http.HandlerFunc(authHandler.ListUsers))))
	mux.Handle("POST /api/v1/users", authMiddleware(adminOnly(http.HandlerFunc(authHandler.CreateUser))))
	mux.Handle("PATCH /api/v1/users/{id}", authMiddleware(adminOnly(http.HandlerFunc(authHandler.UpdateUser))))

	// Apply global middleware stack with RED Metrics
	handlerStack := middleware.Recoverer(log)(
		metrics.HTTPMetricsMiddleware(cfg.ServiceName)(
			middleware.RequestLogger(log)(
				middleware.CORS(mux),
			),
		),
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      handlerStack,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("server starting",
			slog.Int("port", cfg.Port),
			slog.String("version", cfg.Version),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-shutdownChan
	log.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	if db != nil {
		_ = db.Close()
	}

	log.Info("server stopped gracefully")
}
