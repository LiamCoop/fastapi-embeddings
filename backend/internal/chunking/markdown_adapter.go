package chunking

import mdchunking "ragtime-backend/internal/chunking/markdown"

type markdownChunkerAdapter struct {
	inner mdchunking.MarkdownChunker
}

func NewMarkdownChunker(opts mdchunking.MarkdownOptions) (Chunker, error) {
	inner, err := mdchunking.NewMarkdownChunker(opts)
	if err != nil {
		return nil, err
	}
	return markdownChunkerAdapter{inner: inner}, nil
}

func (a markdownChunkerAdapter) Chunk(text string) ([]Chunk, error) {
	chunks, err := a.inner.Chunk(text)
	if err != nil {
		return nil, err
	}
	out := make([]Chunk, 0, len(chunks))
	for _, c := range chunks {
		out = append(out, Chunk{
			Index:      c.Index,
			StartRune:  c.StartRune,
			EndRune:    c.EndRune,
			Content:    c.Content,
			RuneLength: c.RuneLength,
			Metadata:   c.Metadata,
		})
	}
	return out, nil
}
