package ingestion

// IngestDocumentRequest captures a presigned upload document reference.
type IngestDocumentRequest struct {
	KnowledgeBaseID string
	Path            string
	Title           *string
	DocumentType    *string
	SourceMetadata  map[string]any
	RawContentURI   string
}

type IngestDocumentsRequest struct {
	Documents []IngestDocumentRequest
}

type IngestDocumentResult struct {
	KnowledgeBaseID    string
	DocumentID         string
	DocumentVersionID  string
	IngestionJobID     string
	ProcessingStatus   string
	ErrorMessage       *string
	SkippedUnsupported bool
}
