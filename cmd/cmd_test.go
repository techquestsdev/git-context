package cmd

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/cockroachdb/errors"
	"github.com/fatih/color"
	"github.com/techquestsdev/git-context/internal/config"
	"github.com/techquestsdev/git-context/internal/git"
)

func TestRunInit(t *testing.T) {
	t.Parallel()

	// Test initialization when config doesn't exist
	t.Run("InitializeNewConfig", func(t *testing.T) {
		t.Parallel()

		// Create temp directory for test
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		// Ensure config doesn't exist
		_ = os.Remove(configFile)

		// Create and save empty config to simulate init
		cfg := config.NewConfig()

		err := cfg.SaveConfig(configFile)
		if err != nil {
			t.Fatalf("Failed to save initial config: %v", err)
		}

		// Verify file was created
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			t.Error("Init should create config file")
		}

		// Load and verify
		loadedCfg, err := config.LoadConfig(configFile)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		if len(loadedCfg.Profiles) != 0 {
			t.Error("New config should have no profiles")
		}
	})

	// Test initialization when config already exists with profiles
	t.Run("PreserveExistingConfig", func(t *testing.T) {
		t.Parallel()

		// Create temp directory for test
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		// Create config with a profile
		cfg := config.NewConfig()

		profile := &config.Profile{
			User: config.UserConfig{
				Name:  "Existing User",
				Email: "existing@example.com",
			},
		}
		if err := cfg.AddProfile("existing", profile); err != nil {
			t.Fatalf("Failed to add profile: %v", err)
		}

		err := cfg.SaveConfig(configFile)
		if err != nil {
			t.Fatalf("Failed to save config with profile: %v", err)
		}

		// Load config (simulating init on existing config)
		loadedCfg, err := config.LoadConfig(configFile)
		if err != nil {
			t.Fatalf("Failed to load existing config: %v", err)
		}

		// Verify init should preserve existing profiles
		if len(loadedCfg.Profiles) == 0 {
			t.Error("Init should preserve existing profiles")
		}

		p, err := loadedCfg.GetProfile("existing")
		if err != nil {
			t.Error("Existing profile should still be accessible")
		}

		if p.User.Name != "Existing User" {
			t.Error("Profile data should be preserved")
		}
	})
}

func TestAddProfile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Initialize config
	cfg := config.NewConfig()

	err := cfg.SaveConfig(configFile)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	t.Run("AddNewProfile", func(t *testing.T) {
		t.Parallel()

		cfg, _ := config.LoadConfig(configFile)

		profile := &config.Profile{
			User: config.UserConfig{
				Name:  "Test User",
				Email: "test@example.com",
			},
		}

		err := cfg.AddProfile("test", profile)
		if err != nil {
			t.Errorf("Failed to add profile: %v", err)
		}

		err = cfg.SaveConfig(configFile)
		if err != nil {
			t.Errorf("Failed to save config: %v", err)
		}

		// Verify profile was added
		loadedCfg, _ := config.LoadConfig(configFile)

		p, err := loadedCfg.GetProfile("test")
		if err != nil {
			t.Error("Profile should exist after adding")
		}

		if p.User.Name != "Test User" {
			t.Error("Profile data should match")
		}
	})

	t.Run("AddDuplicateProfile", func(t *testing.T) {
		t.Parallel()

		cfg, _ := config.LoadConfig(configFile)

		profile := &config.Profile{
			User: config.UserConfig{
				Name:  "Duplicate",
				Email: "duplicate@example.com",
			},
		}

		// Add first time should succeed
		err := cfg.AddProfile("duplicate", profile)
		if err != nil {
			t.Errorf("First add should succeed: %v", err)
		}

		// Add second time should fail
		err = cfg.AddProfile("duplicate", profile)
		if err == nil {
			t.Error("Adding duplicate profile should fail")
		}
	})
}

