package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/techquestsdev/git-context/internal/config"
	"github.com/techquestsdev/git-context/internal/git"
	"github.com/techquestsdev/git-context/internal/ui"
)

var dirCmd = &cobra.Command{
	Use:   "dir",
	Short: "Manage directory-to-profile assignments",
	Long: `Assign filesystem paths to git profiles. When inside an assigned directory, ` +
		`git uses that profile via includeIf.`,
}

var dirAddCmd = &cobra.Command{
	Use:   "add [path] [profile]",
	Short: "Assign a directory to a profile",
	Args:  cobra.ExactArgs(2),
	RunE:  runDirAdd,
}

var dirRemoveCmd = &cobra.Command{
	Use:   "remove [path]",
	Short: "Remove a directory assignment",
	Args:  cobra.ExactArgs(1),
	RunE:  runDirRemove,
}

var dirListCmd = &cobra.Command{
	Use:   "list",
	Short: "List directory assignments",
	RunE:  runDirList,
}

// runDirAdd handles the 'dir add' command to assign a directory path to a profile.
// It normalizes the path, persists the assignment, and regenerates the git config manifest.
func runDirAdd(cmd *cobra.Command, args []string) error {
	rawPath, profileName := args[0], args[1]

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

	normalized, err := config.NormalizeDir(rawPath)
	if err != nil {
		ui.PrintError(err.Error())

		return errors.Wrap(err, "failed to normalize directory")
	}

	if err := cfg.AssignDir(normalized, profileName); err != nil {
		ui.PrintError(err.Error())

		return errors.Wrap(err, "failed to assign directory")
	}

	if _, err := os.Stat(normalized); os.IsNotExist(err) {
		ui.PrintWarning(fmt.Sprintf("Directory %s does not exist yet (assignment saved anyway)", normalized))
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

	ui.PrintSuccess(fmt.Sprintf("Assigned %s → %s", normalized, profileName))

	if cfg.Current == "" {
		ui.PrintWarning("no default profile set; run 'switch <name>' to apply one outside assigned directories")
	}

	return nil
}

// runDirRemove handles the 'dir remove' command to remove an existing directory
// assignment and regenerate the git config manifest.
func runDirRemove(cmd *cobra.Command, args []string) error {
	rawPath := args[0]

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

	normalized, err := config.NormalizeDir(rawPath)
	if err != nil {
		ui.PrintError(err.Error())

		return errors.Wrap(err, "failed to normalize directory")
	}

	if err := cfg.UnassignDir(normalized); err != nil {
		ui.PrintError(err.Error())

		return errors.Wrap(err, "failed to unassign directory")
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

	ui.PrintSuccess("Removed assignment for " + normalized)

	return nil
}

// runDirList handles the 'dir list' command to display all directory-to-profile
// assignments along with whether each directory currently exists on disk.
func runDirList(cmd *cobra.Command, args []string) error {
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

	assignments := cfg.AssignmentMap()
	if len(assignments) == 0 {
		ui.PrintWarning("No directory assignments. Use 'git-context dir add <path> <profile>' to add one.")

		return nil
	}

	keys := make([]string, 0, len(assignments))
	for k := range assignments {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	rows := make([][]string, 0, len(keys))

	for _, dir := range keys {
		mark := "✓"
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			mark = "✗"
		}

		rows = append(rows, []string{dir, assignments[dir], mark})
	}

	ui.PrintHeader("Directory Assignments")
	ui.PrintTable([]string{"Directory", "Profile", "Exists"}, rows)

	return nil
}

func init() {
	dirCmd.AddCommand(dirAddCmd)
	dirCmd.AddCommand(dirRemoveCmd)
	dirCmd.AddCommand(dirListCmd)
	rootCmd.AddCommand(dirCmd)
}
