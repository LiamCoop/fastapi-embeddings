package markdown

import "testing"

func TestParseBlocks_BasicMarkdown(t *testing.T) {
	input := "---\ntitle: Test\n---\n\n# Heading\n\nParagraph text.\n\n- one\n- two\n\n```go\nfmt.Println(\"x\")\n```\n\n| a | b |\n| - | - |\n| 1 | 2 |\n"
	blocks := ParseBlocks(input, false)
	if len(blocks) < 6 {
		t.Fatalf("expected at least 6 blocks, got %d", len(blocks))
	}
	if blocks[0].Type != BlockFrontmatter {
		t.Fatalf("expected first block frontmatter, got %v", blocks[0].Type)
	}
	if blocks[1].Type != BlockHeading || blocks[1].Level != 1 {
		t.Fatalf("expected heading block, got %+v", blocks[1])
	}
}

func TestParseBlocks_MDX(t *testing.T) {
	input := "import x from 'y'\n\n<Chart data={a} />\n"
	blocks := ParseBlocks(input, true)
	if len(blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(blocks))
	}
	if blocks[0].Type != BlockMDXImport {
		t.Fatalf("expected mdx import, got %v", blocks[0].Type)
	}
	if blocks[1].Type != BlockMDXComponent {
		t.Fatalf("expected mdx component, got %v", blocks[1].Type)
	}
}
