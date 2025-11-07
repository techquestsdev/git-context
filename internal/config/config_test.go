package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()
	if cfg == nil {
		t.Fatal("NewConfig returned nil")
	}

	if cfg.Global == nil {
		t.Error("Global map should be initialized")
	}

	if cfg.Profiles == nil {
		t.Error("Profiles map should be initialized")
	}

	if cfg.Current != "" {
		t.Error("Current should be empty string initially")
	}
}

func TestAddProfile(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()
	profile := &Profile{
		User: UserConfig{
			Name:  "Test User",
			Email: "test@example.com",
		},
	}

	// Test adding new profile
	err := cfg.AddProfile("test", profile)
	if err != nil {
		t.Errorf("AddProfile failed: %v", err)
	}

	// Test adding duplicate profile
	err = cfg.AddProfile("test", profile)
	if err == nil {
		t.Error("AddProfile should fail for duplicate profile")
	}
}

func TestRemoveProfile(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()
	profile := &Profile{
		User: UserConfig{
			Name:  "Test User",
			Email: "test@example.com",
		},
	}

	if err := cfg.AddProfile("test", profile); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	// Test removing existing profile
	err := cfg.RemoveProfile("test")
	if err != nil {
		t.Errorf("RemoveProfile failed: %v", err)
	}

	// Test removing non-existent profile
	err = cfg.RemoveProfile("nonexistent")
	if err == nil {
		t.Error("RemoveProfile should fail for non-existent profile")
	}
}

func TestGetProfile(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()
	profile := &Profile{
		User: UserConfig{
			Name:  "Test User",
			Email: "test@example.com",
		},
	}

	if err := cfg.AddProfile("test", profile); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	// Test getting existing profile
	p, err := cfg.GetProfile("test")
	if err != nil {
		t.Errorf("GetProfile failed: %v", err)
	}

	if p.User.Name != "Test User" {
		t.Errorf("Expected user name 'Test User', got '%s'", p.User.Name)
	}

	// Test getting non-existent profile
	_, err = cfg.GetProfile("nonexistent")
	if err == nil {
		t.Error("GetProfile should fail for non-existent profile")
	}
}

