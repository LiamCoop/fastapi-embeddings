package markdown

import (
	"fmt"
	"strings"
	"testing"

	md "ragtime-backend/internal/markdown"
)

func TestSplitOversized_CodeFencePreservesFenceAndLanguage(t *testing.T) {
	var body strings.Builder
	for i := 0; i < 80; i++ {
		body.WriteString(fmt.Sprintf("fmt.Println(%d)\n", i))
	}
	block := md.Block{
		Type:    md.BlockCodeFence,
		Lang:    "go",
		Content: "```go\n" + body.String() + "```",
	}

	parts := SplitOversized(block, 40, md.BiasBalanced)
	if len(parts) < 2 {
		t.Fatalf("expected split code fence, got %d part(s)", len(parts))
	}
	for i, p := range parts {
		if p.Type != md.BlockCodeFence {
			t.Fatalf("part %d type mismatch: %v", i, p.Type)
		}
		if p.Lang != "go" {
			t.Fatalf("part %d language mismatch: %q", i, p.Lang)
		}
		if !strings.HasPrefix(p.Content, "```go\n") {
			t.Fatalf("part %d missing opening fence", i)
		}
		if !strings.HasSuffix(strings.TrimSpace(p.Content), "```") {
			t.Fatalf("part %d missing closing fence", i)
		}
	}
}

func TestSplitOversized_TableRepeatsHeaderRows(t *testing.T) {
	var rows strings.Builder
	for i := 0; i < 40; i++ {
		rows.WriteString(fmt.Sprintf("| r%d | value%d |\n", i, i))
	}
	block := md.Block{
		Type: md.BlockTable,
		Content: "| h1 | h2 |\n" +
			"| --- | --- |\n" +
			rows.String(),
	}

	parts := SplitOversized(block, 35, md.BiasBalanced)
	if len(parts) < 2 {
		t.Fatalf("expected split table, got %d part(s)", len(parts))
	}
	for i, p := range parts {
		if !strings.Contains(p.Content, "| h1 | h2 |\n| --- | --- |") {
			t.Fatalf("part %d missing repeated table header", i)
		}
	}
}

func TestSplitOversized_ListSplitsByItems(t *testing.T) {
	var list strings.Builder
	for i := 0; i < 50; i++ {
		list.WriteString(fmt.Sprintf("- item %d\n", i))
	}
	block := md.Block{
		Type:    md.BlockList,
		Content: strings.TrimSpace(list.String()),
	}

	parts := SplitOversized(block, 30, md.BiasBalanced)
	if len(parts) < 2 {
		t.Fatalf("expected list to split, got %d part(s)", len(parts))
	}
	for i, p := range parts {
		lines := strings.Split(strings.TrimSpace(p.Content), "\n")
		if len(lines) == 0 || !strings.HasPrefix(strings.TrimSpace(lines[0]), "- ") {
			t.Fatalf("part %d does not start with list item", i)
		}
	}
}

