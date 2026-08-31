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
	"eomp/services/reporting/internal/config"
	"eomp/services/reporting/internal/handler"
	"eomp/services/reporting/internal/repository"
	"eomp/services/reporting/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

	if err := cfg.Validate(); err != nil {
		log.Error("reporting configuration validation failed (fail-fast)", slog.Any("error", err))
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
	repo := repository.NewRepository(db)
	slaAggregator := service.NewSLAAggregator(db, log, 1*time.Hour)
	bus := eventbus.NewResilientEventBus(cfg.RabbitMQURL, cfg.ServiceName)
	for _, topic := range []string{eventbus.TopicTicketCreated, eventbus.TopicTicketStatusChanged, eventbus.TopicTicketAssigned} {
		if err := bus.Subscribe(topic, func(ctx context.Context, event eventbus.Event) error {
			if err := repo.ProjectTicketEvent(ctx, event); err != nil {
				return err
			}
			slaAggregator.RollupOnce(ctx)
			return nil
		}); err != nil {
			log.Error("failed to subscribe reporting ticket projector", slog.String("topic", topic), slog.Any("error", err))
			os.Exit(1)
		}
	}
	svc := service.NewService(repo)
	reportingHandler := handler.NewReportingHandler(svc)
	healthHandler := handler.NewHealthHandler(cfg)

	// 2.1 Start Background SLA Rollup Aggregator Worker (Phase 6)
	slaAggregator.Start()
	defer slaAggregator.Stop()

	// 3. Register HTTP Routes
	mux := http.NewServeMux()

	// Health & Prometheus Metrics
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)
	mux.HandleFunc("GET /ready", database.ReadinessHandler(db))
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	// Reporting & BI Endpoints
	mux.HandleFunc("GET /api/v1/reports/overview", reportingHandler.GetOverview)
	mux.HandleFunc("GET /api/v1/reports/trends", reportingHandler.GetTrends)
	mux.HandleFunc("GET /api/v1/reports/categories", reportingHandler.GetCategories)
	mux.HandleFunc("GET /api/v1/reports/departments-sla", reportingHandler.GetDepartmentsSLA)
	mux.HandleFunc("GET /api/v1/reports/agents", reportingHandler.GetAgents)
	mux.HandleFunc("POST /api/v1/reports/export", reportingHandler.ExportReport)

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
		log.Info("reporting service server starting",
			slog.Int("port", cfg.Port),
			slog.String("version", cfg.Version),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-shutdownChan
	log.Info("shutting down reporting server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	if db != nil {
		_ = db.Close()
	}

	log.Info("reporting server stopped gracefully")
}
