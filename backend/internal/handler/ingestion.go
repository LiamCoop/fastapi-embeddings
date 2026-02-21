package handler

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"strings"

	"ragtime-backend/internal/document"
	"ragtime-backend/internal/ingestion"
)

// IngestionHandler handles document ingestion endpoints.
type IngestionHandler struct {
	service *ingestion.Service
}

func NewIngestionHandler(service *ingestion.Service) *IngestionHandler {
	return &IngestionHandler{service: service}
}

func (h *IngestionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) != 4 || parts[0] != "v1" || parts[1] != "kb" || parts[3] != "ingest" {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	if h.service == nil {
		writeError(w, http.StatusInternalServerError, "ingestion service unavailable")
		return
	}

	kbID := parts[2]
	var payload ingestRequestJSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	documents := payload.Documents
	if payload.Document != nil {
		if len(documents) > 0 {
			writeError(w, http.StatusBadRequest, "provide either document or documents")
			return
		}
		documents = []ingestDocumentJSON{*payload.Document}
	}
	if len(documents) == 0 {
		writeError(w, http.StatusBadRequest, "documents is required")
		return
	}

	req := ingestion.IngestDocumentsRequest{
		Documents: make([]ingestion.IngestDocumentRequest, 0, len(documents)),
	}
	for _, doc := range documents {
		if err := validateIngestDoc(doc); err != "" {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		req.Documents = append(req.Documents, ingestion.IngestDocumentRequest{
			KnowledgeBaseID: kbID,
			Path:            doc.Path,
			Title:           doc.Title,
			DocumentType:    doc.DocumentType,
			SourceMetadata:  doc.SourceMetadata,
			RawContentURI:   doc.RawContentURI,
		})
	}

	results, err := h.service.IngestDocuments(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"documents": results,
	})
}

type ingestRequestJSON struct {
	Document  *ingestDocumentJSON  `json:"document,omitempty"`
	Documents []ingestDocumentJSON `json:"documents,omitempty"`
}

type ingestDocumentJSON struct {
	Path           string         `json:"path"`
	Title          *string        `json:"title"`
	DocumentType   *string        `json:"document_type"`
	SourceMetadata map[string]any `json:"source_metadata"`
	RawContentURI  string         `json:"raw_content_uri"`
}

func validateIngestDoc(doc ingestDocumentJSON) string {
	if strings.TrimSpace(doc.Path) == "" {
		return "path is required"
	}
	if strings.TrimSpace(doc.RawContentURI) == "" {
		return "raw_content_uri is required"
	}
	if !isSupportedRawURI(doc.RawContentURI) {
		return "raw_content_uri must start with s3:// or file://"
	}
	if doc.DocumentType != nil && *doc.DocumentType != "" && *doc.DocumentType != document.DocTypeMarkdown {
		return "document_type must be markdown"
	}
	if !isMarkdownPath(doc.Path) {
		return "path must end with .md or .mdx"
	}
	return ""
}

func isSupportedRawURI(uri string) bool {
	trimmed := strings.TrimSpace(uri)
	return strings.HasPrefix(trimmed, "s3://") || strings.HasPrefix(trimmed, "file://")
}

func isMarkdownPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(strings.TrimSpace(path)))
	return ext == ".md" || ext == ".mdx"
}
