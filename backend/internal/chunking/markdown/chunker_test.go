package markdown

import "testing"

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
