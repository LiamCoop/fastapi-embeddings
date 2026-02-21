package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"ragtime-backend/internal/chunking"
	"ragtime-backend/internal/chunking/cache"
	"ragtime-backend/internal/domain"
	"ragtime-backend/internal/embedding"
	"ragtime-backend/internal/logger"
	"ragtime-backend/internal/objectstore"
)

const (
	DefaultChunkingStrategy = string(chunking.StrategyFixed)
	DefaultMaxRunes         = 1000
	DefaultOverlapRunes     = 100
)

var (
	ErrDocumentNotFound    = errors.New("document not found")
	ErrChunkNotFound       = errors.New("chunk not found")
	ErrEmbedderUnavailable = errors.New("embedder unavailable")
)

type DocumentRequest struct {
	KnowledgeBaseID   string
	DocumentID        string
	DocumentVersionID string
	Content           string
	Strategy          chunking.Strategy
	MaxRunes          int
	OverlapRunes      int
	Separators        []string
	LanguageHints     []chunking.Language
}

type InitiateRequest struct {
	KnowledgeBaseID string
	DocumentID      string
	Strategy        chunking.Strategy
	MaxRunes        int
	OverlapRunes    int
	Separators      []string
	LanguageHints   []chunking.Language
}

type InitiateResult struct {
	DocumentID        string `json:"document_id"`
	DocumentVersionID string `json:"document_version_id"`
	Strategy          string `json:"strategy"`
	ChunkCount        int    `json:"chunk_count"`
}

type Service struct {
	cache    cache.Layer
	chunker  chunking.Chunker
	input    <-chan DocumentRequest
	store    objectstore.Client
	embedder *embedding.Service
	now      func() time.Time
	strategy string
}

func New(
	cacheLayer cache.Layer,
	chunker chunking.Chunker,
	input <-chan DocumentRequest,
	store objectstore.Client,
	embedder *embedding.Service,
) *Service {
	if chunker == nil {
		chunker = chunking.FixedSizeChunker{MaxRunes: DefaultMaxRunes, OverlapRunes: DefaultOverlapRunes}
	}
	return &Service{
		cache:    cacheLayer,
		chunker:  chunker,
		input:    input,
		store:    store,
		embedder: embedder,
		now:      func() time.Time { return time.Now().UTC() },
		strategy: DefaultChunkingStrategy,
	}
}

func (s *Service) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req, ok := <-s.input:
			if !ok {
				return
			}
			if _, err := s.handle(ctx, req); err != nil {
				logger.Error("chunking failed", "error", err, "document_id", req.DocumentID, "version_id", req.DocumentVersionID)
			}
		}
	}
}

func (s *Service) InitiateDocumentChunking(ctx context.Context, req InitiateRequest) (*InitiateResult, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("cache layer is required")
	}
	if s.store == nil {
		return nil, fmt.Errorf("object store is required")
	}
	if strings.TrimSpace(req.KnowledgeBaseID) == "" {
		return nil, fmt.Errorf("knowledgebase_id is required")
	}
	if strings.TrimSpace(req.DocumentID) == "" {
		return nil, fmt.Errorf("document_id is required")
	}

	versionRef, err := s.cache.GetLatestDocumentVersionForDocument(ctx, req.KnowledgeBaseID, req.DocumentID)
	if err != nil {
		return nil, err
	}
	if versionRef == nil {
		return nil, ErrDocumentNotFound
	}

	if err := s.cache.DeleteChunksByDocument(ctx, req.KnowledgeBaseID, req.DocumentID); err != nil {
		errMsg := fmt.Sprintf(
			"stage=CHUNKED document_id=%s version_id=%s error=%v",
			req.DocumentID,
			versionRef.DocumentVersionID,
			err,
		)
		_ = s.cache.UpdateDocumentVersionStatus(ctx, versionRef.DocumentVersionID, string(domain.StatusFailed), &errMsg)
		return nil, err
	}

	reader, _, err := s.store.Get(ctx, versionRef.RawContentURI)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	payload, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	count, err := s.handle(ctx, DocumentRequest{
		KnowledgeBaseID:   req.KnowledgeBaseID,
		DocumentID:        req.DocumentID,
		DocumentVersionID: versionRef.DocumentVersionID,
		Content:           string(payload),
		Strategy:          req.Strategy,
		MaxRunes:          req.MaxRunes,
		OverlapRunes:      req.OverlapRunes,
		Separators:        req.Separators,
		LanguageHints:     req.LanguageHints,
	})
	if err != nil {
		return nil, err
	}

	resolved := req.Strategy
	if resolved == "" {
		resolved = chunking.Strategy(s.strategy)
	}
	return &InitiateResult{
		DocumentID:        req.DocumentID,
		DocumentVersionID: versionRef.DocumentVersionID,
		Strategy:          string(resolved),
		ChunkCount:        count,
	}, nil
}

