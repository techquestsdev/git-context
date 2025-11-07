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

func TestReadConfigNonExistent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent.gitconfig")

	g := NewGit(configPath)

	config, err := g.ReadConfig()
	if err != nil {
		t.Errorf("ReadConfig should not error for non-existent file: %v", err)
	}

	if config == nil {
		t.Fatal("ReadConfig should return empty map for non-existent file")
	}

	if len(config) != 0 {
		t.Errorf("Expected empty config, got %d entries", len(config))
	}
}

func TestWriteAndReadConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.gitconfig")

	g := NewGit(configPath)

	// Write config
	config := map[string]any{
		"user.name":   "Test User",
		"user.email":  "test@example.com",
		"core.editor": "vim",
	}

	err := g.WriteConfig(config)
	if err != nil {
		t.Fatalf("WriteConfig failed: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Read config back
	readConfig, err := g.ReadConfig()
	if err != nil {
		t.Fatalf("ReadConfig failed: %v", err)
	}

	// Verify values
	if readConfig["user.name"] != "Test User" {
		t.Errorf("Expected user.name 'Test User', got '%s'", readConfig["user.name"])
	}

	if readConfig["user.email"] != "test@example.com" {
		t.Errorf("Expected user.email 'test@example.com', got '%s'", readConfig["user.email"])
	}

	if readConfig["core.editor"] != "vim" {
		t.Errorf("Expected core.editor 'vim', got '%s'", readConfig["core.editor"])
	}
}

func TestBackupAndRestoreConfig(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.gitconfig")
	backupPath := filepath.Join(tmpDir, "test.gitconfig.bak")

	g := NewGit(configPath)

	// Create initial config
	originalContent := "[user]\n\tname = Original User\n\temail = original@example.com\n"

	err := os.WriteFile(configPath, []byte(originalContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Backup config
	err = g.BackupConfig(backupPath)
	if err != nil {
		t.Fatalf("BackupConfig failed: %v", err)
	}

	// Check backup was created
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Fatal("Backup file was not created")
	}

	// Modify original config
	modifiedContent := "[user]\n\tname = Modified User\n\temail = modified@example.com\n"

	err = os.WriteFile(configPath, []byte(modifiedContent), 0o644)
	if err != nil {
		t.Fatalf("Failed to modify config: %v", err)
	}

	// Restore from backup
	err = g.RestoreConfig(backupPath)
	if err != nil {
		t.Fatalf("RestoreConfig failed: %v", err)
	}

	// Verify restoration
	restoredContent, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Failed to read restored config: %v", err)
	}

	if string(restoredContent) != originalContent {
		t.Error("Restored content does not match original")
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

func TestRestoreConfigNonExistent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.gitconfig")
	backupPath := filepath.Join(tmpDir, "nonexistent.backup")

	g := NewGit(configPath)

	// Restoring from non-existent backup should error
	err := g.RestoreConfig(backupPath)
	if err == nil {
		t.Error("RestoreConfig should error for non-existent backup")
	}
}

func TestParseGitConfig(t *testing.T) {
	t.Parallel()

	content := `# Comment line
[user]
	name = Test User
	email = test@example.com

[core]
	editor = vim
	autocrlf = input

[url "ssh://git@github.com/"]
	insteadOf = https://github.com/
`

	config := parseGitConfig(content)

	if config["user.name"] != "Test User" {
		t.Errorf("Expected user.name 'Test User', got '%s'", config["user.name"])
	}

	if config["user.email"] != "test@example.com" {
		t.Errorf("Expected user.email 'test@example.com', got '%s'", config["user.email"])
	}

	if config["core.editor"] != "vim" {
		t.Errorf("Expected core.editor 'vim', got '%s'", config["core.editor"])
	}

	if config["core.autocrlf"] != "input" {
		t.Errorf("Expected core.autocrlf 'input', got '%s'", config["core.autocrlf"])
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

func TestParseGitConfigSkipsComments(t *testing.T) {
	t.Parallel()

	content := `# This is a comment
; This is also a comment
[user]
	# Inline comment
	name = Test User
	; Another inline comment
`

	config := parseGitConfig(content)

	if config["user.name"] != "Test User" {
		t.Errorf("Expected user.name 'Test User', got '%s'", config["user.name"])
	}

	// Should only have one entry
	if len(config) != 1 {
		t.Errorf("Expected 1 config entry, got %d", len(config))
	}
}

func TestParseGitConfigHandlesEmptyLines(t *testing.T) {
	t.Parallel()

	content := `

[user]

	name = Test User

	email = test@example.com

`

	config := parseGitConfig(content)

	if config["user.name"] != "Test User" {
		t.Errorf("Expected user.name 'Test User', got '%s'", config["user.name"])
	}

	if config["user.email"] != "test@example.com" {
		t.Errorf("Expected user.email 'test@example.com', got '%s'", config["user.email"])
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

func TestReadConfigError(t *testing.T) {
	t.Parallel()

	// Try to read from a directory instead of a file
	tmpDir := t.TempDir()

	g := NewGit(tmpDir) // tmpDir is a directory, not a file

	_, err := g.ReadConfig()
	if err == nil {
		t.Error("ReadConfig should fail when path is a directory")
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

func TestRestoreConfigError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.gitconfig")

	g := NewGit(configPath)

	// Try to restore when backup doesn't exist
	err := g.RestoreConfig(filepath.Join(tmpDir, "nonexistent.bak"))
	if err == nil {
		t.Error("RestoreConfig should fail when backup doesn't exist")
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