func TestRemoveProfile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Initialize config with a profile
	cfg := config.NewConfig()

	profile := &config.Profile{
		User: config.UserConfig{
			Name:  "To Remove",
			Email: "remove@example.com",
		},
	}
	if err := cfg.AddProfile("remove-test", profile); err != nil {
		t.Fatalf("Failed to add profile: %v", err)
	}

	if err := cfg.SaveConfig(configFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	t.Run("RemoveExistingProfile", func(t *testing.T) {
		t.Parallel()

		cfg, _ := config.LoadConfig(configFile)

		err := cfg.RemoveProfile("remove-test")
		if err != nil {
			t.Errorf("Failed to remove profile: %v", err)
		}

		err = cfg.SaveConfig(configFile)
		if err != nil {
			t.Errorf("Failed to save config: %v", err)
		}

		// Verify profile was removed
		loadedCfg, _ := config.LoadConfig(configFile)

		_, err = loadedCfg.GetProfile("remove-test")
		if err == nil {
			t.Error("Profile should not exist after removal")
		}
	})

	t.Run("RemoveNonExistentProfile", func(t *testing.T) {
		t.Parallel()

		cfg, _ := config.LoadConfig(configFile)

		err := cfg.RemoveProfile("nonexistent")
		if err == nil {
			t.Error("Removing non-existent profile should fail")
		}
	})
}

func TestRemoveRegeneratesAndDropsDirectories(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	paths, err := config.NewPaths()
	if err != nil {
		t.Fatalf("NewPaths: %v", err)
	}

	cfg := config.NewConfig()
	cfg.Profiles["work"] = &config.Profile{
		User:        config.UserConfig{Name: "X", Email: "x@work"},
		Directories: []string{"/tmp/work/"},
	}
	cfg.Current = "work"

	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	// Pre-generate so a stale work.gitconfig exists.
	g := git.NewGit(paths.GitConfigFile)
	if err := g.Regenerate(cfg, paths.ProfilesDir); err != nil {
		t.Fatalf("Regenerate: %v", err)
	}

	// Remove the profile (auto-confirm via removeProfileForTest helper below).
	if err := removeProfileForTest(paths, "work"); err != nil {
		t.Fatalf("removeProfileForTest: %v", err)
	}

	loaded, err := config.LoadConfig(paths.ConfigFile)
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if _, exists := loaded.Profiles["work"]; exists {
		t.Error("profile 'work' still present after remove")
	}

	if _, err := os.Stat(paths.GitConfigFile); err == nil {
		data, _ := os.ReadFile(paths.GitConfigFile)
		if strings.Contains(string(data), "/tmp/work/") {
			t.Errorf("root manifest still references removed dir:\n%s", data)
		}
	}
}

// removeProfileForTest performs the same regeneration logic as runRemove
// but without the interactive prompt — used so tests can exercise the
// post-confirmation code path.
func removeProfileForTest(paths *config.Paths, name string) error {
	cfg, err := config.LoadConfig(paths.ConfigFile)
	if err != nil {
		return errors.Wrap(err, "load config")
	}

	if err := cfg.RemoveProfile(name); err != nil {
		return errors.Wrap(err, "remove profile")
	}

	if cfg.Current == name {
		cfg.Current = ""
	}

	if cfg.Previous == name {
		cfg.Previous = ""
	}

	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		return errors.Wrap(err, "save config")
	}

	if err := git.NewGit(paths.GitConfigFile).Regenerate(cfg, paths.ProfilesDir); err != nil {
		return errors.Wrap(err, "regenerate git config")
	}

	return nil
}

func TestListProfiles(t *testing.T) {
	t.Parallel()

	t.Run("EmptyProfileList", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		cfg := config.NewConfig()
		if err := cfg.SaveConfig(configFile); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		loadedCfg, _ := config.LoadConfig(configFile)
		profiles := loadedCfg.ListProfiles()

		if len(profiles) != 0 {
			t.Errorf("Expected 0 profiles, got %d", len(profiles))
		}
	})

	t.Run("MultipleProfiles", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")

		cfg := config.NewConfig()

		profiles := []string{"work", "personal", "school"}
		for _, name := range profiles {
			if err := cfg.AddProfile(name, &config.Profile{
				User: config.UserConfig{
					Name:  name,
					Email: name + "@example.com",
				},
			}); err != nil {
				t.Fatalf("Failed to add profile: %v", err)
			}
		}

		if err := cfg.SaveConfig(configFile); err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		loadedCfg, _ := config.LoadConfig(configFile)
		listedProfiles := loadedCfg.ListProfiles()

		if len(listedProfiles) != 3 {
			t.Errorf("Expected 3 profiles, got %d", len(listedProfiles))
		}

		// Verify all profiles are present
		profileMap := make(map[string]bool)
		for _, p := range listedProfiles {
			profileMap[p] = true
		}

		for _, expected := range profiles {
			if !profileMap[expected] {
				t.Errorf("Profile '%s' not found in list", expected)
			}
		}
	})
}

