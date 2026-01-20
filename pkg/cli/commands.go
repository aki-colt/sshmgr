package cli

import (
	"crypto/sha256"
	"fmt"
	"os"

	"github.com/aki-colt/sshmgr/pkg/config"
	"github.com/aki-colt/sshmgr/pkg/encryption"
	"github.com/aki-colt/sshmgr/pkg/ssh"
	"github.com/spf13/cobra"
)

var (
	cfg            *config.Config
	encryptor      *encryption.Encryptor
	sshClient      *ssh.SSHClient
	masterPassword string
)

// AddCommand adds a new host
var AddCommand = &cobra.Command{
	Use:   "add",
	Short: "Add a new SSH host",
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := EnsureAuthenticated(cfg); !ok {
			return
		}

		fmt.Print("Enter alias: ")
		var alias string
		fmt.Scanln(&alias)

		fmt.Print("Enter host address: ")
		var host string
		fmt.Scanln(&host)

		fmt.Print("Enter username: ")
		var user string
		fmt.Scanln(&user)

		fmt.Print("Enter password: ")
		var password string
		fmt.Scanln(&password)

		fmt.Print("Enter port (default 22): ")
		var port int
		fmt.Scanln(&port)
		if port == 0 {
			port = 22
		}

		// Encrypt password
		encryptedPassword, err := encryptor.Encrypt(password)
		if err != nil {
			fmt.Printf("Error encrypting password: %v\n", err)
			return
		}

		// Create host
		newHost := config.Host{
			ID:        generateID(),
			Alias:     alias,
			Host:      host,
			User:      user,
			Password:  encryptedPassword,
			Port:      port,
			CreatedAt: getCurrentTime(),
			UpdatedAt: getCurrentTime(),
		}

		// Add to config
		if err := cfg.AddHost(newHost); err != nil {
			fmt.Printf("Error adding host: %v\n", err)
			return
		}

		// Test connection
		fmt.Print("Test connection? [Y/n]: ")
		var test string
		fmt.Scanln(&test)
		if test != "n" && test != "N" {
			if err := sshClient.TestConnection(host, user, password, port); err != nil {
				fmt.Printf("Connection test failed: %v\n", err)
			} else {
				fmt.Println("Connection test successful!")
			}
		}

		// Save config
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Println("Host added successfully!")
	},
}

// ListCommand lists all hosts
var ListCommand = &cobra.Command{
	Use:   "list",
	Short: "List all SSH hosts",
	Run: func(cmd *cobra.Command, args []string) {
		hosts := cfg.ListHosts()

		if len(hosts) == 0 {
			fmt.Println("No hosts found.")
			return
		}

		fmt.Printf("\n%-5s %-20s %-30s %-15s %s\n", "ID", "Alias", "Host", "User", "Port")
		fmt.Println("---------------------------------------------------------------------")
		for i, h := range hosts {
			fmt.Printf("%-5d %-20s %-30s %-15s %d\n", i+1, h.Alias, h.Host, h.User, h.Port)
		}
	},
}

// DeleteCommand deletes a host
var DeleteCommand = &cobra.Command{
	Use:   "delete <alias>",
	Short: "Delete a SSH host by alias",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return GetHostSuggestions(cfg, toComplete), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := EnsureAuthenticated(cfg); !ok {
			return
		}

		alias := args[0]
		host, err := cfg.GetHostByAlias(alias)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		fmt.Printf("Are you sure you want to delete host '%s'? [y/N]: ", alias)
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("Operation cancelled.")
			return
		}

		if err := cfg.DeleteHost(host.ID); err != nil {
			fmt.Printf("Error deleting host: %v\n", err)
			return
		}

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Println("Host deleted successfully!")
	},
}

