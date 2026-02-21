package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"ragtime-backend/internal/document"
)

const (
	maxUploadMemory = int64(32 << 20)
)

// DocumentHandler handles document upload endpoints.
type DocumentHandler struct {
	service *document.Service
}

func NewDocumentHandler(service *document.Service) *DocumentHandler {
	return &DocumentHandler{service: service}
}

func (h *DocumentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "v1" || parts[1] != "kb" {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}
	kbID := parts[2]
	if parts[3] != "documents" {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	if len(parts) == 5 && parts[4] == "batch" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.handleBatchUpload(w, r, kbID)
		return
	}
	if len(parts) == 5 && parts[4] == "presign" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.handlePresign(w, r, kbID)
		return
	}

	if len(parts) != 4 {
		writeError(w, http.StatusNotFound, "route not found")
		return
	}

	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	contentType := r.Header.Get("Content-Type")
	if strings.HasPrefix(contentType, "multipart/form-data") {
		h.handleMultipartUpload(w, r, kbID)
		return
	}

	h.handleJSONUpload(w, r, kbID)
}

func (h *DocumentHandler) handleMultipartUpload(w http.ResponseWriter, r *http.Request, kbID string) {
	if err := r.ParseMultipartForm(maxUploadMemory); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	payload, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "unable to read file")
		return
	}

	path := r.FormValue("path")
	if path == "" {
		path = header.Filename
	}

	title := trimToNil(r.FormValue("title"))
	documentType := trimToNil(r.FormValue("document_type"))
	metadata, err := parseMetadata(r.FormValue("source_metadata"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid source_metadata")
		return
	}

	req := document.UploadRequest{
		KnowledgeBaseID: kbID,
		Path:            path,
		Title:           title,
		DocumentType:    documentType,
		SourceMetadata:  metadata,
		FileName:        header.Filename,
		ContentType:     header.Header.Get("Content-Type"),
		FileContent:     payload,
	}

	res, err := h.service.Upload(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, res)
}

func (h *DocumentHandler) handleJSONUpload(w http.ResponseWriter, r *http.Request, kbID string) {
	var reqPayload uploadJSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reqPayload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	req := document.UploadRequest{
		KnowledgeBaseID: kbID,
		Path:            reqPayload.Path,
		Title:           reqPayload.Title,
		DocumentType:    reqPayload.DocumentType,
		SourceMetadata:  reqPayload.SourceMetadata,
		RawContentURI:   reqPayload.RawContentURI,
		ContentType:     reqPayload.ContentType,
	}

	res, err := h.service.Upload(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, res)
}

func (h *DocumentHandler) handleBatchUpload(w http.ResponseWriter, r *http.Request, kbID string) {
	var reqPayload batchUploadJSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reqPayload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if len(reqPayload.Documents) == 0 {
		writeError(w, http.StatusBadRequest, "documents is required")
		return
	}

	results := make([]*document.UploadResult, 0, len(reqPayload.Documents))
	for i, docReq := range reqPayload.Documents {
		if docReq.RawContentURI == nil || *docReq.RawContentURI == "" {
			writeError(w, http.StatusBadRequest, "raw_content_uri is required for batch uploads")
			return
		}

		res, err := h.service.Upload(r.Context(), document.UploadRequest{
			KnowledgeBaseID: kbID,
			Path:            docReq.Path,
			Title:           docReq.Title,
			DocumentType:    docReq.DocumentType,
			SourceMetadata:  docReq.SourceMetadata,
			RawContentURI:   docReq.RawContentURI,
			ContentType:     docReq.ContentType,
		})
		if err != nil {
			writeError(w, http.StatusBadRequest, "document upload failed at index "+strconv.Itoa(i)+": "+err.Error())
			return
		}
		results = append(results, res)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"documents": results,
	})
}

func (h *DocumentHandler) handlePresign(w http.ResponseWriter, r *http.Request, kbID string) {
	var reqPayload presignJSON
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&reqPayload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if strings.TrimSpace(reqPayload.FileName) == "" {
		writeError(w, http.StatusBadRequest, "file_name is required")
		return
	}

	res, err := h.service.PresignUpload(r.Context(), kbID, reqPayload.FileName, reqPayload.ContentType)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, res)
}

type uploadJSON struct {
	Path           string         `json:"path"`
	Title          *string        `json:"title"`
	DocumentType   *string        `json:"document_type"`
	SourceMetadata map[string]any `json:"source_metadata"`
	RawContentURI  *string        `json:"raw_content_uri"`
	ContentType    string         `json:"content_type"`
}

type batchUploadJSON struct {
	Documents []uploadJSON `json:"documents"`
}

type presignJSON struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
}

func parseMetadata(value string) (map[string]any, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}

	var metadata map[string]any
	if err := json.Unmarshal([]byte(value), &metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

func trimToNil(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	if status >= 500 {
		message = "internal server error"
	}
	writeJSON(w, status, map[string]string{"error": message})
}
