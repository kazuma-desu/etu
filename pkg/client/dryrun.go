package client

import (
	"context"
	"fmt"

	"github.com/kazuma-desu/etu/pkg/models"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Operation struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

type DryRunClient struct {
	operations []Operation
	reader     EtcdReader
}

func NewDryRunClient() *DryRunClient {
	return &DryRunClient{
		operations: make([]Operation, 0),
	}
}

func NewDryRunClientWithReader(reader EtcdReader) *DryRunClient {
	return &DryRunClient{
		operations: make([]Operation, 0),
		reader:     reader,
	}
}

func (d *DryRunClient) Put(ctx context.Context, key, value string) error {
	d.operations = append(d.operations, Operation{
		Type:  "PUT",
		Key:   key,
		Value: value,
	})
	return nil
}

func (d *DryRunClient) PutAll(ctx context.Context, pairs []*models.ConfigPair) error {
	for _, pair := range pairs {
		d.operations = append(d.operations, Operation{
			Type:  "PUT",
			Key:   pair.Key,
			Value: formatValue(pair.Value),
		})
	}
	return nil
}

func (d *DryRunClient) Get(ctx context.Context, key string) (string, error) {
	if d.reader != nil {
		return d.reader.Get(ctx, key)
	}
	return "", fmt.Errorf("dry-run mode: cannot read key %q without connection", key)
}

func (d *DryRunClient) GetWithOptions(ctx context.Context, key string, opts *GetOptions) (*GetResponse, error) {
	if d.reader != nil {
		return d.reader.GetWithOptions(ctx, key, opts)
	}
	return nil, fmt.Errorf("dry-run mode: cannot read keys without connection")
}

func (d *DryRunClient) Close() error {
	return nil
}

func (d *DryRunClient) Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error) {
	return nil, fmt.Errorf("dry-run mode: status check not available")
}

func (d *DryRunClient) Operations() []Operation {
	result := make([]Operation, len(d.operations))
	copy(result, d.operations)
	return result
}

func (d *DryRunClient) OperationCount() int {
	return len(d.operations)
}

var _ EtcdClient = (*DryRunClient)(nil)
