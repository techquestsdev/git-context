package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPaths(t *testing.T) {
	t.Parallel()

	paths, err := NewPaths()
	if err != nil {
		t.Fatalf("NewPaths failed: %v", err)
	}

	if paths == nil {
		t.Fatal("NewPaths returned nil")
	}

	// Check that config directory is set
	if paths.ConfigDir == "" {
		t.Error("ConfigDir should not be empty")
	}

	// Check that config file path is set
	if paths.ConfigFile == "" {
		t.Error("ConfigFile should not be empty")
	}

	// Check that git config file path is set
	if paths.GitConfigFile == "" {
		t.Error("GitConfigFile should not be empty")
	}

	// Check that backup path is set
	if paths.GitConfigBackup == "" {
		t.Error("GitConfigBackup should not be empty")
	}

	// Verify expected paths structure
	home, _ := os.UserHomeDir()

	expectedConfigDir := filepath.Join(home, ".config", "git-context")
	if paths.ConfigDir != expectedConfigDir {
		t.Errorf("Expected ConfigDir %s, got %s", expectedConfigDir, paths.ConfigDir)
	}

	expectedConfigFile := filepath.Join(expectedConfigDir, "config.yaml")
	if paths.ConfigFile != expectedConfigFile {
		t.Errorf("Expected ConfigFile %s, got %s", expectedConfigFile, paths.ConfigFile)
	}

	expectedGitConfig := filepath.Join(home, ".gitconfig")
	if paths.GitConfigFile != expectedGitConfig {
		t.Errorf("Expected GitConfigFile %s, got %s", expectedGitConfig, paths.GitConfigFile)
	}

	expectedBackup := filepath.Join(home, ".gitconfig.bak")
	if paths.GitConfigBackup != expectedBackup {
		t.Errorf("Expected GitConfigBackup %s, got %s", expectedBackup, paths.GitConfigBackup)
	}
}

func TestNewPathsCreatesDirectory(t *testing.T) {
	t.Parallel()

	// This test verifies that NewPaths creates the config directory
	paths, err := NewPaths()
	if err != nil {
		t.Fatalf("NewPaths failed: %v", err)
	}

	// Check that directory exists
	info, err := os.Stat(paths.ConfigDir)
	if err != nil {
		t.Fatalf("Config directory was not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("ConfigDir path exists but is not a directory")
	}
}
