package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/techquestsdev/git-context/internal/config"
)

// TestEndToEndDirectoryAssignment exercises the full lifecycle: switch to a
// default profile, assign a directory to a different profile, verify that
// inside the assigned directory git resolves the right user.email.
func TestEndToEndDirectoryAssignment(t *testing.T) {
	skipOnWindows(t)

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	// Resolve symlinks because git resolves gitdir to its real path; on macOS
	// t.TempDir() lives under /var/folders which is a symlink to /private/var.
	tmpHome, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("EvalSymlinks: %v", err)
	}

	t.Setenv("HOME", tmpHome)

	paths, err := config.NewPaths()
	if err != nil {
		t.Fatalf("NewPaths: %v", err)
	}

	// Set up two profiles in config.yaml.
	cfg := config.NewConfig()
	cfg.Profiles["work"] = &config.Profile{
		User: config.UserConfig{Name: "Andre", Email: "a@work.com"},
	}
	cfg.Profiles["personal"] = &config.Profile{
		User: config.UserConfig{Name: "Andre", Email: "a@home.com"},
	}

	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	// Switch makes 'work' the default and regenerates.
	if err := runSwitch(nil, []string{"work"}); err != nil {
		t.Fatalf("runSwitch: %v", err)
	}

	// Create a real git repo and assign it to 'personal'.
	repo := filepath.Join(tmpHome, "personal-repo")
	if err := os.MkdirAll(repo, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	if out, err := exec.Command("git", "-C", repo, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	if err := runDirAdd(nil, []string{repo, "personal"}); err != nil {
		t.Fatalf("runDirAdd: %v", err)
	}

	// Outside the repo (use HOME), default 'work' should be effective.
	out, err := exec.Command("git", "-C", tmpHome, "config", "user.email").Output()
	if err != nil {
		t.Fatalf("git config user.email outside: %v", err)
	}

	if got := strings.TrimSpace(string(out)); got != "a@work.com" {
		t.Errorf("default user.email = %q, want a@work.com", got)
	}

	// Inside the repo, 'personal' should override.
	out, err = exec.Command("git", "-C", repo, "config", "user.email").Output()
	if err != nil {
		t.Fatalf("git config user.email inside: %v", err)
	}

	if got := strings.TrimSpace(string(out)); got != "a@home.com" {
		t.Errorf("override user.email = %q, want a@home.com", got)
	}

	// dir remove undoes the override.
	if err := runDirRemove(nil, []string{repo}); err != nil {
		t.Fatalf("runDirRemove: %v", err)
	}

	out, err = exec.Command("git", "-C", repo, "config", "user.email").Output()
	if err != nil {
		t.Fatalf("git config user.email after remove: %v", err)
	}

	if got := strings.TrimSpace(string(out)); got != "a@work.com" {
		t.Errorf("after dir remove, user.email = %q, want a@work.com", got)
	}

	// Sanity: profile files exist on disk.
	for _, name := range []string{"work", "personal"} {
		if _, err := os.Stat(filepath.Join(paths.ProfilesDir, name+".gitconfig")); err != nil {
			t.Errorf("expected %s.gitconfig: %v", name, err)
		}
	}

	// Sanity: per-profile file content matches the profile.
	workFile := filepath.Join(paths.ProfilesDir, "work.gitconfig")

	data, _ := os.ReadFile(workFile)
	if !strings.Contains(string(data), "a@work.com") {
		t.Errorf("work.gitconfig missing email:\n%s", data)
	}
}
