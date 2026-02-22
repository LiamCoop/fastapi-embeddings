package service

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
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

var (
	quotedPhrasePattern = regexp.MustCompile(`"[^"]+"|'[^']+'`)
	symbolPattern       = regexp.MustCompile(`[/._]|::|->`)
	camelSnakePattern   = regexp.MustCompile(`[a-z]+[A-Z][a-zA-Z0-9]*|[a-zA-Z]+_[a-zA-Z0-9_]+`)
	errorCodePattern    = regexp.MustCompile(`\b(?:[A-Z]{2,}[_-]?\d+|\d+\.\d+\.\d+|v\d+(?:\.\d+)*)\b`)
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

	profileEffective, semanticWeight, autoSignals := resolveProfileAndWeight(req)
	// Maintain legacy field semantics while adding explicit semantic weight controls.
	req.HybridWeight = semanticWeight

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

	merged := mergeScores(semanticScores, lexicalScores, semanticWeight)
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

		result := buildResult(chunk, item.Score)
		results = append(results, result)

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

	response := &retrieval.Response{
		RequestID:       requestID,
		QueryID:         requestID,
		IndexVersion:    "active-document-versions",
		KnowledgeBaseID: req.KnowledgeBaseID,
		Query:           req.Query,
		TopK:            req.TopK,
		HybridWeight:    semanticWeight,
		ResultCount:     len(results),
		LatencyMS:       latency,
		Results:         results,
		Passages:        results,
	}
	if req.Debug {
		response.Debug = &retrieval.DebugMetadata{
			RetrievalProfileEffective: profileEffective,
			SemanticWeightEffective:   semanticWeight,
			AutoSignalsDetected:       autoSignals,
			LexicalCandidates:         len(lexical),
			SemanticCandidates:        len(semantic),
			RerankerApplied:           false,
			FiltersApplied:            filterPayload,
		}
	}

	return response, nil
}

func (s *Service) Hydrate(ctx context.Context, req retrieval.HydrateRequest) (*retrieval.HydrateResponse, error) {
	if s.cache == nil {
		return nil, retrieval.ErrNilRepository
	}
	if err := retrieval.ValidateHydrateRequest(req); err != nil {
		return nil, err
	}

	baseChunks, err := s.cache.GetChunksWithDocumentsForKB(ctx, req.KnowledgeBaseID, req.ChunkIDs)
	if err != nil {
		return nil, err
	}

	chunkMap := map[string]retrieval.ChunkRecord{}
	for _, chunk := range baseChunks {
		chunkMap[chunk.ChunkID] = chunk
	}

	if req.AdjacentBefore > 0 || req.AdjacentAfter > 0 {
		for _, chunk := range baseChunks {
			start := chunk.SequenceNumber - int32(req.AdjacentBefore)
			if start < 0 {
				start = 0
			}
			end := chunk.SequenceNumber + int32(req.AdjacentAfter)

			adjacent, adjacentErr := s.cache.GetChunksByDocumentVersionRange(ctx, chunk.DocumentVersionID, start, end)
			if adjacentErr != nil {
				return nil, adjacentErr
			}

			for _, expanded := range adjacent {
				chunkMap[expanded.ChunkID] = expanded
			}
		}
	}

	chunks := make([]retrieval.ChunkRecord, 0, len(chunkMap))
	for _, chunk := range chunkMap {
		chunks = append(chunks, chunk)
	}
	sort.Slice(chunks, func(i, j int) bool {
		if chunks[i].DocumentPath != chunks[j].DocumentPath {
			return chunks[i].DocumentPath < chunks[j].DocumentPath
		}
		return chunks[i].SequenceNumber < chunks[j].SequenceNumber
	})

	results := make([]retrieval.Result, 0, len(chunks))
	for _, chunk := range chunks {
		results = append(results, buildResult(chunk, retrieval.Score{}))
	}

	return &retrieval.HydrateResponse{
		KnowledgeBaseID: req.KnowledgeBaseID,
		ChunkCount:      len(results),
		Chunks:          results,
	}, nil
}

func applyDefaults(req *retrieval.Request, topK int, weight float64) {
	if req.TopK == 0 {
		req.TopK = topK
	}
	if !req.HybridWeightSet {
		req.HybridWeight = weight
	}
	if strings.TrimSpace(req.RetrievalProfile) == "" {
		req.RetrievalProfile = retrieval.DefaultRetrievalProfile
	}
	if req.SemanticWeightSet {
		req.HybridWeight = req.SemanticWeight
		req.HybridWeightSet = true
	}
}

