package chunking

import "testing"

func TestFixedSizeChunker_InvalidOptions(t *testing.T) {
	chunker := FixedSizeChunker{MaxRunes: 0}
	if _, err := chunker.Chunk("hello"); err != ErrInvalidMaxRunes {
		t.Fatalf("expected %v, got %v", ErrInvalidMaxRunes, err)
	}

	chunker = FixedSizeChunker{MaxRunes: 10, OverlapRunes: -1}
	if _, err := chunker.Chunk("hello"); err != ErrInvalidOverlap {
		t.Fatalf("expected %v, got %v", ErrInvalidOverlap, err)
	}

	chunker = FixedSizeChunker{MaxRunes: 10, OverlapRunes: 10}
	if _, err := chunker.Chunk("hello"); err != ErrOverlapTooLarge {
		t.Fatalf("expected %v, got %v", ErrOverlapTooLarge, err)
	}
}

func TestFixedSizeChunker_EmptyInput(t *testing.T) {
	chunker := FixedSizeChunker{MaxRunes: 10}
	chunks, err := chunker.Chunk("")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(chunks) != 0 {
		t.Fatalf("expected 0 chunks, got %d", len(chunks))
	}
}

func TestFixedSizeChunker_ChunkSizes(t *testing.T) {
	input := "# Title\n\nThis is a sample markdown document with a few sentences.\n\n" +
		"- Bullet one\n- Bullet two\n- Bullet three\n\n" +
		"```go\nfmt.Println(\"hello\")\n```\n"

	chunker := FixedSizeChunker{MaxRunes: 40, OverlapRunes: 10}
	chunks, err := chunker.Chunk(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(chunks) == 0 {
		t.Fatalf("expected chunks, got 0")
	}

	runes := []rune(input)
	for i, chunk := range chunks {
		if chunk.RuneLength <= 0 {
			t.Fatalf("expected chunk %d to have positive length", i)
		}
		if chunk.RuneLength > chunker.MaxRunes {
			t.Fatalf("expected chunk %d length <= %d, got %d", i, chunker.MaxRunes, chunk.RuneLength)
		}
		if chunk.StartRune < 0 || chunk.EndRune > len(runes) || chunk.StartRune >= chunk.EndRune {
			t.Fatalf("invalid rune offsets for chunk %d: %d-%d", i, chunk.StartRune, chunk.EndRune)
		}

		expected := string(runes[chunk.StartRune:chunk.EndRune])
		if chunk.Content != expected {
			t.Fatalf("chunk %d content mismatch", i)
		}

		if i > 0 {
			prev := chunks[i-1]
			expectedStart := prev.EndRune - chunker.OverlapRunes
			if chunk.StartRune != expectedStart {
				t.Fatalf("chunk %d start rune expected %d, got %d", i, expectedStart, chunk.StartRune)
			}
			if prev.EndRune != chunk.StartRune+chunker.OverlapRunes {
				t.Fatalf("chunk %d overlap mismatch", i)
			}
		}
	}
}
