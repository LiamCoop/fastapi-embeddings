package ingestion

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"ragtime-backend/internal/chunking"
	mdchunking "ragtime-backend/internal/chunking/markdown"
	"ragtime-backend/internal/document"
	"ragtime-backend/internal/domain"
	"ragtime-backend/internal/embedding"
	"ragtime-backend/internal/logger"
	"ragtime-backend/internal/objectstore"
)

const (
	defaultChunkMaxRunes = 2000
	defaultChunkOverlap  = 200
	defaultChunkingName  = "fixed_size"
	markdownChunkingName = "markdown"
)

type Service struct {
	documents *document.Service
	docRepo   document.Repository
	repo      Repository
	store     objectstore.Client
	embedder  *embedding.Service
	chunker   chunking.Chunker
	now       func() time.Time
}

func NewService(
	documents *document.Service,
	docRepo document.Repository,
	repo Repository,
	store objectstore.Client,
	embedder *embedding.Service,
) *Service {
	return &Service{
		documents: documents,
		docRepo:   docRepo,
		repo:      repo,
		store:     store,
		embedder:  embedder,
		chunker: chunking.FixedSizeChunker{
			MaxRunes:     defaultChunkMaxRunes,
			OverlapRunes: defaultChunkOverlap,
		},
		now: func() time.Time { return time.Now().UTC() },
	}
}

func NewServiceWithPostgres(
	db *sql.DB,
	store objectstore.Client,
	embedder *embedding.Service,
) *Service {
	docRepo := document.NewPostgresRepository(db)
	repo := NewPostgresRepository(db)
	documents := document.NewService(docRepo, store)
	return NewService(documents, docRepo, repo, store, embedder)
}

