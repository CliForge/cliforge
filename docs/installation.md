# Installation Guide

This guide covers everything you need to install and configure CliForge on your system.

---

## System Requirements

### Operating Systems

CliForge supports the following operating systems:

| OS | Architecture | Status |
|----|--------------|--------|
| **macOS** | x86_64 (Intel) | ✅ Fully supported |
| **macOS** | arm64 (Apple Silicon) | ✅ Fully supported |
| **Linux** | x86_64 | ✅ Fully supported |
| **Linux** | arm64 | ✅ Fully supported |
| **Windows** | x86_64 | ✅ Fully supported |
| **Windows** | arm64 | 🚧 Experimental |

### Software Requirements

- **Go 1.21 or later** (for building from source)
- **Git** (for cloning repository)
- **curl** or **wget** (for downloading releases)

### Optional Dependencies

- **jq** - JSON processing for advanced scripting
- **AWS CLI** - For plugin integration examples
- **Docker** - For running containerized examples

---

## Installation Methods

Choose the installation method that works best for you:

### Method 1: From Releases (Recommended)

**Coming soon**: Pre-built binaries will be available from GitHub Releases.

Once available, install with:

```bash
# macOS/Linux - Auto-detect platform
curl -L https://github.com/cliforge/cliforge/releases/latest/download/install.sh | sh

# Or manually download for your platform
PLATFORM=$(uname -s)-$(uname -m)
VERSION=v0.9.0

curl -L "https://github.com/cliforge/cliforge/releases/download/${VERSION}/cliforge-${PLATFORM}" \
  -o cliforge

chmod +x cliforge
sudo mv cliforge /usr/local/bin/
```

**Windows (PowerShell):**
```powershell
# Download latest release
Invoke-WebRequest -Uri "https://github.com/cliforge/cliforge/releases/latest/download/cliforge-windows-amd64.exe" `
  -OutFile "cliforge.exe"

# Move to Program Files (requires admin)
Move-Item cliforge.exe "C:\Program Files\CliForge\"

# Add to PATH
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", "Machine") + ";C:\Program Files\CliForge",
    "Machine"
)
```

### Method 2: From Source (Current Method)

Building from source gives you the latest development version.

#### Prerequisites

1. **Install Go 1.21 or later**

   ```bash
   # macOS (Homebrew)
   brew install go

   # Linux (Ubuntu/Debian)
   wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin

   # Windows (Chocolatey)
   choco install golang

   # Verify installation
   go version
   ```

2. **Install Git**

   ```bash
   # macOS (Homebrew)
   brew install git

   # Linux (Ubuntu/Debian)
   sudo apt-get update
   sudo apt-get install git

   # Windows (Chocolatey)
   choco install git

   # Verify installation
   git --version
   ```

#### Build Steps

```bash
# 1. Clone the repository
git clone https://github.com/cliforge/cliforge.git
cd cliforge

# 2. Install dependencies
go mod download

# 3. Build the binary
go build -o cliforge ./cmd/cliforge

# 4. (Optional) Install system-wide
sudo mv cliforge /usr/local/bin/

# Or on Windows (PowerShell as admin):
# Move-Item cliforge.exe "C:\Program Files\CliForge\"

# 5. Verify installation
cliforge --version
```

#### Building for Other Platforms

Cross-compile for different platforms:

```bash
# macOS Intel
GOOS=darwin GOARCH=amd64 go build -o cliforge-darwin-amd64 ./cmd/cliforge

# macOS Apple Silicon
GOOS=darwin GOARCH=arm64 go build -o cliforge-darwin-arm64 ./cmd/cliforge

# Linux x86_64
GOOS=linux GOARCH=amd64 go build -o cliforge-linux-amd64 ./cmd/cliforge

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o cliforge-linux-arm64 ./cmd/cliforge

# Windows x86_64
GOOS=windows GOARCH=amd64 go build -o cliforge-windows-amd64.exe ./cmd/cliforge
```

### Method 3: Via Homebrew (Future)

**Coming soon**: Homebrew tap for easy installation on macOS/Linux.

Once available:

```bash
# Add CliForge tap
brew tap cliforge/tap

# Install CliForge
brew install cliforge

# Update CliForge
brew upgrade cliforge

# Uninstall CliForge
brew uninstall cliforge
```

### Method 4: Via Package Managers (Future)

**Future support planned** for:

**APT (Debian/Ubuntu):**
```bash
# Add repository
echo "deb [trusted=yes] https://apt.cliforge.dev stable main" | \
  sudo tee /etc/apt/sources.list.d/cliforge.list

