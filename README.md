# Git Context Switcher

A powerful command-line tool for managing multiple Git configuration profiles. Seamlessly switch between different Git identities (work, personal, school, etc.) with a single command.

![Go Version](https://img.shields.io/github/go-mod/go-version/techquestsdev/git-context?logo=go&logoColor=white)
[![codecov](https://codecov.io/gh/techquestsdev/git-context/branch/main/graph/badge.svg)](https://codecov.io/gh/techquestsdev/git-context)
![License](https://img.shields.io/github/license/techquestsdev/git-context)

## ✨ Features

- 🎯 **Multiple Profiles** - Create and manage unlimited Git configuration profiles
- ⚡ **Instant Switching** - Switch between profiles with a single command
- 🌍 **Global Settings** - Define shared configuration across all profiles
- 🔗 **URL Rewrites** - Configure Git URL rewrites per profile (SSH <-> HTTPS)
- 🛡️ **Safe Operations** - Automatic backups before switching profiles
- 💬 **Interactive Setup** - Guided profile creation with prompts
- 🎨 **Colored Output** - Beautiful, easy-to-read CLI output
- 📝 **YAML Configuration** - Simple, human-readable configuration format

## 🚀 Installation

### 📦 Quick Install (Recommended)

#### Using Go Install

```bash
go install github.com/techquestsdev/git-context@latest
```

Then run with: `git-context`

#### Using Homebrew (MacOS/Linux)

```bash
brew tap techquestsdev/tap
brew install git-context
```

#### Download Binary

Download the latest release for your platform from the [Releases page](https://github.com/techquestsdev/git-context/releases):

- **Linux** (amd64, arm64)
- **MacOS** (Intel, Apple Silicon)
- **Windows** (amd64)

```bash
# Example: Download and install on Linux/MacOS
curl -L https://github.com/techquestsdev/git-context/releases/latest/download/git-context_Linux_x86_64.tar.gz | tar xz
sudo mv git-context /usr/local/bin/
```

### 🛠️ Build from Source

#### Prerequisites

- **Go** 1.25 or higher
- **Git** 2.0 or higher
- **Make** (optional, for using Makefile commands)

#### Clone and Build

```bash
# Clone the repository
git clone https://github.com/techquestsdev/git-context.git
cd git-context

# Download dependencies
go mod download

# Build the application
go build -o bin/git-context

# Install to your system
make install
# Or without Make:
go install

# Verify installation
git-context --version
```

## 📖 Usage

### Quick Start Guide

#### 1. Initialize Configuration

First-time setup:

```bash
git-context init
```

This creates `~/.config/git-context/config.yaml`.

### Add Your First Profile

```bash
git-context add work
```

You'll be prompted for:

- **Git Name** - Your full name
- **Git Email** - Your email address
- **GPG Signing Key** (optional) - Your GPG key ID
- **URL rewrites** (optional) - SSH/HTTPS URL conversions

#### 3. List All Profiles

```bash
git-context list
```

**Output:**

```text
=== Available Profiles ===

Profile      Email                         Status
-------      -----                         ------
work         andre@work.com                ● (active)
personal     andre@personal.com
school       andre@school.edu
```

#### 4. Switch Between Profiles

```bash
git-context switch personal
```

**Output:**

```text
=== Switching to Profile: personal ===

ℹ Backed up git config to /Users/andre/.gitconfig.bak
✓ Switched to profile 'personal'
ℹ User: Andre Nogueira <andre@personal.com>
```

#### 5. Show Current Profile

```bash
git-context current
```

#### 6. Show Profile Details

```bash
git-context show work
```

#### 7. Remove a Profile

```bash
git-context remove school
```

### All Available Commands

| Command                      | Description                         |
| ---------------------------- | ----------------------------------- |
| `git-context init`            | Initialize configuration            |
| `git-context add <name>`      | Create a new profile                |
| `git-context switch <name>`   | Switch to a profile                 |
| `git-context list`            | List all profiles                   |
| `git-context current`         | Show active profile                 |
| `git-context show <name>`     | Show profile details                |
| `git-context remove <name>`   | Delete a profile                    |
| `git-context --help`          | Show help                           |
| `git-context --version`       | Show version                        |

## ⚙️ Configuration

The configuration is stored in YAML format at `~/.config/git-context/config.yaml`.

### Configuration Structure

```yaml
global:
  # Shared settings across all profiles
  <section>:
    <key>: <value>

profiles:
  <profile-name>:
    user:
      name: "Your Name"
      email: "your.email@example.com"
      signingkey: "GPG_KEY_ID"  # optional
    url:
      - pattern: "ssh://git@gitlab.com/"
        insteadOf: "https://gitlab.com/"
    # Any other Git config sections...
```

### Example Configuration

```yaml
global:
  core:
    pager: delta
    editor: vim
  push:
    autoSetupRemote: true
  merge:
    conflictstyle: diff3
  commit:
    gpgsign: true
  gpg:
    program: /usr/local/bin/gpg
  pull:
    rebase: true

profiles:
  work:
    user:
      name: "Andre Nogueira"
      email: "andre.nogueira@mollie.com"
      signingkey: "A0A90F4231D8B028"
    url:
      - pattern: "git@gitlab.molops.io/"
        insteadOf: "https://gitlab.molops.io/"
      - pattern: "ssh://git@github.com/"
        insteadOf: "https://github.com/"
    http:
      postBuffer: 157286400

  personal:
    user:
      name: "Andre Nogueira"
      email: "andre@personal.com"
    url:
      - pattern: "ssh://git@github.com/"
        insteadOf: "https://github.com/"

  school:
    user:
      name: "Andre N."
      email: "andre@school.edu"
```

### Global vs Profile-Specific Settings

- **Global settings** are applied to all profiles
- **Profile-specific settings** override global settings
- Any Git configuration section can be used (core, push, merge, etc.)

### Common Configuration Sections

| Section        | Purpose                                | Examples                              |
| -------------- | -------------------------------------- | ------------------------------------- |
| `user`         | User identity                          | name, email, signingkey               |
| `core`         | Core Git settings                      | editor, pager, autocrlf               |
| `push`         | Push behavior                          | default, autoSetupRemote              |
| `pull`         | Pull behavior                          | rebase, ff                            |
| `merge`        | Merge strategies                       | conflictStyle, tool                   |
| `commit`       | Commit settings                        | gpgsign, template                     |
| `gpg`          | GPG configuration                      | program, format                       |
| `url`          | URL rewrites                           | pattern, insteadOf                    |
| `http`         | HTTP settings                          | postBuffer, sslVerify                 |
| `delta`        | Delta pager configuration              | navigate, line-numbers                |

> Based on my personal configuration needs.

## 🎯 Use Cases

### Scenario 1: Work vs Personal Repositories

Separate professional and personal Git identities:

```bash
# Create work profile with company email
git-context add work
# Enter: Your Name, you@company.com, work-gpg-key

# Create personal profile with personal email
git-context add personal
# Enter: Your Name, you@personal.com, personal-gpg-key

# Switch based on what you're working on
git-context switch work      # For company projects
git-context switch personal  # For personal projects
```

### Scenario 2: Multiple Companies/Clients

Manage identities for different employers or clients:

```bash
git-context add client-a
git-context add client-b
git-context add freelance

# Switch when changing projects
git-context switch client-a
```

### Scenario 3: Academic and Professional

Separate academic credentials from professional:

```bash
git-context add university   # .edu email, academic GPG key
git-context add work         # Company email, work GPG key
```

### Scenario 4: Different URL Rewrites

Configure SSH access for different Git hosts:

```yaml
profiles:
  work:
    url:
      - pattern: "ssh://git@gitlab.company.com/"
        insteadOf: "https://gitlab.company.com/"
  personal:
    url:
      - pattern: "ssh://git@github.com/"
        insteadOf: "https://github.com/"
```

## 🛡️ Safety Features

- ✅ **Automatic Backups** - Before switching profiles, current Git config is backed up to `~/.gitconfig.bak`
- ✅ **Confirmation Prompts** - Destructive operations require user confirmation
- ✅ **Validation** - Profiles are validated before being applied
- ✅ **Error Handling** - Clear error messages guide you when something goes wrong
- ✅ **Non-Destructive Init** - `init` command preserves existing profiles

## 🐛 Troubleshooting

### Config File Not Found

**Solution:** Initialize the configuration:

```bash
git-context init
```

### Profile Not Found

**Solution:** Check available profiles:

```bash
git-context list
```

### Git Config Not Updating

**Problem:** No write permissions to `~/.gitconfig`

**Solution:** Check file permissions:

```bash
ls -la ~/.gitconfig
chmod 644 ~/.gitconfig  # If needed
```

### Restore from Backup

If something went wrong, restore from the automatic backup:

```bash
cp ~/.gitconfig.bak ~/.gitconfig
```

### Profile Already Exists

**Problem:** `git-context init` reports profiles exist

**Solution:** This is expected! The `init` command now preserves existing profiles instead of clearing them.

## 🧪 Testing

The project has comprehensive test coverage and zero linting issues, leveraging Go's testing framework and `golangci-lint` for code quality.

### Run Tests

```bash
# Run all tests
make test
# Or:
go test -v

# Run tests with coverage
make test-coverage
# Or:
go test -v -cover

# Generate detailed coverage report
make test-coverage-view
# Or:
go test -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run benchmarks
make bench
```

### Linting and Formatting

```bash
# Run linter
make lint

# Auto-fix linting issues
make lint-fix

# Format code
make fmt

# Run vet
make vet

# Check everything (format + lint + test)
make check

# Run all checks and build
make all
```

### Test Categories

- ✅ **Configuration Management** - Profile CRUD, merging, persistence
- ✅ **Git Operations** - Config file parsing, URL rewrites, backups
- ✅ **UI Components** - Colored output, tables, formatting
- ✅ **Command Integration** - Profile operations, error handling
- ✅ **Edge Cases** - Invalid YAML, missing files, error paths

## 🏗️ Architecture

### Project Structure

```text
git-context/
├── cmd/                       # CLI commands
│   ├── root.go               # Root command & init
│   ├── add.go                # Add profile command
│   ├── switch.go             # Switch profile command
│   ├── list.go               # List profiles command
│   ├── remove.go             # Remove profile command
│   ├── current.go            # Show current profile
│   ├── show.go               # Show profile details
│   └── cmd_test.go           # Command tests
├── internal/
│   ├── config/
│   │   ├── config.go         # Configuration management
│   │   ├── config_test.go    # Config tests
│   │   ├── paths.go          # Path management
│   │   └── paths_test.go     # Path tests
│   ├── git/
│   │   ├── git.go            # Git operations
│   │   └── git_test.go       # Git tests
│   └── ui/
│       ├── output.go         # UI/UX helpers
│       └── output_test.go    # UI tests
├── bin/                       # Build output
├── main.go                    # Entry point
├── go.mod                     # Go module dependencies
├── go.sum                     # Dependency checksums
└── README.md                  # This file
```

### Design Principles

- **Separation of Concerns** - Clear boundaries between CLI, config, git, and UI layers
- **Testability** - High test coverage with isolated unit tests
- **User Experience** - Colored output, clear messages, confirmation prompts
- **Safety First** - Automatic backups, validation, error handling

## 📦 Dependencies

- [github.com/spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [github.com/manifoldco/promptui](https://github.com/manifoldco/promptui) - Interactive prompts
- [github.com/fatih/color](https://github.com/fatih/color) - Colored terminal output
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML parsing and serialization

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a [Pull Request](https://github.com/techquestsdev/git-context/pulls)

### Before Submitting

- ✅ Ensure all tests pass (`go test -v ./...`)
- ✅ Maintain or improve code coverage
- ✅ Follow existing code style
- ✅ Add tests for new features
- ✅ Update documentation as needed

### Development Commands

The project includes a comprehensive Makefile with the following commands:

```bash
# Show all available make commands
make help

# Development
make run                  # Run the application
make build                # Build the binary
make install              # Install to GOPATH/bin

# Testing
make test                 # Run tests with race detection
make test-verbose         # Run tests with verbose output
make test-coverage        # Generate coverage report
make test-coverage-view   # Generate and open coverage report in browser
make bench                # Run benchmarks

# Code Quality
make lint                 # Run golangci-lint
make lint-fix             # Run linter with auto-fix
make fmt                  # Format code with gofmt
make vet                  # Run go vet
make check                # Run all checks (fmt, vet, lint, test)

# Dependencies
make deps                 # Download dependencies
make deps-upgrade         # Upgrade all dependencies
make deps-verify          # Verify dependencies

# Cleanup
make clean                # Remove build artifacts and coverage reports
make clean-all            # Full cleanup including Go cache

# Utilities
make mod-graph            # Show module dependency graph
make mod-why PKG=...      # Show why a package is needed
make version              # Show Go version

# All-in-one
make all                  # Run fmt, lint, test, and build
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) for the excellent CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) ecosystem for inspiration
- [PromptUI](https://github.com/manifoldco/promptui) for interactive prompts

## 📧 Contact

André Nogueira - [@aanogueira](https://github.com/aanogueira)

Project Link: [https://github.com/techquestsdev/git-context](https://github.com/techquestsdev/git-context)

---

### Made with ❤️ and Go
