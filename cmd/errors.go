package cmd

import "fmt"

// wrapNotConnectedError returns a standardized error for when etcd
// connection fails due to missing or invalid context configuration.
func wrapNotConnectedError(err error) error {
	return fmt.Errorf("✗ not connected: %w\n\nUse 'etu login' to configure a context", err)
}
