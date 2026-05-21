package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/techquestsdev/git-context/internal/config"
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

func TestWriteProfileFile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "work.gitconfig")

	g := NewGit(filepath.Join(tmpDir, ".gitconfig"))

	settings := map[string]any{
		"user.name":  "Andre",
		"user.email": "andre@work.com",
	}

	if err := g.WriteProfileFile(target, settings); err != nil {
		t.Fatalf("WriteProfileFile error: %v", err)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "[user]") {
		t.Errorf("missing [user] section in:\n%s", content)
	}

	if !strings.Contains(content, "name = Andre") {
		t.Errorf("missing user.name in:\n%s", content)
	}

	if !strings.Contains(content, "email = andre@work.com") {
		t.Errorf("missing user.email in:\n%s", content)
	}
}

func TestWriteProfileFileAtomic(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "work.gitconfig")

	// Pre-create the file with old content to verify replace, not append.
	if err := os.WriteFile(target, []byte("OLD\n"), 0o644); err != nil {
		t.Fatalf("setup write failed: %v", err)
	}

	g := NewGit(filepath.Join(tmpDir, ".gitconfig"))

	if err := g.WriteProfileFile(target, map[string]any{"user.name": "New"}); err != nil {
		t.Fatalf("WriteProfileFile error: %v", err)
	}

	data, _ := os.ReadFile(target)
	if strings.Contains(string(data), "OLD") {
		t.Errorf("old content not replaced:\n%s", data)
	}

	// No leftover .tmp file.
	matches, _ := filepath.Glob(filepath.Join(tmpDir, "*.tmp"))
	if len(matches) > 0 {
		t.Errorf("temp files left behind: %v", matches)
	}
}

func TestWriteProfileFileWriteFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "missing-subdir", "work.gitconfig")

	g := NewGit(filepath.Join(tmpDir, ".gitconfig"))

	err := g.WriteProfileFile(target, map[string]any{"user.name": "X"})
	if err == nil {
		t.Fatal("expected error writing into nonexistent directory, got nil")
	}

	if !strings.Contains(err.Error(), "failed to write temp file") {
		t.Errorf("error = %q, want it to wrap 'failed to write temp file'", err.Error())
	}
}

func TestWriteRootConfigDefaultAndAssignments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	rootPath := filepath.Join(tmpDir, ".gitconfig")
	g := NewGit(rootPath)

	defaultProfilePath := filepath.Join(tmpDir, "profiles", "work.gitconfig")
	assignments := map[string]string{
		"/Users/x/projects/personal/": filepath.Join(tmpDir, "profiles", "personal.gitconfig"),
		"/Users/x/Mollie/":            filepath.Join(tmpDir, "profiles", "work.gitconfig"),
	}

	if err := g.WriteRootConfig(defaultProfilePath, assignments); err != nil {
		t.Fatalf("WriteRootConfig error: %v", err)
	}

	data, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "Generated by git-context") {
		t.Errorf("missing header marker:\n%s", content)
	}

	if !strings.Contains(content, "[include]") {
		t.Errorf("missing [include] block:\n%s", content)
	}

	if !strings.Contains(content, "path = "+toGitPath(defaultProfilePath)) {
		t.Errorf("missing default profile path:\n%s", content)
	}

	if !strings.Contains(content, `[includeIf "gitdir:/Users/x/Mollie/"]`) {
		t.Errorf("missing Mollie includeIf:\n%s", content)
	}

	if !strings.Contains(content, `[includeIf "gitdir:/Users/x/projects/personal/"]`) {
		t.Errorf("missing personal includeIf:\n%s", content)
	}
}

func TestWriteRootConfigDeterministicOrder(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	rootPath := filepath.Join(tmpDir, ".gitconfig")
	g := NewGit(rootPath)

	assignments := map[string]string{
		"/zzz/": "/profiles/c.gitconfig",
		"/aaa/": "/profiles/a.gitconfig",
		"/mmm/": "/profiles/b.gitconfig",
	}

	if err := g.WriteRootConfig("", assignments); err != nil {
		t.Fatalf("WriteRootConfig error: %v", err)
	}

	data, _ := os.ReadFile(rootPath)
	content := string(data)

	idxA := strings.Index(content, "gitdir:/aaa/")
	idxM := strings.Index(content, "gitdir:/mmm/")
	idxZ := strings.Index(content, "gitdir:/zzz/")

	if idxA >= idxM || idxM >= idxZ {
		t.Errorf("includeIf blocks not sorted: aaa=%d mmm=%d zzz=%d", idxA, idxM, idxZ)
	}
}

