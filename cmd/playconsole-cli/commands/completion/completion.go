package completion

import (
	"os"

	"github.com/spf13/cobra"
)

// CompletionCmd generates shell completion scripts
var CompletionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for playconsole-cli.

To load completions:

Bash:
  $ source <(gpc completion bash)

  # To load for each session (Linux):
  $ gpc completion bash > /etc/bash_completion.d/gpc

  # To load for each session (macOS):
  $ gpc completion bash > $(brew --prefix)/etc/bash_completion.d/gpc

Zsh:
  # If shell completion is not already enabled, enable it:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  $ source <(gpc completion zsh)

  # To load for each session:
  $ gpc completion zsh > "${fpath[1]}/_gpc"

Fish:
  $ gpc completion fish | source

  # To load for each session:
  $ gpc completion fish > ~/.config/fish/completions/gpc.fish

PowerShell:
  PS> gpc completion powershell | Out-String | Invoke-Expression

  # To load for each session:
  PS> gpc completion powershell > gpc.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE: func(cmd *cobra.Command, args []string) error {
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
	},
}
