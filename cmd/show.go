package cmd

import (
	"fmt"

	"github.com/aanogueira/git-context/internal/config"
	"github.com/aanogueira/git-context/internal/ui"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show [profile-name]",
	Short: "Show profile details",
	Long:  `Display the configuration details for a specific profile.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runShow,
}

// runShow handles the 'show' command to display details of a specific profile.
// It presents all configured values including user info, signing keys, and URL rewrites.
func runShow(cmd *cobra.Command, args []string) error {
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

	profile, err := cfg.GetProfile(profileName)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Profile not found: %v", err))

		return errors.Wrap(err, "failed to get profile")
	}

	ui.PrintHeader("Profile: " + profileName)

	if profile.User.Name != "" {
		ui.PrintInfo("User Name: " + profile.User.Name)
	}

	if profile.User.Email != "" {
		ui.PrintInfo("Email: " + profile.User.Email)
	}

	if profile.User.SigningKey != "" {
		ui.PrintInfo("Signing Key: " + profile.User.SigningKey)
	}

	if len(profile.URL) > 0 {
		fmt.Println()
		ui.PrintInfo("URL Rewrites:")

		for _, urlCfg := range profile.URL {
			ui.PrintInfo(fmt.Sprintf("  %s → %s", urlCfg.InsteadOf, urlCfg.Pattern))
		}
	}

	return nil
}

func init() {
	rootCmd.AddCommand(showCmd)
}
