package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/models"
)

func TestMockClient_Integration(t *testing.T) {
	t.Run("tracks put operations", func(t *testing.T) {
		mock := client.NewMockClient()
		ctx := context.Background()

		pairs := []*models.ConfigPair{
			{Key: "/app/name", Value: "test-app"},
			{Key: "/app/port", Value: "8080"},
		}

		err := mock.PutAll(ctx, pairs)

		assert.NoError(t, err)
		assert.Len(t, mock.PutAllCalls, 1)
		assert.Equal(t, pairs, mock.PutAllCalls[0])
	})

	t.Run("simulates connection errors", func(t *testing.T) {
		mock := client.NewMockClient()
		mock.PutAllFunc = func(_ context.Context, _ []*models.ConfigPair) error {
			return errors.New("connection refused")
		}

		err := mock.PutAll(context.Background(), []*models.ConfigPair{
			{Key: "/app/name", Value: "test"},
		})

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "connection refused")
	})

	t.Run("simulates timeout", func(t *testing.T) {
		mock := client.NewMockClient()
		mock.GetFunc = func(_ context.Context, _ string) (string, error) {
			return "", context.DeadlineExceeded
		}

		_, err := mock.Get(context.Background(), "/app/config")

		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestDryRunClient_Integration(t *testing.T) {
	t.Run("records all operations", func(t *testing.T) {
		dryClient := client.NewDryRunClient()
		ctx := context.Background()

		pairs := []*models.ConfigPair{
			{Key: "/db/host", Value: "localhost"},
			{Key: "/db/port", Value: "5432"},
		}

		err := dryClient.PutAll(ctx, pairs)

		assert.NoError(t, err)
		ops := dryClient.Operations()
		assert.Len(t, ops, 2)
		assert.Equal(t, "PUT", ops[0].Type)
		assert.Equal(t, "/db/host", ops[0].Key)
		assert.Equal(t, "localhost", ops[0].Value)
	})
}