func TestShowProfile(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Create config with a profile
	cfg := config.NewConfig()

	profile := &config.Profile{
		User: config.UserConfig{
			Name:       "Test User",
			Email:      "test@example.com",
			SigningKey: "ABCD1234",
		},
		URL: []config.URLConfig{
			{
				Pattern:   "ssh://git@github.com/",
				InsteadOf: "https://github.com/",
			},
		},
	}
	if err := cfg.AddProfile("test", profile); err != nil {
		t.Fatalf("Failed to add profile: %v", err)
	}

	if err := cfg.SaveConfig(configFile); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	t.Run("ShowExistingProfile", func(t *testing.T) {
		t.Parallel()

		loadedCfg, _ := config.LoadConfig(configFile)

		p, err := loadedCfg.GetProfile("test")
		if err != nil {
			t.Errorf("Failed to get profile: %v", err)
		}

		if p.User.Name != "Test User" {
			t.Error("Profile name should match")
		}

		if p.User.Email != "test@example.com" {
			t.Error("Profile email should match")
		}

		if p.User.SigningKey != "ABCD1234" {
			t.Error("Profile signing key should match")
		}

		if len(p.URL) != 1 {
			t.Error("Profile should have 1 URL rewrite")
		}
	})

	t.Run("ShowNonExistentProfile", func(t *testing.T) {
		t.Parallel()

		loadedCfg, _ := config.LoadConfig(configFile)

		_, err := loadedCfg.GetProfile("nonexistent")
		if err == nil {
			t.Error("Getting non-existent profile should fail")
		}
	})
}

func TestInitCommandExists(t *testing.T) {
	t.Parallel()

	// Verify init command is registered
	found := false

	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "init" {
			found = true

			break
		}
	}

	if !found {
		t.Error("init command should be registered")
	}
}

func TestListProfilesShowsDirsColumn(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	paths, _ := config.NewPaths()

	cfg := config.NewConfig()
	cfg.Profiles["work"] = &config.Profile{
		User:        config.UserConfig{Email: "x@work"},
		Directories: []string{"/a/", "/b/"},
	}
	cfg.Profiles["personal"] = &config.Profile{User: config.UserConfig{Email: "x@home"}}

	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	out := captureStdout(t, func() {
		if err := runList(nil, nil); err != nil {
			t.Fatalf("runList: %v", err)
		}
	})

	if !strings.Contains(out, "Dirs") {
		t.Errorf("output missing Dirs header:\n%s", out)
	}

	if !strings.Contains(out, "2") {
		t.Errorf("output missing dir count of 2 for work:\n%s", out)
	}
}

// stdoutMu serializes the os.Stdout swap performed by captureStdout so
// concurrent stdout-capturing tests don't race on the global. Tests that
// use captureStdout must NOT call t.Parallel() — t.Setenv is also commonly
// used and is itself incompatible with parallel.
var stdoutMu sync.Mutex

// captureStdout runs fn and returns whatever it wrote to os.Stdout.
// IMPORTANT: do not call from a t.Parallel() test — this swaps the
// global os.Stdout and serializes via stdoutMu. Concurrent capturers
// would still produce correct values because of the mutex, but other
// non-capturing parallel tests printing to stdout during fn will have
// their output silently captured, which is rarely what you want.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	stdoutMu.Lock()
	defer stdoutMu.Unlock()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	old := os.Stdout
	// fatih/color caches its output writer at init, so callers like
	// ui.PrintInfo bypass our os.Stdout swap unless we redirect here too.
	oldColor := color.Output
	os.Stdout = w
	color.Output = w

	defer func() {
		os.Stdout = old
		color.Output = oldColor
	}()

	done := make(chan string)

	go func() {
		var buf bytes.Buffer

		_, _ = buf.ReadFrom(r)
		done <- buf.String()
	}()

	fn()

	_ = w.Close()

	return <-done
}

