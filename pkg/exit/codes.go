// Package exit provides standard exit codes for etu commands.
package exit

// Standard exit codes used by etu commands.
const (
	// Success indicates successful execution.
	Success = 0

	// GeneralError indicates a general error occurred.
	GeneralError = 1

	// ValidationError indicates a validation error (e.g., invalid input, missing arguments).
	ValidationError = 2

	// ConnectionError indicates a connection error to etcd.
	ConnectionError = 3

	// KeyNotFound indicates the requested key was not found in etcd.
	KeyNotFound = 4
)

// CodeDescriptions maps exit codes to their descriptions.
var CodeDescriptions = map[int]string{
	Success:         "Success",
	GeneralError:    "General error",
	ValidationError: "Validation error",
	ConnectionError: "Connection error",
	KeyNotFound:     "Key not found",
}

// GetDescription returns the description for an exit code.
func GetDescription(code int) string {
	if desc, ok := CodeDescriptions[code]; ok {
		return desc
	}
	return "Unknown error"
}
