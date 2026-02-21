package markdown

import (
	"regexp"
	"strings"

	md "ragtime-backend/internal/markdown"
)

var listItemPattern = regexp.MustCompile(`^\s*(?:[-*+]\s+|\d+\.\s+)`)

// SplitOversized breaks a single oversized block into sub-blocks.
func SplitOversized(b md.Block, maxTokens int, bias md.TokenBias) []md.Block {
	if maxTokens <= 0 || md.EstimateTokens(b, bias) <= maxTokens {
		return []md.Block{b}
	}

	switch b.Type {
	case md.BlockCodeFence:
		return splitCodeFence(b, maxTokens, bias)
	case md.BlockParagraph, md.BlockBlockquote:
		return splitProse(b, maxTokens, bias)
	case md.BlockList:
		return splitList(b, maxTokens, bias)
	case md.BlockTable:
		return splitTable(b, maxTokens, bias)
	case md.BlockFrontmatter:
		return splitByLines(b, maxTokens, bias)
	default:
		return splitProse(b, maxTokens, bias)
	}
}

func splitCodeFence(b md.Block, maxTokens int, bias md.TokenBias) []md.Block {
	lines := strings.Split(b.Content, "\n")
	if len(lines) < 3 {
		return splitByLines(b, maxTokens, bias)
	}
	open := lines[0]
	close := lines[len(lines)-1]
	body := lines[1 : len(lines)-1]

	parts := make([]md.Block, 0, 4)
	current := make([]string, 0, 16)
	for _, line := range body {
		candidate := append(append([]string{}, current...), line)
		content := open + "\n" + strings.Join(candidate, "\n") + "\n" + close
		test := b
		test.Content = content
		if len(current) > 0 && md.EstimateTokens(test, bias) > maxTokens {
			parts = append(parts, md.Block{Type: b.Type, Content: open + "\n" + strings.Join(current, "\n") + "\n" + close, Lang: b.Lang, StartLine: b.StartLine, EndLine: b.EndLine})
			current = []string{line}
			continue
		}
		current = candidate
	}
	if len(current) > 0 {
		parts = append(parts, md.Block{Type: b.Type, Content: open + "\n" + strings.Join(current, "\n") + "\n" + close, Lang: b.Lang, StartLine: b.StartLine, EndLine: b.EndLine})
	}
	if len(parts) == 0 {
		return []md.Block{b}
	}
	return parts
}

func splitProse(b md.Block, maxTokens int, bias md.TokenBias) []md.Block {
	parts := splitSentences(b.Content)
	if len(parts) <= 1 {
		return splitByLines(b, maxTokens, bias)
	}

	out := make([]md.Block, 0, 4)
	current := ""
	for _, part := range parts {
		candidate := strings.TrimSpace(strings.TrimSpace(current + " " + part))
		test := b
		test.Content = candidate
		if current != "" && md.EstimateTokens(test, bias) > maxTokens {
			out = append(out, md.Block{Type: b.Type, Content: current, StartLine: b.StartLine, EndLine: b.EndLine})
			current = strings.TrimSpace(part)
			continue
		}
		current = candidate
	}
	if current != "" {
		out = append(out, md.Block{Type: b.Type, Content: current, StartLine: b.StartLine, EndLine: b.EndLine})
	}
	if len(out) == 0 {
		return []md.Block{b}
	}
	return out
}