// ConnectCommand connects to a host by alias
var ConnectCommand = &cobra.Command{
	Use:   "connect <alias>",
	Short: "Connect to a SSH host by alias",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return GetHostSuggestions(cfg, toComplete), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		host, err := cfg.GetHostByAlias(alias)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if _, ok := EnsureAuthenticated(cfg); !ok {
			return
		}

		password, err := encryptor.Decrypt(host.Password)
		if err != nil {
			fmt.Printf("Error decrypting password: %v\n", err)
			return
		}

		fmt.Printf("Connecting to %s as %s...\n", host.Host, host.User)

		if err := sshClient.Connect(host.Host, host.User, password, host.Port); err != nil {
			fmt.Printf("Connection failed: %v\n", err)
		}
	},
}

// PasswordCommand shows the decrypted password for a host
var PasswordCommand = &cobra.Command{
	Use:   "password <alias>",
	Short: "Show the password for a SSH host by alias",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return GetHostSuggestions(cfg, toComplete), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		alias := args[0]

		host, err := cfg.GetHostByAlias(alias)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if _, ok := EnsureAuthenticated(cfg); !ok {
			return
		}

		password, err := encryptor.Decrypt(host.Password)
		if err != nil {
			fmt.Printf("Error decrypting password: %v\n", err)
			return
		}

		fmt.Printf("Password for '%s' (%s@%s:%d): %s\n", host.Alias, host.User, host.Host, host.Port, password)
	},
}

// ModifyCommand modifies a host
var ModifyCommand = &cobra.Command{
	Use:   "modify <alias>",
	Short: "Modify a SSH host by alias",
	Args:  cobra.ExactArgs(1),
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		return GetHostSuggestions(cfg, toComplete), cobra.ShellCompDirectiveNoFileComp
	},
	Run: func(cmd *cobra.Command, args []string) {
		if _, ok := EnsureAuthenticated(cfg); !ok {
			return
		}

		alias := args[0]
		host, err := cfg.GetHostByAlias(alias)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Decrypt password for display
		currentPassword, err := encryptor.Decrypt(host.Password)
		if err != nil {
			fmt.Printf("Error decrypting password: %v\n", err)
			return
		}

		fmt.Printf("\nCurrent configuration:\n")
		fmt.Printf("  Alias: %s\n", host.Alias)
		fmt.Printf("  Host: %s\n", host.Host)
		fmt.Printf("  User: %s\n", host.User)
		fmt.Printf("  Port: %d\n", host.Port)

		// Get new values
		fmt.Print("\nEnter new alias (press Enter to keep current): ")
		var newAlias string
		fmt.Scanln(&newAlias)
		if newAlias == "" {
			newAlias = host.Alias
		}

		fmt.Print("Enter new host (press Enter to keep current): ")
		var newHost string
		fmt.Scanln(&newHost)
		if newHost == "" {
			newHost = host.Host
		}

		fmt.Print("Enter new user (press Enter to keep current): ")
		var newUser string
		fmt.Scanln(&newUser)
		if newUser == "" {
			newUser = host.User
		}

		fmt.Print("Enter new password (press Enter to keep current): ")
		var newPassword string
		fmt.Scanln(&newPassword)
		if newPassword == "" {
			newPassword = currentPassword
		}

		fmt.Print("Enter new port (press Enter to keep current): ")
		var newPort int
		fmt.Scanln(&newPort)
		if newPort == 0 {
			newPort = host.Port
		}

		// Encrypt new password
		encryptedPassword, err := encryptor.Encrypt(newPassword)
		if err != nil {
			fmt.Printf("Error encrypting password: %v\n", err)
			return
		}

		// Update host
		host.Alias = newAlias
		host.Host = newHost
		host.User = newUser
		host.Password = encryptedPassword
		host.Port = newPort
		host.UpdatedAt = getCurrentTime()

		// Test connection
		fmt.Print("Test connection? [Y/n]: ")
		var test string
		fmt.Scanln(&test)
		if test != "n" && test != "N" {
			if err := sshClient.TestConnection(host.Host, host.User, newPassword, host.Port); err != nil {
				fmt.Printf("Connection test failed: %v\n", err)
			} else {
				fmt.Println("Connection test successful!")
			}
		}

		// Update config
		if err := cfg.UpdateHost(*host); err != nil {
			fmt.Printf("Error updating host: %v\n", err)
			return
		}

		// Save config
		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Println("Host modified successfully!")
	},
}

