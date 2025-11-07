package cmd

import (
	"fmt"

	"github.com/aanogueira/git-context/internal/config"
	"github.com/aanogueira/git-context/internal/ui"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [profile-name]",
	Short: "Add a new profile",
	Long:  `Create a new git configuration profile interactively.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runAdd,
}

// runAdd handles the 'add' command to create a new git profile.
// It prompts for profile details and saves them to the configuration.
func runAdd(cmd *cobra.Command, args []string) error {
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

	// Check if profile already exists
	if _, err := cfg.GetProfile(profileName); err == nil {
		ui.PrintError(fmt.Sprintf("Profile '%s' already exists", profileName))

		return errors.New("profile already exists")
	}

	ui.PrintHeader("Creating Profile: " + profileName)

	// Prompt for user details
	name, err := ui.PromptText("Git Name", "")
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to get name: %v", err))

		return errors.Wrap(err, "failed to get name")
	}

	email, err := ui.PromptText("Git Email", "")
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to get email: %v", err))

		return errors.Wrap(err, "failed to get email")
	}

	signingKey, _ := ui.PromptText("GPG Signing Key (optional)", "")

	// Create profile
	profile := &config.Profile{
		User: config.UserConfig{
			Name:       name,
			Email:      email,
			SigningKey: signingKey,
		},
	}

	// Add URL rewrites if needed
	addURLs, _ := ui.PromptConfirm("Add URL rewrites?")
	if addURLs {
		profile.URL = promptURLRewrites()
	}

	// Add the profile
	if err := cfg.AddProfile(profileName, profile); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to add profile: %v", err))

		return errors.Wrap(err, "failed to add profile")
	}

	// Save config
	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to save config: %v", err))

		return errors.Wrap(err, "failed to save config")
	}

	ui.PrintSuccess(fmt.Sprintf("Profile '%s' created successfully", profileName))

	return nil
}

// promptURLRewrites interactively prompts for URL rewrite rules.
// Returns a slice of URLConfig entries for git URL substitution.
func promptURLRewrites() []config.URLConfig {
	var urls []config.URLConfig

	for {
		pattern, err := ui.PromptText("URL pattern (e.g., git@gitlab.com/)", "")
		if err != nil || pattern == "" {
			break
		}

		insteadOf, err := ui.PromptText("Instead of (e.g., https://gitlab.com/)", "")
		if err != nil || insteadOf == "" {
			break
		}

		urls = append(urls, config.URLConfig{
			Pattern:   pattern,
			InsteadOf: insteadOf,
		})

		more, _ := ui.PromptConfirm("Add another URL rewrite?")
		if !more {
			break
		}
	}

	return urls
}

func init() {
	rootCmd.AddCommand(addCmd)
}