# Install
sudo apt update
sudo apt install cliforge
```

**YUM/DNF (RHEL/Fedora):**
```bash
# Add repository
sudo tee /etc/yum.repos.d/cliforge.repo <<EOF
[cliforge]
name=CliForge Repository
baseurl=https://yum.cliforge.dev/stable
enabled=1
gpgcheck=1
EOF

# Install
sudo dnf install cliforge
```

**Scoop (Windows):**
```powershell
# Add bucket
scoop bucket add cliforge https://github.com/cliforge/scoop-bucket

# Install
scoop install cliforge
```

**Chocolatey (Windows):**
```powershell
# Install
choco install cliforge
```

### Method 5: Docker Container (Future)

**Coming soon**: Docker images for running CliForge in containers.

```bash
# Run CliForge in Docker
docker run --rm -it cliforge/cliforge:latest init my-cli

# Build CLI in Docker
docker run --rm -v $(pwd):/workspace cliforge/cliforge:latest \
  build --config /workspace/cli-config.yaml --output /workspace/dist/
```

---

## Verify Installation

After installation, verify CliForge is working correctly.

### Check Version

```bash
cliforge --version
# Expected output:
# CliForge v0.9.0
```

### Check Available Commands

```bash
cliforge --help
# Expected output:
# CliForge - Forge CLIs from APIs
#
# Usage:
#   cliforge [command]
#
# Available Commands:
#   init        Initialize a new CLI project
#   build       Build CLI binary from configuration
#   validate    Validate configuration and OpenAPI spec
#   help        Help about any command
#
# Flags:
#   -h, --help      help for cliforge
#   -v, --version   version for cliforge
```

### Run Example

```bash
# Initialize a test project
cliforge init test-cli
cd test-cli

# Verify files created
ls -la
# cli-config.yaml
# README.md

# Validate configuration
cliforge validate cli-config.yaml
# Expected: ✓ Configuration is valid
```

---

## Post-Installation Configuration

### Shell Completion

Enable shell completion for better CLI experience.

**Bash:**
```bash
# Generate completion script
cliforge completion bash > /etc/bash_completion.d/cliforge

# Or for user-level
cliforge completion bash > ~/.bash_completion.d/cliforge

# Reload shell
source ~/.bashrc
```

**Zsh:**
```bash
# Generate completion script
cliforge completion zsh > ~/.zsh/completions/_cliforge

# Add to .zshrc if not already present
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc

# Reload shell
source ~/.zshrc
```

**Fish:**
```bash
# Generate completion script
cliforge completion fish > ~/.config/fish/completions/cliforge.fish

# Reload shell
source ~/.config/fish/config.fish
```

**PowerShell:**
```powershell
# Add to profile
cliforge completion powershell | Out-String | Invoke-Expression

# Or save to profile
cliforge completion powershell >> $PROFILE
```

### Environment Variables

Set helpful environment variables:

```bash
# Add to ~/.bashrc, ~/.zshrc, or ~/.profile

# CliForge home directory
export CLIFORGE_HOME="$HOME/.cliforge"

# Default output format
export CLIFORGE_OUTPUT="table"

# Disable update checks (if needed)
export CLIFORGE_NO_UPDATE_CHECK="true"

# Enable debug logging
export CLIFORGE_DEBUG="false"
```

### File Locations

CliForge follows the XDG Base Directory Specification:

**macOS/Linux:**
```
~/.config/cliforge/          # Configuration files
~/.cache/cliforge/           # Cache files (OpenAPI specs)
~/.local/share/cliforge/     # Data files (logs, state)
~/.local/state/cliforge/     # State files (last update check)
```

**Windows:**
```
%APPDATA%\cliforge\          # Configuration files
%LOCALAPPDATA%\cliforge\     # Cache and data files
```

**Create directories (optional):**
```bash
# macOS/Linux
mkdir -p ~/.config/cliforge
mkdir -p ~/.cache/cliforge
mkdir -p ~/.local/share/cliforge
mkdir -p ~/.local/state/cliforge

# Windows (PowerShell)
New-Item -ItemType Directory -Force -Path "$env:APPDATA\cliforge"
New-Item -ItemType Directory -Force -Path "$env:LOCALAPPDATA\cliforge"
```

---

## Updating CliForge

### Automatic Updates (Future)

Once released, CliForge will support self-updates:

```bash
# Check for updates
cliforge update check

