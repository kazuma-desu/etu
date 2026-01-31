//go:build integration

package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Watch_Integration(t *testing.T) {
	t.Run("watch single key", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)
		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Start watching
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		opts := &WatchOptions{}
		watchChan := client.Watch(watchCtx, "/watch/test/key", opts)

		// Give watch time to establish
		time.Sleep(100 * time.Millisecond)

		// Put a value
		err := client.Put(ctx, "/watch/test/key", "value1")
		require.NoError(t, err)

		// Wait for event
		select {
		case resp := <-watchChan:
			require.Len(t, resp.Events, 1)
			assert.Equal(t, WatchEventPut, resp.Events[0].Type)
			assert.Equal(t, "/watch/test/key", resp.Events[0].Key)
			assert.Equal(t, "value1", resp.Events[0].Value)
			assert.Greater(t, resp.Events[0].Revision, int64(0))
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for watch event")
		}

		cancel()
		// Drain channel
		for range watchChan {
		}
	})

	t.Run("watch with prefix", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)
		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		opts := &WatchOptions{Prefix: true}
		watchChan := client.Watch(watchCtx, "/watch/prefix/", opts)

		time.Sleep(100 * time.Millisecond)

		// Put multiple keys with same prefix
		err := client.Put(ctx, "/watch/prefix/key1", "value1")
		require.NoError(t, err)
		err = client.Put(ctx, "/watch/prefix/key2", "value2")
		require.NoError(t, err)

		// Collect events
		events := make([]WatchEvent, 0, 2)
		timeout := time.After(5 * time.Second)

		for len(events) < 2 {
			select {
			case resp := <-watchChan:
				events = append(events, resp.Events...)
			case <-timeout:
				t.Fatalf("timeout waiting for events, got %d", len(events))
			}
		}

		assert.Len(t, events, 2)
		keys := []string{events[0].Key, events[1].Key}
		assert.Contains(t, keys, "/watch/prefix/key1")
		assert.Contains(t, keys, "/watch/prefix/key2")

		cancel()
		for range watchChan {
		}
	})

	t.Run("watch delete event", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)
		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Create key first
		err := client.Put(ctx, "/watch/delete/key", "value")
		require.NoError(t, err)

		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		watchChan := client.Watch(watchCtx, "/watch/delete/key", nil)
		time.Sleep(100 * time.Millisecond)

		// Delete the key
		_, err = client.Delete(ctx, "/watch/delete/key")
		require.NoError(t, err)

		select {
		case resp := <-watchChan:
			require.Len(t, resp.Events, 1)
			assert.Equal(t, WatchEventDelete, resp.Events[0].Type)
			assert.Equal(t, "/watch/delete/key", resp.Events[0].Key)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for delete event")
		}

		cancel()
		for range watchChan {
		}
	})

	t.Run("watch with prev_kv", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)
		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Create key
		err := client.Put(ctx, "/watch/prev/key", "original")
		require.NoError(t, err)

		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		opts := &WatchOptions{PrevKV: true}
		watchChan := client.Watch(watchCtx, "/watch/prev/key", opts)
		time.Sleep(100 * time.Millisecond)

		// Update key
		err = client.Put(ctx, "/watch/prev/key", "updated")
		require.NoError(t, err)

		select {
		case resp := <-watchChan:
			require.Len(t, resp.Events, 1)
			assert.Equal(t, WatchEventPut, resp.Events[0].Type)
			assert.Equal(t, "updated", resp.Events[0].Value)
			require.NotNil(t, resp.Events[0].PrevValue)
			assert.Equal(t, "original", *resp.Events[0].PrevValue)
		case <-time.After(5 * time.Second):
			t.Fatal("timeout waiting for event with prev_kv")
		}

		cancel()
		for range watchChan {
		}
	})

	t.Run("watch from specific revision", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)
		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		// Create key and get revision
		err := client.Put(ctx, "/watch/rev/key", "value1")
		require.NoError(t, err)

		// Get the current value to find revision
		resp, err := client.GetWithOptions(ctx, "/watch/rev/key", nil)
		require.NoError(t, err)
		require.Len(t, resp.Kvs, 1)
		firstRev := resp.Kvs[0].ModRevision

		// Update to create a new revision
		err = client.Put(ctx, "/watch/rev/key", "value2")
		require.NoError(t, err)

		// Watch from the first revision
		watchCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		opts := &WatchOptions{Revision: firstRev}
		watchChan := client.Watch(watchCtx, "/watch/rev/key", opts)

		// Should get both events
		events := make([]WatchEvent, 0, 2)
		timeout := time.After(5 * time.Second)

		for len(events) < 2 {
			select {
			case resp := <-watchChan:
				events = append(events, resp.Events...)
			case <-timeout:
				t.Fatalf("timeout waiting for events, got %d", len(events))
			}
		}

		assert.Len(t, events, 2)
		assert.Equal(t, "value1", events[0].Value)
		assert.Equal(t, "value2", events[1].Value)

		cancel()
		for range watchChan {
		}
	})

	t.Run("context cancellation stops watch", func(t *testing.T) {
		endpoint := setupEtcdContainer(t)
		client := newTestClient(t, endpoint)
		ctx := testContext(t)

		watchCtx, cancel := context.WithCancel(ctx)
		watchChan := client.Watch(watchCtx, "/watch/cancel/key", nil)

		// Cancel immediately
		cancel()

		// Channel should close
		select {
		case _, ok := <-watchChan:
			if ok {
				// Drain remaining
				for range watchChan {
				}
			}
			// Success - channel closed
		case <-time.After(2 * time.Second):
			t.Fatal("watch channel should close after context cancellation")
		}
	})
}
