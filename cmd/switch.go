package cmd

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/techquestsdev/git-context/internal/config"
	"github.com/techquestsdev/git-context/internal/git"
	"github.com/techquestsdev/git-context/internal/ui"
)

var switchCmd = &cobra.Command{
	Use:   "switch [profile-name|-]",
	Short: "Switch to a different profile",
	Long:  `Switch the active git configuration to a different profile. Use '-' to return to the previous profile.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSwitch,
}

func runSwitch(cmd *cobra.Command, args []string) error {
	requestedName := args[0]

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

	profileName := requestedName
	if requestedName == "-" {
		if cfg.Previous == "" {
			ui.PrintWarning("No previous profile to switch to")

			return errors.WithStack(errors.New("no previous profile to switch to"))
		}

		profileName = cfg.Previous
	}

	// Check if profile exists
	profile, err := cfg.GetProfile(profileName)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Profile not found: %v", err))

		return errors.Wrap(err, "profile not found")
	}

	ui.PrintHeader("Switching to Profile: " + profileName)

	g := git.NewGit(paths.GitConfigFile)

	if err := g.BackupConfig(paths.GitConfigBackup); err != nil {
		ui.PrintWarning(fmt.Sprintf("Failed to backup git config: %v", err))
	} else {
		ui.PrintInfo("Backed up git config to " + paths.GitConfigBackup)
	}

	if cfg.Current != "" && cfg.Current != profileName {
		cfg.Previous = cfg.Current
	}

	cfg.Current = profileName

	if err := g.Regenerate(cfg, paths.ProfilesDir); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to regenerate git config: %v", err))

		return errors.Wrap(err, "failed to regenerate git config")
	}

	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to save config: %v", err))

		return errors.Wrap(err, "failed to save config")
	}

	ui.PrintSuccess(fmt.Sprintf("Switched to profile '%s'", profileName))
	ui.PrintInfo(fmt.Sprintf("User: %s <%s>", profile.User.Name, profile.User.Email))

	return nil
}

func init() {
	rootCmd.AddCommand(switchCmd)
}
