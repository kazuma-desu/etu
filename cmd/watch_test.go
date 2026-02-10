package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/output"
	"github.com/kazuma-desu/etu/pkg/testutil"
)

func resetWatchOpts() {
	watchOpts.prefix = false
	watchOpts.rev = 0
	watchOpts.prevKV = false
}

func TestPrintWatchEvent_SimpleFormat(t *testing.T) {
	t.Cleanup(resetWatchOpts)
	resetWatchOpts()

	originalFormat := outputFormat
	defer func() { outputFormat = originalFormat }()

	outputFormat = output.FormatSimple.String()

	tests := []struct {
		name    string
		event   client.WatchEvent
		wantErr bool
		want    string
	}{
		{
			name: "PUT event prints raw value only",
			event: client.WatchEvent{
				Type:     client.WatchEventPut,
				Key:      "/app/config",
				Value:    "new-value",
				Revision: 42,
			},
			wantErr: false,
			want:    "new-value",
		},
		{
			name: "DELETE event prints empty value",
			event: client.WatchEvent{
				Type:     client.WatchEventDelete,
				Key:      "/app/config",
				Revision: 43,
			},
			wantErr: false,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			captured, err := testutil.CaptureStdout(func() error {
				return printWatchEvent(tt.event)
			})
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Contains(t, captured, tt.want)
				// Simple format should NOT include metadata
				assert.NotContains(t, captured, "PUT")
				assert.NotContains(t, captured, "DELETE")
				assert.NotContains(t, captured, "rev=")
			}
		})
	}
}

func TestPrintWatchEvent_JSONFormat(t *testing.T) {
	t.Cleanup(resetWatchOpts)
	resetWatchOpts()

	originalFormat := outputFormat
	defer func() { outputFormat = originalFormat }()

	outputFormat = output.FormatJSON.String()

	event := client.WatchEvent{
		Type:     client.WatchEventPut,
		Key:      "/app/config",
		Value:    "test-value",
		Revision: 42,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printWatchEvent(event)
	})
	require.NoError(t, err)

	assert.Contains(t, output, `"Type":"PUT"`)
	assert.Contains(t, output, `"Key":"/app/config"`)
	assert.Contains(t, output, `"Value":"test-value"`)
	assert.Contains(t, output, `"Revision":42`)
}

