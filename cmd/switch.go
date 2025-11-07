package cmd

import (
	"fmt"

	"github.com/aanogueira/git-context/internal/config"
	"github.com/aanogueira/git-context/internal/git"
	"github.com/aanogueira/git-context/internal/ui"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var switchCmd = &cobra.Command{
	Use:   "switch [profile-name]",
	Short: "Switch to a different profile",
	Long:  `Switch the active git configuration to a different profile.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSwitch,
}

func runSwitch(cmd *cobra.Command, args []string) error {
	profileName := args[0]

	paths, err := config.NewPaths()
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to get paths: %v", err))

		return errors.Wrap(err, "failed to get paths")
	}

	cfg, err := config.LoadConfig(paths.ConfigFile)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to load config: %v", err))

		return errors.Wrap(err, "failed to load config")
	}

	// Check if profile exists
	profile, err := cfg.GetProfile(profileName)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Profile not found: %v", err))

		return errors.Wrap(err, "profile not found")
	}

	ui.PrintHeader("Switching to Profile: " + profileName)

	// Create Git instance
	g := git.NewGit(paths.GitConfigFile)

	// Backup current config
	if err := g.BackupConfig(paths.GitConfigBackup); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to backup git config: %v", err))
	} else {
		ui.PrintInfo("Backed up git config to " + paths.GitConfigBackup)
	}

	// Build the merged configuration
	mergedProfile, err := cfg.Merge(profileName)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to merge configurations: %v", err))

		return errors.Wrap(err, "failed to merge configurations")
	}

	// Convert profile to git config format and write
	gitConfig := profileToGitConfig(mergedProfile)
	if err := g.WriteConfig(gitConfig); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to write git config: %v", err))

		return errors.Wrap(err, "failed to write git config")
	}

	// Update current profile
	cfg.Current = profileName
	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to save config: %v", err))

		return errors.Wrap(err, "failed to save config")
	}

	ui.PrintSuccess(fmt.Sprintf("Switched to profile '%s'", profileName))
	ui.PrintInfo(fmt.Sprintf("User: %s <%s>", profile.User.Name, profile.User.Email))

	return nil
}

// profileToGitConfig converts a Profile to a git configuration map.
// It maps profile fields to git config keys (user.name, user.email, etc.).
func profileToGitConfig(profile *config.Profile) map[string]any {
	config := make(map[string]any)

	// User section
	if profile.User.Name != "" {
		config["user.name"] = profile.User.Name
	}

	if profile.User.Email != "" {
		config["user.email"] = profile.User.Email
	}

	if profile.User.SigningKey != "" {
		config["user.signingkey"] = profile.User.SigningKey
	}

	// URL rewrites
	for _, url := range profile.URL {
		key := fmt.Sprintf("url \"%s\".insteadOf", url.Pattern)
		config[key] = url.InsteadOf
	}

	// Add other sections from the profile
	addSectionToConfig(config, "http", profile.HTTP)
	addSectionToConfig(config, "core", profile.Core)
	addSectionToConfig(config, "interactive", profile.Interactive)
	addSectionToConfig(config, "add", profile.Add)
	addSectionToConfig(config, "delta", profile.Delta)
	addSectionToConfig(config, "push", profile.Push)
	addSectionToConfig(config, "merge", profile.Merge)
	addSectionToConfig(config, "commit", profile.Commit)
	addSectionToConfig(config, "gpg", profile.GPG)
	addSectionToConfig(config, "pull", profile.Pull)
	addSectionToConfig(config, "rerere", profile.Rerere)
	addSectionToConfig(config, "column", profile.Column)
	addSectionToConfig(config, "branch", profile.Branch)
	addSectionToConfig(config, "init", profile.Init)

	return config
}

// addSectionToConfig adds a section with values to the git configuration map.
// It delegates to addSectionToConfigRecursive for hierarchical key handling.
func addSectionToConfig(
	config map[string]any,
	section string,
	values map[string]any,
) {
	addSectionToConfigRecursive(config, section, values)
}

// addSectionToConfigRecursive recursively adds nested configuration values.
// It handles dot-separated keys by creating nested maps as needed.
func addSectionToConfigRecursive(
	config map[string]any,
	prefix string,
	values map[string]any,
) {
	for k, v := range values {
		// If the value is a map, it's a subsection - just continue with dot notation
		// e.g., add.interactive, delta.decorations, delta.interactive
		if m, ok := v.(map[string]any); ok {
			key := fmt.Sprintf("%s.%s", prefix, k)
			addSectionToConfigRecursive(config, key, m)
		} else {
			// Leaf value - add it directly
			key := fmt.Sprintf("%s.%s", prefix, k)
			config[key] = v
		}
	}
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
