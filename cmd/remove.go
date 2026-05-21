package cmd

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/techquestsdev/git-context/internal/config"
	"github.com/techquestsdev/git-context/internal/git"
	"github.com/techquestsdev/git-context/internal/ui"
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

	hasDirs := len(cfg.Profiles[profileName].Directories) > 0

	prompt := fmt.Sprintf("Are you sure you want to remove profile '%s'?", profileName)
	if hasDirs {
		prompt = fmt.Sprintf(
			"Profile '%s' has %d assigned director%s. Remove anyway?",
			profileName,
			len(cfg.Profiles[profileName].Directories),
			plural(len(cfg.Profiles[profileName].Directories), "y", "ies"),
		)
	}

	confirm, err := ui.PromptConfirm(prompt)
	if err != nil {
		ui.PrintWarning("Removal canceled")

		return errors.Wrap(err, "failed to confirm removal")
	}

	if !confirm {
		ui.PrintWarning("Removal canceled")

		return nil
	}

	if err := cfg.RemoveProfile(profileName); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to remove profile: %v", err))

		return errors.Wrap(err, "failed to remove profile")
	}

	if cfg.Current == profileName {
		cfg.Current = ""
	}

	if cfg.Previous == profileName {
		cfg.Previous = ""
	}

	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to save config: %v", err))

		return errors.Wrap(err, "failed to save config")
	}

	g := git.NewGit(paths.GitConfigFile)
	if err := g.Regenerate(cfg, paths.ProfilesDir); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to regenerate git config: %v", err))

		return errors.Wrap(err, "failed to regenerate git config")
	}

	ui.PrintSuccess(fmt.Sprintf("Profile '%s' removed successfully", profileName))

	return nil
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}

	return many
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
