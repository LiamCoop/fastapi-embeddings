# Recursive Chunking Contract Requirements

## Purpose
This document defines the minimal recursive chunking contract. It mirrors the style of the fixed-size contract and clarifies the expectations for separator-based splitting.

## Requirements

### Input Validation
- **`max_runes` required**: `max_runes` must be greater than 0.
- **`overlap_runes` non-negative**: `overlap_runes` must be 0 or greater.
- **`overlap_runes` bounds**: `overlap_runes` must be strictly less than `max_runes`.

### Separator Behavior
- Separators are applied in priority order.
- If a separator produces chunks larger than `max_runes`, the next separator is applied recursively to those chunks.
- If no separators match, the chunker falls back to fixed-size splitting.
- Separator matches are included in the preceding chunk to preserve original text.

### Deterministic Output
- Same input string and options must always produce the same chunk boundaries and content.
- Chunk boundaries are measured in runes (not bytes) to handle Unicode safely.

### Chunk Sizing
- Each base chunk must contain at least 1 rune of content.
- Each base chunk must contain **at most** `max_runes` runes.
- Overlap, when enabled, may cause the returned chunk length to exceed `max_runes` by up to `overlap_runes`.

### Overlap Behavior
- For `overlap_runes > 0`, each chunk after the first extends backwards by up to `overlap_runes`.
- Overlap never extends before the start of the previous chunk.

### Offset and Content Integrity
- `StartRune` and `EndRune` must be valid offsets in the original text.
- `Content` must equal the substring defined by `[StartRune, EndRune)`.

### Empty Input
- Empty input returns an empty chunk list (no error).
