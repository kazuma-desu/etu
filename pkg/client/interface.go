package client

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/kazuma-desu/etu/pkg/models"
)

// PutAllResult contains the outcome of a batch put operation.
type PutAllResult struct {
	FailedKey string // Key that caused failure, empty if all succeeded
	Succeeded int    // Number of items successfully applied
	Failed    int    // Number of items that failed (0 or 1, since we stop on first error)
	Total     int    // Total items in the batch
}

// ProgressFunc is called after each successful put operation.
// Parameters: current (1-indexed), total count, and the key just written.
type ProgressFunc func(current, total int, key string)

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
	// Applies items sequentially; partial failures are possible (items before
	// the failed one are committed). For progress feedback or partial failure
	// details, use PutAllWithProgress.
	PutAll(ctx context.Context, pairs []*models.ConfigPair) error

	// PutAllWithProgress writes multiple pairs with optional progress callback.
	// If onProgress is non-nil, it's called after each successful put.
	// Returns PutAllResult with success/failure counts even on error.
	PutAllWithProgress(ctx context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc) (*PutAllResult, error)
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
