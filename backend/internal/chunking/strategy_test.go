package chunking

import "testing"

func TestNewChunker_UnknownStrategy(t *testing.T) {
	_, err := NewChunker(Options{Strategy: Strategy("weird"), MaxRunes: 5})
	if err != ErrUnknownStrategy {
		t.Fatalf("expected %v, got %v", ErrUnknownStrategy, err)
	}
}

func TestNewChunker_DefaultsToFixed(t *testing.T) {
	chunker, err := NewChunker(Options{MaxRunes: 5})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if _, ok := chunker.(FixedSizeChunker); !ok {
		t.Fatalf("expected FixedSizeChunker, got %T", chunker)
	}
}

func TestNewChunker_RecursiveDefaults(t *testing.T) {
	chunker, err := NewChunker(Options{Strategy: StrategyRecursive, MaxRunes: 5})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	recursive, ok := chunker.(RecursiveChunker)
	if !ok {
		t.Fatalf("expected RecursiveChunker, got %T", chunker)
	}
	if len(recursive.Separators) == 0 {
		t.Fatalf("expected default separators")
	}
}
