package embedding

import (
	"context"
	"database/sql"
	"errors"

	"ragtime-backend/internal/logger"
)

var (
	ErrNilEmbedder     = errors.New("embedding client is required")
	ErrNilRepository   = errors.New("repository is required")
	ErrMissingModelID  = errors.New("model id is required")
	ErrNilInputChannel = errors.New("embedding input channel is required")
	ErrNoChunkResult   = errors.New("no embedding result returned for chunk")
)

// TextEmbedder defines the external embedding client contract.
type TextEmbedder interface {
	EmbedTexts(ctx context.Context, texts []string) ([][]float32, int, error)
}

type Service struct {
	embedder       TextEmbedder
	repo           Repository
	defaultModelID string
	inputCh        chan EmbedChunkRequest
}

func NewService(
	embedder TextEmbedder,
	repo Repository,
	defaultModelID string,
	inputCh chan EmbedChunkRequest,
) *Service {
	return &Service{
		embedder:       embedder,
		repo:           repo,
		defaultModelID: defaultModelID,
		inputCh:        inputCh,
	}
}

func NewServiceWithPostgres(
	db *sql.DB,
	embedder TextEmbedder,
	defaultModelID string,
	inputCh chan EmbedChunkRequest,
) *Service {
	repo := NewPostgresRepository(db)
	return NewService(embedder, repo, defaultModelID, inputCh)
}

