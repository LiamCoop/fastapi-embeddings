package document

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	chunkservice "ragtime-backend/internal/chunking/service"
	"ragtime-backend/internal/domain"
	"ragtime-backend/internal/objectstore"
)

const (
	DocTypeMarkdown = "markdown"
	DocTypePDF      = "pdf"
	DocTypeImage    = "image"
	DocTypeDocx     = "docx"
	DocTypeUnknown  = "unknown"
)

var supportedTypes = map[string]struct{}{
	DocTypeMarkdown: {},
}

// Service orchestrates document intake.
type Service struct {
	repo       Repository
	store      objectstore.Client
	chunkingCh chan<- chunkservice.DocumentRequest
	now        func() time.Time
}

func NewService(repo Repository, store objectstore.Client) *Service {
	return NewServiceWithChunking(repo, store, nil)
}

func NewServiceWithChunking(repo Repository, store objectstore.Client, chunkingCh chan<- chunkservice.DocumentRequest) *Service {
	return &Service{
		repo:       repo,
		store:      store,
		chunkingCh: chunkingCh,
		now:        func() time.Time { return time.Now().UTC() },
	}
}

func NewServiceWithPostgres(db *sql.DB, store objectstore.Client, chunkingCh chan<- chunkservice.DocumentRequest) *Service {
	repo := NewPostgresRepository(db)
	return NewServiceWithChunking(repo, store, chunkingCh)
}

func (s *Service) Upload(ctx context.Context, req UploadRequest) (*UploadResult, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository is required")
	}
	if s.store == nil {
		return nil, fmt.Errorf("object store is required")
	}
	if err := ValidateUploadRequest(req); err != nil {
		return nil, err
	}

	docType := req.DocumentType
	if docType == nil || *docType == "" {
		detected := DetectDocumentType(req.Path, req.ContentType)
		docType = &detected
	}

	now := s.now()
	existing, err := s.repo.GetDocumentByKBPath(ctx, req.KnowledgeBaseID, req.Path)
	if err != nil {
		return nil, err
	}

	var doc *DocumentRecord
	if existing == nil {
		doc = &DocumentRecord{
			ID:              uuid.NewString(),
			KnowledgeBaseID: req.KnowledgeBaseID,
			Path:            req.Path,
			Title:           req.Title,
			DocumentType:    *docType,
			SourceMetadata:  req.SourceMetadata,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		inserted, err := s.repo.InsertDocument(ctx, *doc)
		if err != nil {
			return nil, err
		}
		doc = inserted
	} else {
		doc = existing
		if req.Title != nil {
			doc.Title = req.Title
		}
		if docType != nil && *docType != "" {
			doc.DocumentType = *docType
		}
		if req.SourceMetadata != nil {
			doc.SourceMetadata = req.SourceMetadata
		}
		doc.UpdatedAt = now

		updated, err := s.repo.UpdateDocument(ctx, *doc)
		if err != nil {
			return nil, err
		}
		doc = updated
	}

	versionID := uuid.NewString()
	key := buildObjectKey(req.KnowledgeBaseID, doc.ID, versionID, req.FileName, req.Path)
	uri := ""
	if req.RawContentURI != nil && *req.RawContentURI != "" {
		uri = *req.RawContentURI
	} else {
		uri = s.store.URIForKey(key)
	}

	version := &DocumentVersionRecord{
		ID:               versionID,
		DocumentID:       doc.ID,
		KnowledgeBaseID:  req.KnowledgeBaseID,
		RawContentURI:    uri,
		ProcessingStatus: string(domain.StatusReceived),
		IsActive:         false,
		CreatedAt:        now,
	}

	insertedVersion, err := s.repo.InsertDocumentVersion(ctx, *version)
	if err != nil {
		return nil, err
	}
	version = insertedVersion

	status := string(domain.StatusStored)
	if !IsSupportedType(doc.DocumentType) {
		status = string(domain.StatusSkippedUnsupported)
	}

	if req.RawContentURI == nil || *req.RawContentURI == "" {
		if _, _, err := s.store.Put(ctx, key, bytes.NewReader(req.FileContent)); err != nil {
			errMsg := fmt.Sprintf("stage=STORED document_id=%s version_id=%s error=%v", doc.ID, version.ID, err)
			_ = s.repo.UpdateDocumentVersionStatus(ctx, version.ID, string(domain.StatusFailed), &errMsg)
			return nil, err
		}
	}

	if err := s.repo.UpdateDocumentVersionStatus(ctx, version.ID, status, nil); err != nil {
		return nil, err
	}

	if status == string(domain.StatusStored) && s.chunkingCh != nil && len(req.FileContent) > 0 {
		s.chunkingCh <- chunkservice.DocumentRequest{
			KnowledgeBaseID:   req.KnowledgeBaseID,
			DocumentID:        doc.ID,
			DocumentVersionID: version.ID,
			Content:           string(req.FileContent),
		}
	}

	result := &UploadResult{
		DocumentID:        doc.ID,
		DocumentVersionID: version.ID,
		VersionNumber:     version.VersionNumber,
		Path:              doc.Path,
		DocumentType:      doc.DocumentType,
		RawContentURI:     version.RawContentURI,
		ProcessingStatus:  status,
		IsActive:          version.IsActive,
		CreatedAt:         version.CreatedAt,
	}

	return result, nil
}

