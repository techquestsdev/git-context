package config

import (
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
)

// Paths contains all the important path locations for git-context.
type Paths struct {
	ConfigDir       string
	ConfigFile      string
	GitConfigFile   string
	GitConfigBackup string
}

// NewPaths initializes and creates paths with proper defaults.
func NewPaths() (*Paths, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get user home directory")
	}

	configDir := filepath.Join(home, ".config", "git-context")
	configFile := filepath.Join(configDir, "config.yaml")
	gitConfigFile := filepath.Join(home, ".gitconfig")
	gitConfigBackup := filepath.Join(home, ".gitconfig.bak")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		return nil, errors.Wrap(err, "failed to create config directory")
	}

	return &Paths{
		ConfigDir:       configDir,
		ConfigFile:      configFile,
		GitConfigFile:   gitConfigFile,
		GitConfigBackup: gitConfigBackup,
	}, nil
}
