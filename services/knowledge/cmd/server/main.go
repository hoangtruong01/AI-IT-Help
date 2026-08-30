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
	"eomp/packages/shared/pkg/metrics"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/knowledge/internal/config"
	"eomp/services/knowledge/internal/handler"
	"eomp/services/knowledge/internal/repository"
	"eomp/services/knowledge/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

	if err := cfg.Validate(); err != nil {
		log.Error("knowledge configuration validation failed (fail-fast)", slog.Any("error", err))
		os.Exit(1)
	}

	// 1. PostgreSQL Connection & Auto-migrations
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
	log.Info("knowledge database migrations applied successfully")

	// 2. Dependencies
	repo := repository.NewRepository(db)
	knowledgeSvc := service.NewKnowledgeService(repo)
	knowledgeHandler := handler.NewKnowledgeHandler(knowledgeSvc)
	healthHandler := handler.NewHealthHandler(cfg)

	// 3. Routing
	mux := http.NewServeMux()

	// Health & Metrics
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)
	mux.HandleFunc("GET /ready", database.ReadinessHandler(db))
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	// Stats & Search
	mux.HandleFunc("GET /api/v1/knowledge/stats", knowledgeHandler.GetStats)
	mux.HandleFunc("GET /api/v1/knowledge/search", knowledgeHandler.Search)

	// Categories
	mux.HandleFunc("GET /api/v1/knowledge/categories", knowledgeHandler.ListCategories)
	mux.HandleFunc("POST /api/v1/knowledge/categories", knowledgeHandler.CreateCategory)

	// Articles
	mux.HandleFunc("GET /api/v1/knowledge/articles", knowledgeHandler.ListArticles)
	mux.HandleFunc("POST /api/v1/knowledge/articles", knowledgeHandler.CreateArticle)
	mux.HandleFunc("GET /api/v1/knowledge/articles/{id}", knowledgeHandler.GetArticle)
	mux.HandleFunc("PUT /api/v1/knowledge/articles/{id}", knowledgeHandler.UpdateArticle)
	mux.HandleFunc("DELETE /api/v1/knowledge/articles/{id}", knowledgeHandler.DeleteArticle)

	// Runbooks (SOP)
	mux.HandleFunc("GET /api/v1/knowledge/runbooks", knowledgeHandler.ListRunbooks)
	mux.HandleFunc("POST /api/v1/knowledge/runbooks", knowledgeHandler.CreateRunbook)
	mux.HandleFunc("GET /api/v1/knowledge/runbooks/{id}", knowledgeHandler.GetRunbook)

	// Apply middleware stack with RED Metrics
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

	// Graceful shutdown channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("knowledge service starting",
			slog.Int("port", cfg.Port),
			slog.String("version", cfg.Version),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-shutdownChan
	log.Info("shutting down knowledge service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	if db != nil {
		_ = db.Close()
	}

	log.Info("knowledge service stopped gracefully")
}
