package git

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGit(t *testing.T) {
	t.Parallel()

	configPath := "/path/to/.gitconfig"
	g := NewGit(configPath)

	if g == nil {
		t.Fatal("NewGit returned nil")
	}

	if g.globalConfigPath != configPath {
		t.Errorf("Expected globalConfigPath %s, got %s", configPath, g.globalConfigPath)
	}
}

func TestBackupConfigNonExistent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.gitconfig")
	backupPath := filepath.Join(tmpDir, "backup.gitconfig")

	g := NewGit(configPath)

	// Backing up non-existent config should not error
	err := g.BackupConfig(backupPath)
	if err != nil {
		t.Errorf("BackupConfig should not error for non-existent file: %v", err)
	}
}

func TestBuildGitConfig(t *testing.T) {
	t.Parallel()

	config := map[string]any{
		"user.name":     "Test User",
		"user.email":    "test@example.com",
		"core.editor":   "vim",
		"core.autocrlf": "input",
	}

	content := buildGitConfig(config)

	// Verify sections exist
	if !strings.Contains(content, "[user]") {
		t.Error("Config should contain [user] section")
	}

	if !strings.Contains(content, "[core]") {
		t.Error("Config should contain [core] section")
	}

	// Verify values
	if !strings.Contains(content, "name = Test User") {
		t.Error("Config should contain user.name")
	}

	if !strings.Contains(content, "email = test@example.com") {
		t.Error("Config should contain user.email")
	}

	if !strings.Contains(content, "editor = vim") {
		t.Error("Config should contain core.editor")
	}

	if !strings.Contains(content, "autocrlf = input") {
		t.Error("Config should contain core.autocrlf")
	}
}

func TestBuildGitConfigWithQuotedSubsection(t *testing.T) {
	t.Parallel()

	config := map[string]any{
		`url "ssh://git@github.com/".insteadOf`: "https://github.com/",
		"user.name":                             "Test User",
	}

	content := buildGitConfig(config)

	// Verify quoted subsection
	if !strings.Contains(content, `[url "ssh://git@github.com/"]`) {
		t.Error("Config should contain quoted URL subsection")
	}

	if !strings.Contains(content, "insteadOf = https://github.com/") {
		t.Error("Config should contain insteadOf value")
	}
}

func TestWriteConfigWithComplexStructure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.gitconfig")

	g := NewGit(configPath)

	// Write config with nested sections
	config := map[string]any{
		"user.name":                             "Test User",
		"user.email":                            "test@example.com",
		"user.signingkey":                       "ABCD1234",
		"core.editor":                           "vim",
		"core.autocrlf":                         "input",
		"push.default":                          "simple",
		"pull.rebase":                           "false",
		`url "ssh://git@github.com/".insteadOf`: "https://github.com/",
	}

	err := g.WriteConfig(config)
	if err != nil {
		t.Fatalf("WriteConfig failed: %v", err)
	}

	// Read back and verify
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read config: %v", err)
	}

	contentStr := string(content)

	// Verify all sections and values
	expectedPairs := []string{
		"[user]",
		"name = Test User",
		"email = test@example.com",
		"signingkey = ABCD1234",
		"[core]",
		"editor = vim",
		"autocrlf = input",
		"[push]",
		"default = simple",
		"[pull]",
		"rebase = false",
	}

	for _, expected := range expectedPairs {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Config should contain '%s'", expected)
		}
	}
}

func TestWriteConfigError(t *testing.T) {
	t.Parallel()

	g := NewGit("/invalid/path/to/config")

	config := map[string]any{
		"user.name": "Test",
	}

	err := g.WriteConfig(config)
	if err == nil {
		t.Error("WriteConfig should fail for invalid path")
	}
}

func TestBackupConfigError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.gitconfig")

	// Create a directory instead of a file to cause read error
	if err := os.Mkdir(configPath, 0o755); err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	g := NewGit(configPath)

	err := g.BackupConfig(filepath.Join(tmpDir, "backup"))
	if err == nil {
		t.Error("BackupConfig should fail when config is a directory")
	}
}

func TestBackupConfigSuccess(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.gitconfig")
	backupPath := filepath.Join(tmpDir, "backup.gitconfig")

	// Create a config file with content
	content := "[user]\n\tname = Test User\n\temail = test@example.com\n"
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	g := NewGit(configPath)

	// Backup should succeed
	err := g.BackupConfig(backupPath)
	if err != nil {
		t.Fatalf("BackupConfig failed: %v", err)
	}

	// Verify backup content matches original
	backupContent, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("Failed to read backup: %v", err)
	}

	if string(backupContent) != content {
		t.Error("Backup content should match original")
	}
}

func TestBackupConfigInvalidBackupPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.gitconfig")

	// Create a valid config file
	content := "[user]\n\tname = Test\n"
	if err := os.WriteFile(configPath, []byte(content), 0o644); err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	g := NewGit(configPath)

	// Try to backup to an invalid location
	err := g.BackupConfig("/invalid/path/backup")
	if err == nil {
		t.Error("BackupConfig should fail with invalid backup path")
	}
}

func TestBuildGitConfigWithEmptySection(t *testing.T) {
	t.Parallel()

	config := map[string]any{
		"user.name":  "Test User",
		"invalidkey": "value", // Key without section separator
	}

	content := buildGitConfig(config)

	// Should still contain valid entries
	if !strings.Contains(content, "[user]") {
		t.Error("Config should contain [user] section")
	}

	if !strings.Contains(content, "name = Test User") {
		t.Error("Config should contain user.name")
	}
	// Invalid key should be skipped
	if strings.Contains(content, "[invalidkey]") {
		t.Error("Config should skip keys without proper section")
	}
}

func TestBuildGitConfigWithNestedDotNotation(t *testing.T) {
	t.Parallel()

	config := map[string]any{
		"delta.decorations.commit-decoration-style": "bold yellow box ul",
		"delta.interactive.keep-plus-minus-markers": "false",
	}

	content := buildGitConfig(config)

	if !strings.Contains(content, "[delta.decorations]") {
		t.Error("Config should contain nested section")
	}

	if !strings.Contains(content, "[delta.interactive]") {
		t.Error("Config should contain nested section")
	}
}
