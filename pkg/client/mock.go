package client

import (
	"context"

	"github.com/kazuma-desu/etu/pkg/models"
)

type PutCall struct {
	Key   string
	Value string
}

type GetWithOptionsCall struct {
	Opts *GetOptions
	Key  string
}

type PutAllWithProgressCall struct {
	Pairs []*models.ConfigPair
}

type MockClient struct {
	PutFunc                func(ctx context.Context, key, value string) error
	PutAllFunc             func(ctx context.Context, pairs []*models.ConfigPair) error
	PutAllWithProgressFunc func(ctx context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc) (*PutAllResult, error)
	PutAllWithOptionsFunc  func(ctx context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc, opts *BatchOptions) (*PutAllResult, error)
	GetFunc                func(ctx context.Context, key string) (string, error)
	GetWithOptionsFunc     func(ctx context.Context, key string, opts *GetOptions) (*GetResponse, error)
	DeleteFunc             func(ctx context.Context, key string) (int64, error)
	DeletePrefixFunc       func(ctx context.Context, prefix string) (int64, error)
	CloseFunc              func() error
	StatusFunc             func(ctx context.Context, endpoint string) (*StatusResponse, error)

	PutCalls                []PutCall
	PutAllCalls             [][]*models.ConfigPair
	PutAllWithProgressCalls []PutAllWithProgressCall
	GetCalls                []string
	GetWithOptionsCalls     []GetWithOptionsCall
	DeleteCalls             []string
	DeletePrefixCalls       []string
	StatusCalls             []string
	CloseCalled             bool
}

func NewMockClient() *MockClient {
	return &MockClient{
		PutCalls:                make([]PutCall, 0),
		PutAllCalls:             make([][]*models.ConfigPair, 0),
		PutAllWithProgressCalls: make([]PutAllWithProgressCall, 0),
		GetCalls:                make([]string, 0),
		GetWithOptionsCalls:     make([]GetWithOptionsCall, 0),
		DeleteCalls:             make([]string, 0),
		DeletePrefixCalls:       make([]string, 0),
		StatusCalls:             make([]string, 0),
	}
}

func (m *MockClient) Put(ctx context.Context, key, value string) error {
	m.PutCalls = append(m.PutCalls, PutCall{Key: key, Value: value})
	if m.PutFunc != nil {
		return m.PutFunc(ctx, key, value)
	}
	return nil
}

func (m *MockClient) PutAll(ctx context.Context, pairs []*models.ConfigPair) error {
	pairsCopy := make([]*models.ConfigPair, len(pairs))
	copy(pairsCopy, pairs)
	m.PutAllCalls = append(m.PutAllCalls, pairsCopy)
	if m.PutAllFunc != nil {
		return m.PutAllFunc(ctx, pairs)
	}
	return nil
}

func (m *MockClient) PutAllWithProgress(ctx context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc) (*PutAllResult, error) {
	return m.PutAllWithOptions(ctx, pairs, onProgress, nil)
}

func (m *MockClient) PutAllWithOptions(ctx context.Context, pairs []*models.ConfigPair, onProgress ProgressFunc, opts *BatchOptions) (*PutAllResult, error) {
	pairsCopy := make([]*models.ConfigPair, len(pairs))
	copy(pairsCopy, pairs)
	m.PutAllWithProgressCalls = append(m.PutAllWithProgressCalls, PutAllWithProgressCall{Pairs: pairsCopy})

	if m.PutAllWithOptionsFunc != nil {
		return m.PutAllWithOptionsFunc(ctx, pairs, onProgress, opts)
	}

	if m.PutAllWithProgressFunc != nil {
		return m.PutAllWithProgressFunc(ctx, pairs, onProgress)
	}

	result := &PutAllResult{Total: len(pairs)}
	for i, pair := range pairs {
		result.Succeeded++
		if onProgress != nil {
			onProgress(i+1, result.Total, pair.Key)
		}
	}
	return result, nil
}

func (m *MockClient) Get(ctx context.Context, key string) (string, error) {
	m.GetCalls = append(m.GetCalls, key)
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "", nil
}

func (m *MockClient) GetWithOptions(ctx context.Context, key string, opts *GetOptions) (*GetResponse, error) {
	var optsCopy *GetOptions
	if opts != nil {
		copied := *opts
		optsCopy = &copied
	}
	m.GetWithOptionsCalls = append(m.GetWithOptionsCalls, GetWithOptionsCall{Opts: optsCopy, Key: key})
	if m.GetWithOptionsFunc != nil {
		return m.GetWithOptionsFunc(ctx, key, opts)
	}
	return &GetResponse{Kvs: []*KeyValue{}}, nil
}

func (m *MockClient) Delete(ctx context.Context, key string) (int64, error) {
	m.DeleteCalls = append(m.DeleteCalls, key)
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}
	return 1, nil
}

func (m *MockClient) DeletePrefix(ctx context.Context, prefix string) (int64, error) {
	m.DeletePrefixCalls = append(m.DeletePrefixCalls, prefix)
	if m.DeletePrefixFunc != nil {
		return m.DeletePrefixFunc(ctx, prefix)
	}
	return 1, nil
}

func (m *MockClient) Close() error {
	m.CloseCalled = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockClient) Status(ctx context.Context, endpoint string) (*StatusResponse, error) {
	m.StatusCalls = append(m.StatusCalls, endpoint)
	if m.StatusFunc != nil {
		return m.StatusFunc(ctx, endpoint)
	}
	return &StatusResponse{}, nil
}

func (m *MockClient) Reset() {
	m.PutCalls = make([]PutCall, 0)
	m.PutAllCalls = make([][]*models.ConfigPair, 0)
	m.PutAllWithProgressCalls = make([]PutAllWithProgressCall, 0)
	m.GetCalls = make([]string, 0)
	m.GetWithOptionsCalls = make([]GetWithOptionsCall, 0)
	m.DeleteCalls = make([]string, 0)
	m.DeletePrefixCalls = make([]string, 0)
	m.StatusCalls = make([]string, 0)
	m.CloseCalled = false
}

func (m *MockClient) Operations() []Operation {
	ops := make([]Operation, 0, m.OperationCount())

	for _, call := range m.PutCalls {
		ops = append(ops, Operation{Type: "PUT", Key: call.Key, Value: call.Value})
	}

	for _, call := range m.PutAllWithProgressCalls {
		for _, pair := range call.Pairs {
			ops = append(ops, Operation{Type: "PUT", Key: pair.Key, Value: formatValue(pair.Value)})
		}
	}

	return ops
}

func (m *MockClient) OperationCount() int {
	count := len(m.PutCalls)
	for _, call := range m.PutAllWithProgressCalls {
		count += len(call.Pairs)
	}
	return count
}

var (
	_ EtcdClient        = (*MockClient)(nil)
	_ OperationRecorder = (*MockClient)(nil)
)
