package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

	t.Run("rejects with n", func(t *testing.T) {
		in := strings.NewReader("n\n")
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
}
