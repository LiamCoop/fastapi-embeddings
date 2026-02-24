# Markdown Chunker — High-Level Design Spec (Revised)

Note: this is the target/high-level design document. For current behavior implemented in code, see `docs/chunking-markdown-implementation.md`.

## Goals

Design a fast, flexible chunker for Markdown/MDX docs that:

- Preserves semantic structure (headings, paragraphs, lists, code, tables)
- Produces chunks sized for LLM context windows
- Uses fast token estimation heuristics (no tokenizer required)
- Behaves deterministically (same input → same chunks)
- Emits metadata that improves retrieval quality and debugging

---

## Definitions

### Block
A *block* is the smallest unit the chunker will *prefer* not to split:
- Heading
- Paragraph
- List (contiguous list region)
- Code fence (``` … ```)
- Table (contiguous table region)
- Blockquote (optional)
- Frontmatter (YAML)
- MDX imports/exports/components (MDX mode)

### Chunk
A chunk is a sequence of blocks packed to fit a token budget, plus optional overlap.

---

## Token Estimation Heuristics

Use estimated tokens for budgeting:

- Prose-ish text: `tokens ≈ chars / 4`
- Code-ish text (code fences, MDX JS snippets): `tokens ≈ chars / 2.5–3`

The estimator must be:
- O(n) in text length
- Stable / deterministic
- Optionally configurable via a single “bias” knob (see Product Surface)

---

## Chunk Packing Rules (Deterministic)

### Primary budgets
- `target_tokens` (soft goal)
- `max_tokens` (hard ceiling; chunk must not exceed unless forced by oversized block fallback)
- `min_tokens` (undersized chunk handling rule applies)

### Normal packing
1. Build a stream of blocks with associated estimated token counts.
2. Append blocks to the current chunk until the next block would exceed `max_tokens`.
3. If next block would exceed `max_tokens`, finalize current chunk and start the next with overlap (defined below).

---

## Oversized Block Policy (What “absolutely necessary” means)

A block is **oversized** if its estimated tokens exceed `max_tokens` by itself.

**Rule:** oversized blocks MUST be split using a type-specific fallback, rather than emitting a chunk > `max_tokens`.

Fallback splitting by block type:
- Code fence: split by **lines** (preserve fence markers; keep language tag with each split piece)
- Paragraph/prose: split by **sentences** (fallback to whitespace-based split if sentence boundaries are unclear)
- List: split by **top-level items** (preserve item text; avoid splitting inside an item unless item itself is oversized)
- Table: split by **rows** while repeating the header row in each split part
- Frontmatter: usually not oversized; if it is, split by lines

This is the only situation where splitting “mid-block” is allowed:
> **Mid-block splitting is allowed only when a single block exceeds `max_tokens`.**

---

## Overlap Mechanism (Concrete Rule)

Overlap is defined by `overlap_tokens` and operates on **blocks**.

**Default overlap rule:**
- When starting chunk N+1, carry over the **largest suffix** of blocks from chunk N whose total estimated tokens ≤ `overlap_tokens`,
- BUT never start with a heading-only overlap (prefer at least one content block if available).

Optional refinement (recommended):
- Prefer overlap blocks of types: paragraph > list > code > table > heading > blank
- Avoid duplicating frontmatter in overlap

This makes overlap deterministic and implementable.

---

## Heading-Aware Strategy (Default) + Hybrid Handling

### Default behavior: heading-aware chunking
- Headings create section boundaries and breadcrumb context.
- Within a section, blocks are packed into chunks by token budget.

### Hybrid rule for large sections
If a heading section’s body exceeds `max_tokens`:
- Do **not** discard heading semantics.
- Instead, apply **paragraph packing within that section** (i.e., pack the section’s blocks into multiple chunks),
- Each resulting chunk retains the same breadcrumb and section metadata.

This ensures large sections produce multiple chunks rather than one giant chunk or losing structure.

---

## Irregular Heading Hierarchies (Breadcrumb Rules)

Maintain a heading stack with explicit rules for malformed structures:

### When encountering heading level L
- Pop headings from the stack until the top has level < L
- Push the new heading onto the stack
- If the document jumps levels (e.g., ## → ####), allow it; the breadcrumb simply reflects the observed path

### Special case: new H1 (`#`)
- Treat as a new top-level document section
- Clear stack and start a new breadcrumb root for subsequent content

Breadcrumb metadata should be stable even when headings are “weird,” without trying to “fix” the author’s structure.

---

## Minimum Chunk Handling (Merging Rule)

If a finalized chunk has estimated tokens < `min_tokens`, apply:

1. If merging with the **next** chunk keeps total ≤ `max_tokens`, merge forward.
2. Else if merging with the **previous** chunk keeps total ≤ `max_tokens`, merge backward.
3. Else emit as-is (rare; typically occurs with strict `max_tokens` + oversized-block fragments).

This prevents tiny retrieval fragments while remaining deterministic.

---

## YAML Frontmatter Policy

Detect YAML frontmatter at the start of the document (`---` ... `---`).

Configurable modes:
- **Metadata mode (default):** parse frontmatter to populate metadata (title, tags, dates), and do not include it in chunk text.
- **Include mode:** treat frontmatter as a block at the beginning of the first chunk (no overlap).
- **Strip mode:** ignore entirely.

Recommended default: Metadata mode.

---

## MDX Support Policy

When MDX mode is enabled (for `.mdx`):
- Treat `import ...` / `export ...` statements at the top as a distinct block type.
- Treat JSX component blocks as “code-ish” for estimation.
- Preserve atomicity of JSX blocks similarly to code fences where possible.

Guiding principle:
> MDX constructs should not cause parse failure; if unsure, degrade to paragraph/code-ish blocks rather than error.

---

## Tables and Blockquotes (Clarification)

Tables:
- Default: treat contiguous tables as atomic blocks.
- If oversized: split by rows and repeat header.

Blockquotes:
- Optional support; if enabled, treat contiguous blockquotes as atomic blocks.
- If oversized: split by paragraph boundaries within the quote.

---

## Chunk Metadata (Clarified)

Each chunk should emit:
- `chunk_text`
- `est_tokens`
- `breadcrumb` (e.g., `Doc Title > H1 > H2 > H3`)
- `section_title` (most recent heading within depth limit)
- `doc_id` / `path` / `url` (if available)
- `block_range`: **block sequence indices** (start_block_index, end_block_index)
  - (Optional) also emit line offsets if cheap to track

---

## Product Surface Recommendations

### Basic Mode (default)
- Chunk size: small / medium / large (maps to token budgets)
- Overlap: low / medium / high

### Advanced Mode
- Strategy: heading-aware / paragraph / hybrid
- Heading depth limit
- Oversized block policy toggles (enabled by default)
- Frontmatter mode
- MDX mode
- Token estimator bias:
  - prose-heavy (more conservative for prose)
  - balanced (default)
  - code-heavy (more conservative for code/MDX)

“Bias” means adjusting the divisors slightly (e.g., prose /4 → /3.6 or /4.4, code /2.7 → /2.4 or /3.0).

---

## Summary

This chunker:
- Splits by structure first (blocks)
- Packs blocks by a deterministic token estimate
- Handles oversized blocks with explicit fallbacks
- Defines overlap and minimum chunk merging clearly
- Supports YAML frontmatter and MDX without fragility
- Emits metadata that improves retrieval and debugging
