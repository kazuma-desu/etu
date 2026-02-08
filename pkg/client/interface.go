package client

import (
	"context"
	"time"

	"github.com/kazuma-desu/etu/pkg/models"
)

// PutAllResult contains the outcome of a batch put operation.
type PutAllResult struct {
	// FailedKeys contains all keys that failed to be written.
	// When a batch fails, all keys in that batch are included since batches are atomic.
	FailedKeys []string

	// Succeeded is the number of items successfully applied.
	Succeeded int

	// Failed is the number of items that failed.
	Failed int

	// Total is the total number of items in the operation.
	Total int

	// RetryCount is the total number of retry attempts made across all batches.
	RetryCount int

	// UsedFallback indicates whether single-key fallback mode was used.
	UsedFallback bool
}

// FailedKey returns the first failed key for backward compatibility.
// Returns empty string if no keys failed.
func (r *PutAllResult) FailedKey() string {
	if len(r.FailedKeys) == 0 {
		return ""
	}
	return r.FailedKeys[0]
}

// BatchOptions configures batch operation behavior including retry and fallback strategies.
type BatchOptions struct {
	// Logger receives log messages about batch operations.
	// If nil, no logging is performed.
	Logger Logger

	// InitialBackoff is the initial backoff duration before first retry.
	// Subsequent retries use exponential backoff: InitialBackoff * 2^attempt
	// Default: 100ms
	InitialBackoff time.Duration

	// MaxBackoff is the maximum backoff duration between retries.
	// Default: 5s
	MaxBackoff time.Duration

	// MaxRetries is the maximum number of retry attempts for failed batches.
	// Default: 3
	MaxRetries int

	// FallbackToSingleKeys enables falling back to single-key puts when a batch
	// transaction fails after all retries are exhausted.
	// Default: true
	FallbackToSingleKeys bool
}

// DefaultBatchOptions returns BatchOptions with sensible defaults.
func DefaultBatchOptions() *BatchOptions {
	return &BatchOptions{
		MaxRetries:           3,
		InitialBackoff:       100 * time.Millisecond,
		MaxBackoff:           5 * time.Second,
		FallbackToSingleKeys: true,
		Logger:               nil,
	}
}

type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)

	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
}

// ProgressFunc is called after each successful put operation.
// Parameters: current (1-indexed), total count, and the key just written.
type ProgressFunc func(current, total int, key string)

// WatchEventType represents the type of watch event.
type WatchEventType string

const (
	// WatchEventPut indicates a key was created or updated.
	WatchEventPut WatchEventType = "PUT"
	// WatchEventDelete indicates a key was deleted.
	WatchEventDelete WatchEventType = "DELETE"
)

// WatchEvent represents a single event from a watch operation.
type WatchEvent struct {
	// Type is the type of event (PUT or DELETE).
	Type WatchEventType

	// Key is the key that was affected.
	Key string

	// Value is the new value (for PUT events).
	Value string

	// PrevValue is the previous value (if available).
	// nil indicates no previous value was provided.
	PrevValue *string

	// Revision is the revision of the event.
	Revision int64

	// CreateRevision is the revision when the key was created.
	CreateRevision int64

	// ModRevision is the revision when the key was last modified.
	ModRevision int64

	// Version is the version of the key.
	Version int64
}

// WatchResponse contains the response from a watch operation.
type WatchResponse struct {
	// Events is the list of events that occurred.
	Events []WatchEvent

	// CompactRevision is the compaction revision if the watcher was canceled
	// due to compaction.
	CompactRevision int64

	// Err is set if the watch encountered an error.
	Err error
}

// WatchChan is a channel that receives watch responses.
// Consumers should only receive from this channel.
type WatchChan <-chan WatchResponse

// WatchOptions configures watch behavior.
type WatchOptions struct {
	// Prefix watches all keys with the given prefix.
	Prefix bool

	// Revision is the revision to start watching from.
	// If 0, watches from the current revision.
	Revision int64

	// PrevKV indicates whether to include the previous key-value pair
	// in the watch response.
	PrevKV bool
}

// EtcdReader defines read operations on etcd.
// Implementations must be safe for concurrent use.
type EtcdReader interface {
	// Get retrieves a single value from etcd.
	// Returns error if key not found.
	Get(ctx context.Context, key string) (any, error)

	// GetWithOptions retrieves keys with advanced options (prefix, sort, etc.)
	GetWithOptions(ctx context.Context, key string, opts *GetOptions) (*GetResponse, error)

	// Watch watches for changes on a key or prefix.
	// Returns a channel that receives watch responses.
	// The channel is closed when the watch is canceled or encounters an error.
	// Use the context to cancel the watch.
	Watch(ctx context.Context, key string, opts *WatchOptions) WatchChan
}

// EtcdWriter defines write operations on etcd.
type EtcdWriter interface {
	Put(ctx context.Context, key, value string) error
	PutAll(ctx context.Context, pairs []*models.ConfigPair) error
	PutAllWithProgress(ctx context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc) (*PutAllResult, error)
	PutAllWithOptions(ctx context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc, opts *BatchOptions) (*PutAllResult, error)
	Delete(ctx context.Context, key string) (int64, error)
	DeletePrefix(ctx context.Context, prefix string) (int64, error)
}

// StatusResponse contains the status information for an etcd cluster member.
// This is a wrapper type to avoid exposing etcd SDK types directly.
type StatusResponse struct {
	// Version is the etcd server version.
	Version string

	// DbSize is the size of the database in bytes.
	DbSize int64

	// Leader is the member ID of the leader.
	Leader uint64

	// RaftIndex is the current raft index.
	RaftIndex uint64

	// RaftTerm is the current raft term.
	RaftTerm uint64

	// RaftAppliedIndex is the last applied raft index.
	RaftAppliedIndex uint64

	// Errors contains any errors from the cluster.
	Errors []string

	// IsLearner indicates if this member is a learner.
	IsLearner bool
}

// EtcdClient combines read and write operations with lifecycle management.
// This is the primary interface that commands should depend on.
type EtcdClient interface {
	EtcdReader
	EtcdWriter

	// Close releases resources. Must be called when done.
	Close() error

	// Status returns cluster status for the given endpoint.
	Status(ctx context.Context, endpoint string) (*StatusResponse, error)
}

// OperationRecorder is implemented by clients that record operations
// for preview/dry-run purposes. Real clients (Client) do not implement this.
// Use type assertion to check if a client supports operation recording:
//
//	if recorder, ok := client.(OperationRecorder); ok {
//	    ops := recorder.Operations()
//	}
type OperationRecorder interface {
	// Operations returns a copy of all recorded operations.
	// The returned slice is safe to modify.
	Operations() []Operation

	// OperationCount returns the number of recorded operations.
	OperationCount() int
}
