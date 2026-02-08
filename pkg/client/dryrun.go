package client

import (
	"context"
	"fmt"

	"github.com/kazuma-desu/etu/pkg/models"
)

type Operation struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

type DryRunClient struct {
	reader     EtcdReader
	operations []Operation
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

func (d *DryRunClient) Put(_ context.Context, key, value string) error {
	d.operations = append(d.operations, Operation{
		Type:  "PUT",
		Key:   key,
		Value: value,
	})
	return nil
}

func (d *DryRunClient) PutAll(ctx context.Context, pairs []*models.ConfigPair) error {
	_, err := d.PutAllWithProgress(ctx, pairs, nil)
	return err
}

func (d *DryRunClient) PutAllWithProgress(ctx context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc) (*PutAllResult, error) {
	return d.PutAllWithOptions(ctx, pairs, onProgress, nil)
}

func (d *DryRunClient) PutAllWithOptions(_ context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc, opts *BatchOptions) (*PutAllResult, error) {
	result := &PutAllResult{Total: len(pairs)}

	if opts != nil {
		warnLargeValues(opts.Logger, pairs)
	}

	for i, pair := range pairs {
		d.operations = append(d.operations, Operation{
			Type:  "PUT",
			Key:   pair.Key,
			Value: formatValue(pair.Value),
		})
		result.Succeeded++

		if onProgress != nil {
			onProgress(i+1, result.Total, pair.Key)
		}
	}

	return result, nil
}

func (d *DryRunClient) Get(ctx context.Context, key string) (string, error) {
	if d.reader != nil {
		return d.reader.Get(ctx, key)
	}
	return "", fmt.Errorf("dry-run mode: cannot read key %q without connection", key)
}

func (d *DryRunClient) GetTyped(ctx context.Context, key string) (any, error) {
	value, err := d.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return models.InferType(value), nil
}

func (d *DryRunClient) GetWithOptions(ctx context.Context, key string, opts *GetOptions) (*GetResponse, error) {
	if d.reader != nil {
		return d.reader.GetWithOptions(ctx, key, opts)
	}
	return nil, fmt.Errorf("dry-run mode: cannot read keys without connection")
}

func (d *DryRunClient) Delete(_ context.Context, key string) (int64, error) {
	d.operations = append(d.operations, Operation{
		Type: "DELETE",
		Key:  key,
	})
	return 1, nil
}

func (d *DryRunClient) DeletePrefix(_ context.Context, prefix string) (int64, error) {
	d.operations = append(d.operations, Operation{
		Type: "DELETE_PREFIX",
		Key:  prefix,
	})
	return 0, nil
}

func (d *DryRunClient) Close() error {
	return nil
}

func (d *DryRunClient) Status(_ context.Context, _ string) (*StatusResponse, error) {
	return nil, fmt.Errorf("dry-run mode: status check not available")
}

func (d *DryRunClient) Watch(_ context.Context, _ string, _ *WatchOptions) WatchChan {
	// In dry-run mode, return a closed channel immediately
	ch := make(chan WatchResponse)
	close(ch)
	return ch
}

func (d *DryRunClient) Operations() []Operation {
	result := make([]Operation, len(d.operations))
	copy(result, d.operations)
	return result
}

func (d *DryRunClient) OperationCount() int {
	return len(d.operations)
}

var (
	_ EtcdClient        = (*DryRunClient)(nil)
	_ OperationRecorder = (*DryRunClient)(nil)
)