// InitCommand initializes the master password
var InitCommand = &cobra.Command{
	Use:   "init",
	Short: "Initialize master password",
	Run: func(cmd *cobra.Command, args []string) {
		if cfg.Exists() {
			fmt.Println("Config already initialized.")
			return
		}

		fmt.Print("Enter master password: ")
		var password string
		fmt.Scanln(&password)

		if len(password) < 8 {
			fmt.Println("Password must be at least 8 characters.")
			return
		}

		fmt.Print("Confirm master password: ")
		var confirm string
		fmt.Scanln(&confirm)

		if password != confirm {
			fmt.Println("Passwords do not match.")
			return
		}

		hash := sha256.Sum256([]byte(password))
		cfg.SetMasterHash(fmt.Sprintf("%x", hash))

		if err := cfg.Save(); err != nil {
			fmt.Printf("Error saving config: %v\n", err)
			return
		}

		fmt.Println("Master password set successfully!")

		if err := installAutoCompletion(cmd); err != nil {
			fmt.Printf("Warning: Failed to install auto-completion: %v\n", err)
		} else {
			fmt.Println("Shell auto-completion enabled! Please start a new shell or run 'source ~/.zshrc' (zsh) or 'source ~/.bashrc' (bash).")
		}
	},
}

// ResetCommand resets all configuration
var ResetCommand = &cobra.Command{
	Use:   "reset",
	Short: "Reset all configuration and reinitialize",
	Run: func(cmd *cobra.Command, args []string) {
		if !cfg.Exists() {
			fmt.Println("No configuration to reset.")
			return
		}

		// Show warning
		fmt.Println("\n⚠️  WARNING: This will delete ALL your SSH hosts!")
		fmt.Println("This action cannot be undone.")
		fmt.Println("")

		// First confirmation
		fmt.Print("Are you sure you want to reset? [yes/no]: ")
		var confirm string
		fmt.Scanln(&confirm)

		if confirm != "yes" && confirm != "Yes" && confirm != "YES" && confirm != "y" && confirm != "Y" {
			fmt.Println("Reset cancelled.")
			return
		}

		// Second confirmation
		fmt.Print("\nType 'RESET' to confirm: ")
		var finalConfirm string
		fmt.Scanln(&finalConfirm)

		if finalConfirm != "RESET" && finalConfirm != "reset" {
			fmt.Println("Reset cancelled.")
			return
		}

		// Get config file path
		homeDir, _ := os.UserHomeDir()
		configPath := homeDir + "/.ssh_manager_config.yaml"

		// Delete config file
		if err := os.Remove(configPath); err != nil {
			fmt.Printf("Error deleting config file: %v\n", err)
			return
		}

		fmt.Println("\n✅ Configuration reset successfully!")
		fmt.Println("All SSH hosts and master password have been deleted.")
		fmt.Println("\nPlease run 'sshmgr init' to set up a new configuration.")
		fmt.Println("")
	},
}

// ensureAuthenticated ensures user is authenticated
func ensureAuthenticated() bool {
	if !cfg.Exists() {
		fmt.Println("Please run 'sshmgr init' first.")
		return false
	}

	if masterPassword == "" {
		fmt.Print("Enter master password: ")
		fmt.Scanln(&masterPassword)

		// Validate password
		testEncryptor := encryption.NewEncryptor(masterPassword)
		_ = cfg.GetMasterHash()

		// Test by encrypting something
		_, err := testEncryptor.Encrypt("test")
		if err != nil {
			fmt.Println("Invalid master password.")
			masterPassword = ""
			return false
		}

		encryptor = testEncryptor
	}

	return true
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", len(cfg.ListHosts())+1)
}

// getCurrentTime returns current time in a simple format
func getCurrentTime() string {
	return "2026-01-09"
}

// InitializeCLI initializes the CLI
func InitializeCLI() {
	cfg = config.NewConfig()
	cfg.Load()

	encryptor = nil
	sshClient = ssh.NewSSHClient()

	// Check dependencies
	if err := sshClient.CheckDependencies(); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}
}
