package cmd

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/techquestsdev/git-context/internal/config"
	"github.com/techquestsdev/git-context/internal/ui"
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

		dirs := "0"

		if p != nil {
			if p.User.Email != "" {
				email = p.User.Email
			}

			dirs = strconv.Itoa(len(p.Directories))
		}

		rows[i] = []string{profile, email, dirs, status}
	}

	ui.PrintTable([]string{"Profile", "Email", "Dirs", "Status"}, rows)

	return nil
}

func init() {
	rootCmd.AddCommand(listCmd)
}
