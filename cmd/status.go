package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/client"
	"github.com/kazuma-desu/etu/pkg/config"
	"github.com/kazuma-desu/etu/pkg/output"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show etcd cluster status and health information",
	Long: `Display detailed information about the etcd cluster including:
- Endpoints
- Server version
- Database size
- Leader information
- Raft status (index, term, applied index)
- Any cluster errors`,
	Example: `  # Show cluster status
  etu status

  # Output as JSON
  etu status -o json

  # Output as YAML
  etu status -o yaml`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(_ *cobra.Command, _ []string) error {
	// Validate output format early, before attempting connection
	allowedFormats := []string{
		output.FormatSimple.String(),
		output.FormatJSON.String(),
		output.FormatYAML.String(),
	}
	if err := validateOutputFormat(allowedFormats); err != nil {
		return err
	}

	ctx, cancel := getOperationContext()
	defer cancel()

	cfg, err := config.GetEtcdConfigWithContext(contextName)
	if err != nil {
		return wrapNotConnectedError(err)
	}

	etcdClient, cleanup, err := newEtcdClient(cfg)
	if err != nil {
		return err
	}
	defer cleanup()

	statuses := make(map[string]*client.StatusResponse)
	var firstError error

	for _, endpoint := range cfg.Endpoints {
		status, err := etcdClient.Status(ctx, endpoint)
		if err != nil {
			if firstError == nil {
				firstError = fmt.Errorf("failed to get status from %s: %w", endpoint, err)
			}
			statuses[endpoint] = nil
		} else {
			statuses[endpoint] = status
		}
	}

	switch outputFormat {
	case output.FormatSimple.String():
		if err := printStatusSimple(cfg.Endpoints, statuses, firstError); err != nil {
			return err
		}
	case output.FormatJSON.String():
		if err := printStatusJSON(cfg.Endpoints, statuses, firstError); err != nil {
			return err
		}
	case output.FormatYAML.String():
		if err := printStatusYAML(cfg.Endpoints, statuses, firstError); err != nil {
			return err
		}
	default:
		// Safety net: should never reach here due to validateOutputFormat check above
		return fmt.Errorf("✗ invalid output format: %s (use simple, json, or yaml)", outputFormat)
	}

	// Return unified warning if any endpoint was unreachable
	if firstError != nil {
		return fmt.Errorf("warning: some endpoints are unreachable: %w", firstError)
	}
	return nil
}

// printStatusSimple prints cluster status to stdout.
func printStatusSimple(endpoints []string, statuses map[string]*client.StatusResponse, _ error) error {
	fmt.Println("Cluster Status")
	fmt.Println("==============")
	fmt.Println()

	healthyCount := 0
	unhealthyCount := 0

	for _, endpoint := range endpoints {
		status := statuses[endpoint]
		fmt.Printf("Endpoint: %s\n", endpoint)

		if status == nil {
			fmt.Println("  Status: UNHEALTHY")
			fmt.Println("  Error:  Failed to connect")
			unhealthyCount++
		} else {
			fmt.Println("  Status: HEALTHY")
			fmt.Printf("  Version: %s\n", status.Version)
			fmt.Printf("  DB Size: %d bytes (%.2f MB)\n", status.DbSize, float64(status.DbSize)/(1024*1024))
			fmt.Printf("  Leader:  %d\n", status.Leader)
			fmt.Printf("  Raft Index: %d (Term: %d)\n", status.RaftIndex, status.RaftTerm)
			if status.IsLearner {
				fmt.Println("  Role:    Learner")
			}
			if len(status.Errors) > 0 {
				fmt.Println("  Errors:")
				for _, err := range status.Errors {
					fmt.Printf("    - %s\n", err)
				}
			}
			healthyCount++
		}
		fmt.Println()
	}

	fmt.Println("Summary")
	fmt.Println("-------")
	fmt.Printf("Healthy:   %d\n", healthyCount)
	fmt.Printf("Unhealthy: %d\n", unhealthyCount)
	fmt.Printf("Total:     %d\n", len(endpoints))

	return nil
}

func buildStatusData(endpoints []string, statuses map[string]*client.StatusResponse, firstError error) map[string]any {
	endpointList := make([]map[string]any, 0, len(endpoints))
	healthyCount := 0
	for _, endpoint := range endpoints {
		status := statuses[endpoint]
		endpointInfo := map[string]any{
			"endpoint": endpoint,
			"healthy":  status != nil,
		}
		if status != nil {
			healthyCount++
			endpointInfo["version"] = status.Version
			endpointInfo["dbSize"] = status.DbSize
			endpointInfo["leader"] = status.Leader
			endpointInfo["raftIndex"] = status.RaftIndex
			endpointInfo["raftTerm"] = status.RaftTerm
			endpointInfo["raftAppliedIndex"] = status.RaftAppliedIndex
			endpointInfo["isLearner"] = status.IsLearner
			if len(status.Errors) > 0 {
				endpointInfo["errors"] = status.Errors
			}
		}
		endpointList = append(endpointList, endpointInfo)
	}

	result := map[string]any{
		"endpoints": endpointList,
		"summary": map[string]int{
			"healthy":   healthyCount,
			"unhealthy": len(endpoints) - healthyCount,
			"total":     len(endpoints),
		},
	}

	if firstError != nil {
		result["warning"] = "some endpoints are unreachable"
	}

	return result
}

func printStatusJSON(endpoints []string, statuses map[string]*client.StatusResponse, firstError error) error {
	data := buildStatusData(endpoints, statuses, firstError)
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

func printStatusYAML(endpoints []string, statuses map[string]*client.StatusResponse, firstError error) error {
	data := buildStatusData(endpoints, statuses, firstError)
	yamlBytes, err := output.SerializeYAML(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	fmt.Print(string(yamlBytes))
	return nil
}
