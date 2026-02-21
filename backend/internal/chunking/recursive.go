package chunking

import "errors"

var ErrUnknownStrategy = errors.New("unknown chunking strategy")

type runeRange struct {
	start int
	end   int
}

// RecursiveChunker splits text using a hierarchy of separators.
type RecursiveChunker struct {
	MaxRunes     int
	OverlapRunes int
	Separators   []string
}

func (c RecursiveChunker) Chunk(text string) ([]Chunk, error) {
	if c.MaxRunes <= 0 {
		return nil, ErrInvalidMaxRunes
	}
	if c.OverlapRunes < 0 {
		return nil, ErrInvalidOverlap
	}
	if c.OverlapRunes >= c.MaxRunes {
		return nil, ErrOverlapTooLarge
	}

	if text == "" {
		return []Chunk{}, nil
	}

	seps := c.Separators
	if len(seps) == 0 {
		seps = DefaultRecursiveSeparators()
	}

	runes := []rune(text)
	base := c.splitRange(runes, runeRange{start: 0, end: len(runes)}, seps, 0)
	chunks := make([]Chunk, 0, len(base))

	for i, rr := range base {
		start := rr.start
		if c.OverlapRunes > 0 && i > 0 {
			overlapStart := rr.start - c.OverlapRunes
			if overlapStart < base[i-1].start {
				overlapStart = base[i-1].start
			}
			if overlapStart < 0 {
				overlapStart = 0
			}
			start = overlapStart
		}

		content := string(runes[start:rr.end])
		chunks = append(chunks, Chunk{
			Index:      i,
			StartRune:  start,
			EndRune:    rr.end,
			Content:    content,
			RuneLength: rr.end - start,
		})
	}

	return chunks, nil
}

func (c RecursiveChunker) splitRange(runes []rune, rr runeRange, seps []string, sepIndex int) []runeRange {
	if rr.end-rr.start <= c.MaxRunes {
		return []runeRange{rr}
	}
	if sepIndex >= len(seps) {
		return splitFixedRange(rr, c.MaxRunes)
	}

	sep := []rune(seps[sepIndex])
	if len(sep) == 0 {
		return splitFixedRange(rr, c.MaxRunes)
	}

	parts := splitBySeparator(runes, rr, sep)
	if len(parts) == 1 {
		return c.splitRange(runes, rr, seps, sepIndex+1)
	}

	out := make([]runeRange, 0, len(parts))
	for _, part := range parts {
		if part.end-part.start <= c.MaxRunes {
			out = append(out, part)
			continue
		}
		out = append(out, c.splitRange(runes, part, seps, sepIndex+1)...)
	}
	return out
}

func splitFixedRange(rr runeRange, maxRunes int) []runeRange {
	if maxRunes <= 0 || rr.end <= rr.start {
		return []runeRange{}
	}
	out := make([]runeRange, 0, (rr.end-rr.start)/maxRunes+1)
	start := rr.start
	for start < rr.end {
		end := start + maxRunes
		if end > rr.end {
			end = rr.end
		}
		out = append(out, runeRange{start: start, end: end})
		if end == rr.end {
			break
		}
		start = end
	}
	return out
}

func splitBySeparator(runes []rune, rr runeRange, sep []rune) []runeRange {
	if len(sep) == 0 {
		return []runeRange{rr}
	}

	out := make([]runeRange, 0, 8)
	cursor := rr.start
	for cursor < rr.end {
		index := indexOfRunes(runes, sep, cursor, rr.end)
		if index == -1 {
			break
		}
		segmentEnd := index + len(sep)
		out = append(out, runeRange{start: cursor, end: segmentEnd})
		cursor = segmentEnd
	}

	if cursor < rr.end {
		out = append(out, runeRange{start: cursor, end: rr.end})
	}

	if len(out) == 0 {
		return []runeRange{rr}
	}

	return out
}

func indexOfRunes(haystack []rune, needle []rune, start, end int) int {
	if len(needle) == 0 || start >= end || start < 0 || end > len(haystack) {
		return -1
	}
	last := end - len(needle)
	for i := start; i <= last; i++ {
		match := true
		for j := range needle {
			if haystack[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
