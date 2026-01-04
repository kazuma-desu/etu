package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/models"
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

func TestDryRunClient_PutAllWithProgress(t *testing.T) {
	t.Run("records operations with progress", func(t *testing.T) {
		client := NewDryRunClient()
		pairs := []*models.ConfigPair{
			{Key: "/app/name", Value: "myapp"},
			{Key: "/app/port", Value: int64(8080)},
		}

		var progressCalls []int
		onProgress := func(current, _ int, _ string) {
			progressCalls = append(progressCalls, current)
		}

		result, err := client.PutAllWithProgress(context.Background(), pairs, onProgress)

		assert.NoError(t, err)
		assert.Equal(t, 2, result.Succeeded)
		assert.Equal(t, 0, result.Failed)
		assert.Equal(t, 2, result.Total)
		assert.Equal(t, []int{1, 2}, progressCalls)
		assert.Equal(t, 2, client.OperationCount())
	})

	t.Run("nil progress callback is handled", func(t *testing.T) {
		client := NewDryRunClient()
		pairs := []*models.ConfigPair{{Key: "/key", Value: "val"}}

		result, err := client.PutAllWithProgress(context.Background(), pairs, nil)

		assert.NoError(t, err)
		assert.Equal(t, 1, result.Succeeded)
	})
}

func TestDryRunClient_Get_WithoutReader(t *testing.T) {
	client := NewDryRunClient()

	_, err := client.Get(context.Background(), "/app/host")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dry-run mode")
}

func TestDryRunClient_Get_WithReader(t *testing.T) {
	mock := NewMockClient()
	mock.GetFunc = func(_ context.Context, _ string) (string, error) {
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

func TestDryRunClient_GetWithOptions_WithoutReader(t *testing.T) {
	client := NewDryRunClient()

	_, err := client.GetWithOptions(context.Background(), "/prefix/", &GetOptions{Prefix: true})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dry-run mode")
	assert.Contains(t, err.Error(), "without connection")
}

func TestDryRunClient_GetWithOptions_WithReader(t *testing.T) {
	mock := NewMockClient()
	mock.GetWithOptionsFunc = func(_ context.Context, _ string, _ *GetOptions) (*GetResponse, error) {
		return &GetResponse{
			Kvs: []*KeyValue{
				{Key: "/prefix/key1", Value: "val1"},
				{Key: "/prefix/key2", Value: "val2"},
			},
			Count: 2,
		}, nil
	}

	client := NewDryRunClientWithReader(mock)

	resp, err := client.GetWithOptions(context.Background(), "/prefix/", &GetOptions{Prefix: true})

	assert.NoError(t, err)
	assert.Len(t, resp.Kvs, 2)
	assert.Equal(t, "/prefix/key1", resp.Kvs[0].Key)
}

func TestDryRunClient_Status(t *testing.T) {
	client := NewDryRunClient()

	_, err := client.Status(context.Background(), "http://localhost:2379")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dry-run mode")
}

func TestDryRunClient_ImplementsInterface(_ *testing.T) {
	var _ EtcdClient = (*DryRunClient)(nil)
}

func TestDryRunClient_ImplementsOperationRecorder(_ *testing.T) {
	var _ OperationRecorder = (*DryRunClient)(nil)
}