func TestWriteRootConfigNoDefaultNoAssignments(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	rootPath := filepath.Join(tmpDir, ".gitconfig")
	g := NewGit(rootPath)

	if err := g.WriteRootConfig("", map[string]string{}); err != nil {
		t.Fatalf("WriteRootConfig error: %v", err)
	}

	if _, err := os.Stat(rootPath); !os.IsNotExist(err) {
		t.Errorf("expected no file written, got err=%v", err)
	}
}

func TestWriteRootConfigParseableByGit(t *testing.T) {
	t.Parallel()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	tmpDir := t.TempDir()
	rootPath := filepath.Join(tmpDir, ".gitconfig")
	g := NewGit(rootPath)

	defaultPath := filepath.Join(tmpDir, "default.gitconfig")
	assignments := map[string]string{"/Users/x/work/": filepath.Join(tmpDir, "work.gitconfig")}

	if err := g.WriteRootConfig(defaultPath, assignments); err != nil {
		t.Fatalf("WriteRootConfig error: %v", err)
	}

	cmd := exec.Command("git", "config", "-f", rootPath, "--list")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git config rejected file: %v\n%s", err, out)
	}
}

func TestRegenerate(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	profilesDir := filepath.Join(tmpDir, "profiles")

	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatalf("mkdir profiles: %v", err)
	}

	cfg := config.NewConfig()
	cfg.Profiles["work"] = &config.Profile{
		User:        config.UserConfig{Name: "Andre", Email: "a@work.com"},
		Directories: []string{"/Users/x/Mollie/"},
	}
	cfg.Profiles["personal"] = &config.Profile{
		User:        config.UserConfig{Name: "Andre", Email: "a@home.com"},
		Directories: []string{"/Users/x/projects/personal/"},
	}
	cfg.Current = "work"

	rootPath := filepath.Join(tmpDir, ".gitconfig")
	g := NewGit(rootPath)

	if err := g.Regenerate(cfg, profilesDir); err != nil {
		t.Fatalf("Regenerate error: %v", err)
	}

	// Per-profile files exist with expected content.
	for name, wantEmail := range map[string]string{"work": "a@work.com", "personal": "a@home.com"} {
		path := filepath.Join(profilesDir, name+".gitconfig")

		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}

		if !strings.Contains(string(data), "email = "+wantEmail) {
			t.Errorf("%s missing email=%s\n%s", name, wantEmail, data)
		}
	}

	// Root manifest references the right files.
	root, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("read root: %v", err)
	}

	rootStr := string(root)
	wantInclude := toGitPath(filepath.Join(profilesDir, "work.gitconfig"))

	if !strings.Contains(rootStr, "path = "+wantInclude) {
		t.Errorf("root missing default include for %q:\n%s", wantInclude, rootStr)
	}

	if !strings.Contains(rootStr, `gitdir:/Users/x/Mollie/`) {
		t.Errorf("root missing Mollie includeIf:\n%s", rootStr)
	}

	if !strings.Contains(rootStr, `gitdir:/Users/x/projects/personal/`) {
		t.Errorf("root missing personal includeIf:\n%s", rootStr)
	}
}

func TestRegenerateNoOpWhenNothingToWrite(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	rootPath := filepath.Join(tmpDir, ".gitconfig")
	profilesDir := filepath.Join(tmpDir, "profiles")

	if err := os.MkdirAll(profilesDir, 0o755); err != nil {
		t.Fatalf("mkdir profiles: %v", err)
	}

	cfg := config.NewConfig()
	cfg.Profiles["work"] = &config.Profile{User: config.UserConfig{Name: "X"}}
	// No Current, no Directories — nothing to write to root manifest.

	g := NewGit(rootPath)

	if err := g.Regenerate(cfg, profilesDir); err != nil {
		t.Fatalf("Regenerate error: %v", err)
	}

	if _, err := os.Stat(rootPath); !os.IsNotExist(err) {
		t.Errorf("expected no root manifest, got err=%v", err)
	}

	// Per-profile file still gets written, regardless of whether root manifest is.
	if _, err := os.Stat(filepath.Join(profilesDir, "work.gitconfig")); err != nil {
		t.Errorf("expected work.gitconfig to exist: %v", err)
	}
}

func TestBackupConfigSkipsGeneratedManifest(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, ".gitconfig")
	dst := filepath.Join(tmpDir, ".gitconfig.bak")

	// Simulate a git-context-generated source by writing the header marker.
	manifest := rootConfigHeader + "\n[include]\n\tpath = /tmp/x\n"
	if err := os.WriteFile(src, []byte(manifest), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	g := NewGit(src)

	if err := g.BackupConfig(dst); err != nil {
		t.Fatalf("BackupConfig error: %v", err)
	}

	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Errorf("expected NO backup file when source is a generated manifest, got err=%v", err)
	}
}

