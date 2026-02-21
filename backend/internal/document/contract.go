package document

import (
	"context"
	"errors"
	"time"
)

var (
	ErrMissingKnowledgeBaseID = errors.New("knowledgebase_id is required")
	ErrMissingPath            = errors.New("path is required")
	ErrMissingContent         = errors.New("file or raw_content_uri is required")
)

// Repository defines document persistence requirements.
type Repository interface {
	GetDocumentByKBPath(ctx context.Context, kbID, path string) (*DocumentRecord, error)
	InsertDocument(ctx context.Context, doc DocumentRecord) (*DocumentRecord, error)
	UpdateDocument(ctx context.Context, doc DocumentRecord) (*DocumentRecord, error)
	InsertDocumentVersion(ctx context.Context, version DocumentVersionRecord) (*DocumentVersionRecord, error)
	UpdateDocumentVersionStatus(ctx context.Context, versionID, status string, errorMessage *string) error
	ActivateVersion(ctx context.Context, versionID string) error
}

// UploadRequest captures the intake payload for document ingestion.
type UploadRequest struct {
	KnowledgeBaseID string
	Path            string
	Title           *string
	DocumentType    *string
	SourceMetadata  map[string]any
	RawContentURI   *string
	FileName        string
	ContentType     string
	FileContent     []byte
}

// UploadResult summarizes the created document/version state.
type UploadResult struct {
	DocumentID        string
	DocumentVersionID string
	VersionNumber     int32
	Path              string
	DocumentType      string
	RawContentURI     string
	ProcessingStatus  string
	IsActive          bool
	CreatedAt         time.Time
}

// PresignResult returns information needed for client-side uploads.
type PresignResult struct {
	UploadURL     string
	Headers       map[string]string
	RawContentURI string
	ObjectKey     string
}

func ValidateUploadRequest(req UploadRequest) error {
	if req.KnowledgeBaseID == "" {
		return ErrMissingKnowledgeBaseID
	}
	if req.Path == "" {
		return ErrMissingPath
	}
	if len(req.FileContent) == 0 && (req.RawContentURI == nil || *req.RawContentURI == "") {
		return ErrMissingContent
	}

	return nil
}
