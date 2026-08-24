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
	"eomp/services/notification/internal/config"
	"eomp/services/notification/internal/handler"
	"eomp/services/notification/internal/repository"
	"eomp/services/notification/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

	if err := cfg.Validate(); err != nil {
		log.Error("notification configuration validation failed (fail-fast)", slog.Any("error", err))
		os.Exit(1)
	}

	// 1. PostgreSQL Connection & Auto Migrations
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
		log.Warn("could not connect to PostgreSQL (notification_db) - will retry on requests", slog.Any("error", err))
	} else {
		log.Info("connected to PostgreSQL successfully", slog.String("db", cfg.DBName))
		if err := database.RunMigrations(db, cfg.MigrationsPath); err != nil {
			log.Error("failed to run database migrations", slog.Any("error", err))
		} else {
			log.Info("database migrations applied successfully")
		}
	}

	// 2. EventBus & Subscriber
	bus := eventbus.NewMemoryEventBus()
	repo := repository.NewRepository(db)
	notificationSvc := service.NewNotificationService(repo)

	// Subscribe Notification Service to all domain events
	_ = bus.Subscribe("*", func(ctx context.Context, event eventbus.Event) error {
		log.Info("event received via EventBus", slog.String("type", event.Type), slog.String("source", event.Source))
		return notificationSvc.HandleDomainEvent(ctx, event)
	})

	notificationHandler := handler.NewNotificationHandler(notificationSvc)
	healthHandler := handler.NewHealthHandler(cfg)

	// 3. Routes
	mux := http.NewServeMux()

	// Health & Metrics
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	// Notifications API
	mux.HandleFunc("GET /api/v1/notifications/stats", notificationHandler.GetStats)
	mux.HandleFunc("GET /api/v1/notifications", notificationHandler.ListNotifications)
	mux.HandleFunc("POST /api/v1/notifications", notificationHandler.SendNotification)
	mux.HandleFunc("PATCH /api/v1/notifications/{id}/read", notificationHandler.MarkAsRead)
	mux.HandleFunc("POST /api/v1/notifications/read-all", notificationHandler.MarkAllAsRead)

	// Middleware Stack with RED Metrics
	handlerStack := middleware.Recoverer(log)(
		metrics.HTTPMetricsMiddleware(cfg.ServiceName)(
			middleware.RequestLogger(log)(
				middleware.ExtractGatewayHeaders()(
					middleware.CORS(mux),
				),
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
