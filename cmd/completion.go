package cmd

import (
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/kazuma-desu/etu/pkg/config"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for etu.

To load completions:

Bash:
  # Linux:
  $ etu completion bash > /etc/bash_completion.d/etu
  
  # macOS (with Homebrew):
  $ etu completion bash > $(brew --prefix)/etc/bash_completion.d/etu

Zsh:
  # If shell completion is not already enabled, enable it by adding:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc
  
  # Then load the etu completions:
  $ etu completion zsh > "${fpath[1]}/_etu"
  
  # You may need to restart your shell or run:
  $ source ~/.zshrc

Fish:
  $ etu completion fish > ~/.config/fish/completions/etu.fish

PowerShell:
  PS> etu completion powershell | Out-String | Invoke-Expression
  
  # To load on every session:
  PS> etu completion powershell >> $PROFILE
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)

	if err := rootCmd.RegisterFlagCompletionFunc("context", completeContextNames); err != nil {
		_ = err // best-effort
	}
}

func registerFileCompletion(cmd *cobra.Command, flagName string) {
	if err := cmd.RegisterFlagCompletionFunc(flagName, completeConfigFiles); err != nil {
		_ = err // best-effort
	}
}

func completeConfigFiles(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return []string{"txt", "yaml", "yml", "json"}, cobra.ShellCompDirectiveFilterFileExt
}

func runCompletion(cmd *cobra.Command, args []string) error {
	switch args[0] {
	case "bash":
		return cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		return cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		return cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	}
	return nil
}

// completeContextNames returns available context names for shell completion
func completeContextNames(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	contexts := make([]string, 0, len(cfg.Contexts))
	for name := range cfg.Contexts {
		contexts = append(contexts, name)
	}
	sort.Strings(contexts)

	return contexts, cobra.ShellCompDirectiveNoFileComp
}

// CompleteContextNamesForArg is exported for use by other commands that take context as an argument
func CompleteContextNamesForArg(_ *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	// Only complete first argument
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	return completeContextNames(nil, nil, "")
}
