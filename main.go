package main

import (
	"fmt"
	"os"

	"github.com/aki-colt/sshmgr/pkg/cli"
	"github.com/spf13/cobra"
)

func main() {
	cli.InitializeCLI()

	rootCmd := &cobra.Command{
		Use:   "sshmgr",
		Short: "SSH Manager - A modern SSH connection manager",
		Long: `SSH Manager is a tool for managing SSH connections.
It supports CLI mode with encrypted storage and alias management.`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
			} else {
				connectByAlias(args[0])
			}
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) > 0 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return cli.GetHostSuggestions(cli.GetGlobalConfig()), cobra.ShellCompDirectiveNoFileComp
		},
	}

	rootCmd.AddCommand(cli.InitCommand)
	rootCmd.AddCommand(cli.AddCommand)
	rootCmd.AddCommand(cli.ListCommand)
	rootCmd.AddCommand(cli.ConnectCommand)
	rootCmd.AddCommand(cli.PasswordCommand)
	rootCmd.AddCommand(cli.DeleteCommand)
	rootCmd.AddCommand(cli.ModifyCommand)
	rootCmd.AddCommand(cli.ResetCommand)

	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion script",
		Long: `To load completions:

Bash:
  $ source <(sshmgr completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ sshmgr completion bash > /etc/bash_completion.d/sshmgr
  # macOS:
  $ sshmgr completion bash > /usr/local/etc/bash_completion.d/sshmgr

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ sshmgr completion zsh > "${fpath[1]}/_sshmgr"

  # You will need to start a new shell for this setup to take effect.

fish:
  $ sshmgr completion fish | source

  # To load completions for each session, execute once:
  $ sshmgr completion fish > ~/.config/fish/completions/sshmgr.fish

PowerShell:
  PS> sshmgr completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> sshmgr completion powershell > sshmgr.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactValidArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			switch args[0] {
			case "bash":
				cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
		},
	})

	rootCmd.AddCommand(&cobra.Command{
		Use:   "completion-install",
		Short: "Install shell auto-completion automatically",
		Long: `Automatically install shell completion for your detected shell.
This detects your current shell and installs the completion script to the appropriate location.`,
		Run: func(cmd *cobra.Command, args []string) {
			shell := cli.DetectShell()
			if err := cli.InstallCompletion(cmd, shell); err != nil {
				fmt.Printf("Error installing completion: %v\n", err)
			} else {
				fmt.Printf("Shell auto-completion installed successfully! (Shell: %s)\n", shell)
				fmt.Println("Please start a new shell or source your shell configuration file to enable completion.")
				switch shell {
				case "zsh":
					fmt.Println("Run: source ~/.zshrc")
				case "bash":
					fmt.Println("Run: source ~/.bashrc")
				case "fish":
					fmt.Println("Run: source ~/.config/fish/config.fish")
				}
			}
		},
	})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func connectByAlias(alias string) {
	cli.ConnectByAlias(alias)
}
