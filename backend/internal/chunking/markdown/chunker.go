package markdown

import (
	"strings"

	md "ragtime-backend/internal/markdown"
)

// Chunk is the markdown chunking package output prior to adapter conversion.
type Chunk struct {
	Index      int
	StartRune  int
	EndRune    int
	Content    string
	RuneLength int
	Metadata   map[string]any
}

type MarkdownChunker struct {
	Opts MarkdownOptions
}

func NewMarkdownChunker(opts MarkdownOptions) (MarkdownChunker, error) {
	if err := opts.Validate(); err != nil {
		return MarkdownChunker{}, err
	}
	return MarkdownChunker{Opts: opts}, nil
}

func (c MarkdownChunker) Chunk(text string) ([]Chunk, error) {
	if text == "" {
		return []Chunk{}, nil
	}

	blocks := md.ParseBlocks(text, c.Opts.MDX)
	frontmatter := map[string]any(nil)
	if len(blocks) > 0 && blocks[0].Type == md.BlockFrontmatter {
		switch c.Opts.FrontmatterMode {
		case FrontmatterMetadata:
			frontmatter = parseFrontmatter(blocks[0].Content)
			blocks = blocks[1:]
		case FrontmatterStrip:
			blocks = blocks[1:]
		case FrontmatterInclude:
			// keep as block in content
		}
	}

	packed := packBlocks(blocks, c.Opts)
	chunks := make([]Chunk, 0, len(packed))
	for i, p := range packed {
		content := joinBlocks(p.blocks)
		runes := len([]rune(content))
		meta := map[string]any{
			"breadcrumb":    p.breadcrumb,
			"section_title": p.sectionTitle,
			"est_tokens":    p.estTokens,
			"block_start":   p.blockStart,
			"block_end":     p.blockEnd,
		}
		if len(frontmatter) > 0 {
			meta["frontmatter"] = frontmatter
		}
		chunks = append(chunks, Chunk{
			Index:      i,
			StartRune:  0,
			EndRune:    runes,
			Content:    content,
			RuneLength: runes,
			Metadata:   meta,
		})
	}

	return chunks, nil
}

func joinBlocks(blocks []md.Block) string {
	parts := make([]string, 0, len(blocks))
	for _, b := range blocks {
		parts = append(parts, b.Content)
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func parseFrontmatter(content string) map[string]any {
	lines := strings.Split(content, "\n")
	out := make(map[string]any)
	var currentListKey string
	currentList := make([]string, 0)

	flushList := func() {
		if currentListKey != "" {
			items := make([]any, 0, len(currentList))
			for _, v := range currentList {
				items = append(items, v)
			}
			out[currentListKey] = items
			currentListKey = ""
			currentList = currentList[:0]
		}
	}

	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || line == "---" {
			continue
		}
		if strings.HasPrefix(line, "- ") && currentListKey != "" {
			currentList = append(currentList, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
			continue
		}
		flushList()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}
		if value == "" {
			currentListKey = key
			currentList = currentList[:0]
			continue
		}
		out[key] = strings.Trim(value, `"'`)
	}
	flushList()
	if len(out) == 0 {
		return nil
	}
	return out
}
