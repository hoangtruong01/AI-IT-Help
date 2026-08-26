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
	"eomp/services/asset/internal/config"
	"eomp/services/asset/internal/handler"
	"eomp/services/asset/internal/repository"
	"eomp/services/asset/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

	if err := cfg.Validate(); err != nil {
		log.Error("asset configuration validation failed (fail-fast)", slog.Any("error", err))
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
		log.Warn("could not connect to PostgreSQL (asset_db) - will retry on requests", slog.Any("error", err))
	} else {
		log.Info("connected to PostgreSQL successfully", slog.String("db", cfg.DBName))
		if err := database.RunMigrations(db, cfg.MigrationsPath); err != nil {
			log.Error("failed to run database migrations", slog.Any("error", err))
		} else {
			log.Info("database migrations applied successfully")
		}
	}

	// 2. Dependencies & EventBus
	bus := eventbus.NewResilientEventBus(cfg.RabbitMQURL, cfg.ServiceName)
	repo := repository.NewRepository(db)
	assetSvc := service.NewAssetService(repo, cfg.HelpdeskServiceURL, bus)
	cmdbSvc := service.NewCMDBService(repo)

	assetHandler := handler.NewAssetHandler(assetSvc)
	cmdbHandler := handler.NewCMDBHandler(cmdbSvc)
	healthHandler := handler.NewHealthHandler(cfg)

	// 3. Routes
	mux := http.NewServeMux()

	// Health & Metrics
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	// Asset APIs
	mux.HandleFunc("GET /api/v1/assets/stats", assetHandler.GetStats)
	mux.HandleFunc("GET /api/v1/assets", assetHandler.ListAssets)
	mux.HandleFunc("POST /api/v1/assets", assetHandler.CreateAsset)
	mux.HandleFunc("GET /api/v1/assets/employee/{id}/history", assetHandler.GetEmployeeAssetHistory)
	mux.HandleFunc("GET /api/v1/assets/{id}", assetHandler.GetAsset)
	mux.HandleFunc("PATCH /api/v1/assets/{id}/status", assetHandler.UpdateStatus)
	mux.HandleFunc("POST /api/v1/assets/{id}/assign", assetHandler.AssignAsset)
	mux.HandleFunc("POST /api/v1/assets/{id}/return", assetHandler.ReturnAsset)
	mux.HandleFunc("GET /api/v1/assets/{id}/assignments", assetHandler.ListAssignments)
	mux.HandleFunc("GET /api/v1/assets/{id}/incidents", assetHandler.GetAssetIncidents)


	// CMDB APIs
	mux.HandleFunc("GET /api/v1/cmdb/topology", cmdbHandler.GetTopology)
	mux.HandleFunc("GET /api/v1/cmdb/ci", cmdbHandler.ListCIs)
	mux.HandleFunc("POST /api/v1/cmdb/ci", cmdbHandler.CreateCI)
	mux.HandleFunc("GET /api/v1/cmdb/ci/{id}", cmdbHandler.GetCI)
	mux.HandleFunc("PATCH /api/v1/cmdb/ci/{id}/status", cmdbHandler.UpdateCIStatus)
	mux.HandleFunc("GET /api/v1/cmdb/relationships", cmdbHandler.ListRelationships)
	mux.HandleFunc("POST /api/v1/cmdb/relationships", cmdbHandler.CreateRelationship)

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
