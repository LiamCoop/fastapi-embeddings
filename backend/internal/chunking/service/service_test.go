package service

import (
	"testing"

	"ragtime-backend/internal/chunking"
)

func TestResolveDocumentStrategyPrefersMarkdownForMarkdownURI(t *testing.T) {
	strategy := resolveDocumentStrategy(chunking.StrategyFixed, "s3://bucket/kb/documents/doc-1/v1.md")
	if strategy != chunking.StrategyMarkdown {
		t.Fatalf("expected markdown strategy, got %q", strategy)
	}
}

func TestResolveDocumentStrategyKeepsExplicitRecursive(t *testing.T) {
	strategy := resolveDocumentStrategy(chunking.StrategyRecursive, "s3://bucket/kb/documents/doc-1/v1.md")
	if strategy != chunking.StrategyRecursive {
		t.Fatalf("expected recursive strategy, got %q", strategy)
	}
}

func TestResolveDocumentStrategyLeavesFixedForNonMarkdown(t *testing.T) {
	strategy := resolveDocumentStrategy(chunking.StrategyFixed, "s3://bucket/kb/documents/doc-1/v1.pdf")
	if strategy != chunking.StrategyFixed {
		t.Fatalf("expected fixed strategy, got %q", strategy)
	}
}
