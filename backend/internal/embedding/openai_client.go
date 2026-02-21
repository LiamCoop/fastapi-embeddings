package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultOpenAIBaseURL = "https://api.openai.com/v1"

var ErrMissingOpenAIAPIKey = errors.New("OPENAI_API_KEY is required")

// OpenAIClient calls the OpenAI embeddings API.
type OpenAIClient struct {
	apiKey     string
	baseURL    string
	model      string
	dimensions int
	httpClient *http.Client
}

// NewOpenAIClient creates an OpenAI embedding client.
func NewOpenAIClient(apiKey, model string, dimensions int) *OpenAIClient {
	trimmedKey := strings.TrimSpace(apiKey)
	trimmedModel := strings.TrimSpace(model)
	if trimmedModel == "" {
		trimmedModel = "text-embedding-3-small"
	}
	if dimensions < 0 {
		dimensions = 0
	}

	return &OpenAIClient{
		apiKey:     trimmedKey,
		baseURL:    defaultOpenAIBaseURL,
		model:      trimmedModel,
		dimensions: dimensions,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type openAIEmbedRequest struct {
	Input          []string `json:"input"`
	Model          string   `json:"model"`
	EncodingFormat string   `json:"encoding_format,omitempty"`
	Dimensions     int      `json:"dimensions,omitempty"`
}

type openAIEmbedResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

// EmbedTexts requests embeddings from OpenAI.
func (c *OpenAIClient) EmbedTexts(ctx context.Context, texts []string) ([][]float32, int, error) {
	if len(texts) == 0 {
		return nil, 0, ErrEmptyTexts
	}
	if c == nil {
		return nil, 0, errors.New("openai client is required")
	}
	if c.apiKey == "" {
		return nil, 0, ErrMissingOpenAIAPIKey
	}

	reqPayload := openAIEmbedRequest{
		Input:          texts,
		Model:          c.model,
		EncodingFormat: "float",
	}
	if c.dimensions > 0 {
		reqPayload.Dimensions = c.dimensions
	}

	payload, err := json.Marshal(reqPayload)
	if err != nil {
		return nil, 0, fmt.Errorf("marshal openai embed request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embeddings", bytes.NewReader(payload))
	if err != nil {
		return nil, 0, fmt.Errorf("create openai embed request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("openai embed request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, 0, fmt.Errorf("openai embed request returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var decoded openAIEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, 0, fmt.Errorf("decode openai embed response: %w", err)
	}

	if len(decoded.Data) != len(texts) {
		return nil, 0, fmt.Errorf("openai embed response mismatch: expected %d embeddings, got %d", len(texts), len(decoded.Data))
	}

	dim := 0
	vectors := make([][]float32, 0, len(decoded.Data))
	for _, item := range decoded.Data {
		if dim == 0 {
			dim = len(item.Embedding)
		} else if len(item.Embedding) != dim {
			return nil, 0, fmt.Errorf("openai embed response mismatch: expected dimension %d, got %d", dim, len(item.Embedding))
		}

		vector := make([]float32, len(item.Embedding))
		for i, value := range item.Embedding {
			vector[i] = float32(value)
		}
		vectors = append(vectors, vector)
	}

	return vectors, dim, nil
}
