package client

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMockClient_Watch(t *testing.T) {
	t.Run("records watch calls", func(t *testing.T) {
		mock := NewMockClient()

		ch := mock.Watch(context.Background(), "/test/key", &WatchOptions{Prefix: true, Revision: 100})

		// Channel should be closed immediately by default
		_, ok := <-ch
		assert.False(t, ok, "channel should be closed")

		assert.Len(t, mock.WatchCalls, 1)
		assert.Equal(t, "/test/key", mock.WatchCalls[0].Key)
		assert.Equal(t, true, mock.WatchCalls[0].Opts.Prefix)
		assert.Equal(t, int64(100), mock.WatchCalls[0].Opts.Revision)
	})

	t.Run("custom function is called", func(t *testing.T) {
		mock := NewMockClient()
		customCh := make(WatchChan, 1)
		customCh <- WatchResponse{
			Events: []WatchEvent{
				{Type: WatchEventPut, Key: "/test/key", Value: "value"},
			},
		}
		close(customCh)

		mock.WatchFunc = func(_ context.Context, _ string, _ *WatchOptions) WatchChan {
			return customCh
		}

		ch := mock.Watch(context.Background(), "/test/key", nil)

		resp := <-ch
		assert.Len(t, resp.Events, 1)
		assert.Equal(t, "/test/key", resp.Events[0].Key)
	})

	t.Run("nil options handled", func(t *testing.T) {
		mock := NewMockClient()

		ch := mock.Watch(context.Background(), "/test/key", nil)
		<-ch

		assert.Len(t, mock.WatchCalls, 1)
		assert.Nil(t, mock.WatchCalls[0].Opts)
	})

	t.Run("reset clears watch calls", func(t *testing.T) {
		mock := NewMockClient()
		mock.Watch(context.Background(), "/test/key", nil)
		assert.Len(t, mock.WatchCalls, 1)

		mock.Reset()

		assert.Len(t, mock.WatchCalls, 0)
	})
}

func TestDryRunClient_Watch(t *testing.T) {
	t.Run("returns closed channel", func(t *testing.T) {
		client := NewDryRunClient()

		ch := client.Watch(context.Background(), "/test/key", &WatchOptions{Prefix: true})

		select {
		case _, ok := <-ch:
			assert.False(t, ok, "channel should be closed immediately")
		case <-time.After(100 * time.Millisecond):
			t.Error("channel should be closed, not blocking")
		}
	})
}

func TestWatchEventType(t *testing.T) {
	tests := []struct {
		name      string
		eventType WatchEventType
		expected  string
	}{
		{"PUT", WatchEventPut, "PUT"},
		{"DELETE", WatchEventDelete, "DELETE"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.eventType))
		})
	}
}
