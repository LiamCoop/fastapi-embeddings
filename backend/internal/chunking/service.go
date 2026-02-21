package chunking

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"ragtime-backend/internal/domain"
	"ragtime-backend/internal/logger"
	"ragtime-backend/internal/objectstore"
)

const (
	DefaultChunkingStrategy = string(StrategyFixed)
	DefaultMaxRunes         = 1000
	DefaultOverlapRunes     = 100
)

var (
	ErrDocumentNotFound = errors.New("document not found")
)

// DocumentRequest is the payload passed over the chunking channel.
// Content must contain the full document text.
type DocumentRequest struct {
	KnowledgeBaseID   string
	DocumentID        string
	DocumentVersionID string
	Content           string
	Strategy          Strategy
	MaxRunes          int
	OverlapRunes      int
	Separators        []string
	LanguageHints     []Language
}

type InitiateRequest struct {
	KnowledgeBaseID string
	DocumentID      string
	Strategy        Strategy
	MaxRunes        int
	OverlapRunes    int
	Separators      []string
	LanguageHints   []Language
}

type InitiateResult struct {
	DocumentID        string `json:"document_id"`
	DocumentVersionID string `json:"document_version_id"`
	Strategy          string `json:"strategy"`
	ChunkCount        int    `json:"chunk_count"`
}

// Service consumes chunking requests and persists chunks.
type Service struct {
	cache    CacheLayer
	chunker  Chunker
	input    <-chan DocumentRequest
	store    objectstore.Client
	now      func() time.Time
	strategy string
}

func NewService(cache CacheLayer, chunker Chunker, input <-chan DocumentRequest) *Service {
	return NewServiceWithStore(cache, chunker, input, nil)
}

func NewServiceWithStore(cache CacheLayer, chunker Chunker, input <-chan DocumentRequest, store objectstore.Client) *Service {
	if chunker == nil {
		chunker = FixedSizeChunker{MaxRunes: DefaultMaxRunes, OverlapRunes: DefaultOverlapRunes}
	}
	return &Service{
		cache:    cache,
		chunker:  chunker,
		input:    input,
		store:    store,
		now:      func() time.Time { return time.Now().UTC() },
		strategy: DefaultChunkingStrategy,
	}
}

// Run listens on the input channel until the context is canceled.
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
	for i, chunk := range chunks {
		stored = append(stored, domain.Chunk{
			ID:                "",
			DocumentVersionID: req.DocumentVersionID,
			KBID:              req.KnowledgeBaseID,
			SequenceNumber:    i + 1,
			Content:           chunk.Content,
			ContentHash:       hashContent(chunk.Content),
			Metadata: map[string]any{
				"start_rune":  chunk.StartRune,
				"end_rune":    chunk.EndRune,
				"rune_length": chunk.RuneLength,
			},
			ChunkingStrategy: strategyName,
			EmbeddingID:      nil,
			CreatedAt:        s.now(),
		})
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

	return len(stored), nil
}

func hashContent(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

type Request struct {
	Text          string
	Strategy      Strategy
	MaxRunes      int
	OverlapRunes  int
	Separators    []string
	LanguageHints []Language
}

type Result struct {
	Strategy     Strategy
	MaxRunes     int
	OverlapRunes int
	Separators   []string
	Chunks       []Chunk
}

func (s *Service) Chunk(req Request) (*Result, error) {
	if req.MaxRunes <= 0 {
		return nil, ErrInvalidMaxRunes
	}
	if req.OverlapRunes < 0 {
		return nil, ErrInvalidOverlap
	}
	if req.OverlapRunes >= req.MaxRunes {
		return nil, ErrOverlapTooLarge
	}

	strategy := req.Strategy
	if strategy == "" {
		strategy = StrategyFixed
	}

	normalizedHints := normalizeHints(req.LanguageHints)

	chunker, err := NewChunker(Options{
		Strategy:      strategy,
		MaxRunes:      req.MaxRunes,
		OverlapRunes:  req.OverlapRunes,
		Separators:    req.Separators,
		LanguageHints: normalizedHints,
	})
	if err != nil {
		return nil, err
	}

	chunks, err := chunker.Chunk(req.Text)
	if err != nil {
		return nil, err
	}

	seps := req.Separators
	if strategy == StrategyRecursive && len(seps) == 0 {
		seps = separatorsForHints(normalizedHints)
	}

	return &Result{
		Strategy:     strategy,
		MaxRunes:     req.MaxRunes,
		OverlapRunes: req.OverlapRunes,
		Separators:   seps,
		Chunks:       chunks,
	}, nil
}

func normalizeHints(hints []Language) []Language {
	if len(hints) == 0 {
		return nil
	}
	normalized := make([]Language, 0, len(hints))
	for _, hint := range hints {
		value := strings.TrimSpace(string(hint))
		if value == "" {
			continue
		}
		normalized = append(normalized, Language(value))
	}
	return normalized
}

func ParseLanguageHints(values []string) ([]Language, error) {
	if len(values) == 0 {
		return nil, nil
	}
	hints := make([]Language, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		switch strings.ToLower(value) {
		case string(LanguageGeneric),
			string(LanguageGo),
			string(LanguagePython),
			string(LanguageJavaScript),
			string(LanguageJava),
			string(LanguageRust):
			hints = append(hints, Language(value))
		default:
			return nil, fmt.Errorf("unsupported language hint: %s", value)
		}
	}
	return hints, nil
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

	reader, _, err := s.store.Get(ctx, versionRef.RawContentURI)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	payload, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	chunkCount, err := s.handle(ctx, DocumentRequest{
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
		resolved = Strategy(s.strategy)
	}

	return &InitiateResult{
		DocumentID:        req.DocumentID,
		DocumentVersionID: versionRef.DocumentVersionID,
		Strategy:          string(resolved),
		ChunkCount:        chunkCount,
	}, nil
}

func (s *Service) resolveChunker(req DocumentRequest) (Chunker, string, error) {
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
		return nil, "", ErrInvalidOverlap
	}
	if overlap >= maxRunes {
		return nil, "", ErrOverlapTooLarge
	}

	strategy := req.Strategy
	if strategy == "" {
		strategy = StrategyFixed
	}

	chunker, err := NewChunker(Options{
		Strategy:      strategy,
		MaxRunes:      maxRunes,
		OverlapRunes:  overlap,
		Separators:    req.Separators,
		LanguageHints: normalizeHints(req.LanguageHints),
	})
	if err != nil {
		return nil, "", err
	}
	return chunker, string(strategy), nil
}
