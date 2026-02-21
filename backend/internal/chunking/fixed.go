package chunking

import "errors"

var (
	ErrInvalidMaxRunes = errors.New("max_runes must be greater than zero")
	ErrInvalidOverlap  = errors.New("overlap_runes must be zero or greater")
	ErrOverlapTooLarge = errors.New("overlap_runes must be smaller than max_runes")
)

// Chunk holds a fixed-size chunk with rune offsets into the original text.
type Chunk struct {
	Index      int
	StartRune  int
	EndRune    int
	Content    string
	RuneLength int
	Metadata   map[string]any
}

// Chunker abstracts chunking strategies for easy swapping later.
type Chunker interface {
	Chunk(text string) ([]Chunk, error)
}

// FixedSizeChunker splits text into fixed-size rune windows with optional overlap.
type FixedSizeChunker struct {
	MaxRunes     int
	OverlapRunes int
}

func (c FixedSizeChunker) Chunk(text string) ([]Chunk, error) {
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

	runes := []rune(text)
	chunks := make([]Chunk, 0, (len(runes)/c.MaxRunes)+1)
	start := 0
	index := 0

	for start < len(runes) {
		end := start + c.MaxRunes
		if end > len(runes) {
			end = len(runes)
		}

		content := string(runes[start:end])
		chunk := Chunk{
			Index:      index,
			StartRune:  start,
			EndRune:    end,
			Content:    content,
			RuneLength: end - start,
		}
		chunks = append(chunks, chunk)
		index++

		if end == len(runes) {
			break
		}

		start = end - c.OverlapRunes
	}

	return chunks, nil
}
