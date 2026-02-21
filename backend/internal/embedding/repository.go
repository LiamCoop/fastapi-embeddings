package embedding

import "context"

// Repository stores and queries embeddings for a knowledge base.
type Repository interface {
	HasEmbedding(ctx context.Context, knowledgeBaseID, contentHash, modelID string) (bool, error)
	FindEmbeddingID(ctx context.Context, knowledgeBaseID, contentHash, modelID string) (string, bool, error)
	SaveEmbeddings(ctx context.Context, embeddings []EmbeddingResult) ([]EmbeddingResult, error)
}
