package embedding

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

type stubEmbedder struct {
	vectors [][]float32
	dim     int
	calls   int
}

func (s *stubEmbedder) EmbedTexts(_ context.Context, texts []string) ([][]float32, int, error) {
	s.calls++
	return s.vectors, s.dim, nil
}

type stubRepo struct {
	existing map[string]bool
	saved    []EmbeddingResult
}

func (s *stubRepo) HasEmbedding(
	_ context.Context,
	knowledgeBaseID, contentHash, modelID string,
) (bool, error) {
	_ = modelID
	return s.existing[knowledgeBaseID+"|"+contentHash], nil
}

func (s *stubRepo) FindEmbeddingID(
	_ context.Context,
	knowledgeBaseID, contentHash, modelID string,
) (string, bool, error) {
	_ = modelID
	if s.existing[knowledgeBaseID+"|"+contentHash] {
		return "embed-existing", true, nil
	}
	return "", false, nil
}

func (s *stubRepo) SaveEmbeddings(_ context.Context, embeddings []EmbeddingResult) ([]EmbeddingResult, error) {
	for i := range embeddings {
		if embeddings[i].EmbeddingID == "" {
			embeddings[i].EmbeddingID = uuid.NewString()
		}
	}
	s.saved = embeddings
	return embeddings, nil
}

func TestServiceEmbedAndStore(t *testing.T) {
	embedder := &stubEmbedder{
		vectors: [][]float32{{0.1, 0.2}, {0.3, 0.4}},
		dim:     2,
	}
	repo := &stubRepo{
		existing: map[string]bool{"kb-1|hash-b": true},
	}
	service := NewService(embedder, repo, "model-default", make(chan EmbedChunkRequest))

	req := EmbedChunksRequest{
		KnowledgeBaseID: "kb-1",
		Chunks: []ChunkInput{
			{ChunkID: "1", Content: "a", ContentHash: "hash-a"},
			{ChunkID: "2", Content: "b", ContentHash: "hash-b"},
			{ChunkID: "3", Content: "a-dup", ContentHash: "hash-a"},
		},
	}

	results, err := service.EmbedAndStore(context.Background(), req)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if embedder.calls != 1 {
		t.Fatalf("expected embedder called once, got %d", embedder.calls)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 embeddings, got %d", len(results))
	}
	hashes := map[string]bool{}
	for _, result := range results {
		hashes[result.ContentHash] = true
	}
	if !hashes["hash-a"] || !hashes["hash-b"] {
		t.Fatalf("expected hash-a and hash-b results, got %+v", hashes)
	}
	for _, result := range results {
		if result.ModelID != "model-default" {
			t.Fatalf("expected model-default, got %s", result.ModelID)
		}
	}
}

func TestServiceEmbedAndStoreValidation(t *testing.T) {
	baseReq := EmbedChunksRequest{
		KnowledgeBaseID: "kb-1",
		Chunks: []ChunkInput{
			{ChunkID: "1", Content: "a", ContentHash: "hash-a"},
		},
	}

	t.Run("missing embedder", func(t *testing.T) {
		service := NewService(nil, &stubRepo{}, "model-default", make(chan EmbedChunkRequest))
		_, err := service.EmbedAndStore(context.Background(), baseReq)
		if err != ErrNilEmbedder {
			t.Fatalf("expected %v, got %v", ErrNilEmbedder, err)
		}
	})

	t.Run("missing repository", func(t *testing.T) {
		service := NewService(&stubEmbedder{}, nil, "model-default", make(chan EmbedChunkRequest))
		_, err := service.EmbedAndStore(context.Background(), baseReq)
		if err != ErrNilRepository {
			t.Fatalf("expected %v, got %v", ErrNilRepository, err)
		}
	})

	t.Run("missing model id", func(t *testing.T) {
		service := NewService(&stubEmbedder{}, &stubRepo{}, "", make(chan EmbedChunkRequest))
		_, err := service.EmbedAndStore(context.Background(), baseReq)
		if err != ErrMissingModelID {
			t.Fatalf("expected %v, got %v", ErrMissingModelID, err)
		}
	})

	t.Run("request model id overrides default", func(t *testing.T) {
		embedder := &stubEmbedder{vectors: [][]float32{{0.1}}, dim: 1}
		repo := &stubRepo{existing: map[string]bool{}}
		service := NewService(embedder, repo, "model-default", make(chan EmbedChunkRequest))
		modelID := "model-override"
		req := baseReq
		req.ModelID = &modelID

		results, err := service.EmbedAndStore(context.Background(), req)
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(results) != 1 {
			t.Fatalf("expected 1 embedding, got %d", len(results))
		}
		if results[0].ModelID != "model-override" {
			t.Fatalf("expected model-override, got %s", results[0].ModelID)
		}
	})
}

func TestServiceRunProcessesChunks(t *testing.T) {
	embedder := &stubEmbedder{vectors: [][]float32{{0.1}}, dim: 1}
	repo := &stubRepo{existing: map[string]bool{}}
	queue := make(chan EmbedChunkRequest, 1)
	service := NewService(embedder, repo, "model-default", queue)

	queue <- EmbedChunkRequest{
		KnowledgeBaseID: "kb-1",
		Chunk: ChunkInput{
			ChunkID:     "1",
			Content:     "hello",
			ContentHash: "hash-hello",
		},
	}
	close(queue)

	if err := service.Run(context.Background()); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(repo.saved) != 1 {
		t.Fatalf("expected 1 embedding saved, got %d", len(repo.saved))
	}
	if repo.saved[0].KnowledgeBaseID != "kb-1" {
		t.Fatalf("expected kb-1, got %s", repo.saved[0].KnowledgeBaseID)
	}
}

func TestServiceEnqueueChunkAndWait(t *testing.T) {
	embedder := &stubEmbedder{vectors: [][]float32{{0.1}}, dim: 1}
	repo := &stubRepo{existing: map[string]bool{}}
	queue := make(chan EmbedChunkRequest, 1)
	service := NewService(embedder, repo, "model-default", queue)

	done := make(chan error, 1)
	go func() {
		done <- service.Run(context.Background())
	}()

	result, err := service.EnqueueChunkAndWait(context.Background(), EmbedChunkRequest{
		KnowledgeBaseID: "kb-1",
		Chunk: ChunkInput{
			ChunkID:     "1",
			Content:     "hello",
			ContentHash: "hash-hello",
		},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result.EmbeddingID == "" {
		t.Fatalf("expected embedding id to be set")
	}
	if result.ChunkID != "1" {
		t.Fatalf("expected chunk id 1, got %s", result.ChunkID)
	}

	close(queue)
	if err := <-done; err != nil {
		t.Fatalf("expected nil run error, got %v", err)
	}
}
