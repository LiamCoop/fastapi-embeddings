package markdown

import (
	md "ragtime-backend/internal/markdown"
)

type packedChunk struct {
	blocks       []md.Block
	estTokens    int
	breadcrumb   string
	sectionTitle string
	blockStart   int
	blockEnd     int
}

type indexedBlock struct {
	index int
	block md.Block
}

func packBlocks(blocks []md.Block, opts MarkdownOptions) []packedChunk {
	stream := make([]indexedBlock, 0, len(blocks))
	for i, b := range blocks {
		parts := SplitOversized(b, opts.MaxTokens, opts.Bias)
		for _, part := range parts {
			stream = append(stream, indexedBlock{index: i, block: part})
		}
	}
	if len(stream) == 0 {
		return nil
	}

	headings := md.NewHeadingStack()
	result := make([]packedChunk, 0, len(stream)/2+1)
	current := make([]indexedBlock, 0, 8)
	currentTokens := 0

	finalize := func() {
		if len(current) == 0 {
			return
		}
		blocksOnly := make([]md.Block, 0, len(current))
		for _, ib := range current {
			blocksOnly = append(blocksOnly, ib.block)
		}
		result = append(result, packedChunk{
			blocks:       blocksOnly,
			estTokens:    currentTokens,
			breadcrumb:   headings.Breadcrumb(),
			sectionTitle: headings.SectionTitle(),
			blockStart:   current[0].index,
			blockEnd:     current[len(current)-1].index,
		})
	}

	for _, ib := range stream {
		b := ib.block
		tokens := md.EstimateTokens(b, opts.Bias)

		if b.Type == md.BlockHeading && b.Level > 0 && b.Level <= opts.HeadingDepth {
			if len(current) > 0 {
				finalize()
				current = current[:0]
				currentTokens = 0
			}
			headings.Update(b.Level, b.Content)
		}

		if len(current) > 0 && currentTokens+tokens > opts.MaxTokens {
			prev := append([]indexedBlock{}, current...)
			finalize()
			overlap := computeOverlap(prev, opts.OverlapTokens, opts.Bias)
			current = current[:0]
			currentTokens = 0
			for _, ov := range overlap {
				current = append(current, ov)
				currentTokens += md.EstimateTokens(ov.block, opts.Bias)
			}
		}

		current = append(current, ib)
		currentTokens += tokens
	}

	if len(current) > 0 {
		finalize()
	}

	return mergeSmallChunks(result, opts.MinTokens, opts.MaxTokens)
}

func computeOverlap(prevBlocks []indexedBlock, overlapTokens int, bias md.TokenBias) []indexedBlock {
	if overlapTokens <= 0 || len(prevBlocks) == 0 {
		return nil
	}
	selected := make([]indexedBlock, 0, len(prevBlocks))
	total := 0
	for i := len(prevBlocks) - 1; i >= 0; i-- {
		b := prevBlocks[i]
		if b.block.Type == md.BlockFrontmatter {
			continue
		}
		toks := md.EstimateTokens(b.block, bias)
		if total+toks > overlapTokens {
			break
		}
		selected = append(selected, b)
		total += toks
	}
	for i, j := 0, len(selected)-1; i < j; i, j = i+1, j-1 {
		selected[i], selected[j] = selected[j], selected[i]
	}
	if isHeadingOnly(selected) {
		return nil
	}
	return selected
}

func isHeadingOnly(blocks []indexedBlock) bool {
	if len(blocks) == 0 {
		return false
	}
	for _, b := range blocks {
		if b.block.Type != md.BlockHeading {
			return false
		}
	}
	return true
}

func mergeSmallChunks(chunks []packedChunk, minTokens, maxTokens int) []packedChunk {
	if minTokens <= 0 || len(chunks) <= 1 {
		return chunks
	}
	for i := 0; i < len(chunks); i++ {
		if chunks[i].estTokens >= minTokens {
			continue
		}
		if i+1 < len(chunks) && chunks[i].estTokens+chunks[i+1].estTokens <= maxTokens {
			chunks[i+1] = mergePacked(chunks[i], chunks[i+1])
			chunks = append(chunks[:i], chunks[i+1:]...)
			i--
			continue
		}
		if i > 0 && chunks[i].estTokens+chunks[i-1].estTokens <= maxTokens {
			chunks[i-1] = mergePacked(chunks[i-1], chunks[i])
			chunks = append(chunks[:i], chunks[i+1:]...)
			i--
		}
	}
	return chunks
}

func mergePacked(a, b packedChunk) packedChunk {
	mergedBlocks := append(append([]md.Block{}, a.blocks...), b.blocks...)
	return packedChunk{
		blocks:       mergedBlocks,
		estTokens:    a.estTokens + b.estTokens,
		breadcrumb:   b.breadcrumb,
		sectionTitle: b.sectionTitle,
		blockStart:   a.blockStart,
		blockEnd:     b.blockEnd,
	}
}
