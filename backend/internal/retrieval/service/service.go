package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"

	"ragtime-backend/internal/embedding"
	"ragtime-backend/internal/retrieval"
	"ragtime-backend/internal/retrieval/cache"
)

const (
	defaultCandidateFloor = 50
	defaultCandidateCap   = 200
)

// Service orchestrates query embedding, hybrid search, and retrieval observability records.
type Service struct {
	cache         cache.Layer
	embedder      embedding.TextEmbedder
	now           func() time.Time
	defaultTopK   int
	defaultHybrid float64
}

func New(cacheLayer cache.Layer, embedder embedding.TextEmbedder) *Service {
	return &Service{
		cache:         cacheLayer,
		embedder:      embedder,
		now:           func() time.Time { return time.Now().UTC() },
		defaultTopK:   retrieval.DefaultTopK,
		defaultHybrid: retrieval.DefaultHybridWeight,
	}
}

func (s *Service) Retrieve(ctx context.Context, req retrieval.Request) (*retrieval.Response, error) {
	if s.cache == nil {
		return nil, retrieval.ErrNilRepository
	}
	if s.embedder == nil {
		return nil, retrieval.ErrNilEmbedder
	}

	applyDefaults(&req, s.defaultTopK, s.defaultHybrid)
	if err := retrieval.ValidateRequest(req); err != nil {
		return nil, err
	}

	start := s.now()
	requestID := uuid.NewString()

	filterPayload := buildFilterPayload(req.Filters)
	_, err := s.cache.InsertRetrievalRequest(ctx, retrieval.RetrievalRequestRecord{
		ID:            requestID,
		KnowledgeBase: req.KnowledgeBaseID,
		Query:         req.Query,
		Filters:       filterPayload,
		TopK:          req.TopK,
		HybridWeight:  req.HybridWeight,
		ResultCount:   0,
		LatencyMS:     0,
		EmptyResult:   false,
		CreatedAt:     start,
	})
	if err != nil {
		return nil, err
	}

	embeddings, dim, err := s.embedder.EmbedTexts(ctx, []string{req.Query})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("embedding service returned no vectors")
	}

	searchParams := retrieval.SearchParams{
		KnowledgeBaseID: req.KnowledgeBaseID,
		Query:           req.Query,
		QueryVector:     embeddings[0],
		VectorDimension: dim,
		DocumentType:    req.Filters.DocumentType,
		PathPrefix:      normalizePathPrefix(req.Filters.PathPrefix),
		Source:          req.Filters.Source,
		TagsFilter:      buildTagsFilter(req.Filters.Tags),
		CreatedAfter:    req.Filters.CreatedAfter,
		CreatedBefore:   req.Filters.CreatedBefore,
		Limit:           candidateLimit(req.TopK),
	}

	semantic, err := s.cache.SearchSemantic(ctx, searchParams)
	if err != nil {
		return nil, err
	}
	lexical, err := s.cache.SearchLexical(ctx, searchParams)
	if err != nil {
		return nil, err
	}

	semanticScores := normalizeScores(semantic)
	lexicalScores := normalizeScores(lexical)

	merged := mergeScores(semanticScores, lexicalScores, req.HybridWeight)
	sortResults(merged)

	if len(merged) > req.TopK {
		merged = merged[:req.TopK]
	}

	chunkIDs := make([]string, 0, len(merged))
	for _, item := range merged {
		chunkIDs = append(chunkIDs, item.ChunkID)
	}

	chunks, err := s.cache.GetChunksWithDocuments(ctx, chunkIDs)
	if err != nil {
		return nil, err
	}

	chunkMap := make(map[string]retrieval.ChunkRecord, len(chunks))
	for _, chunk := range chunks {
		chunkMap[chunk.ChunkID] = chunk
	}

	results := make([]retrieval.Result, 0, len(merged))
	resultRecords := make([]retrieval.RetrievalResultRecord, 0, len(merged))

	for i, item := range merged {
		chunk, ok := chunkMap[item.ChunkID]
		if !ok {
			continue
		}

		citation := buildCitation(chunk)
		results = append(results, retrieval.Result{
			ChunkID:           chunk.ChunkID,
			DocumentID:        chunk.DocumentID,
			DocumentVersionID: chunk.DocumentVersionID,
			DocumentPath:      chunk.DocumentPath,
			DocumentTitle:     chunk.DocumentTitle,
			DocumentType:      chunk.DocumentType,
			Content:           chunk.Content,
			Metadata:          chunk.Metadata,
			Scores:            item.Score,
			Citation:          citation,
		})

		resultRecords = append(resultRecords, retrieval.RetrievalResultRecord{
			ID:                 uuid.NewString(),
			RetrievalRequestID: requestID,
			ChunkID:            chunk.ChunkID,
			Rank:               i + 1,
			SemanticScore:      item.Score.Semantic,
			LexicalScore:       item.Score.Lexical,
			FinalScore:         item.Score.Final,
			CreatedAt:          s.now(),
		})
	}

	latency := s.now().Sub(start).Milliseconds()
	emptyResult := len(results) == 0

	if err := s.cache.InsertRetrievalResults(ctx, resultRecords); err != nil {
		return nil, err
	}
	if err := s.cache.UpdateRetrievalRequest(ctx, requestID, len(results), latency, emptyResult); err != nil {
		return nil, err
	}

	return &retrieval.Response{
		RequestID:       requestID,
		KnowledgeBaseID: req.KnowledgeBaseID,
		Query:           req.Query,
		TopK:            req.TopK,
		HybridWeight:    req.HybridWeight,
		ResultCount:     len(results),
		LatencyMS:       latency,
		Results:         results,
	}, nil
}

