# Markdown Chunking (Current Implementation)

This document describes how markdown chunking currently works in code today.

Related design doc: `docs/chunking-high-level-spec.md` (original target design).

## Which strategy is used

- For `.md` and `.mdx` documents, the chunking service auto-selects `markdown` strategy when:
  - requested strategy is empty, or
  - requested strategy is `fixed`.
- If strategy is explicitly set to `recursive`, markdown auto-selection is not applied.
- Non-markdown files use the requested/default non-markdown strategy.

Implementation reference:
- `backend/internal/chunking/service/service.go` (`resolveDocumentStrategy`, `isMarkdownURI`)

## Core approach

1. Parse document into structural blocks.
2. Optionally process YAML frontmatter.
3. Split only oversized blocks (type-specific fallback rules).
4. Pack blocks into chunks up to `max_tokens`.
5. Apply overlap between consecutive chunks.
6. Merge chunks that are too small (`min_tokens`) when possible.
7. Emit chunk metadata (`breadcrumb`, `section_title`, token estimate, block range).

Implementation references:
- `backend/internal/chunking/markdown/chunker.go`
- `backend/internal/chunking/markdown/packer.go`
- `backend/internal/chunking/markdown/splitter.go`

## Default markdown options

- `target_tokens`: 750
- `max_tokens`: 1000
- `min_tokens`: 200
- `overlap_tokens`: 80
- `heading_depth`: 3
- `frontmatter_mode`: metadata
- `mdx`: false
- `bias`: balanced

Implementation reference:
- `backend/internal/chunking/markdown/options.go`

## Block parsing rules

Blocks are parsed in a single pass from markdown text:

- YAML frontmatter: only if first line is `---` and closed by another `---`.
- Code fence: lines starting with triple-backticks; collected through closing fence.
- Heading: `#{1,6} + text`.
- List: contiguous list items (unordered or numbered), including indented continuation lines.
- Table: contiguous lines beginning with `|`.
- Blockquote: contiguous lines beginning with `>`.
- Paragraph: fallback for contiguous non-empty lines that are not matched by other block types.
- MDX (only when `mdx=true`): `import`/`export` blocks and component-like blocks (`<Component ...>`).

Implementation reference:
- `backend/internal/markdown/block.go`

## Token estimation

Token counts are estimated deterministically per block from rune length:

- Prose-like blocks: `ceil(chars / prose_divisor)`
- Code-like blocks (code fence, MDX import/component): `ceil(chars / code_divisor)`

Bias presets:

- `balanced`: prose `4.0`, code `2.75`
- `prose`: prose `4.4`, code `3.0`
- `code`: prose `3.6`, code `2.4`

Implementation reference:
- `backend/internal/markdown/token.go`

## Splitting rules (when a block is oversized)

A block is oversized when its own estimated tokens exceed `max_tokens`.

Type-specific splitting:

- Code fence: split by lines, preserving opening/closing fences and language tag.
- Paragraph and blockquote: split by sentence boundaries; fallback to whitespace words; final fallback line split.
- List: split by top-level items.
- Table: split by rows; repeats header (first two lines) in each split part.
- Frontmatter: split by lines.
- Other/unknown: prose split fallback.

Implementation reference:
- `backend/internal/chunking/markdown/splitter.go`

## Packing rules

- Blocks are packed in order until adding the next block would exceed `max_tokens`.
- Heading boundary rule:
  - if a heading level is within `heading_depth`, current chunk is finalized before that heading, and heading stack is updated.
- Chunk overlap:
  - next chunk starts with the largest suffix of prior chunk blocks within `overlap_tokens`.
  - frontmatter is never included in overlap.
  - heading-only overlap is dropped.
- Small chunk merge:
  - if chunk tokens `< min_tokens`, merge forward if within `max_tokens`, else merge backward if within `max_tokens`.

Implementation reference:
- `backend/internal/chunking/markdown/packer.go`

## Frontmatter handling

Modes:

- `metadata` (default): frontmatter parsed into key/value metadata and removed from chunk content.
- `include`: frontmatter remains as a content block.
- `strip`: frontmatter removed and not retained.

Current parser behavior:

- Supports simple `key: value`.
- Supports simple list form:
  - `key:`
  - `- item1`
  - `- item2`
- Does not implement full YAML parsing.

Implementation reference:
- `backend/internal/chunking/markdown/chunker.go` (`parseFrontmatter`)

## Metadata emitted per chunk

- `breadcrumb` (heading path)
- `section_title` (last heading in stack)
- `est_tokens`
- `block_start` (original block index)
- `block_end` (original block index)
- `frontmatter` (only when mode is metadata and parsed values exist)

Note: chunk adapter/service currently persists `start_rune`, `end_rune`, `rune_length` at chunk record level separately.

Implementation references:
- `backend/internal/chunking/markdown/chunker.go`
- `backend/internal/chunking/service/service.go`

## Known implementation notes

- `target_tokens` is validated and stored in options but not currently used by packing logic (packing uses `max_tokens`/`min_tokens`/`overlap_tokens`).
- Metadata generated by markdown chunker is currently replaced by service-level metadata when storing chunks (`start_rune`, `end_rune`, `rune_length` only).

