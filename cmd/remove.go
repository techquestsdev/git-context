package cmd

import (
	"fmt"

	"github.com/aanogueira/git-context/internal/config"
	"github.com/aanogueira/git-context/internal/ui"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [profile-name]",
	Short: "Remove a profile",
	Long:  `Delete a git configuration profile.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runRemove,
}

// runRemove handles the 'remove' command to delete a profile.
// It prompts for confirmation before removing the specified profile.
func runRemove(cmd *cobra.Command, args []string) error {
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
	if _, err := cfg.GetProfile(profileName); err != nil {
		ui.PrintError(fmt.Sprintf("Profile not found: %v", err))

		return errors.Wrap(err, "profile not found")
	}

	// Confirm removal
	confirm, err := ui.PromptConfirm(
		fmt.Sprintf("Are you sure you want to remove profile '%s'?", profileName),
	)
	if err != nil {
		ui.PrintWarning("Removal canceled")

		return errors.Wrap(err, "failed to confirm removal")
	}

	if !confirm {
		ui.PrintWarning("Removal canceled")

		return nil
	}

	// Remove the profile
	if err := cfg.RemoveProfile(profileName); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to remove profile: %v", err))

		return errors.Wrap(err, "failed to remove profile")
	}

	// Save config
	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to save config: %v", err))

		return errors.Wrap(err, "failed to save config")
	}

	ui.PrintSuccess(fmt.Sprintf("Profile '%s' removed successfully", profileName))

	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