func (s *Service) IngestDocuments(ctx context.Context, req IngestDocumentsRequest) ([]IngestDocumentResult, error) {
	if s.documents == nil || s.docRepo == nil || s.repo == nil || s.store == nil || s.embedder == nil {
		return nil, fmt.Errorf("ingestion service is not fully configured")
	}
	results := make([]IngestDocumentResult, 0, len(req.Documents))
	for _, doc := range req.Documents {
		result := s.ingestOne(ctx, doc)
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) ingestOne(ctx context.Context, doc IngestDocumentRequest) IngestDocumentResult {
	result := IngestDocumentResult{
		KnowledgeBaseID: doc.KnowledgeBaseID,
	}
	if doc.RawContentURI == "" {
		msg := "raw_content_uri is required"
		result.ErrorMessage = &msg
		result.ProcessingStatus = string(domain.StatusFailed)
		return result
	}

	uploadResult, err := s.documents.Upload(ctx, document.UploadRequest{
		KnowledgeBaseID: doc.KnowledgeBaseID,
		Path:            doc.Path,
		Title:           doc.Title,
		DocumentType:    doc.DocumentType,
		SourceMetadata:  doc.SourceMetadata,
		RawContentURI:   &doc.RawContentURI,
	})
	if err != nil {
		msg := err.Error()
		result.ErrorMessage = &msg
		result.ProcessingStatus = string(domain.StatusFailed)
		return result
	}

	result.DocumentID = uploadResult.DocumentID
	result.DocumentVersionID = uploadResult.DocumentVersionID
	result.ProcessingStatus = uploadResult.ProcessingStatus

	jobID := uuid.NewString()
	result.IngestionJobID = jobID
	started := s.now()
	if err := s.repo.InsertIngestionJob(ctx, IngestionJobRecord{
		ID:                jobID,
		DocumentVersionID: uploadResult.DocumentVersionID,
		KnowledgeBaseID:   doc.KnowledgeBaseID,
		JobStatus:         JobInProgress,
		StartedAt:         &started,
	}); err != nil {
		msg := err.Error()
		result.ErrorMessage = &msg
		result.ProcessingStatus = string(domain.StatusFailed)
		return result
	}

	if uploadResult.ProcessingStatus == string(domain.StatusSkippedUnsupported) {
		result.SkippedUnsupported = true
		s.completeJob(ctx, jobID, JobSuccess, nil)
		return result
	}

	if err := s.processContent(ctx, uploadResult, doc); err != nil {
		msg := err.Error()
		result.ErrorMessage = &msg
		result.ProcessingStatus = string(domain.StatusFailed)
		_ = s.docRepo.UpdateDocumentVersionStatus(ctx, uploadResult.DocumentVersionID, string(domain.StatusFailed), &msg)
		s.completeJob(ctx, jobID, JobFailed, &msg)
		return result
	}

	result.ProcessingStatus = string(domain.StatusActivated)
	s.completeJob(ctx, jobID, JobSuccess, nil)
	return result
}

func (s *Service) processContent(
	ctx context.Context,
	uploadResult *document.UploadResult,
	doc IngestDocumentRequest,
) error {
	reader, _, err := s.store.Get(ctx, uploadResult.RawContentURI)
	if err != nil {
		return fmt.Errorf("fetch raw content: %w", err)
	}
	defer reader.Close()

	payload, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("read raw content: %w", err)
	}
	content := string(payload)
	if err := s.docRepo.UpdateDocumentVersionStatus(ctx, uploadResult.DocumentVersionID, string(domain.StatusExtracted), nil); err != nil {
		return err
	}

	chunker, err := s.chunkerForDoc(doc)
	if err != nil {
		return err
	}
	chunks, err := chunker.Chunk(content)
	if err != nil {
		return err
	}
	chunkingStrategy := s.chunkingStrategyForDoc(doc)

	chunkRecords := make([]ChunkRecord, 0, len(chunks))
	for _, chunk := range chunks {
		hash := sha256.Sum256([]byte(chunk.Content))
		meta := map[string]any{
			"start_rune": chunk.StartRune,
			"end_rune":   chunk.EndRune,
		}
		for k, v := range chunk.Metadata {
			meta[k] = v
		}
		chunkRecords = append(chunkRecords, ChunkRecord{
			ID:                uuid.NewString(),
			DocumentVersionID: uploadResult.DocumentVersionID,
			KnowledgeBaseID:   doc.KnowledgeBaseID,
			SequenceNumber:    chunk.Index,
			Content:           chunk.Content,
			ContentHash:       hex.EncodeToString(hash[:]),
			Metadata:          meta,
			ChunkingStrategy:  chunkingStrategy,
			CreatedAt:         s.now(),
		})
	}

	if err := s.repo.InsertChunks(ctx, chunkRecords); err != nil {
		return err
	}
	if err := s.docRepo.UpdateDocumentVersionStatus(ctx, uploadResult.DocumentVersionID, string(domain.StatusChunked), nil); err != nil {
		return err
	}

	for _, record := range chunkRecords {
		result, err := s.embedder.EnqueueChunkAndWait(ctx, embedding.EmbedChunkRequest{
			KnowledgeBaseID: doc.KnowledgeBaseID,
			Chunk: embedding.ChunkInput{
				ChunkID:     record.ID,
				Content:     record.Content,
				ContentHash: record.ContentHash,
				Metadata:    record.Metadata,
			},
		})
		if err != nil {
			return err
		}
		if result.EmbeddingID == "" {
			return fmt.Errorf("missing embedding id for chunk %s", record.ID)
		}
		if err := s.repo.UpdateChunkEmbedding(ctx, record.ID, result.EmbeddingID); err != nil {
			logger.Error("update chunk embedding failed", "chunk_id", record.ID, "error", err)
		}
	}

	if err := s.docRepo.UpdateDocumentVersionStatus(ctx, uploadResult.DocumentVersionID, string(domain.StatusEmbedded), nil); err != nil {
		return err
	}

	if err := s.docRepo.ActivateVersion(ctx, uploadResult.DocumentVersionID); err != nil {
		return err
	}

	return nil
}

func (s *Service) completeJob(ctx context.Context, jobID string, status JobStatus, errMsg *string) {
	completed := s.now()
	if err := s.repo.UpdateIngestionJob(ctx, jobID, status, errMsg, &completed); err != nil {
		logger.Error("update ingestion job failed", "job_id", jobID, "error", err)
	}
}

func (s *Service) chunkerForDoc(doc IngestDocumentRequest) (chunking.Chunker, error) {
	if !isMarkdownDoc(doc) {
		return s.chunker, nil
	}
	opts := mdchunking.DefaultMarkdownOptions()
	opts.MDX = isMDXPath(doc.Path)
	return chunking.NewMarkdownChunker(opts)
}

func (s *Service) chunkingStrategyForDoc(doc IngestDocumentRequest) string {
	if isMarkdownDoc(doc) {
		return markdownChunkingName
	}
	return defaultChunkingName
}

func isMarkdownDoc(doc IngestDocumentRequest) bool {
	if doc.DocumentType != nil && strings.EqualFold(*doc.DocumentType, document.DocTypeMarkdown) {
		return true
	}
	ext := strings.ToLower(filepath.Ext(doc.Path))
	return ext == ".md" || ext == ".mdx"
}

func isMDXPath(path string) bool {
	return strings.EqualFold(filepath.Ext(path), ".mdx")
}
