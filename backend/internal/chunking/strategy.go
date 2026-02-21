package chunking

import (
	"strings"

	mdchunking "ragtime-backend/internal/chunking/markdown"
)

// Strategy identifies a chunking approach.
type Strategy string

const (
	StrategyFixed     Strategy = "fixed"
	StrategyRecursive Strategy = "recursive"
	StrategyMarkdown  Strategy = "markdown"
)

// Language captures language-specific separator presets.
type Language string

const (
	LanguageGeneric    Language = "generic"
	LanguageGo         Language = "go"
	LanguagePython     Language = "python"
	LanguageJavaScript Language = "javascript"
	LanguageJava       Language = "java"
	LanguageRust       Language = "rust"
)

// Options configures chunker selection and behavior.
type Options struct {
	Strategy      Strategy
	MaxRunes      int
	OverlapRunes  int
	Separators    []string
	LanguageHints []Language
}

// NewChunker returns a chunker for the requested strategy.
func NewChunker(opts Options) (Chunker, error) {
	switch opts.Strategy {
	case StrategyFixed, "":
		return FixedSizeChunker{MaxRunes: opts.MaxRunes, OverlapRunes: opts.OverlapRunes}, nil
	case StrategyRecursive:
		seps := opts.Separators
		if len(seps) == 0 {
			seps = separatorsForHints(opts.LanguageHints)
		}
		return RecursiveChunker{
			MaxRunes:     opts.MaxRunes,
			OverlapRunes: opts.OverlapRunes,
			Separators:   seps,
		}, nil
	case StrategyMarkdown:
		return NewMarkdownChunker(mdchunking.DefaultMarkdownOptions())
	default:
		return nil, ErrUnknownStrategy
	}
}

// separatorsForHints merges language separator presets with the generic fallback.
func separatorsForHints(hints []Language) []string {
	if len(hints) == 0 {
		return DefaultRecursiveSeparators()
	}
	seen := make(map[string]struct{})
	merged := make([]string, 0, 16)
	for _, hint := range hints {
		for _, sep := range SeparatorsForLanguage(hint) {
			if _, ok := seen[sep]; ok {
				continue
			}
			seen[sep] = struct{}{}
			merged = append(merged, sep)
		}
	}
	for _, sep := range DefaultRecursiveSeparators() {
		if _, ok := seen[sep]; ok {
			continue
		}
		seen[sep] = struct{}{}
		merged = append(merged, sep)
	}
	return merged
}

// DefaultRecursiveSeparators returns the generic recursive separator list.
func DefaultRecursiveSeparators() []string {
	return []string{"\n\n", "\n", " ", ""}
}

// SeparatorsForLanguage returns separators tuned for a specific language.
func SeparatorsForLanguage(language Language) []string {
	switch strings.ToLower(string(language)) {
	case string(LanguagePython):
		return []string{"\nclass ", "\ndef ", "\n\n", "\n", " ", ""}
	case string(LanguageJavaScript):
		return []string{"\nclass ", "\nfunction ", "\nconst ", "\nlet ", "\n\n", "\n", " ", ""}
	case string(LanguageJava):
		return []string{"\nclass ", "\ninterface ", "\npublic ", "\nprivate ", "\n\n", "\n", " ", ""}
	case string(LanguageGo):
		return []string{"\nfunc ", "\ntype ", "\nvar ", "\nconst ", "\n\n", "\n", " ", ""}
	case string(LanguageRust):
		return []string{"\nfn ", "\nstruct ", "\nenum ", "\nimpl ", "\n\n", "\n", " ", ""}
	default:
		return DefaultRecursiveSeparators()
	}
}
