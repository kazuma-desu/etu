// Package testutil provides utilities for testing.
package testutil

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

// CaptureStdout captures stdout output from the provided function.
// It returns the captured output and any error returned by f, or an error
// if f() panics (in which case the panic is recovered and converted to an error).
//
// Usage:
//
//	output, err := CaptureStdout(func() error {
//	    return someFunctionThatWritesToStdout()
//	})
func CaptureStdout(f func() error) (string, error) {
	old := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return "", fmt.Errorf("captureStdout: failed to create pipe: %w", pipeErr)
	}

	// Read from pipe in goroutine to avoid blocking on large outputs
	outCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		_ = r.Close()
		outCh <- buf.String()
	}()

	os.Stdout = w

	var fErr error
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				fErr = fmt.Errorf("captureStdout: f() panicked: %v", rec)
			}
		}()
		fErr = f()
	}()

	// Close writer to signal EOF to goroutine, then restore stdout
	_ = w.Close()
	os.Stdout = old

	// Wait for goroutine to finish reading
	output := <-outCh

	return output, fErr
}

// CaptureStdoutFunc captures stdout output from a function that doesn't return an error.
// It returns the captured output and nil error on success, or an error if f() panics.
//
// Usage:
//
//	output, err := CaptureStdoutFunc(func() {
//	    someFunctionThatWritesToStdout()
//	})
func CaptureStdoutFunc(f func()) (string, error) {
	return CaptureStdout(func() error {
		f()
		return nil
	})
}
