# External Embedding Service Contract Requirements

This contract defines what the backend expects from the external embedding service and what the backend must enforce before calling it.

## Interface (MVP)

### Request

- Endpoint accepts a `knowledgebase_id` (KBID) and a list of chunks.
- Each chunk includes:
  - `chunk_id`
  - `content`
  - `content_hash`
  - optional `metadata`
- A `model_id` may be supplied, but the default model is used if missing.
- Internal support for multiple embedding models exists, but model selection is not exposed beyond the optional `model_id` field.

### Response

- Returns a list of embeddings, one per input chunk that is not skipped by dedup.
- Each returned embedding includes at least:
  - `embedding_id`
  - `chunk_id`
  - `knowledgebase_id`
  - `content_hash`
  - `model_id`
  - `vector`
  - `model_version`
  - `vector_dimension`

## Core Requirements

- **KBID required**: requests without a valid `knowledgebase_id` are rejected.
- **Chunk fields required**: each chunk must include `chunk_id`, `content`, and `content_hash`.
- **Deterministic output size**: the response list size equals the number of input chunks minus dedup-skipped chunks.
- **Dedup by KBID + ContentHash (+ model)**:
  - If an embedding already exists for the same `knowledgebase_id`, `content_hash`, `model_id`, and `model_version`, the backend must skip re-embedding it.
  - Skipped chunks are not sent to the service and do not appear in the returned list.
- **Default model behavior**: if `model_id` is absent, the active default model is used.
- **Model identity recorded**: each embedding records the `model_id` and `model_version` used.
- **Model version required**: the service requires a configured `model_version`.
- **Model ID required**: if no default model is configured, a request `model_id` must be provided.

## Suggested Test Cases

### Validation

- Missing `knowledgebase_id` returns a validation error.
- Chunk missing `chunk_id` returns a validation error.
- Chunk missing `content` returns a validation error.
- Empty chunk list returns an empty embedding list.
- Chunk missing `content_hash` returns a validation error.

### Model Selection

- No `model_id` uses the default model.
- Missing default model and no request `model_id` returns a validation error.
- Request `model_id` overrides the default model.

### Deduplication

- Existing embedding with same `knowledgebase_id`, `content_hash`, `model_id`, and `model_version` is skipped and not returned.
- Two chunks with identical `content_hash` in the same request result in a single embedding (second is skipped).
- Same `content_hash` but different `knowledgebase_id` results in separate embeddings.

### Output Integrity

- Returned list includes one embedding per non-skipped input chunk.
- Each embedding includes correct `knowledgebase_id`, `content_hash`, and `model_id`.
- Embedding vectors have the dimension specified by the selected model.
