package storage

import "context"

type ObjectStore interface {
	Put(ctx context.Context, key string, contentType string, data []byte) (string, error)
	Get(ctx context.Context, key string) ([]byte, error)
}
