package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/aki-colt/sshmgr/pkg/config"
	"github.com/aki-colt/sshmgr/pkg/encryption"
	"github.com/aki-colt/sshmgr/pkg/ssh"
	"github.com/sahilm/fuzzy"
	"github.com/spf13/cobra"
)

func GetGlobalConfig() *config.Config {
	return cfg
}

func EnsureAuthenticated(cfg *config.Config) (string, bool) {
	if !cfg.Exists() {
		fmt.Println("Please run 'sshmgr init' first.")
		return "", false
	}

	if masterPassword == "" {
		fmt.Print("Enter master password: ")
		var masterPassword string
		fmt.Scanln(&masterPassword)

		testEncryptor := encryption.NewEncryptor(masterPassword)

		_, err := testEncryptor.Encrypt("test")
		if err != nil {
			fmt.Println("Invalid master password.")
			return "", false
		}

		encryptor = testEncryptor
	}

	return masterPassword, true
}

// GetEncryptor returns an encryptor with the given master password
func GetEncryptor(masterPassword string) *encryption.Encryptor {
	return encryption.NewEncryptor(masterPassword)
}

// GetSSHClient returns a new SSH client
func GetSSHClient() *ssh.SSHClient {
	return ssh.NewSSHClient()
}

// generateID generates a unique ID
func GenerateID(cfg *config.Config) string {
	return fmt.Sprintf("%d", len(cfg.ListHosts())+1)
}

// getCurrentTime returns current time in a simple format
func GetCurrentTime() string {
	return "2026-01-09"
}

func ConnectByAlias(alias string) {
	host, err := findHostByAlias(cfg, alias)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	masterPassword, ok := EnsureAuthenticated(cfg)
	if !ok {
		return
	}

	encryptor := GetEncryptor(masterPassword)

	password, err := encryptor.Decrypt(host.Password)
	if err != nil {
		fmt.Printf("Error decrypting password: %v\n", err)
		return
	}

	sshClient := GetSSHClient()
	fmt.Printf("Connecting to %s as %s...\n", host.Host, host.User)

	if err := sshClient.Connect(host.Host, host.User, password, host.Port); err != nil {
		fmt.Printf("Connection failed: %v\n", err)
	}
}

func findHostByAlias(cfg *config.Config, alias string) (config.Host, error) {
	host, err := cfg.GetHostByAlias(alias)
	if err == nil {
		return *host, nil
	}

	hosts := cfg.ListHosts()
	if len(hosts) == 0 {
		return config.Host{}, config.ErrHostNotFound
	}

	aliases := make([]string, len(hosts))
	for i, h := range hosts {
		aliases[i] = h.Alias
	}

	matches := fuzzy.Find(alias, aliases)
	if len(matches) == 0 {
		return config.Host{}, config.ErrHostNotFound
	}

	bestMatch := matches[0]
	return hosts[bestMatch.Index], nil
}

func GetHostSuggestions(cfg *config.Config, toComplete ...string) []string {
	hosts := cfg.ListHosts()
	aliases := make([]string, len(hosts))
	for i, h := range hosts {
		aliases[i] = h.Alias
	}

	if len(toComplete) > 0 && toComplete[0] != "" {
		matches := fuzzy.Find(strings.ToLower(toComplete[0]), aliases)
		filtered := make([]string, len(matches))
		for i, match := range matches {
			filtered[i] = aliases[match.Index]
		}
		return filtered
	}

	return aliases
}

func FuzzySearchHosts(cfg *config.Config, query string) []config.Host {
	hosts := cfg.ListHosts()
	if len(hosts) == 0 {
		return []config.Host{}
	}

	aliases := make([]string, len(hosts))
	for i, h := range hosts {
		aliases[i] = h.Alias
	}

	matches := fuzzy.Find(strings.ToLower(query), aliases)
	if len(matches) == 0 {
		return []config.Host{}
	}

	result := make([]config.Host, len(matches))
	for i, match := range matches {
		result[i] = hosts[match.Index]
	}

	return result
}

func DetectShell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return "bash"
	}

	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "fish") {
		return "fish"
	}
	if strings.Contains(shell, "powershell") || strings.Contains(shell, "pwsh") {
		return "powershell"
	}

	return "bash"
}

func InstallCompletion(cmd *cobra.Command, shell string) error {
	switch shell {
	case "bash":
		return installBashCompletion(cmd)
	case "zsh":
		return installZshCompletion(cmd)
	case "fish":
		return installFishCompletion(cmd)
	case "powershell":
		return installPowerShellCompletion(cmd)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func installAutoCompletion(cmd *cobra.Command) error {
	shell := DetectShell()
	fmt.Printf("Detected shell: %s\n", shell)
	return InstallCompletion(cmd, shell)
}

func installBashCompletion(cmd *cobra.Command) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	completionFile := homeDir + "/.sshmgr-completion.bash"
	rcFile := homeDir + "/.bashrc"

	if err := cmd.Root().GenBashCompletionFile(completionFile); err != nil {
		return fmt.Errorf("failed to generate completion: %w", err)
	}

	sourceLine := fmt.Sprintf("\n# sshmgr auto-completion\nsource %s\n", completionFile)

	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open bashrc: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(sourceLine); err != nil {
		return fmt.Errorf("failed to write to bashrc: %w", err)
	}

	return nil
}

func installZshCompletion(cmd *cobra.Command) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	completionDir := homeDir + "/.zsh/completions"
	completionFile := completionDir + "/_sshmgr"

	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("failed to create completions directory: %w", err)
	}

	if err := cmd.Root().GenZshCompletionFile(completionFile); err != nil {
		return fmt.Errorf("failed to generate completion: %w", err)
	}

	rcFile := homeDir + "/.zshrc"

	content, err := os.ReadFile(rcFile)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read zshrc: %w", err)
	}

	rcContent := string(content)
	fpathLine := "fpath=(~/.zsh/completions $fpath)"
	autoloadLine := "autoload -U compinit && compinit"
	compdefLine := "compdef _sshmgr sshmgr"

	linesToAdd := []string{}
	if !strings.Contains(rcContent, fpathLine) {
		linesToAdd = append(linesToAdd, fpathLine)
	}
	if !strings.Contains(rcContent, autoloadLine) && !strings.Contains(rcContent, "compinit") {
		linesToAdd = append(linesToAdd, autoloadLine)
	}
	if !strings.Contains(rcContent, compdefLine) {
		linesToAdd = append(linesToAdd, compdefLine)
	}

	if len(linesToAdd) > 0 {
		f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return fmt.Errorf("failed to open zshrc: %w", err)
		}
		defer f.Close()

		sourceLines := "\n# sshmgr auto-completion\n" + strings.Join(linesToAdd, "\n") + "\n"
		if _, err := f.WriteString(sourceLines); err != nil {
			return fmt.Errorf("failed to write to zshrc: %w", err)
		}
	}

	return nil
}

func installFishCompletion(cmd *cobra.Command) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	completionDir := homeDir + "/.config/fish/completions"
	completionFile := completionDir + "/sshmgr.fish"

	if err := os.MkdirAll(completionDir, 0755); err != nil {
		return fmt.Errorf("failed to create completions directory: %w", err)
	}

	if err := cmd.Root().GenFishCompletionFile(completionFile, true); err != nil {
		return fmt.Errorf("failed to generate completion: %w", err)
	}

	return nil
}

func installPowerShellCompletion(cmd *cobra.Command) error {
	return fmt.Errorf("PowerShell completion auto-install not supported yet. Please run 'sshmgr completion powershell' and add it to your profile")
}
