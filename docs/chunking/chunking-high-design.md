# Markdown Chunker — High-Level Design

## Context

The project currently uses a `FixedSizeChunker` (rune-window) for all ingested content. The goal is to add a `MarkdownChunker` that respects document structure per the spec in `docs/chunking-high-level-spec.md`. The underlying block-parsing and token-estimation utilities are extracted into a standalone `internal/markdown/` package so a future document-analyzer service (pre-chunking) can call them directly without touching chunking code.

---

## Package Structure

```
internal/
├── markdown/                        # NEW — reusable utilities (no chunking coupling)
│   ├── block.go                     # BlockType, Block, ParseBlocks()
│   ├── token.go                     # EstimateTokens(), TokenBias
│   └── breadcrumb.go                # HeadingStack, Breadcrumb(), SectionTitle()
│
├── chunking/
│   ├── fixed.go                     # MODIFY — add Metadata map[string]any to Chunk
│   ├── strategy.go                  # MODIFY — add StrategyMarkdown, register in NewChunker
│   └── markdown/                    # NEW — MarkdownChunker (uses internal/markdown)
│       ├── options.go               # MarkdownOptions, FrontmatterMode, defaults
│       ├── packer.go                # packBlocks() — block→chunk packing algorithm
│       ├── splitter.go              # SplitOversized() — type-specific fallback splits
│       └── chunker.go               # MarkdownChunker struct, implements chunking.Chunker
│
└── ingestion/
    └── service.go                   # MODIFY — merge chunk.Metadata; select chunker by doc type
```

---

## Key Types & Signatures

### `internal/markdown/block.go`
```go
type BlockType int
const (
    BlockFrontmatter BlockType = iota
    BlockHeading
    BlockParagraph
    BlockCodeFence
    BlockList
    BlockTable
    BlockBlockquote
    BlockMDXImport
    BlockMDXComponent
)

type Block struct {
    Type      BlockType
    Content   string
    Level     int    // heading depth 1–6; 0 for non-headings
    Lang      string // code fence language tag
    StartLine int
    EndLine   int
}

// ParseBlocks scans markdown text into a flat slice of structural blocks.
// Set mdx=true to enable MDX import/export/component detection.
// Uses a simple state-machine scanner; no external parser dependency.
func ParseBlocks(text string, mdx bool) []Block
```

### `internal/markdown/token.go`
```go
type TokenBias int
const (
    BiasBalanced TokenBias = iota  // prose /4, code /2.75
    BiasProse                       // prose /4.4, code /3.0
    BiasCode                        // prose /3.6, code /2.4
)

// EstimateTokens returns a fast O(n) token estimate for a block.
// Code-ish blocks (CodeFence, MDXImport, MDXComponent) use the code divisor.
func EstimateTokens(b Block, bias TokenBias) int
```

### `internal/markdown/breadcrumb.go`
```go
type HeadingStack struct { /* unexported */ }

func NewHeadingStack() *HeadingStack
// Update applies the heading level rules from the spec:
//   - pops entries with level >= L, then pushes the new heading
//   - new H1 clears the stack entirely
func (h *HeadingStack) Update(level int, title string)
func (h *HeadingStack) Breadcrumb() string     // e.g. "Doc Title > Section > Sub"
func (h *HeadingStack) SectionTitle() string   // most recent heading
func (h *HeadingStack) Clone() *HeadingStack   // snapshot for chunk metadata
```

### `internal/chunking/markdown/options.go`
```go
type FrontmatterMode int
const (
    FrontmatterMetadata FrontmatterMode = iota // parse to metadata, omit from chunk text
    FrontmatterInclude                          // include as first block, never overlap
    FrontmatterStrip                            // discard entirely
)

type MarkdownOptions struct {
    TargetTokens    int             // soft goal (default 750)
    MaxTokens       int             // hard ceiling (default 1000)
    MinTokens       int             // merge-forward threshold (default 200)
    OverlapTokens   int             // overlap budget in tokens (default 80)
    HeadingDepth    int             // deepest heading level to split on, e.g. 3 = up to ### (default 3)
    FrontmatterMode FrontmatterMode // default: FrontmatterMetadata
    MDX             bool            // enable MDX block detection
    Bias            markdown.TokenBias
}

func DefaultMarkdownOptions() MarkdownOptions
```

### `internal/chunking/markdown/packer.go`
```go
// packedChunk is an internal intermediate before conversion to chunking.Chunk
type packedChunk struct {
    blocks       []markdown.Block
    estTokens    int
    breadcrumb   string
    sectionTitle string
    blockStart   int // index of first block in source slice
    blockEnd     int // index of last block (inclusive)
}

// packBlocks implements the deterministic packing algorithm from the spec:
//   - heading-aware: headings update breadcrumb, may trigger section boundary
//   - hybrid: large sections fall back to paragraph packing within the section
//   - min-token merge: forward-then-backward merge for undersized chunks
func packBlocks(blocks []markdown.Block, opts MarkdownOptions) []packedChunk
```

### `internal/chunking/markdown/chunker.go`
```go
type MarkdownChunker struct {
    Opts MarkdownOptions
}

func NewMarkdownChunker(opts MarkdownOptions) (MarkdownChunker, error)

// Chunk implements chunking.Chunker.
// Each returned chunking.Chunk has Metadata populated with:
//   "breadcrumb", "section_title", "est_tokens", "block_start", "block_end"
//   (plus "frontmatter" map if FrontmatterMetadata mode and doc has frontmatter)
func (c MarkdownChunker) Chunk(text string) ([]chunking.Chunk, error)
```

---

## Existing File Modifications

