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

	"eomp/packages/shared/pkg/logger"
	"eomp/packages/shared/pkg/metrics"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/ai/internal/config"
	"eomp/services/ai/internal/handler"
	"eomp/services/ai/internal/provider"
	"eomp/services/ai/internal/rag"
	"eomp/services/ai/internal/service"
)

func main() {
	cfg := config.Load()
	log := logger.InitLogger(cfg.ServiceName, cfg.Environment)

	if err := cfg.Validate(); err != nil {
		log.Error("ai configuration validation failed (fail-fast)", slog.Any("error", err))
		os.Exit(1)
	}

	// 1. Initialize AI Providers & Resilient RAG Retriever
	mockProvider := provider.NewMockProvider()
	smartRetriever := rag.NewSmartRetriever("localhost", 6333, "knowledge_base")

	// 2. Initialize Service and Handlers
	aiService := service.NewAIService(mockProvider, mockProvider, smartRetriever)
	aiHandler := handler.NewAIHandler(aiService)
	healthHandler := handler.NewHealthHandler(cfg)

	// 3. Routing
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler.Check)
	mux.HandleFunc("GET /api/health", healthHandler.Check)
	mux.HandleFunc("GET /metrics", metrics.PrometheusHandler())

	// Standard API v1 routes
	mux.HandleFunc("POST /api/v1/ai/chat", aiHandler.Chat)
	mux.HandleFunc("POST /api/v1/ai/analyze-ticket", aiHandler.AnalyzeTicket)

	// Legacy / Direct routes
	mux.HandleFunc("POST /api/ai/chat", aiHandler.Chat)
	mux.HandleFunc("POST /api/ai/analyze-ticket", aiHandler.AnalyzeTicket)

	// 4. Apply middleware stack with RED Metrics
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
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown channel
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Info("ai service starting",
			slog.Int("port", cfg.Port),
			slog.String("version", cfg.Version),
			slog.String("model", cfg.AIModel),
			slog.String("embedding_model", cfg.EmbeddingModel),
		)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server failed to start", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-shutdownChan
	log.Info("shutting down ai service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced shutdown", slog.Any("error", err))
		os.Exit(1)
	}

	log.Info("ai service stopped gracefully")
}
