package embedding

import "testing"

func TestValidateEmbedChunksRequest(t *testing.T) {
	newReq := func() EmbedChunksRequest {
		return EmbedChunksRequest{
			KnowledgeBaseID: "kb-123",
			Chunks: []ChunkInput{
				{
					ChunkID:     "chunk-1",
					Content:     "hello",
					ContentHash: "hash-1",
				},
			},
		}
	}

	t.Run("missing knowledgebase", func(t *testing.T) {
		req := newReq()
		req.KnowledgeBaseID = ""
		if err := ValidateEmbedChunksRequest(req); err != ErrMissingKnowledgeBaseID {
			t.Fatalf("expected %v, got %v", ErrMissingKnowledgeBaseID, err)
		}
	})

	t.Run("missing chunk id", func(t *testing.T) {
		req := newReq()
		req.Chunks[0].ChunkID = ""
		if err := ValidateEmbedChunksRequest(req); err != ErrMissingChunkID {
			t.Fatalf("expected %v, got %v", ErrMissingChunkID, err)
		}
	})

	t.Run("missing content", func(t *testing.T) {
		req := newReq()
		req.Chunks[0].Content = ""
		if err := ValidateEmbedChunksRequest(req); err != ErrMissingContent {
			t.Fatalf("expected %v, got %v", ErrMissingContent, err)
		}
	})

	t.Run("missing content hash", func(t *testing.T) {
		req := newReq()
		req.Chunks[0].ContentHash = ""
		if err := ValidateEmbedChunksRequest(req); err != ErrMissingContentHash {
			t.Fatalf("expected %v, got %v", ErrMissingContentHash, err)
		}
	})

	t.Run("valid request", func(t *testing.T) {
		if err := ValidateEmbedChunksRequest(newReq()); err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
	})
}

func TestFilterDedupedChunks(t *testing.T) {
	chunks := []ChunkInput{
		{ChunkID: "1", Content: "a", ContentHash: "hash-a"},
		{ChunkID: "2", Content: "b", ContentHash: "hash-b"},
		{ChunkID: "3", Content: "a-dup", ContentHash: "hash-a"},
	}

	seenExisting := func(knowledgeBaseID, contentHash string) bool {
		return knowledgeBaseID == "kb-1" && contentHash == "hash-b"
	}

	filtered := FilterDedupedChunks("kb-1", chunks, seenExisting)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 chunk after dedup, got %d", len(filtered))
	}
	if filtered[0].ContentHash != "hash-a" {
		t.Fatalf("expected remaining hash-a, got %s", filtered[0].ContentHash)
	}
}
