package embedding

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

const defaultEmbeddingProvider = "fastapi"

// NewEmbedderFromEnv selects an embedding client based on environment variables.
// EMBEDDING_PROVIDER: fastapi (default) | openai
// EMBEDDING_BASE_URL: base URL for the FastAPI embedding service
// FASTAPI_EMBEDDINGS_URL: alternate base URL for the FastAPI embedding service
// OPENAI_API_KEY: API key for OpenAI (required when provider=openai)
// OPENAI_EMBEDDING_MODEL: OpenAI model id (default text-embedding-3-small)
// OPENAI_EMBEDDING_DIMENSIONS: optional integer to request specific dimensions
func NewEmbedderFromEnv() (TextEmbedder, error) {
	provider := strings.ToLower(strings.TrimSpace(os.Getenv("EMBEDDING_PROVIDER")))
	if provider == "" {
		provider = defaultEmbeddingProvider
	}

	switch provider {
	case "fastapi", "http", "local":
		baseURL := strings.TrimSpace(os.Getenv("EMBEDDING_BASE_URL"))
		if baseURL == "" {
			baseURL = strings.TrimSpace(os.Getenv("FASTAPI_EMBEDDINGS_URL"))
		}
		return NewClient(baseURL), nil
	case "openai":
		apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
		model := strings.TrimSpace(os.Getenv("OPENAI_EMBEDDING_MODEL"))
		dimensions := 0

		if raw := strings.TrimSpace(os.Getenv("OPENAI_EMBEDDING_DIMENSIONS")); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				return nil, fmt.Errorf("invalid OPENAI_EMBEDDING_DIMENSIONS: %w", err)
			}
			dimensions = parsed
		} else if model == "" || model == "text-embedding-3-small" {
			dimensions = 1536
		}

		return NewOpenAIClient(apiKey, model, dimensions), nil
	default:
		return nil, fmt.Errorf("unsupported EMBEDDING_PROVIDER: %s", provider)
	}
}
