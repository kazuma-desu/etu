package client

import (
	"context"
	"errors"
	"testing"

	"github.com/kazuma-desu/etu/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockClient_Put(t *testing.T) {
	t.Run("default behavior returns nil", func(t *testing.T) {
		mock := NewMockClient()
		err := mock.Put(context.Background(), "/key", "value")

		assert.NoError(t, err)
		assert.Len(t, mock.PutCalls, 1)
		assert.Equal(t, "/key", mock.PutCalls[0].Key)
		assert.Equal(t, "value", mock.PutCalls[0].Value)
	})

	t.Run("custom function is called", func(t *testing.T) {
		expectedErr := errors.New("put failed")
		mock := NewMockClient()
		mock.PutFunc = func(ctx context.Context, key, value string) error {
			return expectedErr
		}

		err := mock.Put(context.Background(), "/key", "value")

		assert.Equal(t, expectedErr, err)
		assert.Len(t, mock.PutCalls, 1)
	})
}

func TestMockClient_PutAll(t *testing.T) {
	mock := NewMockClient()
	pairs := []*models.ConfigPair{
		{Key: "/app/name", Value: "test"},
		{Key: "/app/port", Value: "8080"},
	}

	err := mock.PutAll(context.Background(), pairs)

	assert.NoError(t, err)
	require.Len(t, mock.PutAllCalls, 1)
	assert.Equal(t, pairs, mock.PutAllCalls[0])
}

func TestMockClient_Get(t *testing.T) {
	t.Run("custom function returns value", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetFunc = func(ctx context.Context, key string) (string, error) {
			return "test-value", nil
		}

		value, err := mock.Get(context.Background(), "/test/key")

		assert.NoError(t, err)
		assert.Equal(t, "test-value", value)
		assert.Equal(t, []string{"/test/key"}, mock.GetCalls)
	})

	t.Run("tracks multiple calls", func(t *testing.T) {
		mock := NewMockClient()

		mock.Get(context.Background(), "/key1")
		mock.Get(context.Background(), "/key2")
		mock.Get(context.Background(), "/key3")

		assert.Equal(t, []string{"/key1", "/key2", "/key3"}, mock.GetCalls)
	})
}

func TestMockClient_Close(t *testing.T) {
	mock := NewMockClient()
	assert.False(t, mock.CloseCalled)

	err := mock.Close()

	assert.NoError(t, err)
	assert.True(t, mock.CloseCalled)
}

func TestMockClient_Reset(t *testing.T) {
	mock := NewMockClient()
	mock.Put(context.Background(), "/key", "value")
	mock.Get(context.Background(), "/key")
	mock.Close()

	mock.Reset()

	assert.Empty(t, mock.PutCalls)
	assert.Empty(t, mock.GetCalls)
	assert.False(t, mock.CloseCalled)
}

func TestMockClient_ImplementsInterface(t *testing.T) {
	var _ EtcdClient = (*MockClient)(nil)
}
