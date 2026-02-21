package embedding

import "errors"

var (
	ErrMissingKnowledgeBaseID = errors.New("knowledgebase_id is required")
	ErrMissingChunkID         = errors.New("chunk_id is required")
	ErrMissingContent         = errors.New("content is required")
	ErrMissingContentHash     = errors.New("content_hash is required")
)

type ChunkInput struct {
	ChunkID     string
	Content     string
	ContentHash string
	Metadata    map[string]any
}

type EmbedChunksRequest struct {
	KnowledgeBaseID string
	Chunks          []ChunkInput
	ModelID         *string
}

// EmbedChunkRequest is the payload sent over the embedding input channel.
type EmbedChunkRequest struct {
	KnowledgeBaseID string
	Chunk           ChunkInput
	ModelID         *string
	ResultCh        chan<- EmbedChunkResult
}

// EmbedChunkResult carries the outcome of processing one chunk.
type EmbedChunkResult struct {
	Result *EmbeddingResult
	Err    error
}

type EmbeddingResult struct {
	EmbeddingID     string
	ChunkID         string
	KnowledgeBaseID string
	ContentHash     string
	ModelID         string
	Vector          []float32
	VectorDimension int
}

// ValidateEmbedChunksRequest enforces the minimal backend contract before calling the embedding service.
func ValidateEmbedChunksRequest(req EmbedChunksRequest) error {
	if req.KnowledgeBaseID == "" {
		return ErrMissingKnowledgeBaseID
	}

	for _, chunk := range req.Chunks {
		if chunk.ChunkID == "" {
			return ErrMissingChunkID
		}
		if chunk.Content == "" {
			return ErrMissingContent
		}
		if chunk.ContentHash == "" {
			return ErrMissingContentHash
		}
	}

	return nil
}

// FilterDedupedChunks removes chunks that already have embeddings for the same KBID + content hash.
func FilterDedupedChunks(
	knowledgeBaseID string,
	chunks []ChunkInput,
	hasExisting func(knowledgeBaseID, contentHash string) bool,
) []ChunkInput {
	if len(chunks) == 0 {
		return nil
	}

	filtered := make([]ChunkInput, 0, len(chunks))
	seen := make(map[string]struct{}, len(chunks))

	for _, chunk := range chunks {
		if _, ok := seen[chunk.ContentHash]; ok {
			continue
		}
		seen[chunk.ContentHash] = struct{}{}

		if hasExisting != nil && hasExisting(knowledgeBaseID, chunk.ContentHash) {
			continue
		}

		filtered = append(filtered, chunk)
	}

	return filtered
}
