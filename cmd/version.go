package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/output"
)

// Version info - set via ldflags at build time
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

type versionInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the version, commit, build date, Go version, and platform.",
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// runVersion prints CLI version information either as indented JSON to stdout
// or as a human-readable block. When the package-level Version is "dev", it
// attempts to enrich Version, Commit, and BuildDate from the module build
// information (vcs.revision and vcs.time) if available. If JSON output is
// selected via outputFormat, it returns any error produced by the encoder;
// otherwise it returns nil.
func runVersion(_ *cobra.Command, _ []string) error {
	info := versionInfo{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}

	// Fallback to Go module info if version is not set (e.g. go install)
	if info.Version == "dev" {
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			// Prefer Main.Version if available and not (devel), otherwise keep "dev"
			if len(buildInfo.Main.Version) > 0 && buildInfo.Main.Version != "(devel)" {
				info.Version = buildInfo.Main.Version
			}
			// Iterate over settings to find vcs info
			for _, setting := range buildInfo.Settings {
				switch setting.Key {
				case "vcs.revision":
					info.Commit = setting.Value
				case "vcs.time":
					info.BuildDate = setting.Value
				}
			}
		}
	}

	if outputFormat == output.FormatJSON.String() {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}

	fmt.Printf("etu version %s\n", info.Version)
	fmt.Printf("  commit:     %s\n", info.Commit)
	fmt.Printf("  built:      %s\n", info.BuildDate)
	fmt.Printf("  go version: %s\n", info.GoVersion)
	fmt.Printf("  platform:   %s\n", info.Platform)
	return nil
}