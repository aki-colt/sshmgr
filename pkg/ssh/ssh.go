package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// SSHClient handles SSH connections using sshpass
type SSHClient struct {
	sshpassPath string
	sshPath     string
}

// NewSSHClient creates a new SSHClient
func NewSSHClient() *SSHClient {
	return &SSHClient{
		sshpassPath: "sshpass",
		sshPath:     "ssh",
	}
}

// Connect connects to a host using password authentication
func (c *SSHClient) Connect(host, user, password string, port int) error {
	return c.connect(host, user, password, port, nil)
}

// ConnectWithCommand connects to a host and executes a command
func (c *SSHClient) ConnectWithCommand(host, user, password string, port int, command string) error {
	return c.connect(host, user, password, port, []string{command})
}

// TestConnection tests if a connection can be established
func (c *SSHClient) TestConnection(host, user, password string, port int) error {
	return c.connect(host, user, password, port, []string{"exit"})
}

// connect performs the actual SSH connection
func (c *SSHClient) connect(host, user, password string, port int, command []string) error {
	// Build SSH command
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "ConnectTimeout=5",
		"-p", fmt.Sprintf("%d", port),
		fmt.Sprintf("%s@%s", user, host),
	}

	// Add command if provided
	args = append(args, command...)

	// Build sshpass command
	cmd := exec.Command(c.sshpassPath, "-p", password, c.sshPath)
	cmd.Args = append(cmd.Args, args...)

	// Set up stdin/stdout/stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("SSH connection failed: %w", err)
	}

	return nil
}

// CheckDependencies checks if sshpass and ssh are available
func (c *SSHClient) CheckDependencies() error {
	// Check sshpass
	if _, err := exec.LookPath(c.sshpassPath); err != nil {
		return fmt.Errorf("sshpass not found: %w", err)
	}

	// Check ssh
	if _, err := exec.LookPath(c.sshPath); err != nil {
		return fmt.Errorf("ssh not found: %w", err)
	}

	return nil
}

// GetSSHKeyPath returns the path to SSH key if available
func (c *SSHClient) GetSSHKeyPath() string {
	homeDir, _ := os.UserHomeDir()
	keyPath := fmt.Sprintf("%s/.ssh/id_rsa", homeDir)
	if _, err := os.Stat(keyPath); err == nil {
		return keyPath
	}
	return ""
}

// ParseHostString parses a host string in the format "user@host:port"
func ParseHostString(hostStr string) (host, user string, port int, err error) {
	// Default values
	port = 22

	// Parse user@host
	parts := strings.Split(hostStr, "@")
	if len(parts) == 2 {
		user = parts[0]
		hostStr = parts[1]
	}

	// Parse host:port
	hostParts := strings.Split(hostStr, ":")
	if len(hostParts) >= 1 {
		host = hostParts[0]
	}

	if len(hostParts) == 2 {
		_, err := fmt.Sscanf(hostParts[1], "%d", &port)
		if err != nil {
			return "", "", 0, fmt.Errorf("invalid port: %w", err)
		}
	}

	if host == "" {
		return "", "", 0, fmt.Errorf("invalid host string")
	}

	return host, user, port, nil
}

// TestConnectionWithTimeout tests connection with a timeout
func TestConnectionWithTimeout(host, user, password string, port int, timeout time.Duration) error {
	client := NewSSHClient()

	done := make(chan error, 1)
	go func() {
		done <- client.TestConnection(host, user, password, port)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("connection timeout after %v", timeout)
	}
}