func TestBackupConfigCopiesUserContent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	src := filepath.Join(tmpDir, ".gitconfig")
	dst := filepath.Join(tmpDir, ".gitconfig.bak")

	if err := os.WriteFile(src, []byte("[user]\n\tname = Old\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	g := NewGit(src)

	if err := g.BackupConfig(dst); err != nil {
		t.Fatalf("BackupConfig error: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}

	if !strings.Contains(string(data), "name = Old") {
		t.Errorf("backup missing original content:\n%s", data)
	}
}

func TestProfileToFlatConfigCoreFields(t *testing.T) {
	t.Parallel()

	p := &config.Profile{
		User: config.UserConfig{
			Name:       "Andre",
			Email:      "a@x.com",
			SigningKey: "ABC123",
		},
	}

	got := profileToFlatConfig(p)

	if got["user.name"] != "Andre" {
		t.Errorf("user.name = %v, want Andre", got["user.name"])
	}

	if got["user.email"] != "a@x.com" {
		t.Errorf("user.email = %v, want a@x.com", got["user.email"])
	}

	if got["user.signingkey"] != "ABC123" {
		t.Errorf("user.signingkey = %v, want ABC123", got["user.signingkey"])
	}
}

func TestProfileToFlatConfigOmitsEmptyUserFields(t *testing.T) {
	t.Parallel()

	p := &config.Profile{User: config.UserConfig{Name: "Andre"}}
	got := profileToFlatConfig(p)

	if _, ok := got["user.email"]; ok {
		t.Error("user.email should not be set when empty")
	}

	if _, ok := got["user.signingkey"]; ok {
		t.Error("user.signingkey should not be set when empty")
	}
}

func TestProfileToFlatConfigURLRewrites(t *testing.T) {
	t.Parallel()

	p := &config.Profile{
		URL: []config.URLConfig{
			{Pattern: "ssh://git@github.com/", InsteadOf: "https://github.com/"},
		},
	}

	got := profileToFlatConfig(p)
	key := `url "ssh://git@github.com/".insteadOf`

	if got[key] != "https://github.com/" {
		t.Errorf("URL rewrite key %q = %v, want https://github.com/", key, got[key])
	}
}

func TestProfileToFlatConfigNestedSection(t *testing.T) {
	t.Parallel()

	p := &config.Profile{
		Delta: map[string]any{
			"decorations": map[string]any{
				"file-style": "bold yellow",
			},
		},
	}

	got := profileToFlatConfig(p)

	if got["delta.decorations.file-style"] != "bold yellow" {
		t.Errorf("delta.decorations.file-style = %v, want 'bold yellow'\nfull map: %#v",
			got["delta.decorations.file-style"], got)
	}
}

func TestFlattenIntoLeafValues(t *testing.T) {
	t.Parallel()

	out := make(map[string]any)
	flattenInto(out, "core", map[string]any{
		"editor":   "nvim",
		"autocrlf": "input",
	})

	if out["core.editor"] != "nvim" {
		t.Errorf("core.editor = %v", out["core.editor"])
	}

	if out["core.autocrlf"] != "input" {
		t.Errorf("core.autocrlf = %v", out["core.autocrlf"])
	}
}

func TestFlattenIntoRecursesNestedMaps(t *testing.T) {
	t.Parallel()

	out := make(map[string]any)
	flattenInto(out, "delta", map[string]any{
		"decorations": map[string]any{
			"hunk-header-style": "syntax bold",
		},
		"interactive": map[string]any{
			"keep-plus-minus-markers": true,
		},
	})

	if out["delta.decorations.hunk-header-style"] != "syntax bold" {
		t.Errorf("missing nested key, got map: %#v", out)
	}

	if out["delta.interactive.keep-plus-minus-markers"] != true {
		t.Errorf("missing nested bool, got map: %#v", out)
	}
}

func TestWriteRootConfigNormalizesBackslashes(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	rootPath := filepath.Join(tmpDir, ".gitconfig")
	g := NewGit(rootPath)

	// Simulate Windows-style absolute paths even on POSIX hosts.
	defaultPath := `C:\Users\runner\AppData\profiles\work.gitconfig`
	assignments := map[string]string{
		`C:\Users\runner\Mollie\`: `C:\Users\runner\AppData\profiles\work.gitconfig`,
	}

	if err := g.WriteRootConfig(defaultPath, assignments); err != nil {
		t.Fatalf("WriteRootConfig error: %v", err)
	}

	data, err := os.ReadFile(rootPath)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}

	content := string(data)

	if strings.Contains(content, `\`) {
		t.Errorf("manifest contains backslash; git config would reject:\n%s", content)
	}

	if !strings.Contains(content, "C:/Users/runner/AppData/profiles/work.gitconfig") {
		t.Errorf("path not slash-normalized:\n%s", content)
	}

	if !strings.Contains(content, `gitdir:C:/Users/runner/Mollie/`) {
		t.Errorf("gitdir not slash-normalized:\n%s", content)
	}
}
