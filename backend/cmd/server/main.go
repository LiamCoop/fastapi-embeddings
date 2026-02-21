package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"

	chunkcache "ragtime-backend/internal/chunking/cache"
	chunkhttp "ragtime-backend/internal/chunking/http"
	chunkrepo "ragtime-backend/internal/chunking/repository"
	chunkservice "ragtime-backend/internal/chunking/service"
	"ragtime-backend/internal/document"
	"ragtime-backend/internal/embedding"
	"ragtime-backend/internal/ingestion"
	"ragtime-backend/internal/logger"
	"ragtime-backend/internal/objectstore"
	retrievalcache "ragtime-backend/internal/retrieval/cache"
	retrievalhttp "ragtime-backend/internal/retrieval/http"
	retrievalrepo "ragtime-backend/internal/retrieval/repository"
	retrievalservice "ragtime-backend/internal/retrieval/service"
	"ragtime-backend/internal/storage"
)

func main() {
	// Parse command line flags
	port := flag.Int("p", 8080, "Port to run the server on")
	flag.Parse()

	dsn := requiredEnv("DATABASE_URL")
	db, err := storage.OpenDB(dsn)
	if err != nil {
		logger.Fatal("Database connection failed", "error", err)
	}

	store := mustObjectStoreClient()

	services := newServices(db, store)
	router := chi.NewRouter()
	chunkingHandler := chunkhttp.NewHandler(services.chunking)
	retrievalHandler := retrievalhttp.NewHandler(services.retrieval)
	router.Post("/v1/kb/{kbID}/documents/{documentID}/chunking", chunkingHandler.InitiateDocumentChunking)
	router.Post("/v1/kb/{kbID}/chunks/{chunkID}/embed", chunkingHandler.EmbedChunkByID)
	router.Post("/v1/kb/{kbID}/retrieve", retrievalHandler.Retrieve)
	go services.chunking.Run(context.Background())

	addr := fmt.Sprintf(":%d", *port)
	logger.Info("Starting server", "port", *port)

	// Start server
	if err := http.ListenAndServe(addr, router); err != nil {
		logger.Fatal("Server failed to start", "error", err)
	}
}

type appServices struct {
	documents      *document.Service
	embeddings     *embedding.Service
	ingestion      *ingestion.Service
	embeddingQueue chan embedding.EmbedChunkRequest
	chunking       *chunkservice.Service
	retrieval      *retrievalservice.Service
}

func newServices(db *sql.DB, store objectstore.Client) appServices {
	embedder, err := embedding.NewEmbedderFromEnv()
	if err != nil {
		logger.Fatal("Embedding client configuration failed", "error", err)
	}
	modelID := strings.TrimSpace(os.Getenv("EMBEDDING_MODEL_ID"))
	chunkingCh := make(chan chunkservice.DocumentRequest, 128)
	embeddingQueue := make(chan embedding.EmbedChunkRequest, 128)

	embedService := embedding.NewServiceWithPostgres(db, embedder, modelID, embeddingQueue)
	ingestionService := ingestion.NewServiceWithPostgres(db, store, embedService)
	go func() {
		if err := embedService.Run(context.Background()); err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("embedding worker stopped", "error", err)
		}
	}()

	chunkRepo := chunkrepo.NewPostgresStore(db)
	chunkCache := chunkcache.NewNoopLayer(chunkRepo)
	retrievalRepo := retrievalrepo.NewPostgresStore(db)
	retrievalCache := retrievalcache.NewNoopLayer(retrievalRepo)

	return appServices{
		documents:      document.NewServiceWithPostgres(db, store, chunkingCh),
		chunking:       chunkservice.New(chunkCache, nil, chunkingCh, store, embedService),
		embeddings:     embedService,
		ingestion:      ingestionService,
		embeddingQueue: embeddingQueue,
		retrieval:      retrievalservice.New(retrievalCache, embedder),
	}
}

func mustObjectStoreClient() objectstore.Client {
	storeType := strings.ToLower(strings.TrimSpace(os.Getenv("OBJECT_STORE_TYPE")))
	if storeType == "" {
		storeType = "s3"
	}

	switch storeType {
	case "local":
		root := strings.TrimSpace(os.Getenv("OBJECT_STORE_ROOT"))
		if root == "" {
			root = "/tmp/ragtime-objects"
		}
		return objectstore.NewLocalClient(root)
	case "s3":
		forcePathStyle := strings.ToLower(strings.TrimSpace(os.Getenv("S3_FORCE_PATH_STYLE"))) == "true"
		s3Client, err := objectstore.NewS3Client(context.Background(), objectstore.S3Config{
			Region:          requiredEnv("REGION"),
			Bucket:          requiredEnv("BUCKET_NAME"),
			Endpoint:        requiredEnv("ENDPOINT_URL"),
			AccessKeyID:     requiredEnv("ACCESS_KEY_ID"),
			SecretAccessKey: requiredEnv("SECRET_ACCESS_KEY"),
			ForcePathStyle:  forcePathStyle,
		})
		if err != nil {
			logger.Fatal("Object store connection failed", "error", err)
		}
		return s3Client
	default:
		logger.Fatal("Unsupported OBJECT_STORE_TYPE", "type", storeType)
		return nil
	}
}

func requiredEnv(key string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		logger.Fatal("Missing required environment variable", "key", key)
	}
	return value
}
