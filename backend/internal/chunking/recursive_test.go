package chunking

import "testing"

func TestRecursiveChunker_InvalidOptions(t *testing.T) {
	chunker := RecursiveChunker{MaxRunes: 0}
	if _, err := chunker.Chunk("hello"); err != ErrInvalidMaxRunes {
		t.Fatalf("expected %v, got %v", ErrInvalidMaxRunes, err)
	}

	chunker = RecursiveChunker{MaxRunes: 10, OverlapRunes: -1}
	if _, err := chunker.Chunk("hello"); err != ErrInvalidOverlap {
		t.Fatalf("expected %v, got %v", ErrInvalidOverlap, err)
	}

	chunker = RecursiveChunker{MaxRunes: 10, OverlapRunes: 10}
	if _, err := chunker.Chunk("hello"); err != ErrOverlapTooLarge {
		t.Fatalf("expected %v, got %v", ErrOverlapTooLarge, err)
	}
}

func TestRecursiveChunker_EmptyInput(t *testing.T) {
	chunker := RecursiveChunker{MaxRunes: 10}
	chunks, err := chunker.Chunk("")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks, got %d", len(chunks))
	}
}

func TestRecursiveChunker_SplitsOnSeparators(t *testing.T) {
	input := "section1\n\nsection2\n\nsection3"
	chunker := RecursiveChunker{
		MaxRunes:   10,
		Separators: []string{"\n\n", "\n", " ", ""},
	}

	chunks, err := chunker.Chunk(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	expected := []string{"section1\n\n", "section2\n\n", "section3"}
	for i, chunk := range chunks {
		if chunk.Content != expected[i] {
			t.Fatalf("chunk %d content mismatch: %q", i, chunk.Content)
		}
	}
}

func TestRecursiveChunker_FallbackToFixed(t *testing.T) {
	input := "abcdefghijklmnop"
	chunker := RecursiveChunker{
		MaxRunes:   5,
		Separators: []string{"\n\n"},
	}

	chunks, err := chunker.Chunk(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(chunks) != 4 {
		t.Fatalf("expected 4 chunks, got %d", len(chunks))
	}

	for i, chunk := range chunks {
		if chunk.RuneLength != 5 && i != len(chunks)-1 {
			t.Fatalf("expected chunk %d length 5, got %d", i, chunk.RuneLength)
		}
	}
}

func TestRecursiveChunker_Overlap(t *testing.T) {
	input := "alpha\n\nbeta\n\ngamma"
	chunker := RecursiveChunker{
		MaxRunes:     10,
		OverlapRunes: 3,
		Separators:   []string{"\n\n"},
	}

	chunks, err := chunker.Chunk(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks, got %d", len(chunks))
	}

	for i := 1; i < len(chunks); i++ {
		prev := chunks[i-1]
		curr := chunks[i]
		if curr.StartRune >= curr.EndRune {
			t.Fatalf("invalid overlap range for chunk %d", i)
		}
		if prev.EndRune-curr.StartRune < 1 {
			t.Fatalf("expected overlap for chunk %d", i)
		}
	}
}

func TestSeparatorsForLanguage(t *testing.T) {
	seps := SeparatorsForLanguage(LanguageGo)
	if len(seps) == 0 {
		t.Fatal("expected language separators")
	}
}
