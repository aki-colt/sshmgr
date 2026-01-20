# SSH Manager (sshmgr)

A modern SSH connection manager written in Go with CLI mode support.

## Features

- ✅ **Encrypted Storage**: All passwords encrypted with AES-256-GCM using a master password
- ✅ **Alias Support**: Connect to servers using easy-to-remember aliases
- ✅ **Fuzzy Search**: Find hosts by partial alias matches
- ✅ **Password Management**: Secure password encryption and decryption
- ✅ **Connection Testing**: Test SSH connections before saving
- ✅ **Full CRUD**: Add, list, modify, and delete SSH hosts
- ✅ **Shell Completion**: Auto-completion for bash, zsh, fish, and PowerShell
- ✅ **Cross-platform**: Single binary for macOS, Linux, and Windows

## Installation

### Prerequisites

- **Go 1.23+** for building from source or installing via `go install`
- **sshpass** for password-based SSH connections
  - macOS: `brew install hudochenkov/sshpass/sshpass`
  - Linux: `sudo apt-get install sshpass`

### Dependencies

Go will automatically download all required dependencies when building:
- `gopkg.in/yaml.v3` - YAML parsing
- `github.com/spf13/cobra` - CLI framework
- `github.com/sahilm/fuzzy` - Fuzzy search library
- Standard library for encryption and SSH connections

### Install via go install (Recommended)

```bash
go install github.com/aki-colt/sshmgr@latest
```

This will install the binary to `~/go/bin/` (or `$(go env GOPATH)/bin`). Make sure this directory is in your `$PATH`:

```bash
# Add to your shell profile (~/.bashrc, ~/.zshrc, etc.)
export PATH=$PATH:$(go env GOPATH)/bin
```

### Build from Source

```bash
cd ssh-manager-go
go build -o sshmgr .
```

### Install System-wide

```bash
go build -o sshmgr .
sudo mv sshmgr /usr/local/bin/
```

### Setup Shell Completion

Auto-completion is automatically enabled when you run `sshmgr init` for the first time.

#### Automatic Installation (Recommended)

```bash
# Auto-completion is installed automatically during init
sshmgr init
# Follow the instructions to source your shell config
```

If you need to reinstall completion later:

```bash
sshmgr completion-install
```

#### Manual Installation

If you prefer manual setup or need to configure manually:

Bash:
```bash
# Load for current session
$ source <(sshmgr completion bash)

# Load for all sessions (Linux)
$ sshmgr completion bash > /etc/bash_completion.d/sshmgr

# Load for all sessions (macOS)
$ sshmgr completion bash > /usr/local/etc/bash_completion.d/sshmgr
```

Zsh:
```bash
# Load for current session
$ sshmgr completion zsh > "${fpath[1]}/_sshmgr"

# Start a new shell for changes to take effect
```

Fish:
```bash
# Load for current session
$ sshmgr completion fish | source

# Load for all sessions
$ sshmgr completion fish > ~/.config/fish/completions/sshmgr.fish
```

PowerShell:
```powershell
# Load for current session
PS> sshmgr completion powershell | Out-String | Invoke-Expression

# Load for all sessions
PS> sshmgr completion powershell > sshmgr.ps1
# Then add to your PowerShell profile
```

## Usage

### Initial Setup

First, initialize the master password:

```bash
$ sshmgr init
Enter master password: ********
Confirm master password: ********
Master password set successfully!
```

### CLI Mode

#### Add a New Host

```bash
$ sshmgr add
Enter alias: myserver
Enter host address: 192.168.1.100
Enter username: admin
Enter password: ********
Enter port (default 22): 22
Test connection? [Y/n]: Y
Connection test successful!
Host added successfully!
```

#### List All Hosts

```bash
$ sshmgr list

ID    Alias                Host                           User             Port
---------------------------------------------------------------------
1     myserver             192.168.1.100                   admin             22
```

#### Connect to a Host

```bash
$ sshmgr connect myserver
Enter master password: ********
Connecting to 192.168.1.100 as admin...
```

#### Fuzzy Search and Connect

If you don't remember exact alias, you can use partial matches:

```bash
# These all work for alias "myserver"
$ sshmgr connect myserver
$ sshmgr connect myser
$ sshmgr connect mysv

# Or use shortcut with fuzzy matching
$ sshmgr myser
```

#### Quick Connect (using alias directly)

```bash
$ sshmgr myserver
Enter master password: ********
Connecting to 192.168.1.100 as admin...
```

#### Modify a Host

```bash
$ sshmgr modify myserver
Enter master password: ********

Current configuration:
  Alias: myserver
  Host: 192.168.1.100
  User: admin
  Port: 22

Enter new alias (press Enter to keep current):
Enter new host (press Enter to keep current):
Enter new user (press Enter to keep current):
Enter new password (press Enter to keep current):
Enter new port (press Enter to keep current): 2222
Test connection? [Y/n]: Y
Host modified successfully!
```

#### Delete a Host

```bash
$ sshmgr delete myserver
Enter master password: ********
Are you sure you want to delete host 'myserver'? [y/N]: y
Host deleted successfully!
```

## Configuration

Configuration is stored in `~/.ssh_manager_config.yaml` in the following format:

```yaml
version: "1.0"
master_hash: a1b2c3...
hosts:
  - id: "1"
    alias: myserver
    host: 192.168.1.100
    user: admin
    password: encrypted_base64_string
    port: 22
    created_at: "2026-01-09"
    updated_at: "2026-01-09"
```

### Password Encryption

- **Algorithm**: AES-256-GCM (Galois/Counter Mode)
- **Key Derivation**: SHA-256 hash of master password
- **Storage**: Base64-encoded encrypted ciphertext

## Project Structure

```
ssh-manager-go/
├── main.go              # Application entry point
├── pkg/
│   ├── cli/
│   │   ├── commands.go   # CLI command definitions
│   │   └── helpers.go   # CLI helper functions
│   ├── config/
│   │   └── config.go    # Configuration management
│   ├── encryption/
│   │   └── encryption.go # AES-256-GCM encryption
│   └── ssh/
│       └── ssh.go       # SSH connection handling
├── go.mod
├── go.sum
└── README.md
```

## Security Considerations

- ✅ Passwords are encrypted before storage
- ✅ Master password required for decryption
- ✅ Configuration file permissions set to 0600 (owner read/write only)
- ✅ No passwords in command-line arguments
- ⚠️ **Important**: Keep your master password secure - it cannot be recovered if lost

## Comparison with Original Bash Script

| Feature | Bash Script | Go Implementation |
|----------|-------------|-------------------|
| Password Storage | Plaintext | Encrypted (AES-256-GCM) |
| Config Format | Plaintext | YAML |
| Alias Support | ❌ | ✅ |
| Fuzzy Search | ❌ | ✅ |
| Shell Completion | ❌ | ✅ |
| CLI Mode | ✅ | ✅ |
| Password Management | Plaintext | Encrypted |
| Cross-platform | Linux/macOS | Linux/macOS/Windows |
| Single Binary | ❌ | ✅ |

## Dependencies

- **github.com/spf13/cobra** - CLI framework
- **github.com/sahilm/fuzzy** - Fuzzy search library
- **Standard Library** - crypto/aes, crypto/cipher, crypto/sha256

## Roadmap

- [ ] SSH key support
- [ ] Host groups/tags
- [ ] Port forwarding configuration
- [ ] Command execution on remote hosts
- [ ] Configuration export/import
- [ ] History tracking
- [ ] Batch operations

## License

MIT License - Feel free to use and modify as needed.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