# Update to latest version
cliforge update

# Update to specific version
cliforge update --version v1.0.0

# Update to beta channel
cliforge update --channel beta
```

### Manual Updates

#### From Releases

```bash
# Download latest release
curl -L https://github.com/cliforge/cliforge/releases/latest/download/cliforge-$(uname -s)-$(uname -m) \
  -o cliforge-new

# Verify it works
chmod +x cliforge-new
./cliforge-new --version

# Replace existing binary
sudo mv cliforge-new /usr/local/bin/cliforge
```

#### From Source

```bash
cd cliforge

# Pull latest changes
git pull origin main

# Rebuild
go build -o cliforge ./cmd/cliforge

# Reinstall
sudo mv cliforge /usr/local/bin/
```

#### Via Package Managers

```bash
# Homebrew
brew upgrade cliforge

# APT
sudo apt update
sudo apt upgrade cliforge

# Chocolatey
choco upgrade cliforge
```

---

## Uninstallation

### Remove Binary

```bash
# macOS/Linux
sudo rm /usr/local/bin/cliforge

# Windows (PowerShell as admin)
Remove-Item "C:\Program Files\CliForge\cliforge.exe"
```

### Remove Data Files

**Keep your CLI projects**, but remove CliForge data:

```bash
# macOS/Linux
rm -rf ~/.config/cliforge
rm -rf ~/.cache/cliforge
rm -rf ~/.local/share/cliforge
rm -rf ~/.local/state/cliforge

# Windows (PowerShell)
Remove-Item -Recurse -Force "$env:APPDATA\cliforge"
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\cliforge"
```

### Remove Shell Completions

```bash
# Bash
rm /etc/bash_completion.d/cliforge
rm ~/.bash_completion.d/cliforge

# Zsh
rm ~/.zsh/completions/_cliforge

# Fish
rm ~/.config/fish/completions/cliforge.fish
```

### Via Package Managers

```bash
# Homebrew
brew uninstall cliforge

# APT
sudo apt remove cliforge

# Chocolatey
choco uninstall cliforge
```

---

## Troubleshooting Common Installation Issues

### Issue: "cliforge: command not found"

**Cause**: Binary not in PATH

**Solution**:
```bash
# Check if binary exists
which cliforge

# If not found, check installation location
ls -la /usr/local/bin/cliforge

# Add to PATH
export PATH=$PATH:/usr/local/bin

# Make permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc
```

### Issue: "permission denied" when running cliforge

**Cause**: Binary not executable

**Solution**:
```bash
# Make executable
chmod +x /usr/local/bin/cliforge

# Or if in current directory
chmod +x ./cliforge
```

### Issue: "cannot execute binary file"

**Cause**: Wrong architecture or corrupted download

**Solution**:
```bash
# Check your architecture
uname -m
# x86_64, arm64, etc.

# Check file type
file /usr/local/bin/cliforge
# Should show: "Mach-O 64-bit executable" (macOS) or "ELF 64-bit LSB executable" (Linux)

# Re-download correct version
PLATFORM=$(uname -s)-$(uname -m)
curl -L "https://github.com/cliforge/cliforge/releases/latest/download/cliforge-${PLATFORM}" \
  -o cliforge
chmod +x cliforge
sudo mv cliforge /usr/local/bin/
```

### Issue: Go build fails with "missing dependencies"

**Cause**: Dependencies not downloaded

**Solution**:
```bash
# Clear module cache
go clean -modcache

# Download dependencies
go mod download

# Verify dependencies
go mod verify

# Build again
go build -o cliforge ./cmd/cliforge
```

### Issue: "unsupported Go version"

**Cause**: Go version too old

**Solution**:
```bash
# Check Go version
go version

# Should be 1.21 or later
# Update Go from https://go.dev/dl/

# Or via Homebrew (macOS)
brew upgrade go

# Or via package manager (Linux)
sudo apt-get update
sudo apt-get install golang-1.21
```

### Issue: "certificate verify failed" during download

**Cause**: SSL/TLS certificate issues

**Solution**:
```bash
# Update CA certificates
# macOS
brew update && brew upgrade ca-certificates

# Linux (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install ca-certificates

