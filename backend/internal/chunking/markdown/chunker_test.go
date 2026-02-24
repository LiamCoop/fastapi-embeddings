package markdown

import (
	"strings"
	"testing"
)

func TestMarkdownChunker_Deterministic(t *testing.T) {
	opts := DefaultMarkdownOptions()
	opts.TargetTokens = 20
	opts.MaxTokens = 20
	opts.MinTokens = 0
	c, err := NewMarkdownChunker(opts)
	if err != nil {
		t.Fatalf("new chunker: %v", err)
	}

	input := "# Title\n\nFirst paragraph with enough text to split chunks.\n\nSecond paragraph also long enough to force packing."
	a, err := c.Chunk(input)
	if err != nil {
		t.Fatalf("chunk a: %v", err)
	}
	b, err := c.Chunk(input)
	if err != nil {
		t.Fatalf("chunk b: %v", err)
	}
	if len(a) != len(b) {
		t.Fatalf("chunk count mismatch: %d vs %d", len(a), len(b))
	}
	for i := range a {
		if a[i].Content != b[i].Content {
			t.Fatalf("chunk %d content mismatch", i)
		}
	}
}

func TestMarkdownChunker_FrontmatterMetadata(t *testing.T) {
	opts := DefaultMarkdownOptions()
	opts.FrontmatterMode = FrontmatterMetadata
	c, err := NewMarkdownChunker(opts)
	if err != nil {
		t.Fatalf("new chunker: %v", err)
	}

	chunks, err := c.Chunk("---\ntitle: Doc\ntags:\n- a\n- b\n---\n\n# H1\n\nBody")
	if err != nil {
		t.Fatalf("chunk: %v", err)
	}
	if len(chunks) == 0 {
		t.Fatalf("expected chunks")
	}
	fm, ok := chunks[0].Metadata["frontmatter"].(map[string]any)
	if !ok || fm["title"] != "Doc" {
		t.Fatalf("expected frontmatter metadata")
	}
	if chunks[0].Metadata["breadcrumb"] == "" {
		t.Fatalf("expected breadcrumb metadata")
	}
}

func TestMarkdownChunker_FrontmatterModes(t *testing.T) {
	input := "---\ntitle: Doc\nowner: Eng\n---\n\n# H1\n\nBody"

	t.Run("include", func(t *testing.T) {
		opts := DefaultMarkdownOptions()
		opts.FrontmatterMode = FrontmatterInclude
		c, err := NewMarkdownChunker(opts)
		if err != nil {
			t.Fatalf("new chunker: %v", err)
		}
		chunks, err := c.Chunk(input)
		if err != nil {
			t.Fatalf("chunk: %v", err)
		}
		if len(chunks) == 0 {
			t.Fatalf("expected chunks")
		}
		if !strings.Contains(chunks[0].Content, "title: Doc") {
			t.Fatalf("expected frontmatter in chunk content")
		}
		if _, ok := chunks[0].Metadata["frontmatter"]; ok {
			t.Fatalf("did not expect frontmatter metadata in include mode")
		}
	})

	t.Run("strip", func(t *testing.T) {
		opts := DefaultMarkdownOptions()
		opts.FrontmatterMode = FrontmatterStrip
		c, err := NewMarkdownChunker(opts)
		if err != nil {
			t.Fatalf("new chunker: %v", err)
		}
		chunks, err := c.Chunk(input)
		if err != nil {
			t.Fatalf("chunk: %v", err)
		}
		if len(chunks) == 0 {
			t.Fatalf("expected chunks")
		}
		if strings.Contains(chunks[0].Content, "title: Doc") {
			t.Fatalf("did not expect frontmatter in chunk content")
		}
		if _, ok := chunks[0].Metadata["frontmatter"]; ok {
			t.Fatalf("did not expect frontmatter metadata in strip mode")
		}
	})
}

func TestMarkdownChunker_MetadataFieldsPresent(t *testing.T) {
	opts := DefaultMarkdownOptions()
	opts.TargetTokens = 20
	opts.MaxTokens = 20
	opts.MinTokens = 0
	opts.OverlapTokens = 0
	c, err := NewMarkdownChunker(opts)
	if err != nil {
		t.Fatalf("new chunker: %v", err)
	}

	input := "# Title\n\nOne paragraph with enough content to split.\n\nAnother paragraph with enough content to split too."
	chunks, err := c.Chunk(input)
	if err != nil {
		t.Fatalf("chunk: %v", err)
	}
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(chunks))
	}

	for i, ch := range chunks {
		if ch.Metadata["breadcrumb"] == nil {
			t.Fatalf("chunk %d missing breadcrumb", i)
		}
		if ch.Metadata["section_title"] == nil {
			t.Fatalf("chunk %d missing section_title", i)
		}
		if _, ok := ch.Metadata["est_tokens"].(int); !ok {
			t.Fatalf("chunk %d est_tokens not int", i)
		}
		if _, ok := ch.Metadata["block_start"].(int); !ok {
			t.Fatalf("chunk %d block_start not int", i)
		}
		if _, ok := ch.Metadata["block_end"].(int); !ok {
			t.Fatalf("chunk %d block_end not int", i)
		}
	}
}

func TestMarkdownChunker_MDXToggleChangesTokenEstimate(t *testing.T) {
	input := "<Widget prop=\"x\" />"

	base := DefaultMarkdownOptions()
	base.TargetTokens = 200
	base.MaxTokens = 200
	base.MinTokens = 0
	base.OverlapTokens = 0

	optsNoMDX := base
	optsNoMDX.MDX = false
	cNoMDX, err := NewMarkdownChunker(optsNoMDX)
	if err != nil {
		t.Fatalf("new chunker (mdx=false): %v", err)
	}
	noMDXChunks, err := cNoMDX.Chunk(input)
	if err != nil {
		t.Fatalf("chunk (mdx=false): %v", err)
	}

	optsMDX := base
	optsMDX.MDX = true
	cMDX, err := NewMarkdownChunker(optsMDX)
	if err != nil {
		t.Fatalf("new chunker (mdx=true): %v", err)
	}
	mdxChunks, err := cMDX.Chunk(input)
	if err != nil {
		t.Fatalf("chunk (mdx=true): %v", err)
	}

	if len(noMDXChunks) != 1 || len(mdxChunks) != 1 {
		t.Fatalf("expected one chunk in each mode")
	}

	noMDXEst, ok := noMDXChunks[0].Metadata["est_tokens"].(int)
	if !ok {
		t.Fatalf("mdx=false est_tokens not int")
	}
	mdxEst, ok := mdxChunks[0].Metadata["est_tokens"].(int)
	if !ok {
		t.Fatalf("mdx=true est_tokens not int")
	}
	if mdxEst <= noMDXEst {
		t.Fatalf("expected mdx=true to estimate more tokens for MDX component: got mdx=%d non-mdx=%d", mdxEst, noMDXEst)
	}
}

func TestMarkdownChunker_EmptyInput(t *testing.T) {
	c, err := NewMarkdownChunker(DefaultMarkdownOptions())
	if err != nil {
		t.Fatalf("new chunker: %v", err)
	}
	chunks, err := c.Chunk("")
	if err != nil {
		t.Fatalf("chunk: %v", err)
	}
	if len(chunks) != 0 {
		t.Fatalf("expected no chunks")
	}
}
