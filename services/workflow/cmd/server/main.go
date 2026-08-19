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
	"eomp/packages/shared/pkg/logger"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/workflow/internal/config"
	"eomp/services/workflow/internal/handler"
	"eomp/services/workflow/internal/repository"
	"eomp/services/workflow/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

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
		log.Warn("could not connect to PostgreSQL (workflow_db) - will retry on requests", slog.Any("error", err))
	} else {
		log.Info("connected to PostgreSQL successfully", slog.String("db", cfg.DBName))
		if err := database.RunMigrations(db, cfg.MigrationsPath); err != nil {
			log.Error("failed to run database migrations", slog.Any("error", err))
		} else {
			log.Info("database migrations applied successfully")
		}
	}

	// 2. Dependencies
	repo := repository.NewRepository(db)
	changeRepo := repository.NewChangeRepository(db)
	workflowSvc := service.NewWorkflowService(repo)
	changeSvc := service.NewChangeService(changeRepo)

	workflowHandler := handler.NewWorkflowHandler(workflowSvc)
	approvalHandler := handler.NewApprovalHandler(workflowSvc)
	changeHandler := handler.NewChangeHandler(changeSvc)
	healthHandler := handler.NewHealthHandler(cfg)

	// 3. Routes
	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)

	// Workflow APIs
	mux.HandleFunc("GET /api/v1/workflows/stats", workflowHandler.GetStats)
	mux.HandleFunc("GET /api/v1/workflows/definitions", workflowHandler.ListDefinitions)
	mux.HandleFunc("GET /api/v1/workflows/definitions/{id}", workflowHandler.GetDefinition)
	mux.HandleFunc("GET /api/v1/workflows/instances", workflowHandler.ListInstances)
	mux.HandleFunc("POST /api/v1/workflows/instances", workflowHandler.StartWorkflow)
	mux.HandleFunc("GET /api/v1/workflows/instances/{id}", workflowHandler.GetInstance)
	mux.HandleFunc("GET /api/v1/workflows/instances/{id}/logs", workflowHandler.ListLogs)

	// Approval APIs
	mux.HandleFunc("GET /api/v1/approvals", approvalHandler.ListApprovals)
	mux.HandleFunc("POST /api/v1/approvals/{id}/decision", approvalHandler.ProcessDecision)

	// Change Management & CAB APIs (ITIL v4)
	mux.HandleFunc("GET /api/v1/changes/stats", changeHandler.GetStats)
	mux.HandleFunc("GET /api/v1/changes/calendar", changeHandler.GetCalendar)
	mux.HandleFunc("GET /api/v1/changes", changeHandler.ListChanges)
	mux.HandleFunc("POST /api/v1/changes", changeHandler.CreateChange)
	mux.HandleFunc("GET /api/v1/changes/{id}", changeHandler.GetChange)
	mux.HandleFunc("PATCH /api/v1/changes/{id}/status", changeHandler.UpdateStatus)
	mux.HandleFunc("POST /api/v1/changes/{id}/cab-vote", changeHandler.SubmitCABVote)

	// Middleware Stack
	handlerStack := middleware.Recoverer(log)(
		middleware.RequestLogger(log)(
			middleware.ExtractGatewayHeaders()(
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
