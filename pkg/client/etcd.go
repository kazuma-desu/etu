package client

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/grpclog"

	"github.com/kazuma-desu/etu/pkg/models"
)

func init() {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(io.Discard, io.Discard, io.Discard))
}

const (
	// DefaultMaxOpsPerTxn is etcd's server limit (embed.DefaultMaxTxnOps).
	DefaultMaxOpsPerTxn = 128

	// WarnValueSize threshold for performance warnings (100KB).
	// TODO: Wire this into Put/PutAll methods to emit warnings when value sizes
	// exceed this threshold. Reserved for future implementation of large value
	// detection and performance optimization warnings.
	WarnValueSize = 100 * 1024
)

type Client struct {
	client *clientv3.Client
	config *Config
}

type Config struct {
	Username    string
	Password    string
	Endpoints   []string
	DialTimeout time.Duration
}

func NewClient(cfg *Config) (*Client, error) {
	if err := validateAndPrepareConfig(cfg); err != nil {
		return nil, err
	}

	clientConfig := clientv3.Config{
		Endpoints:           cfg.Endpoints,
		DialTimeout:         cfg.DialTimeout,
		PermitWithoutStream: true,
	}

	if cfg.Username != "" {
		clientConfig.Username = cfg.Username
		clientConfig.Password = cfg.Password
	}

	cli, err := clientv3.New(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	return &Client{
		client: cli,
		config: cfg,
	}, nil
}

func validateAndPrepareConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}

	if len(cfg.Endpoints) == 0 {
		return fmt.Errorf("at least one endpoint is required")
	}

	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}

	return nil
}

func (c *Client) Put(ctx context.Context, key, value string) error {
	_, err := c.client.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to put key %s: %w", key, err)
	}
	return nil
}

func (c *Client) PutAll(ctx context.Context, pairs []*models.ConfigPair) error {
	_, err := c.PutAllWithProgress(ctx, pairs, nil)
	return err
}

func (c *Client) PutAllWithProgress(ctx context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc) (*PutAllResult, error) {
	return c.PutAllWithOptions(ctx, pairs, onProgress, nil)
}

func (c *Client) PutAllWithOptions(ctx context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc, opts *BatchOptions) (*PutAllResult, error) {
	if opts == nil {
		opts = DefaultBatchOptions()
	}

	result := &PutAllResult{Total: len(pairs)}

	if len(pairs) == 0 {
		return result, nil
	}

	for i := 0; i < len(pairs); i += DefaultMaxOpsPerTxn {
		end := min(i+DefaultMaxOpsPerTxn, len(pairs))
		chunk := pairs[i:end]
		batchNum := (i / DefaultMaxOpsPerTxn) + 1

		if opts.Logger != nil {
			opts.Logger.Debug("attempting batch", "batch", batchNum, "keys", len(chunk), "startIdx", i+1, "endIdx", end)
		}

		err := c.executeBatchWithRetry(ctx, chunk, opts, result, batchNum)
		if err != nil {
			if opts.FallbackToSingleKeys {
				if opts.Logger != nil {
					opts.Logger.Warn("batch failed, falling back to single-key mode", "batch", batchNum, "error", err)
				}
				fallbackErr := c.executeSingleKeyFallback(ctx, chunk, opts, result, i, onProgress)
				if fallbackErr != nil {
					return result, fallbackErr
				}
				result.UsedFallback = true
				continue
			}

			for _, pair := range chunk {
				result.FailedKeys = append(result.FailedKeys, pair.Key)
			}
			result.Failed += len(chunk)
			return result, fmt.Errorf("batch %d (items %d-%d) failed: %w", batchNum, i+1, end, err)
		}

		result.Succeeded += len(chunk)

		if onProgress != nil {
			for j, pair := range chunk {
				onProgress(i+j+1, result.Total, pair.Key)
			}
		}
	}

	if opts.Logger != nil {
		opts.Logger.Info("PutAll complete", "succeeded", result.Succeeded, "failed", result.Failed, "total", result.Total, "retries", result.RetryCount, "usedFallback", result.UsedFallback)
	}

	return result, nil
}