type EmbedChunkResult struct {
	ChunkID     string `json:"chunk_id"`
	EmbeddingID string `json:"embedding_id"`
	Reused      bool   `json:"reused"`
}

func (s *Service) EmbedChunkByID(ctx context.Context, knowledgeBaseID, chunkID string) (*EmbedChunkResult, error) {
	if s.cache == nil {
		return nil, fmt.Errorf("cache layer is required")
	}
	if s.embedder == nil {
		return nil, ErrEmbedderUnavailable
	}
	if strings.TrimSpace(knowledgeBaseID) == "" {
		return nil, fmt.Errorf("knowledgebase_id is required")
	}
	if strings.TrimSpace(chunkID) == "" {
		return nil, fmt.Errorf("chunk_id is required")
	}

	chunk, err := s.cache.GetChunkByID(ctx, knowledgeBaseID, chunkID)
	if err != nil {
		return nil, err
	}
	if chunk == nil {
		return nil, ErrChunkNotFound
	}

	result, err := s.embedder.EnqueueChunkAndWait(ctx, embedding.EmbedChunkRequest{
		KnowledgeBaseID: knowledgeBaseID,
		Chunk: embedding.ChunkInput{
			ChunkID:     chunk.ID,
			Content:     chunk.Content,
			ContentHash: chunk.ContentHash,
			Metadata:    chunk.Metadata,
		},
	})
	if err != nil {
		return nil, err
	}
	if result.EmbeddingID == "" {
		return nil, fmt.Errorf("missing embedding id for chunk %s", chunk.ID)
	}
	if err := s.cache.UpdateChunkEmbedding(ctx, knowledgeBaseID, chunk.ID, result.EmbeddingID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChunkNotFound
		}
		return nil, err
	}

	return &EmbedChunkResult{
		ChunkID:     chunk.ID,
		EmbeddingID: result.EmbeddingID,
		Reused:      len(result.Vector) == 0,
	}, nil
}

func ParseLanguageHints(values []string) ([]chunking.Language, error) {
	if len(values) == 0 {
		return nil, nil
	}
	hints := make([]chunking.Language, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		switch strings.ToLower(value) {
		case string(chunking.LanguageGeneric),
			string(chunking.LanguageGo),
			string(chunking.LanguagePython),
			string(chunking.LanguageJavaScript),
			string(chunking.LanguageJava),
			string(chunking.LanguageRust):
			hints = append(hints, chunking.Language(value))
		default:
			return nil, fmt.Errorf("unsupported language hint: %s", value)
		}
	}
	return hints, nil
}

