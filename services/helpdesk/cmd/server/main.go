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
	"eomp/services/helpdesk/internal/config"
	"eomp/services/helpdesk/internal/handler"
	"eomp/services/helpdesk/internal/repository"
	"eomp/services/helpdesk/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

	if err := cfg.Validate(); err != nil {
		log.Error("helpdesk configuration validation failed (fail-fast)", slog.Any("error", err))
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
		log.Warn("could not connect to PostgreSQL (helpdesk_db) - will retry on requests", slog.Any("error", err))
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
	slaEngine := service.NewSLAEngine()
	repo := repository.NewRepository(db)
	problemRepo := repository.NewProblemRepository(db)
	ticketSvc := service.NewTicketService(repo, slaEngine, bus)
	problemSvc := service.NewProblemService(problemRepo, repo)
	ticketHandler := handler.NewTicketHandler(ticketSvc)
	problemHandler := handler.NewProblemHandler(problemSvc)
	healthHandler := handler.NewHealthHandler(cfg)

	// 3. Routes
	mux := http.NewServeMux()

	// Health & Metrics
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	// Tickets API
	mux.HandleFunc("GET /api/v1/tickets", ticketHandler.ListTickets)
	mux.HandleFunc("POST /api/v1/tickets", ticketHandler.CreateTicket)
	mux.HandleFunc("GET /api/v1/tickets/asset/{assetId}", ticketHandler.ListTicketsByAsset)
	mux.HandleFunc("GET /api/v1/tickets/{id}", ticketHandler.GetTicket)
	mux.HandleFunc("PATCH /api/v1/tickets/{id}/status", ticketHandler.UpdateStatus)
	mux.HandleFunc("PATCH /api/v1/tickets/{id}/assign", ticketHandler.AssignTicket)
	mux.HandleFunc("POST /api/v1/tickets/{id}/comments", ticketHandler.AddComment)
	mux.HandleFunc("GET /api/v1/tickets/{id}/comments", ticketHandler.ListComments)
	mux.HandleFunc("GET /api/v1/tickets/{id}/timeline", ticketHandler.ListTimeline)


	// Problem Management API (ITIL v4)
	mux.HandleFunc("GET /api/v1/problems/stats", problemHandler.GetStats)
	mux.HandleFunc("GET /api/v1/problems", problemHandler.ListProblems)
	mux.HandleFunc("POST /api/v1/problems", problemHandler.CreateProblem)
	mux.HandleFunc("GET /api/v1/problems/{id}", problemHandler.GetProblem)
	mux.HandleFunc("PATCH /api/v1/problems/{id}/status", problemHandler.UpdateStatus)
	mux.HandleFunc("PATCH /api/v1/problems/{id}/rca", problemHandler.UpdateRCA)
	mux.HandleFunc("POST /api/v1/problems/{id}/link-incident", problemHandler.LinkIncident)
	mux.HandleFunc("DELETE /api/v1/problems/{id}/unlink-incident/{ticketId}", problemHandler.UnlinkIncident)

	// Service Catalog API
	mux.HandleFunc("GET /api/v1/services/categories", ticketHandler.ListServiceCategories)
	mux.HandleFunc("GET /api/v1/services/items", ticketHandler.ListServiceCatalogItems)

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