func (c *Client) executeBatchWithRetry(ctx context.Context, chunk []*models.ConfigPair, opts *BatchOptions, result *PutAllResult, batchNum int) error {
	ops := make([]clientv3.Op, 0, len(chunk))
	for _, pair := range chunk {
		value := formatValue(pair.Value)
		ops = append(ops, clientv3.OpPut(pair.Key, value))
	}

	var lastErr error
	backoff := opts.InitialBackoff

	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		if attempt > 0 {
			result.RetryCount++
			if opts.Logger != nil {
				opts.Logger.Warn("retrying batch", "batch", batchNum, "attempt", attempt, "maxRetries", opts.MaxRetries, "backoff", backoff)
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}

			backoff = min(backoff*2, opts.MaxBackoff)
		}

		resp, err := c.client.Txn(ctx).Then(ops...).Commit()
		if err != nil {
			lastErr = err
			if opts.Logger != nil {
				opts.Logger.Debug("batch transaction failed", "batch", batchNum, "attempt", attempt, "error", err)
			}
			continue
		}

		if !resp.Succeeded {
			lastErr = fmt.Errorf("transaction did not succeed")
			if opts.Logger != nil {
				opts.Logger.Debug("batch transaction returned false", "batch", batchNum, "attempt", attempt)
			}
			continue
		}

		return nil
	}

	return lastErr
}

func (c *Client) executeSingleKeyFallback(ctx context.Context, chunk []*models.ConfigPair, opts *BatchOptions, result *PutAllResult, baseIdx int, onProgress ProgressFunc) error {
	for j, pair := range chunk {
		value := formatValue(pair.Value)

		if opts.Logger != nil {
			opts.Logger.Debug("single-key put", "key", pair.Key, "idx", baseIdx+j+1)
		}

		_, err := c.client.Put(ctx, pair.Key, value)
		if err != nil {
			result.FailedKeys = append(result.FailedKeys, pair.Key)
			result.Failed++

			if opts.Logger != nil {
				opts.Logger.Error("single-key put failed", "key", pair.Key, "error", err)
			}

			return fmt.Errorf("single-key fallback failed for key %s: %w", pair.Key, err)
		}

		result.Succeeded++

		if onProgress != nil {
			onProgress(baseIdx+j+1, result.Total, pair.Key)
		}
	}

	return nil
}

type GetOptions struct {
	SortOrder    string // ASCEND or DESCEND
	SortTarget   string // CREATE, KEY, MODIFY, VALUE, or VERSION
	RangeEnd     string // End of key range
	Limit        int64  // Maximum number of results
	Revision     int64  // Get at specific revision
	MinModRev    int64  // Minimum modify revision
	MaxModRev    int64  // Maximum modify revision
	MinCreateRev int64  // Minimum create revision
	MaxCreateRev int64  // Maximum create revision
	Prefix       bool   // Get keys with matching prefix
	FromKey      bool   // Get keys >= given key
	KeysOnly     bool   // Return only keys, not values
	CountOnly    bool   // Return only count
}

type KeyValue struct {
	Key            string
	Value          string
	CreateRevision int64
	ModRevision    int64
	Version        int64
	Lease          int64
}

type GetResponse struct {
	Kvs   []*KeyValue
	Count int64
	More  bool
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	opts := &GetOptions{}
	resp, err := c.GetWithOptions(ctx, key, opts)
	if err != nil {
		return "", err
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return resp.Kvs[0].Value, nil
}

func (c *Client) GetWithOptions(ctx context.Context, key string, opts *GetOptions) (*GetResponse, error) {
	clientOpts, err := buildClientOptions(opts)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Get(ctx, key, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}

	result := &GetResponse{
		Count: resp.Count,
		More:  resp.More,
		Kvs:   make([]*KeyValue, len(resp.Kvs)),
	}

	for i, kv := range resp.Kvs {
		result.Kvs[i] = &KeyValue{
			Key:            string(kv.Key),
			Value:          string(kv.Value),
			CreateRevision: kv.CreateRevision,
			ModRevision:    kv.ModRevision,
			Version:        kv.Version,
			Lease:          kv.Lease,
		}
	}

	return result, nil
}

func (c *Client) Delete(ctx context.Context, key string) (int64, error) {
	resp, err := c.client.Delete(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to delete key %s: %w", key, err)
	}
	return resp.Deleted, nil
}

func (c *Client) DeletePrefix(ctx context.Context, prefix string) (int64, error) {
	resp, err := c.client.Delete(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return 0, fmt.Errorf("failed to delete prefix %s: %w", prefix, err)
	}
	return resp.Deleted, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}

func (c *Client) Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error) {
	return c.client.Status(ctx, endpoint)
}

func formatValue(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case int, int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case map[string]any:
		var lines []string
		for k, val := range v {
			lines = append(lines, fmt.Sprintf("%s: %v", k, val))
		}
		return strings.Join(lines, "\n")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Compile-time verification that Client implements EtcdClient
var _ EtcdClient = (*Client)(nil)
