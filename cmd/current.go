package cmd

import (
	"fmt"

	"github.com/aanogueira/git-context/internal/config"
	"github.com/aanogueira/git-context/internal/ui"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the currently active profile",
	Long:  `Display which git configuration profile is currently active.`,
	RunE:  runCurrent,
}

// runCurrent handles the 'current' command to show the currently active profile.
// It compares git configuration with saved profiles to determine which is active.
func runCurrent(cmd *cobra.Command, args []string) error {
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

	if cfg.Current == "" {
		ui.PrintWarning("No active profile set")

		return nil
	}

	profile, err := cfg.GetProfile(cfg.Current)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Active profile not found: %v", err))

		return errors.Wrap(err, "failed to get active profile")
	}

	ui.PrintHeader("Current Profile")
	ui.PrintInfo("Profile: " + cfg.Current)
	ui.PrintInfo("Name: " + profile.User.Name)
	ui.PrintInfo("Email: " + profile.User.Email)

	return nil
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
