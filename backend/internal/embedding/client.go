package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "http://localhost:8000"

var ErrEmptyTexts = errors.New("texts are required")

// Client calls the external embedding service.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	trimmed := strings.TrimRight(baseURL, "/")
	if trimmed == "" {
		trimmed = defaultBaseURL
	}

	return &Client{
		baseURL: trimmed,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type embedRequest struct {
	Texts     []string `json:"texts"`
	Normalize bool     `json:"normalize"`
}

type embedResponse struct {
	Dim        int         `json:"dim"`
	Embeddings [][]float32 `json:"embeddings"`
}

// EmbedTexts requests embeddings for raw texts from the external embedding service.
func (c *Client) EmbedTexts(ctx context.Context, texts []string) ([][]float32, int, error) {
	if len(texts) == 0 {
		return nil, 0, ErrEmptyTexts
	}

	payload, err := json.Marshal(embedRequest{Texts: texts, Normalize: true})
	if err != nil {
		return nil, 0, fmt.Errorf("marshal embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embed", bytes.NewReader(payload))
	if err != nil {
		return nil, 0, fmt.Errorf("create embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("embed request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, 0, fmt.Errorf("embed request returned status %d", resp.StatusCode)
	}

	var decoded embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, 0, fmt.Errorf("decode embed response: %w", err)
	}

	if len(decoded.Embeddings) != len(texts) {
		return nil, 0, fmt.Errorf("embed response mismatch: expected %d embeddings, got %d", len(texts), len(decoded.Embeddings))
	}

	return decoded.Embeddings, decoded.Dim, nil
}