func TestRootCommandMetadata(t *testing.T) {
	t.Parallel()

	if rootCmd.Use != "git-context" {
		t.Errorf("Expected Use 'git-context', got '%s'", rootCmd.Use)
	}

	if rootCmd.Version == "" {
		t.Error("Version should be set")
	}

	if rootCmd.Short == "" {
		t.Error("Short description should be set")
	}
}

func TestShowDisplaysDirectories(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	paths, _ := config.NewPaths()

	cfg := config.NewConfig()
	cfg.Profiles["work"] = &config.Profile{
		User:        config.UserConfig{Name: "X", Email: "x@work"},
		Directories: []string{"/Users/x/work/", "/Users/x/Mollie/"},
	}

	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	out := captureStdout(t, func() {
		if err := runShow(nil, []string{"work"}); err != nil {
			t.Fatalf("runShow: %v", err)
		}
	})

	if !strings.Contains(out, "Directories") {
		t.Errorf("output missing Directories label:\n%s", out)
	}

	if !strings.Contains(out, "/Users/x/work/") {
		t.Errorf("output missing assigned dir:\n%s", out)
	}

	if !strings.Contains(out, "/Users/x/Mollie/") {
		t.Errorf("output missing assigned dir:\n%s", out)
	}
}

func TestCurrentShowsEffectiveProfileInCwd(t *testing.T) {
	skipOnWindows(t)

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	paths, _ := config.NewPaths()

	cfg := config.NewConfig()
	cfg.Profiles["work"] = &config.Profile{
		User:        config.UserConfig{Name: "Andre", Email: "a@work.com"},
		Directories: []string{},
	}
	cfg.Profiles["personal"] = &config.Profile{
		User:        config.UserConfig{Name: "Andre", Email: "a@home.com"},
		Directories: []string{},
	}
	cfg.Current = "work"

	repoDir := filepath.Join(tmpHome, "personal-repo")
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	if out, err := exec.Command("git", "-C", repoDir, "init").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v\n%s", err, out)
	}

	cfg.Profiles["personal"].Directories = []string{repoDir + "/"}

	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	g := git.NewGit(paths.GitConfigFile)
	if err := g.Regenerate(cfg, paths.ProfilesDir); err != nil {
		t.Fatalf("Regenerate: %v", err)
	}

	t.Chdir(repoDir)

	out := captureStdout(t, func() {
		if err := runCurrent(nil, nil); err != nil {
			t.Fatalf("runCurrent: %v", err)
		}
	})

	if !strings.Contains(out, "Effective in") {
		t.Errorf("output missing 'Effective in' line:\n%s", out)
	}

	if !strings.Contains(out, "personal") {
		t.Errorf("expected 'personal' to be effective in this dir:\n%s", out)
	}
}

func TestEffectiveProfileInCwdRejectsNonFileOrigin(t *testing.T) {
	// Hard to fake `git config --show-origin` without actually running git,
	// so this test serves as a placeholder asserting the function returns
	// "" when no git command is available (i.e. when we're in a directory
	// where the git invocation fails). The richer behavior is exercised by
	// TestCurrentShowsEffectiveProfileInCwd.
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}

	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	// Run from a temp dir that is NOT a git repo. git config will fail.
	t.Chdir(tmpHome)

	paths, _ := config.NewPaths()

	cfg := config.NewConfig()
	cfg.Profiles["work"] = &config.Profile{User: config.UserConfig{Name: "X"}}

	if got := effectiveProfileInCwd(paths, cfg); got != "" {
		t.Errorf("effectiveProfileInCwd outside a repo = %q, want empty", got)
	}
}