func (s *Service) handle(ctx context.Context, req DocumentRequest) (int, error) {
	if s.cache == nil {
		return 0, fmt.Errorf("cache layer is required")
	}
	if req.KnowledgeBaseID == "" {
		return 0, fmt.Errorf("knowledgebase_id is required")
	}
	if req.DocumentVersionID == "" {
		return 0, fmt.Errorf("document_version_id is required")
	}

	chunker, strategyName, err := s.resolveChunker(req)
	if err != nil {
		errMsg := fmt.Sprintf("stage=CHUNKED document_id=%s version_id=%s error=%v", req.DocumentID, req.DocumentVersionID, err)
		_ = s.cache.UpdateDocumentVersionStatus(ctx, req.DocumentVersionID, string(domain.StatusFailed), &errMsg)
		return 0, err
	}

	chunks, err := chunker.Chunk(req.Content)
	if err != nil {
		errMsg := fmt.Sprintf("stage=CHUNKED document_id=%s version_id=%s error=%v", req.DocumentID, req.DocumentVersionID, err)
		_ = s.cache.UpdateDocumentVersionStatus(ctx, req.DocumentVersionID, string(domain.StatusFailed), &errMsg)
		return 0, err
	}

	stored := make([]domain.Chunk, 0, len(chunks))
	for i, ch := range chunks {
		chunkID := uuid.NewString()
		stored = append(stored, domain.Chunk{
			ID:                chunkID,
			DocumentVersionID: req.DocumentVersionID,
			KBID:              req.KnowledgeBaseID,
			SequenceNumber:    i + 1,
			Content:           ch.Content,
			ContentHash:       hashContent(ch.Content),
			Metadata: map[string]any{
				"start_rune":  ch.StartRune,
				"end_rune":    ch.EndRune,
				"rune_length": ch.RuneLength,
			},
			ChunkingStrategy: strategyName,
			CreatedAt:        s.now(),
		})
	}

	if s.embedder != nil {
		for i := range stored {
			result, err := s.embedder.EnqueueChunkAndWait(ctx, embedding.EmbedChunkRequest{
				KnowledgeBaseID: req.KnowledgeBaseID,
				Chunk: embedding.ChunkInput{
					ChunkID:     stored[i].ID,
					Content:     stored[i].Content,
					ContentHash: stored[i].ContentHash,
					Metadata:    stored[i].Metadata,
				},
			})
			if err != nil {
				errMsg := fmt.Sprintf(
					"stage=EMBEDDED document_id=%s version_id=%s chunk_id=%s error=%v",
					req.DocumentID,
					req.DocumentVersionID,
					stored[i].ID,
					err,
				)
				_ = s.cache.UpdateDocumentVersionStatus(ctx, req.DocumentVersionID, string(domain.StatusFailed), &errMsg)
				return 0, err
			}
			if result.EmbeddingID == "" {
				err := fmt.Errorf("missing embedding id for chunk %s", stored[i].ID)
				errMsg := fmt.Sprintf(
					"stage=EMBEDDED document_id=%s version_id=%s chunk_id=%s error=%v",
					req.DocumentID,
					req.DocumentVersionID,
					stored[i].ID,
					err,
				)
				_ = s.cache.UpdateDocumentVersionStatus(ctx, req.DocumentVersionID, string(domain.StatusFailed), &errMsg)
				return 0, err
			}
			stored[i].EmbeddingID = &result.EmbeddingID
			logger.Info(
				"chunk embedding linked",
				"knowledge_base_id", req.KnowledgeBaseID,
				"document_id", req.DocumentID,
				"document_version_id", req.DocumentVersionID,
				"chunk_id", stored[i].ID,
				"embedding_id", result.EmbeddingID,
				"model_id", result.ModelID,
				"vector_dimension", result.VectorDimension,
				"reused", len(result.Vector) == 0,
			)
		}
	}

	if err := s.cache.DeleteChunksByDocumentVersion(ctx, req.DocumentVersionID); err != nil {
		errMsg := fmt.Sprintf("stage=CHUNKED document_id=%s version_id=%s error=%v", req.DocumentID, req.DocumentVersionID, err)
		_ = s.cache.UpdateDocumentVersionStatus(ctx, req.DocumentVersionID, string(domain.StatusFailed), &errMsg)
		return 0, err
	}
	if err := s.cache.InsertChunks(ctx, stored); err != nil {
		errMsg := fmt.Sprintf("stage=CHUNKED document_id=%s version_id=%s error=%v", req.DocumentID, req.DocumentVersionID, err)
		_ = s.cache.UpdateDocumentVersionStatus(ctx, req.DocumentVersionID, string(domain.StatusFailed), &errMsg)
		return 0, err
	}
	if err := s.cache.UpdateDocumentVersionStatus(ctx, req.DocumentVersionID, string(domain.StatusChunked), nil); err != nil {
		return 0, err
	}
	if s.embedder != nil {
		if err := s.cache.UpdateDocumentVersionStatus(ctx, req.DocumentVersionID, string(domain.StatusEmbedded), nil); err != nil {
			return 0, err
		}
	}
	if err := s.cache.ActivateDocumentVersion(ctx, req.DocumentVersionID); err != nil {
		return 0, err
	}
	return len(stored), nil
}

func (s *Service) resolveChunker(req DocumentRequest) (chunking.Chunker, string, error) {
	if req.Strategy == "" &&
		req.MaxRunes <= 0 &&
		req.OverlapRunes == 0 &&
		len(req.Separators) == 0 &&
		len(req.LanguageHints) == 0 {
		return s.chunker, s.strategy, nil
	}

	maxRunes := req.MaxRunes
	if maxRunes <= 0 {
		maxRunes = DefaultMaxRunes
	}
	overlap := req.OverlapRunes
	if overlap < 0 {
		return nil, "", chunking.ErrInvalidOverlap
	}
	if overlap >= maxRunes {
		return nil, "", chunking.ErrOverlapTooLarge
	}

	strategy := req.Strategy
	if strategy == "" {
		strategy = chunking.StrategyFixed
	}
	ch, err := chunking.NewChunker(chunking.Options{
		Strategy:      strategy,
		MaxRunes:      maxRunes,
		OverlapRunes:  overlap,
		Separators:    req.Separators,
		LanguageHints: normalizeHints(req.LanguageHints),
	})
	if err != nil {
		return nil, "", err
	}
	return ch, string(strategy), nil
}

func normalizeHints(hints []chunking.Language) []chunking.Language {
	if len(hints) == 0 {
		return nil
	}
	normalized := make([]chunking.Language, 0, len(hints))
	for _, hint := range hints {
		value := strings.TrimSpace(string(hint))
		if value == "" {
			continue
		}
		normalized = append(normalized, chunking.Language(value))
	}
	return normalized
}

func hashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}
