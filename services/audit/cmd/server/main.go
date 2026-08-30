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

	"eomp/packages/shared/pkg/database"
	"eomp/packages/shared/pkg/eventbus"
	"eomp/packages/shared/pkg/logger"
	"eomp/packages/shared/pkg/metrics"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/audit/internal/config"
	"eomp/services/audit/internal/handler"
	"eomp/services/audit/internal/repository"
	"eomp/services/audit/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

	if err := cfg.Validate(); err != nil {
		log.Error("audit configuration validation failed (fail-fast)", slog.Any("error", err))
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

	// 2. Instantiate Dependencies & EventBus Consumer
	bus := eventbus.NewResilientEventBus(cfg.RabbitMQURL, cfg.ServiceName)
	repo := repository.NewRepository(db)
	svc := service.NewService(repo)

	// Subscribe Audit Service to all domain events (#) for tamper-evident cryptographic log sealing
	_ = bus.Subscribe("*", func(ctx context.Context, event eventbus.Event) error {
		log.Info("audit service received domain event",
			slog.String("event_type", event.Type),
			slog.String("source", event.Source),
			slog.String("id", event.ID),
		)
		return svc.IngestDomainEvent(ctx, event)
	})

	auditHandler := handler.NewAuditHandler(svc)
	healthHandler := handler.NewHealthHandler(cfg)

	// 3. Register HTTP Routes
	mux := http.NewServeMux()

	// Health & Prometheus Metrics
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)
	mux.HandleFunc("GET /ready", database.ReadinessHandler(db))
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	// Audit Trail Endpoints
	mux.HandleFunc("GET /api/v1/audit/logs", auditHandler.ListAuditLogs)
	mux.HandleFunc("GET /api/v1/audit/logs/{id}", auditHandler.GetAuditLogByID)
	mux.HandleFunc("POST /api/v1/audit/logs", auditHandler.CreateAuditLog)
	mux.HandleFunc("GET /api/v1/audit/stats", auditHandler.GetStats)
	mux.HandleFunc("GET /api/v1/audit/security-events", auditHandler.GetSecurityEvents)

	// Protected routes gateway header extraction
	gatewayMiddleware := middleware.ExtractGatewayHeaders()

	// Apply middleware stack with RED Metrics
	handlerStack := middleware.Recoverer(log)(
		metrics.HTTPMetricsMiddleware(cfg.ServiceName)(
			middleware.RequestLogger(log)(
				gatewayMiddleware(
					middleware.CORS(mux),
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

	// Graceful shutdown channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("audit service server starting",
			slog.Int("port", cfg.Port),
			slog.String("version", cfg.Version),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-shutdownChan
	log.Info("shutting down audit server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	if db != nil {
		_ = db.Close()
	}

	log.Info("audit server stopped gracefully")
}
