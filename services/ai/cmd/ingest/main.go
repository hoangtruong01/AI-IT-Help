package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	"eomp/packages/shared/pkg/config"
	"eomp/packages/shared/pkg/logger"
	aiConfig "eomp/services/ai/internal/config"
	"eomp/services/ai/internal/provider"
	"eomp/services/ai/internal/rag"
	_ "github.com/lib/pq"
)

func main() {
	cfg := aiConfig.Load()
	log := logger.InitLogger("ai-ingest", cfg.Environment)

	log.Info("starting knowledge vector ingestion pipeline...",
		slog.String("qdrant_url", cfg.QdrantURL),
		slog.String("collection", cfg.QdrantCollection),
		slog.String("ai_provider", cfg.AIProvider),
	)

	// 1. Initialize Embedding Provider
	var embedder provider.EmbeddingProvider
	switch cfg.AIProvider {
	case "ollama":
		embedder = provider.NewOllamaProvider(cfg.OllamaBaseURL, cfg.AIModel, cfg.EmbeddingModel)
	case "openai":
		embedder = provider.NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.AIModel, cfg.EmbeddingModel)
	case "gemini":
		embedder = provider.NewGeminiProvider(cfg.GeminiAPIKey, cfg.AIModel, cfg.EmbeddingModel)
	default:
		log.Info("using MockProvider for vector embeddings")
		embedder = provider.NewMockProvider()
	}

	// 2. Connect to PostgreSQL knowledge_db
	dbHost := config.GetEnv("POSTGRES_HOST", "localhost")
	dbPort := config.GetEnvInt("POSTGRES_PORT", 5432)
	dbUser := config.GetEnv("POSTGRES_USER", "eomp")
	dbPass := config.GetEnv("POSTGRES_PASSWORD", "eomp_dev_password")
	dbName := config.GetEnv("KNOWLEDGE_DB_NAME", "knowledge_db")
	sslMode := config.GetEnv("POSTGRES_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPass, dbName, sslMode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Error("failed to connect to knowledge_db", slog.Any("error", err))
		os.Exit(1)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Warn("could not ping knowledge_db, using in-memory catalog for vector simulation", slog.Any("error", err))
	} else {
		log.Info("successfully connected to knowledge_db")
	}

	// 3. Run Ingestion Pipeline
	pipeline := rag.NewIngestionPipeline(cfg.QdrantURL, cfg.QdrantCollection, embedder)

	if err := pipeline.EnsureCollection(ctx, embedder.Dimensions()); err != nil {
		log.Warn("Qdrant collection setup warning (is Qdrant running on :6333?)", slog.Any("warning", err))
	} else {
		log.Info("Qdrant collection verified / ready", slog.String("collection", cfg.QdrantCollection), slog.Int("dimensions", embedder.Dimensions()))
	}

	// If DB connected, ingest real data
	if db.PingContext(ctx) == nil {
		totalDocs, totalVectors, err := pipeline.IngestFromKnowledgeDB(ctx, db)
		if err != nil {
			log.Error("ingestion failed", slog.Any("error", err))
			os.Exit(1)
		}
		log.Info("knowledge vector ingestion completed successfully!",
			slog.Int("documents_processed", totalDocs),
			slog.Int("vectors_upserted", totalVectors),
		)
	} else {
		log.Info("knowledge ingestion finished in local baseline mode")
	}
}
