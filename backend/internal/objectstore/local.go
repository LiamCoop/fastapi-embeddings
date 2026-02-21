package objectstore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalClient stores objects on the local filesystem for development.
type LocalClient struct {
	root string
}

func NewLocalClient(root string) *LocalClient {
	return &LocalClient{root: root}
}

func (c *LocalClient) URIForKey(key string) string {
	path := filepath.Join(c.root, filepath.FromSlash(key))
	return "file://" + path
}

func (c *LocalClient) Put(ctx context.Context, key string, r io.Reader) (string, int64, error) {
	path := filepath.Join(c.root, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", 0, err
	}

	file, err := os.Create(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	written, err := io.Copy(file, r)
	if err != nil {
		return "", 0, err
	}

	_ = ctx
	return "file://" + path, written, nil
}

func (c *LocalClient) PresignPut(ctx context.Context, key, contentType string) (string, map[string]string, string, error) {
	_ = ctx
	_ = key
	_ = contentType
	return "", nil, "", fmt.Errorf("presign not supported for local object store")
}

func (c *LocalClient) Get(ctx context.Context, uri string) (io.ReadCloser, int64, error) {
	_ = ctx
	if !strings.HasPrefix(uri, "file://") {
		return nil, 0, fmt.Errorf("unsupported local uri: %s", uri)
	}

	path := strings.TrimPrefix(uri, "file://")
	file, err := os.Open(path)
	if err != nil {
		return nil, 0, err
	}

	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, 0, err
	}

	return file, info.Size(), nil
}
