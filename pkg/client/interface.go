package client

import (
	"context"

	"github.com/kazuma-desu/etu/pkg/models"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// EtcdReader defines read operations on etcd.
// Implementations must be safe for concurrent use.
type EtcdReader interface {
	// Get retrieves a single value from etcd.
	// Returns error if key not found.
	Get(ctx context.Context, key string) (string, error)

	// GetWithOptions retrieves keys with advanced options (prefix, sort, etc.)
	GetWithOptions(ctx context.Context, key string, opts *GetOptions) (*GetResponse, error)
}

// EtcdWriter defines write operations on etcd.
type EtcdWriter interface {
	// Put writes a single key-value pair.
	Put(ctx context.Context, key, value string) error

	// PutAll writes multiple configuration pairs.
	// May use transactions for atomicity (implementation-dependent).
	PutAll(ctx context.Context, pairs []*models.ConfigPair) error
}

// EtcdClient combines read and write operations with lifecycle management.
// This is the primary interface that commands should depend on.
type EtcdClient interface {
	EtcdReader
	EtcdWriter

	// Close releases resources. Must be called when done.
	Close() error

	// Status returns cluster status for the given endpoint.
	Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error)
}
