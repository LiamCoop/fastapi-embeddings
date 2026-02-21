package objectstore

import (
	"context"
	"io"
)

// Client defines the minimal object storage contract for raw document bytes.
type Client interface {
	Put(ctx context.Context, key string, r io.Reader) (uri string, size int64, err error)
	URIForKey(key string) string
	PresignPut(ctx context.Context, key, contentType string) (url string, headers map[string]string, uri string, err error)
	Get(ctx context.Context, uri string) (io.ReadCloser, int64, error)
}
