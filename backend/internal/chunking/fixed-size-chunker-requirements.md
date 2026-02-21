# Chunking Contract Requirements

## Purpose
This document defines the minimal fixed-size chunking contract we validate in tests. It intentionally mirrors the style of `backend/internal/embedding/contract-requirements.md` and captures the current behavior of the fixed-size chunker.

## Requirements

### Input Validation
- **`max_runes` required**: `max_runes` must be greater than 0.
- **`overlap_runes` non-negative**: `overlap_runes` must be 0 or greater.
- **`overlap_runes` bounds**: `overlap_runes` must be strictly less than `max_runes`.

### Deterministic Output
- Same input string and options must always produce the same chunk boundaries and content.
- Chunk boundaries are measured in runes (not bytes) to handle Unicode safely.

### Chunk Sizing
- Each chunk must contain at least 1 rune of content.
- Each chunk must contain **at most** `max_runes` runes.
- The final chunk may be smaller than `max_runes`.

### Overlap Behavior
- For `overlap_runes > 0`, each chunk after the first must start at:
  - `prev.EndRune - overlap_runes`
- Overlap is only applied between consecutive chunks and never exceeds the previous chunk’s size.

### Offset and Content Integrity
- `StartRune` and `EndRune` must be valid offsets in the original text.
- `Content` must equal the substring defined by `[StartRune, EndRune)`.

### Empty Input
- Empty input returns an empty chunk list (no error).

---

## Tests (Implemented)

### `TestFixedSizeChunker_InvalidOptions`
- Verifies input validation for invalid `max_runes` and `overlap_runes` values.

### `TestFixedSizeChunker_EmptyInput`
- Verifies empty input returns zero chunks and no error.

### `TestFixedSizeChunker_ChunkSizes`
- Verifies chunk sizing, overlap boundaries, rune offsets, and content integrity against a sample Markdown document.
