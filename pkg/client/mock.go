package client

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/kazuma-desu/etu/pkg/models"
)

type PutCall struct {
	Key   string
	Value string
}

type MockClient struct {
	PutFunc            func(ctx context.Context, key, value string) error
	PutAllFunc         func(ctx context.Context, pairs []*models.ConfigPair) error
	GetFunc            func(ctx context.Context, key string) (string, error)
	GetWithOptionsFunc func(ctx context.Context, key string, opts *GetOptions) (*GetResponse, error)
	CloseFunc          func() error
	StatusFunc         func(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error)

	PutCalls    []PutCall
	PutAllCalls [][]*models.ConfigPair
	GetCalls    []string
	CloseCalled bool
}

func NewMockClient() *MockClient {
	return &MockClient{
		PutCalls:    make([]PutCall, 0),
		PutAllCalls: make([][]*models.ConfigPair, 0),
		GetCalls:    make([]string, 0),
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
	m.PutAllCalls = append(m.PutAllCalls, pairs)
	if m.PutAllFunc != nil {
		return m.PutAllFunc(ctx, pairs)
	}
	return nil
}

func (m *MockClient) Get(ctx context.Context, key string) (string, error) {
	m.GetCalls = append(m.GetCalls, key)
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "", nil
}

func (m *MockClient) GetWithOptions(ctx context.Context, key string, opts *GetOptions) (*GetResponse, error) {
	if m.GetWithOptionsFunc != nil {
		return m.GetWithOptionsFunc(ctx, key, opts)
	}
	return &GetResponse{Kvs: []*KeyValue{}}, nil
}

func (m *MockClient) Close() error {
	m.CloseCalled = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func (m *MockClient) Status(ctx context.Context, endpoint string) (*clientv3.StatusResponse, error) {
	if m.StatusFunc != nil {
		return m.StatusFunc(ctx, endpoint)
	}
	return &clientv3.StatusResponse{}, nil
}

func (m *MockClient) Reset() {
	m.PutCalls = make([]PutCall, 0)
	m.PutAllCalls = make([][]*models.ConfigPair, 0)
	m.GetCalls = make([]string, 0)
	m.CloseCalled = false
}

var _ EtcdClient = (*MockClient)(nil)
