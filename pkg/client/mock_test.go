package client

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/kazuma-desu/etu/pkg/models"
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
		mock.PutFunc = func(_ context.Context, _, _ string) error {
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

func TestMockClient_PutAllWithProgress(t *testing.T) {
	t.Run("default behavior with progress callback", func(t *testing.T) {
		mock := NewMockClient()
		pairs := []*models.ConfigPair{
			{Key: "/app/name", Value: "test"},
			{Key: "/app/port", Value: "8080"},
		}

		var progressCalls []string
		onProgress := func(_, _ int, key string) {
			progressCalls = append(progressCalls, key)
		}

		result, err := mock.PutAllWithProgress(context.Background(), pairs, onProgress)

		assert.NoError(t, err)
		assert.Equal(t, 2, result.Succeeded)
		assert.Equal(t, 0, result.Failed)
		assert.Equal(t, 2, result.Total)
		assert.Empty(t, result.FailedKey())
		assert.Equal(t, []string{"/app/name", "/app/port"}, progressCalls)
		require.Len(t, mock.PutAllWithProgressCalls, 1)
	})

	t.Run("custom function simulates partial failure", func(t *testing.T) {
		mock := NewMockClient()
		pairs := []*models.ConfigPair{
			{Key: "/key1", Value: "val1"},
			{Key: "/key2", Value: "val2"},
		}

		mock.PutAllWithProgressFunc = func(_ context.Context, _ []*models.ConfigPair, _ ProgressFunc) (*PutAllResult, error) {
			return &PutAllResult{
				Succeeded:  1,
				Failed:     1,
				Total:      2,
				FailedKeys: []string{"/key2"},
			}, errors.New("connection lost")
		}

		result, err := mock.PutAllWithProgress(context.Background(), pairs, nil)

		assert.Error(t, err)
		assert.Equal(t, 1, result.Succeeded)
		assert.Equal(t, "/key2", result.FailedKey())
	})

	t.Run("nil progress callback is handled", func(t *testing.T) {
		mock := NewMockClient()
		pairs := []*models.ConfigPair{{Key: "/key", Value: "val"}}

		result, err := mock.PutAllWithProgress(context.Background(), pairs, nil)

		assert.NoError(t, err)
		assert.Equal(t, 1, result.Succeeded)
	})
}

func TestMockClient_Get(t *testing.T) {
	t.Run("custom function returns value", func(t *testing.T) {
		mock := NewMockClient()
		mock.GetFunc = func(_ context.Context, _ string) (string, error) {
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

func TestMockClient_GetWithOptions(t *testing.T) {
	t.Run("records calls with key and options", func(t *testing.T) {
		mock := NewMockClient()
		opts := &GetOptions{Prefix: true, Limit: 10}

		_, err := mock.GetWithOptions(context.Background(), "/prefix/", opts)

		assert.NoError(t, err)
		require.Len(t, mock.GetWithOptionsCalls, 1)
		assert.Equal(t, "/prefix/", mock.GetWithOptionsCalls[0].Key)
		assert.Equal(t, opts, mock.GetWithOptionsCalls[0].Opts)
	})

	t.Run("tracks multiple calls", func(t *testing.T) {
		mock := NewMockClient()

		mock.GetWithOptions(context.Background(), "/key1", nil)
		mock.GetWithOptions(context.Background(), "/key2", &GetOptions{Prefix: true})

		require.Len(t, mock.GetWithOptionsCalls, 2)
		assert.Equal(t, "/key1", mock.GetWithOptionsCalls[0].Key)
		assert.Equal(t, "/key2", mock.GetWithOptionsCalls[1].Key)
	})
}

func TestMockClient_Status(t *testing.T) {
	t.Run("records endpoint calls", func(t *testing.T) {
		mock := NewMockClient()

		_, err := mock.Status(context.Background(), "http://localhost:2379")

		assert.NoError(t, err)
		require.Len(t, mock.StatusCalls, 1)
		assert.Equal(t, "http://localhost:2379", mock.StatusCalls[0])
	})

	t.Run("tracks multiple calls", func(t *testing.T) {
		mock := NewMockClient()

		mock.Status(context.Background(), "http://node1:2379")
		mock.Status(context.Background(), "http://node2:2379")

		assert.Equal(t, []string{"http://node1:2379", "http://node2:2379"}, mock.StatusCalls)
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
	mock.PutAll(context.Background(), []*models.ConfigPair{{Key: "/all", Value: "value"}})
	mock.PutAllWithProgress(context.Background(), []*models.ConfigPair{{Key: "/progress", Value: "val"}}, nil)
	mock.Get(context.Background(), "/key")
	mock.GetWithOptions(context.Background(), "/prefix/", &GetOptions{Prefix: true})
	mock.Status(context.Background(), "http://localhost:2379")
	mock.Close()

	mock.Reset()

	assert.Empty(t, mock.PutCalls)
	assert.Empty(t, mock.PutAllCalls)
	assert.Empty(t, mock.PutAllWithProgressCalls)
	assert.Empty(t, mock.GetCalls)
	assert.Empty(t, mock.GetWithOptionsCalls)
	assert.Empty(t, mock.StatusCalls)
	assert.False(t, mock.CloseCalled)
}

func TestMockClient_Operations(t *testing.T) {
	t.Run("returns empty slice when no puts", func(t *testing.T) {
		mock := NewMockClient()

		ops := mock.Operations()

		assert.Empty(t, ops)
		assert.Equal(t, 0, mock.OperationCount())
	})

	t.Run("converts put calls to operations", func(t *testing.T) {
		mock := NewMockClient()
		mock.Put(context.Background(), "/key1", "value1")
		mock.Put(context.Background(), "/key2", "value2")

		ops := mock.Operations()

		assert.Len(t, ops, 2)
		assert.Equal(t, 2, mock.OperationCount())
		assert.Equal(t, Operation{Type: "PUT", Key: "/key1", Value: "value1"}, ops[0])
		assert.Equal(t, Operation{Type: "PUT", Key: "/key2", Value: "value2"}, ops[1])
	})

	t.Run("returns copy that is safe to modify", func(t *testing.T) {
		mock := NewMockClient()
		mock.Put(context.Background(), "/key", "value")

		ops1 := mock.Operations()
		ops1[0].Key = "modified"
		ops2 := mock.Operations()

		assert.Equal(t, "/key", ops2[0].Key)
	})

	t.Run("includes PutAllWithProgress calls", func(t *testing.T) {
		mock := NewMockClient()
		mock.Put(context.Background(), "/single", "value")
		mock.PutAllWithProgress(context.Background(), []*models.ConfigPair{
			{Key: "/batch/key1", Value: "batch1"},
			{Key: "/batch/key2", Value: int64(42)},
		}, nil)

		ops := mock.Operations()

		assert.Len(t, ops, 3)
		assert.Equal(t, 3, mock.OperationCount())
		assert.Equal(t, Operation{Type: "PUT", Key: "/single", Value: "value"}, ops[0])
		assert.Equal(t, Operation{Type: "PUT", Key: "/batch/key1", Value: "batch1"}, ops[1])
		assert.Equal(t, Operation{Type: "PUT", Key: "/batch/key2", Value: "42"}, ops[2])
	})
}

func TestMockClient_ImplementsInterface(_ *testing.T) {
	var _ EtcdClient = (*MockClient)(nil)
}

func TestMockClient_ImplementsOperationRecorder(_ *testing.T) {
	var _ OperationRecorder = (*MockClient)(nil)
}

func TestMockClient_Delete(t *testing.T) {
	t.Run("default behavior returns 1", func(t *testing.T) {
		mock := NewMockClient()

		deleted, err := mock.Delete(context.Background(), "/key")

		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
		assert.Equal(t, []string{"/key"}, mock.DeleteCalls)
	})

	t.Run("custom function is called", func(t *testing.T) {
		expectedErr := errors.New("delete failed")
		mock := NewMockClient()
		mock.DeleteFunc = func(_ context.Context, key string) (int64, error) {
			if key == "/protected" {
				return 0, expectedErr
			}
			return 1, nil
		}

		deleted, err := mock.Delete(context.Background(), "/protected")

		assert.Equal(t, expectedErr, err)
		assert.Equal(t, int64(0), deleted)
		assert.Len(t, mock.DeleteCalls, 1)
	})

	t.Run("tracks multiple calls", func(t *testing.T) {
		mock := NewMockClient()

		mock.Delete(context.Background(), "/key1")
		mock.Delete(context.Background(), "/key2")
		mock.Delete(context.Background(), "/key3")

		assert.Equal(t, []string{"/key1", "/key2", "/key3"}, mock.DeleteCalls)
	})
}

func TestMockClient_DeletePrefix(t *testing.T) {
	t.Run("default behavior returns 1", func(t *testing.T) {
		mock := NewMockClient()

		deleted, err := mock.DeletePrefix(context.Background(), "/prefix/")

		assert.NoError(t, err)
		assert.Equal(t, int64(1), deleted)
		assert.Equal(t, []string{"/prefix/"}, mock.DeletePrefixCalls)
	})

	t.Run("custom function is called", func(t *testing.T) {
		mock := NewMockClient()
		mock.DeletePrefixFunc = func(_ context.Context, prefix string) (int64, error) {
			if prefix == "/config/" {
				return 5, nil
			}
			return 0, nil
		}

		deleted, err := mock.DeletePrefix(context.Background(), "/config/")

		assert.NoError(t, err)
		assert.Equal(t, int64(5), deleted)
		assert.Len(t, mock.DeletePrefixCalls, 1)
	})

	t.Run("custom function returns error", func(t *testing.T) {
		expectedErr := errors.New("prefix delete failed")
		mock := NewMockClient()
		mock.DeletePrefixFunc = func(_ context.Context, _ string) (int64, error) {
			return 0, expectedErr
		}

		deleted, err := mock.DeletePrefix(context.Background(), "/any/")

		assert.Equal(t, expectedErr, err)
		assert.Equal(t, int64(0), deleted)
	})

	t.Run("tracks multiple calls", func(t *testing.T) {
		mock := NewMockClient()

		mock.DeletePrefix(context.Background(), "/prefix1/")
		mock.DeletePrefix(context.Background(), "/prefix2/")

		assert.Equal(t, []string{"/prefix1/", "/prefix2/"}, mock.DeletePrefixCalls)
	})
}

func TestMockClient_PutAllWithOptions(t *testing.T) {
	t.Run("custom PutAllWithOptionsFunc is called", func(t *testing.T) {
		mock := NewMockClient()
		mock.PutAllWithOptionsFunc = func(_ context.Context, pairs []*models.ConfigPair, _ ProgressFunc, _ *BatchOptions) (*PutAllResult, error) {
			return &PutAllResult{
				Total:     len(pairs),
				Succeeded: len(pairs),
			}, nil
		}

		pairs := []*models.ConfigPair{{Key: "/key", Value: "val"}}
		opts := &BatchOptions{MaxRetries: 5}
		result, err := mock.PutAllWithOptions(context.Background(), pairs, nil, opts)

		assert.NoError(t, err)
		assert.Equal(t, 1, result.Succeeded)
	})

	t.Run("falls back to PutAllWithProgressFunc when PutAllWithOptionsFunc is nil", func(t *testing.T) {
		mock := NewMockClient()
		var calledWithPairs []*models.ConfigPair
		mock.PutAllWithProgressFunc = func(_ context.Context, pairs []*models.ConfigPair, _ ProgressFunc) (*PutAllResult, error) {
			calledWithPairs = pairs
			return &PutAllResult{Total: len(pairs), Succeeded: len(pairs)}, nil
		}

		pairs := []*models.ConfigPair{{Key: "/key", Value: "val"}}
		result, err := mock.PutAllWithOptions(context.Background(), pairs, nil, nil)

		assert.NoError(t, err)
		assert.Equal(t, 1, result.Succeeded)
		assert.Equal(t, pairs, calledWithPairs)
	})
}

func TestMockClient_CloseWithCustomFunc(t *testing.T) {
	expectedErr := errors.New("close failed")
	mock := NewMockClient()
	mock.CloseFunc = func() error {
		return expectedErr
	}

	err := mock.Close()

	assert.Equal(t, expectedErr, err)
	assert.True(t, mock.CloseCalled)
}

func TestMockClient_StatusWithCustomFunc(t *testing.T) {
	expectedErr := errors.New("status check failed")
	mock := NewMockClient()
	mock.StatusFunc = func(_ context.Context, _ string) (*clientv3.StatusResponse, error) {
		return nil, expectedErr
	}

	resp, err := mock.Status(context.Background(), "http://localhost:2379")

	assert.Equal(t, expectedErr, err)
	assert.Nil(t, resp)
	assert.Len(t, mock.StatusCalls, 1)
}

func TestMockClient_PutAllWithCustomFunc(t *testing.T) {
	expectedErr := errors.New("batch put failed")
	mock := NewMockClient()
	mock.PutAllFunc = func(_ context.Context, _ []*models.ConfigPair) error {
		return expectedErr
	}

	pairs := []*models.ConfigPair{{Key: "/key", Value: "val"}}
	err := mock.PutAll(context.Background(), pairs)

	assert.Equal(t, expectedErr, err)
	assert.Len(t, mock.PutAllCalls, 1)
}