func applyDefaults(req *retrieval.Request, topK int, weight float64) {
	if req.TopK == 0 {
		req.TopK = topK
	}
	if !req.HybridWeightSet {
		req.HybridWeight = weight
	}
}

func candidateLimit(topK int) int {
	limit := topK * 5
	if limit < defaultCandidateFloor {
		limit = defaultCandidateFloor
	}
	if limit > defaultCandidateCap {
		limit = defaultCandidateCap
	}
	return limit
}

func normalizePathPrefix(prefix *string) *string {
	if prefix == nil {
		return nil
	}
	value := *prefix
	if value == "" {
		return nil
	}
	if value[len(value)-1] != '%' {
		value = value + "%"
	}
	return &value
}

func buildTagsFilter(tags []string) map[string]any {
	if len(tags) == 0 {
		return nil
	}
	return map[string]any{"tags": tags}
}

func buildFilterPayload(filters retrieval.Filters) map[string]any {
	payload := map[string]any{}
	if filters.DocumentType != nil {
		payload["document_type"] = *filters.DocumentType
	}
	if filters.PathPrefix != nil {
		payload["path_prefix"] = *filters.PathPrefix
	}
	if filters.Source != nil {
		payload["source"] = *filters.Source
	}
	if len(filters.Tags) > 0 {
		payload["tags"] = filters.Tags
	}
	if filters.CreatedAfter != nil {
		payload["created_after"] = filters.CreatedAfter.Format(time.RFC3339)
	}
	if filters.CreatedBefore != nil {
		payload["created_before"] = filters.CreatedBefore.Format(time.RFC3339)
	}
	return payload
}

type mergedScore struct {
	ChunkID string
	Score   retrieval.Score
}

func normalizeScores(items []retrieval.ScoredChunk) map[string]float64 {
	scores := make(map[string]float64, len(items))
	var max float64
	for _, item := range items {
		value := item.Score
		if value < 0 {
			value = 0
		}
		scores[item.ChunkID] = value
		if value > max {
			max = value
		}
	}
	if max <= 0 {
		return scores
	}
	for id, value := range scores {
		scores[id] = value / max
	}
	return scores
}

func mergeScores(semantic map[string]float64, lexical map[string]float64, weight float64) []mergedScore {
	merged := make([]mergedScore, 0, len(semantic)+len(lexical))
	seen := map[string]struct{}{}

	add := func(id string) {
		if _, ok := seen[id]; ok {
			return
		}
		seen[id] = struct{}{}
		sem := semantic[id]
		lex := lexical[id]
		final := (weight * sem) + ((1 - weight) * lex)
		merged = append(merged, mergedScore{
			ChunkID: id,
			Score: retrieval.Score{
				Semantic: sem,
				Lexical:  lex,
				Final:    final,
			},
		})
	}

	for id := range semantic {
		add(id)
	}
	for id := range lexical {
		add(id)
	}

	return merged
}

func sortResults(results []mergedScore) {
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score.Final != results[j].Score.Final {
			return results[i].Score.Final > results[j].Score.Final
		}
		if results[i].Score.Semantic != results[j].Score.Semantic {
			return results[i].Score.Semantic > results[j].Score.Semantic
		}
		if results[i].Score.Lexical != results[j].Score.Lexical {
			return results[i].Score.Lexical > results[j].Score.Lexical
		}
		return results[i].ChunkID < results[j].ChunkID
	})
}

func buildCitation(chunk retrieval.ChunkRecord) retrieval.Citation {
	startRune := extractInt(chunk.Metadata, "start_rune")
	endRune := extractInt(chunk.Metadata, "end_rune")
	runeLength := extractInt(chunk.Metadata, "rune_length")

	return retrieval.Citation{
		DocumentID:        chunk.DocumentID,
		DocumentVersionID: chunk.DocumentVersionID,
		Path:              chunk.DocumentPath,
		Title:             chunk.DocumentTitle,
		VersionNumber:     chunk.VersionNumber,
		ChunkSequence:     chunk.SequenceNumber,
		StartRune:         startRune,
		EndRune:           endRune,
		RuneLength:        runeLength,
	}
}

func extractInt(metadata map[string]any, key string) *int {
	if metadata == nil {
		return nil
	}
	value, ok := metadata[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case int:
		return &typed
	case int32:
		v := int(typed)
		return &v
	case int64:
		v := int(typed)
		return &v
	case float64:
		v := int(typed)
		return &v
	case json.Number:
		parsed, err := typed.Int64()
		if err != nil {
			return nil
		}
		v := int(parsed)
		return &v
	default:
		return nil
	}
}