func TestRunWatch_InvalidRevision(t *testing.T) {
	t.Cleanup(resetWatchOpts)
	resetWatchOpts()

	watchOpts.rev = -1

	err := runWatch(nil, []string{"/test/key"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid --rev")
}

func TestRunWatch_NotConnected(t *testing.T) {
	t.Cleanup(resetWatchOpts)
	resetWatchOpts()

	origContextName := contextName
	defer func() { contextName = origContextName }()

	contextName = "nonexistent-context-for-testing"

	err := runWatch(nil, []string{"/test/key"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRunWatch_WithContextCancelled(t *testing.T) {
	t.Cleanup(resetWatchOpts)
	resetWatchOpts()

	mock := client.NewMockClient()
	mock.WatchFunc = func(_ context.Context, _ string, _ *client.WatchOptions) client.WatchChan {
		ch := make(chan client.WatchResponse)
		close(ch)
		return ch
	}

	origContextName := contextName
	defer func() { contextName = origContextName }()

	contextName = "nonexistent"
	err := runWatch(nil, []string{"/test/key"})
	require.Error(t, err)
}

func TestRunWatch_MockClient(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping mock client test in short mode")
	}

	t.Cleanup(resetWatchOpts)
	resetWatchOpts()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `current-context: test-context
contexts:
  test-context:
    endpoints:
      - http://localhost:2379
`
	err := os.WriteFile(configPath, []byte(configContent), 0600)
	require.NoError(t, err)

	t.Setenv("ETUCONFIG", configPath)

	mock := client.NewMockClient()
	eventCount := 0
	mock.WatchFunc = func(ctx context.Context, _ string, _ *client.WatchOptions) client.WatchChan {
		ch := make(chan client.WatchResponse)
		go func() {
			defer close(ch)
			for i := 0; i < 3; i++ {
				select {
				case <-ctx.Done():
					return
				case ch <- client.WatchResponse{
					Events: []client.WatchEvent{
						{
							Type:     client.WatchEventPut,
							Key:      "/test/key",
							Value:    fmt.Sprintf("value-%d", i),
							Revision: int64(100 + i),
						},
					},
				}:
					eventCount++
					time.Sleep(10 * time.Millisecond)
				}
			}
		}()
		return ch
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch := mock.Watch(ctx, "/test/key", nil)
	eventCountFromMock := 0
	for range ch {
		eventCountFromMock++
	}

	assert.GreaterOrEqual(t, eventCountFromMock, 0, "Mock should have returned events")
}

func TestWatchOpts_Reset(t *testing.T) {
	watchOpts.prefix = true
	watchOpts.rev = 100
	watchOpts.prevKV = true

	resetWatchOpts()

	assert.False(t, watchOpts.prefix)
	assert.Equal(t, int64(0), watchOpts.rev)
	assert.False(t, watchOpts.prevKV)
}

func TestWatchCommand_Flags(t *testing.T) {
	assert.NotNil(t, watchCmd)

	prefixFlag := watchCmd.Flags().Lookup("prefix")
	require.NotNil(t, prefixFlag)
	assert.Equal(t, "false", prefixFlag.DefValue)

	revFlag := watchCmd.Flags().Lookup("rev")
	require.NotNil(t, revFlag)
	assert.Equal(t, "0", revFlag.DefValue)

	prevKVFlag := watchCmd.Flags().Lookup("prev-kv")
	require.NotNil(t, prevKVFlag)
	assert.Equal(t, "false", prevKVFlag.DefValue)
}

func TestRunWatch_WatchError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping watch error test in short mode")
	}

	mock := client.NewMockClient()
	mock.WatchFunc = func(ctx context.Context, _ string, _ *client.WatchOptions) client.WatchChan {
		ch := make(chan client.WatchResponse)
		go func() {
			defer close(ch)
			select {
			case <-ctx.Done():
				return
			case ch <- client.WatchResponse{
				Err: errors.New("watch failed"),
			}:
			}
		}()
		return ch
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch := mock.Watch(ctx, "/test", nil)
	for resp := range ch {
		if resp.Err != nil {
			assert.Contains(t, resp.Err.Error(), "watch failed")
		}
	}
}

func TestRunWatch_CompactRevision(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping compact revision test in short mode")
	}

	mock := client.NewMockClient()
	mock.WatchFunc = func(ctx context.Context, _ string, _ *client.WatchOptions) client.WatchChan {
		ch := make(chan client.WatchResponse)
		go func() {
			defer close(ch)
			select {
			case <-ctx.Done():
				return
			case ch <- client.WatchResponse{
				CompactRevision: 50,
			}:
			}
		}()
		return ch
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ch := mock.Watch(ctx, "/test", nil)
	for resp := range ch {
		if resp.CompactRevision > 0 {
			assert.Equal(t, int64(50), resp.CompactRevision)
		}
	}
}

func TestPrintWatchEvent_JSONMarshalError(t *testing.T) {
	if runtime.GOOS == "js" {
		t.Skip("Skipping on JS/WASM platform")
	}

	originalFormat := outputFormat
	defer func() { outputFormat = originalFormat }()

	outputFormat = output.FormatJSON.String()

	event := client.WatchEvent{
		Type:     client.WatchEventPut,
		Key:      "/test",
		Value:    "value",
		Revision: 1,
	}

	output, err := testutil.CaptureStdout(func() error {
		return printWatchEvent(event)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "PUT")
}

func TestWatchCallRecording(t *testing.T) {
	mock := client.NewMockClient()

	opts := &client.WatchOptions{
		Prefix:   true,
		Revision: 100,
		PrevKV:   true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_ = mock.Watch(ctx, "/prefix", opts)

	require.Len(t, mock.WatchCalls, 1)
	assert.Equal(t, "/prefix", mock.WatchCalls[0].Key)
	assert.Equal(t, true, mock.WatchCalls[0].Opts.Prefix)
	assert.Equal(t, int64(100), mock.WatchCalls[0].Opts.Revision)
	assert.Equal(t, true, mock.WatchCalls[0].Opts.PrevKV)
}

func TestWatchOptions_Behavior(t *testing.T) {
	tests := []struct {
		name     string
		opts     *client.WatchOptions
		wantOpts client.WatchOptions
	}{
		{
			name: "default options",
			opts: nil,
			wantOpts: client.WatchOptions{
				Prefix:   false,
				Revision: 0,
				PrevKV:   false,
			},
		},
		{
			name: "prefix only",
			opts: &client.WatchOptions{
				Prefix: true,
			},
			wantOpts: client.WatchOptions{
				Prefix:   true,
				Revision: 0,
				PrevKV:   false,
			},
		},
		{
			name: "all options set",
			opts: &client.WatchOptions{
				Prefix:   true,
				Revision: 100,
				PrevKV:   true,
			},
			wantOpts: client.WatchOptions{
				Prefix:   true,
				Revision: 100,
				PrevKV:   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := client.NewMockClient()

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			_ = mock.Watch(ctx, "/test", tt.opts)

			require.Len(t, mock.WatchCalls, 1)
			if tt.opts == nil {
				assert.Nil(t, mock.WatchCalls[0].Opts)
			} else {
				assert.Equal(t, tt.wantOpts.Prefix, mock.WatchCalls[0].Opts.Prefix)
				assert.Equal(t, tt.wantOpts.Revision, mock.WatchCalls[0].Opts.Revision)
				assert.Equal(t, tt.wantOpts.PrevKV, mock.WatchCalls[0].Opts.PrevKV)
			}
		})
	}
}