func splitList(b md.Block, maxTokens int, bias md.TokenBias) []md.Block {
	lines := strings.Split(b.Content, "\n")
	items := make([]string, 0)
	current := make([]string, 0)
	for _, line := range lines {
		if listItemPattern.MatchString(line) {
			if len(current) > 0 {
				items = append(items, strings.Join(current, "\n"))
			}
			current = []string{line}
			continue
		}
		if len(current) == 0 {
			current = []string{line}
			continue
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		items = append(items, strings.Join(current, "\n"))
	}
	if len(items) <= 1 {
		return splitByLines(b, maxTokens, bias)
	}

	out := make([]md.Block, 0, 4)
	curItems := make([]string, 0)
	for _, item := range items {
		candidateItems := append(append([]string{}, curItems...), item)
		test := b
		test.Content = strings.Join(candidateItems, "\n")
		if len(curItems) > 0 && md.EstimateTokens(test, bias) > maxTokens {
			out = append(out, md.Block{Type: b.Type, Content: strings.Join(curItems, "\n"), StartLine: b.StartLine, EndLine: b.EndLine})
			curItems = []string{item}
			continue
		}
		curItems = candidateItems
	}
	if len(curItems) > 0 {
		out = append(out, md.Block{Type: b.Type, Content: strings.Join(curItems, "\n"), StartLine: b.StartLine, EndLine: b.EndLine})
	}
	if len(out) == 0 {
		return []md.Block{b}
	}
	return out
}

func splitTable(b md.Block, maxTokens int, bias md.TokenBias) []md.Block {
	lines := strings.Split(b.Content, "\n")
	if len(lines) <= 2 {
		return splitByLines(b, maxTokens, bias)
	}
	head := lines[:2]
	rows := lines[2:]

	out := make([]md.Block, 0, 4)
	curRows := make([]string, 0)
	for _, row := range rows {
		candidateRows := append(append([]string{}, curRows...), row)
		test := b
		test.Content = strings.Join(append(append([]string{}, head...), candidateRows...), "\n")
		if len(curRows) > 0 && md.EstimateTokens(test, bias) > maxTokens {
			content := strings.Join(append(append([]string{}, head...), curRows...), "\n")
			out = append(out, md.Block{Type: b.Type, Content: content, StartLine: b.StartLine, EndLine: b.EndLine})
			curRows = []string{row}
			continue
		}
		curRows = candidateRows
	}
	if len(curRows) > 0 {
		content := strings.Join(append(append([]string{}, head...), curRows...), "\n")
		out = append(out, md.Block{Type: b.Type, Content: content, StartLine: b.StartLine, EndLine: b.EndLine})
	}
	if len(out) == 0 {
		return []md.Block{b}
	}
	return out
}

func splitByLines(b md.Block, maxTokens int, bias md.TokenBias) []md.Block {
	lines := strings.Split(b.Content, "\n")
	out := make([]md.Block, 0, 4)
	cur := make([]string, 0)
	for _, line := range lines {
		candidate := append(append([]string{}, cur...), line)
		test := b
		test.Content = strings.Join(candidate, "\n")
		if len(cur) > 0 && md.EstimateTokens(test, bias) > maxTokens {
			out = append(out, md.Block{Type: b.Type, Content: strings.Join(cur, "\n"), StartLine: b.StartLine, EndLine: b.EndLine, Lang: b.Lang, Level: b.Level})
			cur = []string{line}
			continue
		}
		cur = candidate
	}
	if len(cur) > 0 {
		out = append(out, md.Block{Type: b.Type, Content: strings.Join(cur, "\n"), StartLine: b.StartLine, EndLine: b.EndLine, Lang: b.Lang, Level: b.Level})
	}
	if len(out) == 0 {
		return []md.Block{b}
	}
	return out
}

func splitSentences(text string) []string {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil
	}
	parts := make([]string, 0, 8)
	start := 0
	runes := []rune(trimmed)
	for i := 0; i < len(runes); i++ {
		switch runes[i] {
		case '.', '!', '?':
			if i+1 < len(runes) && runes[i+1] != ' ' && runes[i+1] != '\n' {
				continue
			}
			parts = append(parts, strings.TrimSpace(string(runes[start:i+1])))
			for i+1 < len(runes) && (runes[i+1] == ' ' || runes[i+1] == '\n' || runes[i+1] == '\t') {
				i++
			}
			start = i + 1
		}
	}
	if start < len(runes) {
		parts = append(parts, strings.TrimSpace(string(runes[start:])))
	}
	if len(parts) <= 1 {
		return strings.Fields(trimmed)
	}
	return parts
}