### 1. `internal/chunking/fixed.go` — extend `Chunk`
Add `Metadata map[string]any` to the `Chunk` struct. No behavior change for existing chunkers (field is nil/zero).

```go
type Chunk struct {
    Index      int
    StartRune  int
    EndRune    int
    Content    string
    RuneLength int
    Metadata   map[string]any  // populated by structured chunkers; nil for fixed/recursive
}
```

### 2. `internal/chunking/strategy.go` — register MarkdownChunker
```go
const StrategyMarkdown Strategy = "markdown"

// In NewChunker():
case StrategyMarkdown:
    return markdown.NewMarkdownChunker(markdown.DefaultMarkdownOptions())
```

### 3. `internal/ingestion/service.go` — two small changes

**a. Document-type-aware chunker selection** (replaces hardcoded FixedSizeChunker):
```go
func (s *Service) chunkerForDoc(doc IngestDocumentRequest) chunking.Chunker {
    switch doc.DocumentType {
    case "md", "mdx":
        opts := mdchunking.DefaultMarkdownOptions()
        opts.MDX = doc.DocumentType == "mdx"
        c, _ := mdchunking.NewMarkdownChunker(opts)
        return c
    default:
        return s.chunker // existing fixed-size default
    }
}
```

**b. Metadata pass-through** in `processContent`:
```go
meta := map[string]any{
    "start_rune": chunk.StartRune,
    "end_rune":   chunk.EndRune,
}
for k, v := range chunk.Metadata {
    meta[k] = v
}
```

---

## Block Parser Design (no external deps)

The block scanner uses a single-pass state machine over lines:
- `---` at line 0 → frontmatter state until closing `---`
- `` ``` `` → code fence state until closing `` ``` ``
- `|` prefix → table accumulation
- `^#+ ` → heading (extract level + title)
- `^[-*+] ` or `^\d+\. ` → list accumulation
- `^> ` → blockquote (if enabled)
- Blank line → flush current paragraph
- MDX mode: `^import ` / `^export ` / `^<[A-Z]` → MDX blocks

All block types are **atomic** by default; oversized fallback is handled in `splitter.go`.

---

## Splitter (oversized fallback)

```go
// SplitOversized breaks a single block that exceeds maxTokens into sub-blocks
// using type-specific rules from the spec.
func SplitOversized(b markdown.Block, maxTokens int, bias markdown.TokenBias) []markdown.Block
```

- **CodeFence**: split by lines; repeat opening fence + language tag on each piece
- **Paragraph/Blockquote**: split by sentences (period/question/exclamation + space); fallback to whitespace
- **List**: split by top-level items
- **Table**: split by rows; repeat header row on each piece
- **Frontmatter**: split by lines (rare edge case)

---

## Overlap Mechanism

```go
// computeOverlap returns the largest suffix of prevBlocks whose total estimated
// tokens ≤ overlapTokens, skipping heading-only suffixes and frontmatter blocks.
func computeOverlap(prevBlocks []markdown.Block, overlapTokens int, bias markdown.TokenBias) []markdown.Block
```

---

## How It Composes (Call Flow)

```
MarkdownChunker.Chunk(text)
  │
  ├─ markdown.ParseBlocks(text, mdx)       → []Block
  │
  ├─ (frontmatter handling per mode)
  │
  ├─ packBlocks(blocks, opts)
  │   ├─ HeadingStack.Update() per heading block
  │   ├─ EstimateTokens() per block
  │   ├─ SplitOversized() when block > MaxTokens
  │   ├─ computeOverlap() when starting a new chunk
  │   └─ forward/backward merge for MinTokens
  │
  └─ convert []packedChunk → []chunking.Chunk
      (Content = joined block text, Metadata = breadcrumb, section_title, est_tokens, block range)
```

Each utility function (`ParseBlocks`, `EstimateTokens`, `NewHeadingStack`) is independently importable by any other service (e.g. a future document analyzer doing pre-chunking structure analysis).

---

## Testing Plan

| File | What to test |
|---|---|
| `markdown/block_test.go` | Frontmatter detection, heading levels, code fences (nested, lang tags), lists, tables, blank-line paragraphs, MDX blocks, irregular inputs |
| `markdown/token_test.go` | EstimateTokens values for prose vs code blocks across BiasBalanced/Prose/Code |
| `markdown/breadcrumb_test.go` | Normal hierarchy, level-skipping, H1 reset, clone isolation |
| `chunking/markdown/packer_test.go` | Packing stops at MaxTokens, min-token merge forward/backward, large section hybrid fallback |
| `chunking/markdown/splitter_test.go` | Each block type oversized fallback; header row repetition for tables; fence markers preserved for code |
| `chunking/markdown/chunker_test.go` | Determinism (same input → same output), frontmatter modes, MDX mode toggle, metadata fields present and correct, empty input |
| `ingestion/service_test.go` (existing) | No regression; md/mdx doc types route to MarkdownChunker |

---

## Files to Create / Modify

| Action | Path |
|---|---|
| CREATE | `internal/markdown/block.go` |
| CREATE | `internal/markdown/token.go` |
| CREATE | `internal/markdown/breadcrumb.go` |
| CREATE | `internal/chunking/markdown/options.go` |
| CREATE | `internal/chunking/markdown/packer.go` |
| CREATE | `internal/chunking/markdown/splitter.go` |
| CREATE | `internal/chunking/markdown/chunker.go` |
| MODIFY | `internal/chunking/fixed.go` (add `Metadata` to `Chunk`) |
| MODIFY | `internal/chunking/strategy.go` (add `StrategyMarkdown`) |
| MODIFY | `internal/ingestion/service.go` (doc-type chunker selection + metadata pass-through) |
