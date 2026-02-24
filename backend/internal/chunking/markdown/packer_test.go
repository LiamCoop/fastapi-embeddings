package markdown

import (
	"strings"
	"testing"

	md "ragtime-backend/internal/markdown"
)

func TestPackBlocks_RespectsMaxTokens(t *testing.T) {
	opts := DefaultMarkdownOptions()
	opts.MaxTokens = 20
	opts.TargetTokens = 20
	opts.MinTokens = 0
	opts.OverlapTokens = 0

	blocks := []md.Block{
		{Type: md.BlockHeading, Content: "Title", Level: 1},
		{Type: md.BlockParagraph, Content: strings.Repeat("alpha ", 40)},
		{Type: md.BlockParagraph, Content: strings.Repeat("beta ", 40)},
	}

	chunks := packBlocks(blocks, opts)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}
	for i, ch := range chunks {
		if ch.estTokens > opts.MaxTokens {
			t.Fatalf("chunk %d exceeds max tokens: %d > %d", i, ch.estTokens, opts.MaxTokens)
		}
	}
}

func TestComputeOverlap_SkipsFrontmatter(t *testing.T) {
	prev := []indexedBlock{
		{index: 0, block: md.Block{Type: md.BlockFrontmatter, Content: "---\na: b\n---"}},
		{index: 1, block: md.Block{Type: md.BlockParagraph, Content: "useful trailing paragraph"}},
	}

	overlap := computeOverlap(prev, 1000, md.BiasBalanced)
	if len(overlap) != 1 {
		t.Fatalf("expected one overlap block, got %d", len(overlap))
	}
	if overlap[0].block.Type != md.BlockParagraph {
		t.Fatalf("expected paragraph overlap, got %v", overlap[0].block.Type)
	}
}

func TestComputeOverlap_DropsHeadingOnlySuffix(t *testing.T) {
	prev := []indexedBlock{
		{index: 0, block: md.Block{Type: md.BlockHeading, Content: "Section", Level: 2}},
	}

	overlap := computeOverlap(prev, 1000, md.BiasBalanced)
	if len(overlap) != 0 {
		t.Fatalf("expected heading-only overlap to be dropped, got %d blocks", len(overlap))
	}
}

