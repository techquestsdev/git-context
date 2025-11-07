package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aanogueira/git-context/internal/config"
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

func TestProfileToGitConfig(t *testing.T) {
	t.Parallel()

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
		Core: map[string]any{
			"editor": "vim",
		},
		Push: map[string]any{
			"default": "simple",
		},
	}

	gitConfig := profileToGitConfig(profile)

	// Verify user config
	if gitConfig["user.name"] != "Test User" {
		t.Error("Git config should contain user.name")
	}

	if gitConfig["user.email"] != "test@example.com" {
		t.Error("Git config should contain user.email")
	}

	if gitConfig["user.signingkey"] != "ABCD1234" {
		t.Error("Git config should contain user.signingkey")
	}

	// Verify URL rewrite
	urlKey := `url "ssh://git@github.com/".insteadOf`
	if gitConfig[urlKey] != "https://github.com/" {
		t.Error("Git config should contain URL rewrite")
	}

	// Verify other sections
	if gitConfig["core.editor"] != "vim" {
		t.Error("Git config should contain core.editor")
	}

	if gitConfig["push.default"] != "simple" {
		t.Error("Git config should contain push.default")
	}
}

func TestAddSectionToConfig(t *testing.T) {
	t.Parallel()

	gitConfig := make(map[string]any)

	section := map[string]any{
		"editor":   "vim",
		"autocrlf": "input",
	}

	addSectionToConfig(gitConfig, "core", section)

	if gitConfig["core.editor"] != "vim" {
		t.Error("Should add core.editor")
	}

	if gitConfig["core.autocrlf"] != "input" {
		t.Error("Should add core.autocrlf")
	}
}

func TestAddSectionToConfigNested(t *testing.T) {
	t.Parallel()

	gitConfig := make(map[string]any)

	section := map[string]any{
		"interactive": map[string]any{
			"diffFilter": "delta --color-only",
		},
	}

	addSectionToConfig(gitConfig, "add", section)

	if gitConfig["add.interactive.diffFilter"] != "delta --color-only" {
		t.Error("Should add nested config values")
	}
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

func TestProfileToGitConfigAllSections(t *testing.T) {
	t.Parallel()

	profile := &config.Profile{
		User: config.UserConfig{
			Name:       "Test User",
			Email:      "test@example.com",
			SigningKey: "KEY123",
		},
		HTTP: map[string]any{
			"postBuffer": "524288000",
		},
		Core: map[string]any{
			"editor": "vim",
		},
		Interactive: map[string]any{
			"singleKey": "true",
		},
		Add: map[string]any{
			"interactive": map[string]any{
				"useBuiltin": "false",
			},
		},
		Delta: map[string]any{
			"navigate": "true",
		},
		Push: map[string]any{
			"default": "current",
		},
		Merge: map[string]any{
			"conflictStyle": "diff3",
		},
		Commit: map[string]any{
			"gpgsign": "true",
		},
		GPG: map[string]any{
			"program": "gpg2",
		},
		Pull: map[string]any{
			"rebase": "true",
		},
		Rerere: map[string]any{
			"enabled": "true",
		},
		Column: map[string]any{
			"ui": "auto",
		},
		Branch: map[string]any{
			"autoSetupRebase": "always",
		},
		Init: map[string]any{
			"defaultBranch": "main",
		},
	}

	gitConfig := profileToGitConfig(profile)

	// Verify all sections are present
	sections := []string{
		"user.name", "user.email", "user.signingkey",
		"http.postBuffer",
		"core.editor",
		"interactive.singleKey",
		"add.interactive.useBuiltin",
		"delta.navigate",
		"push.default",
		"merge.conflictStyle",
		"commit.gpgsign",
		"gpg.program",
		"pull.rebase",
		"rerere.enabled",
		"column.ui",
		"branch.autoSetupRebase",
		"init.defaultBranch",
	}

	for _, key := range sections {
		if _, exists := gitConfig[key]; !exists {
			t.Errorf("Git config should contain key: %s", key)
		}
	}
}

func TestProfileToGitConfigEmptySections(t *testing.T) {
	t.Parallel()

	profile := &config.Profile{
		User: config.UserConfig{
			Name:  "Test",
			Email: "test@test.com",
		},
		// All other sections empty/nil
	}

	gitConfig := profileToGitConfig(profile)

	// Should have user config
	if gitConfig["user.name"] != "Test" {
		t.Error("Should have user.name")
	}

	// Other sections should not add keys
	if val, exists := gitConfig["http.something"]; exists {
		t.Errorf("Should not have http keys, got: %v", val)
	}
}

func TestAddSectionToConfigRecursiveDeepNesting(t *testing.T) {
	t.Parallel()

	gitConfig := make(map[string]any)

	values := map[string]any{
		"level1": map[string]any{
			"level2": map[string]any{
				"level3": "deepvalue",
			},
		},
	}

	addSectionToConfigRecursive(gitConfig, "section", values)

	if gitConfig["section.level1.level2.level3"] != "deepvalue" {
		t.Error("Should handle deep nesting")
	}
}
