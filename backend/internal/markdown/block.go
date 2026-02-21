package markdown

import (
	"regexp"
	"strings"
)

// BlockType identifies a parsed markdown structural block.
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

// Block is a structural unit parsed from markdown text.
type Block struct {
	Type      BlockType
	Content   string
	Level     int
	Lang      string
	StartLine int
	EndLine   int
}

var (
	headingPattern   = regexp.MustCompile(`^(#{1,6})\s+(.+?)\s*$`)
	listItemPattern  = regexp.MustCompile(`^\s*(?:[-*+]\s+|\d+\.\s+)`)
	mdxImportPattern = regexp.MustCompile(`^(import|export)\s+`)
	mdxComponentOpen = regexp.MustCompile(`^<[A-Z][A-Za-z0-9]*(?:\s|>|/)`)
)

// ParseBlocks scans markdown text into a flat slice of structural blocks.
func ParseBlocks(text string, mdx bool) []Block {
	if text == "" {
		return nil
	}

	lines := strings.Split(text, "\n")
	blocks := make([]Block, 0, len(lines)/2)
	i := 0

	// YAML frontmatter at the beginning only.
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		j := 1
		for j < len(lines) {
			if strings.TrimSpace(lines[j]) == "---" {
				break
			}
			j++
		}
		if j < len(lines) {
			blocks = append(blocks, Block{
				Type:      BlockFrontmatter,
				Content:   strings.Join(lines[:j+1], "\n"),
				StartLine: 1,
				EndLine:   j + 1,
			})
			i = j + 1
		}
	}

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			i++
			continue
		}

		if strings.HasPrefix(trimmed, "```") {
			lang := strings.TrimSpace(strings.TrimPrefix(trimmed, "```"))
			start := i
			i++
			for i < len(lines) {
				if strings.HasPrefix(strings.TrimSpace(lines[i]), "```") {
					i++
					break
				}
				i++
			}
			blocks = append(blocks, Block{
				Type:      BlockCodeFence,
				Content:   strings.Join(lines[start:i], "\n"),
				Lang:      lang,
				StartLine: start + 1,
				EndLine:   i,
			})
			continue
		}

		if m := headingPattern.FindStringSubmatch(trimmed); m != nil {
			blocks = append(blocks, Block{
				Type:      BlockHeading,
				Content:   strings.TrimSpace(m[2]),
				Level:     len(m[1]),
				StartLine: i + 1,
				EndLine:   i + 1,
			})
			i++
			continue
		}

		if mdx && mdxImportPattern.MatchString(trimmed) {
			start := i
			i++
			for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
				i++
			}
			blocks = append(blocks, Block{
				Type:      BlockMDXImport,
				Content:   strings.Join(lines[start:i], "\n"),
				StartLine: start + 1,
				EndLine:   i,
			})
			continue
		}

		if mdx && mdxComponentOpen.MatchString(trimmed) {
			start := i
			i++
			for i < len(lines) {
				if strings.TrimSpace(lines[i]) == "" {
					break
				}
				nextTrimmed := strings.TrimSpace(lines[i])
				if headingPattern.MatchString(nextTrimmed) || listItemPattern.MatchString(nextTrimmed) {
					break
				}
				i++
			}
			blocks = append(blocks, Block{
				Type:      BlockMDXComponent,
				Content:   strings.Join(lines[start:i], "\n"),
				StartLine: start + 1,
				EndLine:   i,
			})
			continue
		}

		if listItemPattern.MatchString(line) {
			start := i
			i++
			for i < len(lines) {
				next := lines[i]
				nextTrimmed := strings.TrimSpace(next)
				if nextTrimmed == "" {
					i++
					break
				}
				if listItemPattern.MatchString(next) || strings.HasPrefix(next, " ") || strings.HasPrefix(next, "\t") {
					i++
					continue
				}
				break
			}
			blocks = append(blocks, Block{
				Type:      BlockList,
				Content:   strings.TrimRight(strings.Join(lines[start:i], "\n"), "\n"),
				StartLine: start + 1,
				EndLine:   i,
			})
			continue
		}

		if strings.HasPrefix(trimmed, "|") {
			start := i
			i++
			for i < len(lines) {
				nextTrimmed := strings.TrimSpace(lines[i])
				if !strings.HasPrefix(nextTrimmed, "|") {
					break
				}
				i++
			}
			blocks = append(blocks, Block{
				Type:      BlockTable,
				Content:   strings.Join(lines[start:i], "\n"),
				StartLine: start + 1,
				EndLine:   i,
			})
			continue
		}

		if strings.HasPrefix(trimmed, ">") {
			start := i
			i++
			for i < len(lines) {
				nextTrimmed := strings.TrimSpace(lines[i])
				if !strings.HasPrefix(nextTrimmed, ">") {
					break
				}
				i++
			}
			blocks = append(blocks, Block{
				Type:      BlockBlockquote,
				Content:   strings.Join(lines[start:i], "\n"),
				StartLine: start + 1,
				EndLine:   i,
			})
			continue
		}

		start := i
		i++
		for i < len(lines) {
			nextTrimmed := strings.TrimSpace(lines[i])
			if nextTrimmed == "" {
				break
			}
			if strings.HasPrefix(nextTrimmed, "```") || headingPattern.MatchString(nextTrimmed) || listItemPattern.MatchString(lines[i]) || strings.HasPrefix(nextTrimmed, "|") || strings.HasPrefix(nextTrimmed, ">") {
				break
			}
			if mdx && (mdxImportPattern.MatchString(nextTrimmed) || mdxComponentOpen.MatchString(nextTrimmed)) {
				break
			}
			i++
		}
		blocks = append(blocks, Block{
			Type:      BlockParagraph,
			Content:   strings.Join(lines[start:i], "\n"),
			StartLine: start + 1,
			EndLine:   i,
		})
	}

	return blocks
}
