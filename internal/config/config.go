package config

import (
	"fmt"
	"maps"
	"os"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/cockroachdb/errors"
)

// Profile represents a git configuration profile.
type Profile struct {
	User        UserConfig     `yaml:"user,omitempty"`
	URL         []URLConfig    `yaml:"url,omitempty"`
	HTTP        map[string]any `yaml:"http,omitempty"`
	Core        map[string]any `yaml:"core,omitempty"`
	Interactive map[string]any `yaml:"interactive,omitempty"`
	Add         map[string]any `yaml:"add,omitempty"`
	Delta       map[string]any `yaml:"delta,omitempty"`
	Push        map[string]any `yaml:"push,omitempty"`
	Merge       map[string]any `yaml:"merge,omitempty"`
	Commit      map[string]any `yaml:"commit,omitempty"`
	GPG         map[string]any `yaml:"gpg,omitempty"`
	Pull        map[string]any `yaml:"pull,omitempty"`
	Rerere      map[string]any `yaml:"rerere,omitempty"`
	Column      map[string]any `yaml:"column,omitempty"`
	Branch      map[string]any `yaml:"branch,omitempty"`
	Init        map[string]any `yaml:"init,omitempty"`
	Custom      map[string]any `yaml:"custom,omitempty"`
}

// UserConfig represents git user section.
type UserConfig struct {
	Name       string `yaml:"name"`
	Email      string `yaml:"email"`
	SigningKey string `yaml:"signingkey,omitempty"`
}

// URLConfig represents git url rewrite rules.
type URLConfig struct {
	Pattern   string `yaml:"pattern"`
	InsteadOf string `yaml:"insteadOf"`
}

// Config represents the entire configuration.
type Config struct {
	Global   map[string]any      `yaml:"global"`
	Profiles map[string]*Profile `yaml:"profiles"`
	Current  string              `yaml:"-"` // Not saved, determined at runtime
}

// NewConfig creates a new empty config.
func NewConfig() *Config {
	return &Config{
		Global:   make(map[string]any),
		Profiles: make(map[string]*Profile),
		Current:  "",
	}
}

// LoadConfig loads the configuration from file.
func LoadConfig(configFile string) (*Config, error) {
	// If file doesn't exist, return empty config
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return NewConfig(), nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read config file")
	}

	config := NewConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, errors.Wrap(err, "failed to parse config file")
	}

	// Determine current profile by checking git config
	config.determineCurrent()

	return config, nil
}

// SaveConfig saves the configuration to file.
func (c *Config) SaveConfig(configFile string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "failed to marshal config")
	}

	if err := os.WriteFile(configFile, data, 0o644); err != nil {
		return errors.Wrap(err, "failed to write config file")
	}

	return nil
}

// AddProfile adds a new profile.
func (c *Config) AddProfile(name string, profile *Profile) error {
	if _, exists := c.Profiles[name]; exists {
		return errors.WithStack(errors.Newf("profile '%s' already exists", name))
	}

	c.Profiles[name] = profile

	return nil
}

// RemoveProfile removes a profile.
func (c *Config) RemoveProfile(name string) error {
	if _, exists := c.Profiles[name]; !exists {
		return errors.WithStack(errors.Newf("profile '%s' does not exist", name))
	}

	delete(c.Profiles, name)

	return nil
}

// GetProfile gets a profile by name.
func (c *Config) GetProfile(name string) (*Profile, error) {
	profile, exists := c.Profiles[name]
	if !exists {
		return nil, errors.WithStack(errors.Newf("profile '%s' does not exist", name))
	}

	return profile, nil
}

// ListProfiles returns a list of all profile names.
func (c *Config) ListProfiles() []string {
	profiles := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		profiles = append(profiles, name)
	}

	return profiles
}

// Merge combines global config with profile config.
func (c *Config) Merge(profileName string) (*Profile, error) {
	profile, err := c.GetProfile(profileName)
	if err != nil {
		return nil, err
	}

	// Helper to get section from global config
	getGlobalSection := func(section string) map[string]any {
		if val, exists := c.Global[section]; exists {
			if m, ok := val.(map[string]any); ok {
				return m
			}
		}

		return make(map[string]any)
	}

	// Merge URLs: profile URLs override global URLs
	mergedURLs := profile.URL
	if len(mergedURLs) == 0 && c.Global["url"] != nil {
		// If profile has no URLs, use global URLs
		if urlList, ok := c.Global["url"].([]URLConfig); ok {
			mergedURLs = urlList
		} else if urlList, ok := c.Global["url"].([]any); ok {
			// Handle case where URL is unmarshalled as []interface{}
			for _, item := range urlList {
				if urlMap, ok := item.(map[string]any); ok {
					mergedURLs = append(mergedURLs, URLConfig{
						Pattern:   fmt.Sprintf("%v", urlMap["pattern"]),
						InsteadOf: fmt.Sprintf("%v", urlMap["insteadOf"]),
					})
				}
			}
		}
	}

	// Create a new merged profile
	merged := &Profile{
		User:        profile.User,
		URL:         mergedURLs,
		HTTP:        mergeMap(getGlobalSection("http"), profile.HTTP),
		Core:        mergeMap(getGlobalSection("core"), profile.Core),
		Interactive: mergeMap(getGlobalSection("interactive"), profile.Interactive),
		Add:         mergeMap(getGlobalSection("add"), profile.Add),
		Delta:       mergeMap(getGlobalSection("delta"), profile.Delta),
		Push:        mergeMap(getGlobalSection("push"), profile.Push),
		Merge:       mergeMap(getGlobalSection("merge"), profile.Merge),
		Commit:      mergeMap(getGlobalSection("commit"), profile.Commit),
		GPG:         mergeMap(getGlobalSection("gpg"), profile.GPG),
		Pull:        mergeMap(getGlobalSection("pull"), profile.Pull),
		Rerere:      mergeMap(getGlobalSection("rerere"), profile.Rerere),
		Column:      mergeMap(getGlobalSection("column"), profile.Column),
		Branch:      mergeMap(getGlobalSection("branch"), profile.Branch),
		Init:        mergeMap(getGlobalSection("init"), profile.Init),
		Custom:      profile.Custom,
	}

	return merged, nil
}

// determineCurrent determines which profile is currently active by checking git config.
func (c *Config) determineCurrent() {
	// Get current git user.name from git config
	cmd := exec.Command("git", "config", "--global", "user.name")

	output, err := cmd.Output()
	if err != nil {
		return // Can't determine current profile
	}

	currentName := strings.TrimSpace(string(output))

	// Get current git user.email from git config
	cmd = exec.Command("git", "config", "--global", "user.email")

	output, err = cmd.Output()
	if err != nil {
		return // Can't determine current profile
	}

	currentEmail := strings.TrimSpace(string(output))

	// Match against profiles
	for profileName, profile := range c.Profiles {
		if profile.User.Name == currentName && profile.User.Email == currentEmail {
			c.Current = profileName

			return
		}
	}
}

// mergeMap merges two maps, with values from profileConfig overriding globalConfig.
func mergeMap(
	globalConfig map[string]any,
	profileConfig map[string]any,
) map[string]any {
	result := make(map[string]any)

	// First add global config values
	maps.Copy(result, globalConfig)

	// Then override with profile-specific values
	maps.Copy(result, profileConfig)

	return result
}
