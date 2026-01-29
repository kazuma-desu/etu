package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// captureStdoutFunc captures stdout from f() with panic recovery.
func captureStdoutFunc(f func()) (string, error) {
	old := os.Stdout
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return "", fmt.Errorf("captureStdout: failed to create pipe: %w", pipeErr)
	}

	// Read from pipe in goroutine to avoid blocking
	outCh := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()
		outCh <- buf.String()
	}()

	os.Stdout = w

	var panicked bool
	func() {
		defer func() {
			if rec := recover(); rec != nil {
				panicked = true
			}
		}()
		f()
	}()

	// Close writer to signal EOF to goroutine, then restore stdout
	w.Close()
	os.Stdout = old

	// Wait for goroutine to finish reading
	output := <-outCh

	if panicked {
		return output, fmt.Errorf("f() panicked")
	}
	return output, nil
}

func TestConfirmDeletion(t *testing.T) {
	keys := []string{"/app/config/a", "/app/config/b", "/app/config/c"}
	prefix := "/app/config/"

	t.Run("confirms with y", func(t *testing.T) {
		in := strings.NewReader("y\n")
		out := &bytes.Buffer{}

		result := confirmDeletion(keys, prefix, in, out)

		assert.True(t, result)
		assert.Contains(t, out.String(), "3 keys will be deleted")
		assert.Contains(t, out.String(), "/app/config/a")
		assert.Contains(t, out.String(), "[y/N]")
	})

	t.Run("confirms with yes", func(t *testing.T) {
		in := strings.NewReader("yes\n")
		out := &bytes.Buffer{}

		result := confirmDeletion(keys, prefix, in, out)

		assert.True(t, result)
	})

	t.Run("confirms with Y uppercase", func(t *testing.T) {
		in := strings.NewReader("Y\n")
		out := &bytes.Buffer{}

		result := confirmDeletion(keys, prefix, in, out)

		assert.True(t, result)
	})

	t.Run("confirms with YES uppercase", func(t *testing.T) {
		in := strings.NewReader("YES\n")
		out := &bytes.Buffer{}

		result := confirmDeletion(keys, prefix, in, out)

		assert.True(t, result)
	})

	t.Run("confirms with whitespace around y", func(t *testing.T) {
		in := strings.NewReader("  y  \n")
		out := &bytes.Buffer{}

		result := confirmDeletion(keys, prefix, in, out)

		assert.True(t, result)
	})

	t.Run("rejects with n", func(t *testing.T) {
		in := strings.NewReader("n\n")
		out := &bytes.Buffer{}

		result := confirmDeletion(keys, prefix, in, out)

		assert.False(t, result)
	})

	t.Run("rejects with no", func(t *testing.T) {
		in := strings.NewReader("no\n")
		out := &bytes.Buffer{}

		result := confirmDeletion(keys, prefix, in, out)

		assert.False(t, result)
	})

	t.Run("rejects with empty input", func(t *testing.T) {
		in := strings.NewReader("\n")
		out := &bytes.Buffer{}

		result := confirmDeletion(keys, prefix, in, out)

		assert.False(t, result)
	})

	t.Run("rejects with random text", func(t *testing.T) {
		in := strings.NewReader("maybe\n")
		out := &bytes.Buffer{}

		result := confirmDeletion(keys, prefix, in, out)

		assert.False(t, result)
	})

	t.Run("rejects on EOF", func(t *testing.T) {
		in := strings.NewReader("")
		out := &bytes.Buffer{}

		result := confirmDeletion(keys, prefix, in, out)

		assert.False(t, result)
	})

	t.Run("lists all keys in output", func(t *testing.T) {
		in := strings.NewReader("n\n")
		out := &bytes.Buffer{}

		confirmDeletion(keys, prefix, in, out)

		outputStr := out.String()
		for _, k := range keys {
			assert.Contains(t, outputStr, k)
		}
	})

	t.Run("shows prefix in prompt", func(t *testing.T) {
		in := strings.NewReader("n\n")
		out := &bytes.Buffer{}

		confirmDeletion(keys, prefix, in, out)

		assert.Contains(t, out.String(), prefix)
	})

	t.Run("handles single key", func(t *testing.T) {
		in := strings.NewReader("y\n")
		out := &bytes.Buffer{}

		result := confirmDeletion([]string{"/single"}, "/single", in, out)

		assert.True(t, result)
		assert.Contains(t, out.String(), "1 keys will be deleted")
	})

	t.Run("handles many keys", func(t *testing.T) {
		manyKeys := make([]string, 100)
		for i := range manyKeys {
			manyKeys[i] = "/key/" + string(rune('a'+i%26))
		}
		in := strings.NewReader("y\n")
		out := &bytes.Buffer{}

		result := confirmDeletion(manyKeys, "/key/", in, out)

		assert.True(t, result)
		assert.Contains(t, out.String(), "100 keys will be deleted")
	})
}

func TestPrintKeysToDelete(t *testing.T) {
	t.Run("prints keys with prefix", func(t *testing.T) {
		output, err := captureStdoutFunc(func() {
			printKeysToDelete([]string{"/a", "/b", "/c"}, "/prefix/")
		})
		assert.NoError(t, err)
		assert.Contains(t, output, "Would delete 3 keys")
		assert.Contains(t, output, `"/prefix/"`)
		assert.Contains(t, output, "/a")
		assert.Contains(t, output, "/b")
		assert.Contains(t, output, "/c")
	})

	t.Run("prints single key", func(t *testing.T) {
		output, err := captureStdoutFunc(func() {
			printKeysToDelete([]string{"/only"}, "/only")
		})
		assert.NoError(t, err)
		assert.Contains(t, output, "Would delete 1 keys")
		assert.Contains(t, output, "/only")
	})

	t.Run("handles empty keys slice", func(t *testing.T) {
		output, err := captureStdoutFunc(func() {
			printKeysToDelete([]string{}, "/empty/")
		})
		assert.NoError(t, err)
		assert.Contains(t, output, "Would delete 0 keys")
	})
}