# Or use --insecure (not recommended)
curl -L --insecure https://... -o cliforge
```

### Issue: Windows SmartScreen warning

**Cause**: Unsigned binary (expected for now)

**Solution**:
1. Click "More info"
2. Click "Run anyway"

Future releases will be code-signed to prevent this warning.

### Issue: macOS Gatekeeper blocking execution

**Cause**: Binary not code-signed

**Solution**:
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine /usr/local/bin/cliforge

# Or via System Preferences
# System Preferences > Security & Privacy > General
# Click "Allow Anyway" next to the blocked app message
```

### Issue: "disk space" error during installation

**Cause**: Insufficient disk space

**Solution**:
```bash
# Check available space
df -h

# Clear cache
rm -rf ~/.cache/*

# Or use a different location
mkdir ~/my-cliforge
cd ~/my-cliforge
# Install here instead
```

---

## Platform-Specific Notes

### macOS

**Apple Silicon (M1/M2/M3):**
- Use the `arm64` binary for native performance
- Rosetta 2 can run `x86_64` binaries if needed
- Check architecture: `uname -m` (should show `arm64`)

**Homebrew:**
- CliForge will be available via Homebrew tap (coming soon)
- For now, build from source or use releases

**Keychain Access:**
- CliForge stores OAuth2 tokens in macOS Keychain
- You may be prompted to allow access on first auth

### Linux

**Distribution-Specific Notes:**

**Ubuntu/Debian:**
- Install Go from apt: `sudo apt install golang-go`
- Or download latest from golang.org for newer versions

**Fedora/RHEL:**
- Install Go: `sudo dnf install golang`

**Arch Linux:**
- Install Go: `sudo pacman -S go`

**Keyring Support:**
- Install `gnome-keyring` or `kwallet` for secure token storage
- Ubuntu: `sudo apt install gnome-keyring`
- Fedora: `sudo dnf install gnome-keyring`

**Running as non-root:**
- Use `$HOME/.local/bin` instead of `/usr/local/bin`
- Add to PATH: `export PATH=$PATH:$HOME/.local/bin`

### Windows

**PowerShell Execution Policy:**
```powershell
# Check current policy
Get-ExecutionPolicy

# If restricted, set to RemoteSigned (run as admin)
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
```

**PATH Configuration:**
- Add to User PATH (no admin required)
- Or System PATH (requires admin)

**Credential Manager:**
- CliForge stores tokens in Windows Credential Manager
- Access via: Control Panel > Credential Manager

**WSL (Windows Subsystem for Linux):**
- Install Linux version in WSL for better compatibility
- Or use Windows version with PowerShell

---

## Building Your First CLI

Now that CliForge is installed, you're ready to create your first CLI:

```bash
# Initialize a new project
cliforge init my-api-cli

# Edit the configuration
cd my-api-cli
# Edit cli-config.yaml with your API details

# Build the CLI
cliforge build --config cli-config.yaml --output dist/

# Test it
./dist/my-api-cli-$(uname -s)-$(uname -m) --help
```

**Next steps:**
- Read the [Getting Started Guide](getting-started.md) for a complete tutorial
- Review the [Configuration DSL](configuration-dsl.md) for all options
- Explore the Petstore Example in the [GitHub repository](https://github.com/CliForge/cliforge/tree/main/examples/petstore) for real-world patterns

---

## Getting Help

### Documentation

- [Getting Started Guide](getting-started.md)
- [Configuration DSL Reference](configuration-dsl.md)
- [Technical Specification](technical-specification.md)
- [Contributing Guide](https://github.com/CliForge/cliforge/blob/main/CONTRIBUTING.md)

### Community

- **Issues**: [github.com/cliforge/cliforge/issues](https://github.com/cliforge/cliforge/issues)
- **Discussions**: [github.com/cliforge/cliforge/discussions](https://github.com/cliforge/cliforge/discussions)

### Reporting Issues

When reporting installation issues, please include:

```bash
# System information
uname -a

# Go version
go version

# CliForge version (if installed)
cliforge --version

# Installation method used
# (from source, from release, etc.)

# Error messages (full output)
# Copy the complete error message
```

---

## What's Next?

Installation complete! Here's what to do next:

1. **Tutorial**: Follow the [Getting Started Guide](getting-started.md) to build your first CLI
2. **Example**: See the Petstore Example in the [GitHub repository](https://github.com/CliForge/cliforge/tree/main/examples/petstore) to explore all features
3. **Configuration**: Review the [Configuration DSL](configuration-dsl.md) for customization options
4. **Community**: Join discussions and share your CLI creations

Happy building with CliForge!