// Run consumes chunk embedding requests from the configured channel.
func (s *Service) Run(ctx context.Context) error {
	if s.inputCh == nil {
		return ErrNilInputChannel
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case req, ok := <-s.inputCh:
			if !ok {
				return nil
			}
			logger.Info(
				"embedding job started",
				"knowledge_base_id", req.KnowledgeBaseID,
				"chunk_id", req.Chunk.ChunkID,
			)
			result, err := s.embedChunk(ctx, req)
			if req.ResultCh != nil {
				outcome := EmbedChunkResult{Err: err}
				if err == nil {
					outcome.Result = &result
				}
				select {
				case req.ResultCh <- outcome:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			if err != nil {
				logger.Error(
					"embedding job failed",
					"knowledge_base_id", req.KnowledgeBaseID,
					"chunk_id", req.Chunk.ChunkID,
					"error", err,
				)
				continue
			}
			logger.Info(
				"embedding job completed",
				"knowledge_base_id", req.KnowledgeBaseID,
				"chunk_id", req.Chunk.ChunkID,
				"embedding_id", result.EmbeddingID,
				"model_id", result.ModelID,
				"vector_dimension", result.VectorDimension,
				"reused", len(result.Vector) == 0,
			)
		}
	}
}

func (s *Service) embedChunk(ctx context.Context, req EmbedChunkRequest) (EmbeddingResult, error) {
	results, err := s.EmbedAndStore(ctx, EmbedChunksRequest{
		KnowledgeBaseID: req.KnowledgeBaseID,
		Chunks:          []ChunkInput{req.Chunk},
		ModelID:         req.ModelID,
	})
	if err != nil {
		return EmbeddingResult{}, err
	}
	for _, result := range results {
		if result.ChunkID == req.Chunk.ChunkID {
			return result, nil
		}
	}
	if len(results) > 0 {
		return results[0], nil
	}
	return EmbeddingResult{}, ErrNoChunkResult
}

// EnqueueChunkAndWait sends one chunk to the embedding worker and waits for completion.
func (s *Service) EnqueueChunkAndWait(ctx context.Context, req EmbedChunkRequest) (EmbeddingResult, error) {
	if s.inputCh == nil {
		return s.embedChunk(ctx, req)
	}

	resultCh := make(chan EmbedChunkResult, 1)
	req.ResultCh = resultCh

	select {
	case s.inputCh <- req:
	case <-ctx.Done():
		return EmbeddingResult{}, ctx.Err()
	}

	select {
	case outcome := <-resultCh:
		if outcome.Err != nil {
			return EmbeddingResult{}, outcome.Err
		}
		if outcome.Result == nil {
			return EmbeddingResult{}, ErrNoChunkResult
		}
		return *outcome.Result, nil
	case <-ctx.Done():
		return EmbeddingResult{}, ctx.Err()
	}
}

// EmbedAndStore validates, deduplicates, embeds, and persists embeddings.
func (s *Service) EmbedAndStore(ctx context.Context, req EmbedChunksRequest) ([]EmbeddingResult, error) {
	if s.embedder == nil {
		return nil, ErrNilEmbedder
	}
	if s.repo == nil {
		return nil, ErrNilRepository
	}

	if err := ValidateEmbedChunksRequest(req); err != nil {
		return nil, err
	}

	modelID := s.defaultModelID
	if req.ModelID != nil && *req.ModelID != "" {
		modelID = *req.ModelID
	}
	if modelID == "" {
		return nil, ErrMissingModelID
	}
	logger.Info(
		"embedding request received",
		"knowledge_base_id", req.KnowledgeBaseID,
		"chunk_count", len(req.Chunks),
		"model_id", modelID,
	)

	filtered := make([]ChunkInput, 0, len(req.Chunks))
	results := make([]EmbeddingResult, 0, len(req.Chunks))
	seen := make(map[string]struct{}, len(req.Chunks))
	reusedCount := 0
	duplicateCount := 0
	for _, chunk := range req.Chunks {
		if _, ok := seen[chunk.ContentHash]; ok {
			duplicateCount++
			continue
		}
		seen[chunk.ContentHash] = struct{}{}

		existingID, exists, err := s.repo.FindEmbeddingID(ctx, req.KnowledgeBaseID, chunk.ContentHash, modelID)
		if err != nil {
			return nil, err
		}
		if exists {
			reusedCount++
			results = append(results, EmbeddingResult{
				EmbeddingID:     existingID,
				ChunkID:         chunk.ChunkID,
				KnowledgeBaseID: req.KnowledgeBaseID,
				ContentHash:     chunk.ContentHash,
				ModelID:         modelID,
			})
			continue
		}

		filtered = append(filtered, chunk)
	}

	if len(filtered) == 0 {
		logger.Info(
			"embedding request resolved from existing vectors",
			"knowledge_base_id", req.KnowledgeBaseID,
			"model_id", modelID,
			"requested_chunks", len(req.Chunks),
			"reused_embeddings", reusedCount,
			"deduped_in_request", duplicateCount,
		)
		return results, nil
	}
	logger.Info(
		"embedding generation started",
		"knowledge_base_id", req.KnowledgeBaseID,
		"model_id", modelID,
		"new_embeddings", len(filtered),
		"reused_embeddings", reusedCount,
		"deduped_in_request", duplicateCount,
	)

	texts := make([]string, 0, len(filtered))
	for _, chunk := range filtered {
		texts = append(texts, chunk.Content)
	}

	vectors, dim, err := s.embedder.EmbedTexts(ctx, texts)
	if err != nil {
		return nil, err
	}

	newResults := make([]EmbeddingResult, 0, len(filtered))
	for i, chunk := range filtered {
		newResults = append(newResults, EmbeddingResult{
			ChunkID:         chunk.ChunkID,
			KnowledgeBaseID: req.KnowledgeBaseID,
			ContentHash:     chunk.ContentHash,
			ModelID:         modelID,
			Vector:          vectors[i],
			VectorDimension: dim,
		})
	}

	stored, err := s.repo.SaveEmbeddings(ctx, newResults)
	if err != nil {
		return nil, err
	}
	if stored != nil {
		results = append(results, stored...)
		logger.Info(
			"embeddings persisted",
			"knowledge_base_id", req.KnowledgeBaseID,
			"model_id", modelID,
			"created_embeddings", len(stored),
			"vector_dimension", dim,
			"reused_embeddings", reusedCount,
			"deduped_in_request", duplicateCount,
			"total_results", len(results),
		)
		return results, nil
	}

	results = append(results, newResults...)
	logger.Info(
		"embedding request completed without repository return payload",
		"knowledge_base_id", req.KnowledgeBaseID,
		"model_id", modelID,
		"created_embeddings", len(newResults),
		"vector_dimension", dim,
		"reused_embeddings", reusedCount,
		"deduped_in_request", duplicateCount,
		"total_results", len(results),
	)
	return results, nil
}
