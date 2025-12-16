package client

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/kazuma-desu/etu/pkg/models"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/grpclog"
)

func init() {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(io.Discard, io.Discard, io.Discard))
}

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

// NewClient creates a new etcd client
func NewClient(cfg *Config) (*Client, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, fmt.Errorf("at least one endpoint is required")
	}

	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 5 * time.Second
	}

	clientConfig := clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
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

// Put writes a single key-value pair to etcd
func (c *Client) Put(ctx context.Context, key, value string) error {
	_, err := c.client.Put(ctx, key, value)
	if err != nil {
		return fmt.Errorf("failed to put key %s: %w", key, err)
	}
	return nil
}

// PutAll writes multiple configuration pairs to etcd
func (c *Client) PutAll(ctx context.Context, pairs []*models.ConfigPair) error {
	for _, pair := range pairs {
		value := formatValue(pair.Value)
		if err := c.Put(ctx, pair.Key, value); err != nil {
			return err
		}
	}
	return nil
}

// GetOptions contains options for Get operations
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

// KeyValue represents a key-value pair with metadata
type KeyValue struct {
	Key            string
	Value          string
	CreateRevision int64
	ModRevision    int64
	Version        int64
	Lease          int64
}

// GetResponse contains the response from a Get operation
type GetResponse struct {
	Kvs   []*KeyValue
	Count int64
	More  bool
}

// Get retrieves a value from etcd (simple version for backward compatibility)
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

// GetWithOptions retrieves keys from etcd with various options
func (c *Client) GetWithOptions(ctx context.Context, key string, opts *GetOptions) (*GetResponse, error) {
	// Build options array
	var clientOpts []clientv3.OpOption

	// Handle prefix mode
	if opts.Prefix {
		clientOpts = append(clientOpts, clientv3.WithPrefix())
	}

	// Handle from-key mode
	if opts.FromKey {
		clientOpts = append(clientOpts, clientv3.WithFromKey())
	}

	// Handle range
	if opts.RangeEnd != "" {
		clientOpts = append(clientOpts, clientv3.WithRange(opts.RangeEnd))
	}

	// Handle limit
	if opts.Limit > 0 {
		clientOpts = append(clientOpts, clientv3.WithLimit(opts.Limit))
	}

	// Handle revision
	if opts.Revision > 0 {
		clientOpts = append(clientOpts, clientv3.WithRev(opts.Revision))
	}

	// Handle sort order
	if opts.SortOrder != "" || opts.SortTarget != "" {
		var order clientv3.SortOrder
		var target clientv3.SortTarget

		// Parse order
		switch opts.SortOrder {
		case "ASCEND", "":
			order = clientv3.SortAscend
		case "DESCEND":
			order = clientv3.SortDescend
		default:
			return nil, fmt.Errorf("invalid sort order: %s (use ASCEND or DESCEND)", opts.SortOrder)
		}

		// Parse target
		switch opts.SortTarget {
		case "KEY", "":
			target = clientv3.SortByKey
		case "VERSION":
			target = clientv3.SortByVersion
		case "CREATE":
			target = clientv3.SortByCreateRevision
		case "MODIFY":
			target = clientv3.SortByModRevision
		case "VALUE":
			target = clientv3.SortByValue
		default:
			return nil, fmt.Errorf("invalid sort target: %s", opts.SortTarget)
		}

		clientOpts = append(clientOpts, clientv3.WithSort(target, order))
	}

	// Handle keys-only
	if opts.KeysOnly {
		clientOpts = append(clientOpts, clientv3.WithKeysOnly())
	}

	// Handle count-only
	if opts.CountOnly {
		clientOpts = append(clientOpts, clientv3.WithCountOnly())
	}

	// Handle revision filters
	if opts.MinModRev > 0 {
		clientOpts = append(clientOpts, clientv3.WithMinModRev(opts.MinModRev))
	}
	if opts.MaxModRev > 0 {
		clientOpts = append(clientOpts, clientv3.WithMaxModRev(opts.MaxModRev))
	}
	if opts.MinCreateRev > 0 {
		clientOpts = append(clientOpts, clientv3.WithMinCreateRev(opts.MinCreateRev))
	}
	if opts.MaxCreateRev > 0 {
		clientOpts = append(clientOpts, clientv3.WithMaxCreateRev(opts.MaxCreateRev))
	}

	// Execute get
	resp, err := c.client.Get(ctx, key, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}

	// Convert response
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

// Close closes the etcd client connection
func (c *Client) Close() error {
	return c.client.Close()
}

// Status gets the status of an etcd endpoint
func (c *Client) Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error) {
	return c.client.Status(ctx, endpoint)
}

// formatValue converts various value types to string format for etcd
func formatValue(val any) string {
	switch v := val.(type) {
	case string:
		return v
	case int, int64:
		return fmt.Sprintf("%d", v)
	case float64:
		return fmt.Sprintf("%f", v)
	case map[string]any:
		// Format as key: value lines
		var lines []string
		for k, val := range v {
			lines = append(lines, fmt.Sprintf("%s: %v", k, val))
		}
		return fmt.Sprintf("%v", lines)
	default:
		return fmt.Sprintf("%v", v)
	}
}
