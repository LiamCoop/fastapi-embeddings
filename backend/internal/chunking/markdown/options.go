package markdown

import (
	"fmt"

	md "ragtime-backend/internal/markdown"
)

// FrontmatterMode controls how YAML frontmatter is treated.
type FrontmatterMode int

const (
	FrontmatterMetadata FrontmatterMode = iota
	FrontmatterInclude
	FrontmatterStrip
)

// MarkdownOptions configures markdown chunking behavior.
type MarkdownOptions struct {
	TargetTokens    int
	MaxTokens       int
	MinTokens       int
	OverlapTokens   int
	HeadingDepth    int
	FrontmatterMode FrontmatterMode
	MDX             bool
	Bias            md.TokenBias
}

func DefaultMarkdownOptions() MarkdownOptions {
	return MarkdownOptions{
		TargetTokens:    750,
		MaxTokens:       1000,
		MinTokens:       200,
		OverlapTokens:   80,
		HeadingDepth:    3,
		FrontmatterMode: FrontmatterMetadata,
		MDX:             false,
		Bias:            md.BiasBalanced,
	}
}

func (o MarkdownOptions) Validate() error {
	if o.TargetTokens <= 0 {
		return fmt.Errorf("target_tokens must be greater than zero")
	}
	if o.MaxTokens <= 0 {
		return fmt.Errorf("max_tokens must be greater than zero")
	}
	if o.MinTokens < 0 {
		return fmt.Errorf("min_tokens must be zero or greater")
	}
	if o.OverlapTokens < 0 {
		return fmt.Errorf("overlap_tokens must be zero or greater")
	}
	if o.TargetTokens > o.MaxTokens {
		return fmt.Errorf("target_tokens must be <= max_tokens")
	}
	if o.HeadingDepth <= 0 || o.HeadingDepth > 6 {
		return fmt.Errorf("heading_depth must be between 1 and 6")
	}
	return nil
}