func resolveProfileAndWeight(req retrieval.Request) (string, float64, []string) {
	profile := strings.ToLower(strings.TrimSpace(req.RetrievalProfile))
	if profile == "" {
		profile = retrieval.RetrievalProfileAuto
	}

	if req.SemanticWeightSet {
		return profile, req.SemanticWeight, []string{"semantic_weight_override"}
	}
	if req.HybridWeightSet {
		return profile, req.HybridWeight, []string{"hybrid_weight_override"}
	}

	switch profile {
	case retrieval.RetrievalProfileExact:
		return profile, 0.2, nil
	case retrieval.RetrievalProfileBalanced:
		return profile, 0.5, nil
	case retrieval.RetrievalProfileSemantic:
		return profile, 0.8, nil
	default:
		autoProfile, signals := classifyAutoProfile(req.Query)
		switch autoProfile {
		case retrieval.RetrievalProfileExact:
			return autoProfile, 0.2, signals
		case retrieval.RetrievalProfileSemantic:
			return autoProfile, 0.8, signals
		default:
			return retrieval.RetrievalProfileBalanced, 0.5, signals
		}
	}
}

func classifyAutoProfile(query string) (string, []string) {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return retrieval.RetrievalProfileBalanced, nil
	}

	tokens := strings.Fields(trimmed)
	lower := strings.ToLower(trimmed)
	lexicalSignals := make([]string, 0)
	semanticSignals := make([]string, 0)

	if quotedPhrasePattern.MatchString(trimmed) {
		lexicalSignals = append(lexicalSignals, "quoted_phrase")
	}
	if symbolPattern.MatchString(trimmed) {
		lexicalSignals = append(lexicalSignals, "symbols")
	}
	if camelSnakePattern.MatchString(trimmed) {
		lexicalSignals = append(lexicalSignals, "identifier_tokens")
	}
	if errorCodePattern.MatchString(trimmed) {
		lexicalSignals = append(lexicalSignals, "error_or_version_pattern")
	}
	if len(tokens) <= 4 {
		lexicalSignals = append(lexicalSignals, "short_query")
	}

	if strings.HasPrefix(lower, "how ") || strings.HasPrefix(lower, "why ") || strings.HasPrefix(lower, "when ") || strings.HasPrefix(lower, "what ") {
		semanticSignals = append(semanticSignals, "question_form")
	}
	if len(tokens) >= 9 {
		semanticSignals = append(semanticSignals, "long_natural_language")
	}
	if !symbolPattern.MatchString(trimmed) && !camelSnakePattern.MatchString(trimmed) {
		semanticSignals = append(semanticSignals, "conversational_phrasing")
	}

	if len(lexicalSignals) > len(semanticSignals) {
		return retrieval.RetrievalProfileExact, lexicalSignals
	}
	if len(semanticSignals) > len(lexicalSignals) {
		return retrieval.RetrievalProfileSemantic, semanticSignals
	}

	signals := append([]string{}, lexicalSignals...)
	signals = append(signals, semanticSignals...)
	return retrieval.RetrievalProfileBalanced, dedupeStrings(signals)
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

func buildResult(chunk retrieval.ChunkRecord, score retrieval.Score) retrieval.Result {
	citation := buildCitation(chunk)
	offsets := &retrieval.Offsets{
		StartRune:  citation.StartRune,
		EndRune:    citation.EndRune,
		RuneLength: citation.RuneLength,
	}
	if offsets.StartRune == nil && offsets.EndRune == nil && offsets.RuneLength == nil {
		offsets = nil
	}
	sectionPath := extractStringSlice(chunk.Metadata, "section_path")

	return retrieval.Result{
		ChunkID:           chunk.ChunkID,
		DocumentID:        chunk.DocumentID,
		DocumentVersionID: chunk.DocumentVersionID,
		DocumentPath:      chunk.DocumentPath,
		DocumentTitle:     chunk.DocumentTitle,
		DocumentType:      chunk.DocumentType,
		Content:           chunk.Content,
		Metadata:          chunk.Metadata,
		Scores:            score,
		Citation:          citation,
		SourceURI:         chunk.DocumentPath,
		Title:             chunk.DocumentTitle,
		SectionPath:       sectionPath,
		Text:              chunk.Content,
		Score:             score.Final,
		ScoreDetail:       score,
		Offsets:           offsets,
	}
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

func extractStringSlice(metadata map[string]any, key string) []string {
	if metadata == nil {
		return nil
	}
	raw, ok := metadata[key]
	if !ok {
		return nil
	}

	switch typed := raw.(type) {
	case []string:
		if len(typed) == 0 {
			return nil
		}
		return typed
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			value, ok := item.(string)
			if !ok {
				continue
			}
			out = append(out, value)
		}
		if len(out) == 0 {
			return nil
		}
		return out
	default:
		return nil
	}
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
