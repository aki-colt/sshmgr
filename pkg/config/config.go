package config

import (
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Host represents an SSH host configuration
type Host struct {
	ID        string `yaml:"id"`
	Alias     string `yaml:"alias"`
	Host      string `yaml:"host"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"` // encrypted
	Port      int    `yaml:"port"`
	CreatedAt string `yaml:"created_at"`
	UpdatedAt string `yaml:"updated_at"`
}

// Config represents SSH manager configuration
type Config struct {
	Version    string `yaml:"version"`
	MasterHash string `yaml:"master_hash"` // hash of master password for validation
	Hosts      []Host `yaml:"hosts"`
	mu         sync.RWMutex
	configPath string
}

// NewConfig creates a new Config instance
func NewConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".ssh_manager_config.yaml")

	return &Config{
		Version:    "1.0",
		Hosts:      make([]Host, 0),
		configPath: configPath,
	}
}

// Load loads configuration from file
func (c *Config) Load() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := os.ReadFile(c.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return empty config
			return nil
		}
		return err
	}

	return yaml.Unmarshal(data, c)
}

// Save saves configuration to file
func (c *Config) Save() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(c.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(c.configPath, data, 0600)
}

// AddHost adds a new host to the configuration
func (c *Config) AddHost(host Host) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check for duplicate alias
	for _, h := range c.Hosts {
		if h.Alias == host.Alias {
			return ErrAliasExists
		}
	}

	c.Hosts = append(c.Hosts, host)
	return nil
}

// DeleteHost deletes a host by ID
func (c *Config) DeleteHost(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, h := range c.Hosts {
		if h.ID == id {
			c.Hosts = append(c.Hosts[:i], c.Hosts[i+1:]...)
			return nil
		}
	}

	return ErrHostNotFound
}

// UpdateHost updates an existing host
func (c *Config) UpdateHost(host Host) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check for duplicate alias (excluding current host)
	for _, h := range c.Hosts {
		if h.Alias == host.Alias && h.ID != host.ID {
			return ErrAliasExists
		}
	}

	for i, h := range c.Hosts {
		if h.ID == host.ID {
			c.Hosts[i] = host
			return nil
		}
	}

	return ErrHostNotFound
}

// GetHostByID retrieves a host by ID
func (c *Config) GetHostByID(id string) (*Host, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, h := range c.Hosts {
		if h.ID == id {
			return &h, nil
		}
	}

	return nil, ErrHostNotFound
}

// GetHostByAlias retrieves a host by alias
func (c *Config) GetHostByAlias(alias string) (*Host, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, h := range c.Hosts {
		if h.Alias == alias {
			return &h, nil
		}
	}

	return nil, ErrHostNotFound
}

// ListHosts returns all hosts
func (c *Config) ListHosts() []Host {
	c.mu.RLock()
	defer c.mu.RUnlock()

	hosts := make([]Host, len(c.Hosts))
	copy(hosts, c.Hosts)
	return hosts
}

// Exists checks if config file exists
func (c *Config) Exists() bool {
	_, err := os.Stat(c.configPath)
	return !os.IsNotExist(err)
}

// SetMasterHash sets master password hash
func (c *Config) SetMasterHash(hash string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.MasterHash = hash
}

// GetMasterHash returns master password hash
func (c *Config) GetMasterHash() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.MasterHash
}

// Errors
var (
	ErrHostNotFound = &ConfigError{Message: "host not found"}
	ErrAliasExists  = &ConfigError{Message: "alias already exists"}
)

// ConfigError represents a configuration error
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}
