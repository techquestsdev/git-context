package cmd

import (
	"fmt"
	"sort"

	"github.com/aanogueira/git-context/internal/config"
	"github.com/aanogueira/git-context/internal/ui"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available profiles",
	Long:  `Display all available git configuration profiles.`,
	RunE:  runList,
}

// runList handles the 'list' command to display all saved profiles.
// It shows a formatted table with profile names, emails, and signing key status.
func runList(cmd *cobra.Command, args []string) error {
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

	profiles := cfg.ListProfiles()
	if len(profiles) == 0 {
		ui.PrintWarning("No profiles found. Create one with 'git-context add <name>'")

		return nil
	}

	sort.Strings(profiles)

	ui.PrintHeader("Available Profiles")

	rows := make([][]string, len(profiles))
	for i, profile := range profiles {
		status := ""
		if profile == cfg.Current {
			status = "● (active)"
		}

		p, _ := cfg.GetProfile(profile)

		email := ""
		if p != nil && p.User.Email != "" {
			email = p.User.Email
		}

		rows[i] = []string{profile, email, status}
	}

	ui.PrintTable([]string{"Profile", "Email", "Status"}, rows)

	return nil
}

func init() {
	rootCmd.AddCommand(listCmd)
}
