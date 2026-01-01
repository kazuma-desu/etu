package client

import (
	"context"
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDryRunClient_Put(t *testing.T) {
	client := NewDryRunClient()

	err := client.Put(context.Background(), "/app/host", "localhost")

	assert.NoError(t, err)
	ops := client.Operations()
	require.Len(t, ops, 1)
	assert.Equal(t, "PUT", ops[0].Type)
	assert.Equal(t, "/app/host", ops[0].Key)
	assert.Equal(t, "localhost", ops[0].Value)
}

func TestDryRunClient_PutAll(t *testing.T) {
	client := NewDryRunClient()
	pairs := []*models.ConfigPair{
		{Key: "/app/name", Value: "myapp"},
		{Key: "/app/port", Value: int64(8080)},
	}

	err := client.PutAll(context.Background(), pairs)

	assert.NoError(t, err)
	assert.Equal(t, 2, client.OperationCount())

	ops := client.Operations()
	assert.Equal(t, "/app/name", ops[0].Key)
	assert.Equal(t, "myapp", ops[0].Value)
	assert.Equal(t, "/app/port", ops[1].Key)
	assert.Equal(t, "8080", ops[1].Value)
}

func TestDryRunClient_Get_WithoutReader(t *testing.T) {
	client := NewDryRunClient()

	_, err := client.Get(context.Background(), "/app/host")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dry-run mode")
}

func TestDryRunClient_Get_WithReader(t *testing.T) {
	mock := NewMockClient()
	mock.GetFunc = func(ctx context.Context, key string) (string, error) {
		return "test-value", nil
	}

	client := NewDryRunClientWithReader(mock)

	value, err := client.Get(context.Background(), "/app/host")

	assert.NoError(t, err)
	assert.Equal(t, "test-value", value)
}

func TestDryRunClient_Close(t *testing.T) {
	client := NewDryRunClient()
	err := client.Close()
	assert.NoError(t, err)
}

func TestDryRunClient_Operations_ReturnsCopy(t *testing.T) {
	client := NewDryRunClient()
	client.Put(context.Background(), "/key", "value")

	ops1 := client.Operations()
	ops2 := client.Operations()

	ops1[0].Key = "modified"

	assert.Equal(t, "/key", ops2[0].Key)
}

func TestDryRunClient_ImplementsInterface(t *testing.T) {
	var _ EtcdClient = (*DryRunClient)(nil)
}