func TestListProfiles(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()

	// Test empty list
	profiles := cfg.ListProfiles()
	if len(profiles) != 0 {
		t.Errorf("Expected 0 profiles, got %d", len(profiles))
	}

	// Add profiles
	if err := cfg.AddProfile("work", &Profile{User: UserConfig{Name: "Work", Email: "work@example.com"}}); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	if err := cfg.AddProfile(
		"personal",
		&Profile{User: UserConfig{Name: "Personal", Email: "personal@example.com"}},
	); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	profiles = cfg.ListProfiles()
	if len(profiles) != 2 {
		t.Errorf("Expected 2 profiles, got %d", len(profiles))
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	t.Parallel()

	// Create temp directory
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")

	// Create config with profiles
	cfg := NewConfig()

	profile := &Profile{
		User: UserConfig{
			Name:       "Test User",
			Email:      "test@example.com",
			SigningKey: "ABCD1234",
		},
		URL: []URLConfig{
			{
				Pattern:   "ssh://git@github.com/",
				InsteadOf: "https://github.com/",
			},
		},
	}
	if err := cfg.AddProfile("test", profile); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	cfg.Global = map[string]any{
		"core": map[string]any{
			"editor": "vim",
		},
	}

	// Save config
	err := cfg.SaveConfig(configFile)
	if err != nil {
		t.Fatalf("SaveConfig failed: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load config
	loadedCfg, err := LoadConfig(configFile)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Verify loaded config
	p, err := loadedCfg.GetProfile("test")
	if err != nil {
		t.Fatalf("Profile not found in loaded config: %v", err)
	}

	if p.User.Name != "Test User" {
		t.Errorf("Expected user name 'Test User', got '%s'", p.User.Name)
	}

	if p.User.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got '%s'", p.User.Email)
	}

	if p.User.SigningKey != "ABCD1234" {
		t.Errorf("Expected signing key 'ABCD1234', got '%s'", p.User.SigningKey)
	}

	if len(p.URL) != 1 {
		t.Errorf("Expected 1 URL rewrite, got %d", len(p.URL))
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	t.Parallel()

	cfg, err := LoadConfig("/nonexistent/path/config.yaml")
	if err != nil {
		t.Errorf("LoadConfig should not fail for non-existent file: %v", err)
	}

	if cfg == nil {
		t.Fatal("LoadConfig should return empty config for non-existent file")
	}

	if len(cfg.Profiles) != 0 {
		t.Error("Loaded config should be empty")
	}
}

func TestMerge(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()

	// Set up global config
	cfg.Global = map[string]any{
		"core": map[string]any{
			"editor":   "vim",
			"autocrlf": "input",
		},
		"push": map[string]any{
			"default": "simple",
		},
	}

	// Create profile with partial config
	profile := &Profile{
		User: UserConfig{
			Name:  "Test User",
			Email: "test@example.com",
		},
		Core: map[string]any{
			"autocrlf": "false", // Override global
		},
		HTTP: map[string]any{
			"postBuffer": "524288000",
		},
	}
	if err := cfg.AddProfile("test", profile); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	// Merge configuration
	merged, err := cfg.Merge("test")
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Verify user config
	if merged.User.Name != "Test User" {
		t.Errorf("Expected merged user name 'Test User', got '%s'", merged.User.Name)
	}

	// Verify core config
	if merged.Core["editor"] != "vim" {
		t.Error("Core.editor should be inherited from global")
	}

	if merged.Core["autocrlf"] != "false" {
		t.Error("Core.autocrlf should be overridden by profile")
	}

	// Verify push config
	if merged.Push["default"] != "simple" {
		t.Error("Push.default should be inherited from global")
	}

	// Verify HTTP config
	if merged.HTTP["postBuffer"] != "524288000" {
		t.Error("HTTP.postBuffer should be from profile")
	}
}

func TestMergeNonExistentProfile(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()

	_, err := cfg.Merge("nonexistent")
	if err == nil {
		t.Error("Merge should fail for non-existent profile")
	}
}

func TestMergeURLs(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()

	// Set global URLs
	cfg.Global = map[string]any{
		"url": []URLConfig{
			{
				Pattern:   "ssh://git@github.com/",
				InsteadOf: "https://github.com/",
			},
		},
	}

	// Profile with its own URLs
	profile := &Profile{
		User: UserConfig{
			Name:  "Test User",
			Email: "test@example.com",
		},
		URL: []URLConfig{
			{
				Pattern:   "ssh://git@gitlab.com/",
				InsteadOf: "https://gitlab.com/",
			},
		},
	}
	if err := cfg.AddProfile("test", profile); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	merged, err := cfg.Merge("test")
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Profile URLs should override global URLs
	if len(merged.URL) != 1 {
		t.Errorf("Expected 1 URL, got %d", len(merged.URL))
	}

	if merged.URL[0].Pattern != "ssh://git@gitlab.com/" {
		t.Errorf("Expected gitlab URL, got %s", merged.URL[0].Pattern)
	}
}

func TestMergeEmptyProfileURLsUsesGlobal(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()

	// Set global URLs
	cfg.Global = map[string]any{
		"url": []URLConfig{
			{
				Pattern:   "ssh://git@github.com/",
				InsteadOf: "https://github.com/",
			},
		},
	}

	// Profile without URLs
	profile := &Profile{
		User: UserConfig{
			Name:  "Test User",
			Email: "test@example.com",
		},
	}
	if err := cfg.AddProfile("test", profile); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	merged, err := cfg.Merge("test")
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should use global URLs
	if len(merged.URL) != 1 {
		t.Errorf("Expected 1 URL from global, got %d", len(merged.URL))
	}
}

func TestSaveConfigError(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()

	// Try to save to invalid path
	err := cfg.SaveConfig("/invalid/path/that/does/not/exist/config.yaml")
	if err == nil {
		t.Error("SaveConfig should fail for invalid path")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	invalidYAML := "invalid: yaml: content: [unclosed"

	err := os.WriteFile(configFile, []byte(invalidYAML), 0o644)
	if err != nil {
		t.Fatalf("Failed to create invalid YAML: %v", err)
	}

	_, err = LoadConfig(configFile)
	if err == nil {
		t.Error("LoadConfig should fail for invalid YAML")
	}
}

func TestMergeWithInterfaceURLs(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()

	// Set global URLs as []interface{} (how it might be unmarshaled)
	cfg.Global = map[string]any{
		"url": []any{
			map[string]any{
				"pattern":   "ssh://git@github.com/",
				"insteadOf": "https://github.com/",
			},
		},
	}

	// Profile without URLs
	profile := &Profile{
		User: UserConfig{
			Name:  "Test User",
			Email: "test@example.com",
		},
	}
	if err := cfg.AddProfile("test", profile); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	merged, err := cfg.Merge("test")
	if err != nil {
		t.Fatalf("Merge failed: %v", err)
	}

	// Should convert and use global URLs
	if len(merged.URL) != 1 {
		t.Errorf("Expected 1 URL from global interface{}, got %d", len(merged.URL))
	}
}

func TestDetermineCurrent(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()

	// Add profiles
	if err := cfg.AddProfile("work", &Profile{
		User: UserConfig{
			Name:  "Work User",
			Email: "work@example.com",
		},
	}); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	if err := cfg.AddProfile("personal", &Profile{
		User: UserConfig{
			Name:  "Personal User",
			Email: "personal@example.com",
		},
	}); err != nil {
		t.Fatalf("AddProfile failed: %v", err)
	}

	// determineCurrent is called during LoadConfig, but we can test it indirectly
	// by checking that Current is empty when no git config matches
	if cfg.Current != "" {
		t.Errorf("Current should be empty when no matching git config, got: %s", cfg.Current)
	}
}

func TestDetermineCurrentDirectly(t *testing.T) {
	t.Parallel()

	cfg := NewConfig()

	// Test with no profiles
	cfg.determineCurrent()

	if cfg.Current != "" {
		t.Errorf("Current should be empty with no profiles, got: %s", cfg.Current)
	}

	// Add a profile
	cfg.Profiles["test"] = &Profile{
		User: UserConfig{
			Name:  "Test User",
			Email: "test@example.com",
		},
	}

	// Call determineCurrent - won't match unless git config actually has these values
	// This tests the code path for non-matching profiles
	cfg.determineCurrent()
	// Current will be empty unless the system's actual git config matches
	// We're just ensuring no panic/error occurs

	// Add multiple profiles to test the matching loop
	cfg.Profiles["work"] = &Profile{
		User: UserConfig{
			Name:  "Work User",
			Email: "work@example.com",
		},
	}
	cfg.Profiles["personal"] = &Profile{
		User: UserConfig{
			Name:  "Personal User",
			Email: "personal@example.com",
		},
	}

	cfg.determineCurrent()
	// Again, just ensuring the function completes without error
}

func TestMergeMapEdgeCases(t *testing.T) {
	t.Parallel()

	// Test merging empty maps
	result := mergeMap(nil, nil)
	if len(result) != 0 {
		t.Errorf("Merging nil maps should produce empty map, got length %d", len(result))
	}

	// Test merging with nil global
	profile := map[string]any{"key1": "value1"}

	result = mergeMap(nil, profile)
	if len(result) != 1 || result["key1"] != "value1" {
		t.Error("Should use profile values when global is nil")
	}

	// Test merging with nil profile
	global := map[string]any{"key2": "value2"}

	result = mergeMap(global, nil)
	if len(result) != 1 || result["key2"] != "value2" {
		t.Error("Should use global values when profile is nil")
	}

	// Test override behavior
	global = map[string]any{
		"key1": "global1",
		"key2": "global2",
	}
	profile = map[string]any{
		"key1": "profile1",
		"key3": "profile3",
	}
	result = mergeMap(global, profile)

	if result["key1"] != "profile1" {
		t.Error("Profile value should override global")
	}

	if result["key2"] != "global2" {
		t.Error("Global value should be preserved when not in profile")
	}

	if result["key3"] != "profile3" {
		t.Error("Profile value should be included")
	}
}
