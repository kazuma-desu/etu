package cmd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestWrapContextError_Integration tests error handling in command contexts
func TestWrapContextError_Integration(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
		wantNil bool
	}{
		{
			name:    "nil error returns nil",
			err:     nil,
			wantNil: true,
		},
		{
			name:    "deadline exceeded gets wrapped with timeout message",
			err:     context.DeadlineExceeded,
			wantMsg: "operation timed out",
			wantNil: false,
		},
		{
			name:    "context canceled gets wrapped with user message",
			err:     context.Canceled,
			wantMsg: "operation canceled by user",
			wantNil: false,
		},
		{
			name:    "other errors pass through unchanged",
			err:     errors.New("some other error"),
			wantMsg: "some other error",
			wantNil: false,
		},
		{
			name:    "wrapped deadline exceeded is detected",
			err:     fmt.Errorf("wrapped: %w", context.DeadlineExceeded),
			wantMsg: "operation timed out",
			wantNil: false,
		},
		{
			name:    "wrapped context canceled is detected",
			err:     fmt.Errorf("wrapped: %w", context.Canceled),
			wantMsg: "operation canceled by user",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapContextError(tt.err)
			if tt.wantNil {
				assert.Nil(t, got)
				return
			}
			assert.NotNil(t, got)
			assert.True(t, strings.Contains(got.Error(), tt.wantMsg),
				"expected error to contain %q, got %q", tt.wantMsg, got.Error())
		})
	}
}

// TestGetOperationContext_Cancellation tests that signal handling works correctly
func TestGetOperationContext_Cancellation(t *testing.T) {
	// Test that context is created with deadline
	ctx, cancel := getOperationContext()
	defer cancel()

	assert.NotNil(t, ctx)

	// Verify context has a deadline
	deadline, hasDeadline := ctx.Deadline()
	assert.True(t, hasDeadline, "context should have a deadline")
	assert.False(t, deadline.IsZero(), "deadline should not be zero")

	// Verify context is not already canceled
	select {
	case <-ctx.Done():
		t.Error("context should not be canceled immediately")
	default:
		// Expected - context is not canceled
	}
}

// TestContextErrorInCommandFlow tests error propagation through command flow
func TestContextErrorInCommandFlow(t *testing.T) {
	// Test that wrapContextError properly handles errors from operations

	t.Run("timeout error from operation", func(t *testing.T) {
		// Simulate an operation that times out
		simulatedErr := context.DeadlineExceeded
		wrappedErr := wrapContextError(simulatedErr)

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "operation timed out")
		assert.Contains(t, wrappedErr.Error(), "consider increasing --timeout")
	})

	t.Run("cancellation error from operation", func(t *testing.T) {
		// Simulate an operation that was canceled
		simulatedErr := context.Canceled
		wrappedErr := wrapContextError(simulatedErr)

		assert.NotNil(t, wrappedErr)
		assert.Contains(t, wrappedErr.Error(), "operation canceled by user")
	})

	t.Run("regular error unchanged", func(t *testing.T) {
		// Simulate a regular error
		simulatedErr := errors.New("connection refused")
		wrappedErr := wrapContextError(simulatedErr)

		assert.NotNil(t, wrappedErr)
		assert.Equal(t, "connection refused", wrappedErr.Error())
	})
}

// TestOperationTimeoutBehavior tests timeout configuration
func TestOperationTimeoutBehavior(t *testing.T) {
	// Save original timeout
	originalTimeout := operationTimeout
	defer func() {
		operationTimeout = originalTimeout
	}()

	// Set a short timeout for testing
	operationTimeout = 100 * time.Millisecond

	ctx, cancel := getOperationContext()
	defer cancel()

	// Verify deadline is set correctly
	deadline, hasDeadline := ctx.Deadline()
	assert.True(t, hasDeadline)

	// Deadline should be approximately operationTimeout from now
	expectedDeadline := time.Now().Add(operationTimeout)
	deadlineDiff := deadline.Sub(expectedDeadline)
	assert.True(t, deadlineDiff < time.Second && deadlineDiff > -time.Second,
		"deadline should be approximately %v from now, got diff %v", operationTimeout, deadlineDiff)
}

// BenchmarkWrapContextError benchmarks the error wrapping function
func BenchmarkWrapContextError(b *testing.B) {
	errs := []error{
		nil,
		context.DeadlineExceeded,
		context.Canceled,
		errors.New("some error"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapContextError(errs[i%len(errs)])
	}
}
