package output

// ContextView represents a context for display purposes.
// It contains only the fields needed for output rendering.
type ContextView struct {
	Username  string
	Endpoints []string
}

// ConfigView represents the configuration for display purposes.
// It contains only the fields needed for output rendering.
type ConfigView struct {
	CurrentContext string
	LogLevel       string
	DefaultFormat  string
	Strict         bool
	NoValidate     bool
	Contexts       map[string]*ContextView
}

// DryRunOperation represents a dry-run operation for display purposes.
// This mirrors client.Operation but is defined in the output package
// to avoid coupling output to the client package.
type DryRunOperation struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}