// PresignUpload returns a presigned URL and storage URI for client-side uploads.
func (s *Service) PresignUpload(ctx context.Context, knowledgeBaseID, fileName, contentType string) (*PresignResult, error) {
	if s.store == nil {
		return nil, fmt.Errorf("object store is required")
	}
	if knowledgeBaseID == "" {
		return nil, ErrMissingKnowledgeBaseID
	}
	if strings.TrimSpace(fileName) == "" {
		return nil, fmt.Errorf("file_name is required")
	}

	uploadID := uuid.NewString()
	key := buildUploadKey(knowledgeBaseID, uploadID, fileName)
	url, headers, uri, err := s.store.PresignPut(ctx, key, contentType)
	if err != nil {
		return nil, err
	}

	return &PresignResult{
		UploadURL:     url,
		Headers:       headers,
		RawContentURI: uri,
		ObjectKey:     key,
	}, nil
}

func IsSupportedType(docType string) bool {
	_, ok := supportedTypes[docType]
	return ok
}

// IsKnownType returns true if the provided document type is recognized.
func IsKnownType(docType string) bool {
	switch docType {
	case DocTypeMarkdown, DocTypePDF, DocTypeImage, DocTypeDocx, DocTypeUnknown:
		return true
	default:
		return false
	}
}

func DetectDocumentType(path, contentType string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md", ".mdx":
		return DocTypeMarkdown
	case ".pdf":
		return DocTypePDF
	case ".docx":
		return DocTypeDocx
	}

	if strings.HasPrefix(contentType, "text/markdown") {
		return DocTypeMarkdown
	}
	if strings.HasPrefix(contentType, "application/pdf") {
		return DocTypePDF
	}
	if strings.HasPrefix(contentType, "image/") {
		return DocTypeImage
	}

	return DocTypeUnknown
}

func buildObjectKey(kbID, docID, versionID, fileName, docPath string) string {
	name := strings.TrimSpace(fileName)
	if name == "" {
		name = filepath.Base(docPath)
	}
	if name == "" {
		name = "document"
	}
	name = strings.ReplaceAll(name, string(filepath.Separator), "_")
	return fmt.Sprintf("kb/%s/documents/%s/versions/%s/%s", kbID, docID, versionID, name)
}

func buildUploadKey(kbID, uploadID, fileName string) string {
	name := strings.TrimSpace(fileName)
	if name == "" {
		name = "upload"
	}
	name = strings.ReplaceAll(name, string(filepath.Separator), "_")
	return fmt.Sprintf("kb/%s/uploads/%s/%s", kbID, uploadID, name)
}

// DocumentRecord captures the repository representation of a document.
type DocumentRecord struct {
	ID              string
	KnowledgeBaseID string
	Path            string
	Title           *string
	DocumentType    string
	SourceMetadata  map[string]any
	ActiveVersionID *string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// DocumentVersionRecord captures the repository representation of a version.
type DocumentVersionRecord struct {
	ID               string
	DocumentID       string
	KnowledgeBaseID  string
	VersionNumber    int32
	RawContentURI    string
	ProcessingStatus string
	ErrorMessage     *string
	IsActive         bool
	CreatedAt        time.Time
}
