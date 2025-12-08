package client

import (
	"context"
	"fmt"
	"time"

	"github.com/kazuma-desu/etu/pkg/models"

	"github.com/charmbracelet/log"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/grpclog"
)

func init() {
	// Redirect gRPC logs through charmbracelet/log at package init
	// This only needs to be done once globally
	grpclog.SetLoggerV2(&grpcLogger{})
}

// Client wraps the etcd client v3 with convenience methods
type Client struct {
	client *clientv3.Client
	config *Config
}

// Config holds etcd client configuration
type Config struct {
	Username    string
	Password    string
	Endpoints   []string
	DialTimeout time.Duration
}

// grpcLogger wraps charmbracelet/log to implement grpclog.LoggerV2
type grpcLogger struct{}

func (g *grpcLogger) Info(args ...any) {
	if len(args) > 0 {
		log.Debug(fmt.Sprint(args...))
	}
}

func (g *grpcLogger) Infoln(args ...any) {
	if len(args) > 0 {
		log.Debug(fmt.Sprint(args...))
	}
}

func (g *grpcLogger) Infof(format string, args ...any) {
	log.Debugf(format, args...)
}

func (g *grpcLogger) Warning(_ ...any) {
	// Suppress gRPC warnings - they're too verbose for user-facing output
	// Users will see our clean error messages instead
}

func (g *grpcLogger) Warningln(_ ...any) {
	// Suppress gRPC warnings - they're too verbose for user-facing output
}

func (g *grpcLogger) Warningf(_ string, _ ...any) {
	// Suppress gRPC warnings - they're too verbose for user-facing output
}

func (g *grpcLogger) Error(args ...any) {
	if len(args) > 0 {
		log.Error(fmt.Sprint(args...))
	}
}

func (g *grpcLogger) Errorln(args ...any) {
	if len(args) > 0 {
		log.Error(fmt.Sprint(args...))
	}
}

func (g *grpcLogger) Errorf(format string, args ...any) {
	log.Errorf(format, args...)
}

func (g *grpcLogger) Fatal(args ...any) {
	if len(args) > 0 {
		log.Fatal(fmt.Sprint(args...))
	}
}

func (g *grpcLogger) Fatalln(args ...any) {
	if len(args) > 0 {
		log.Fatal(fmt.Sprint(args...))
	}
}

func (g *grpcLogger) Fatalf(format string, args ...any) {
	log.Fatalf(format, args...)
}

func (g *grpcLogger) V(l int) bool { return l <= 0 }

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

// Get retrieves a value from etcd
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	if len(resp.Kvs) == 0 {
		return "", fmt.Errorf("key not found: %s", key)
	}

	return string(resp.Kvs[0].Value), nil
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
