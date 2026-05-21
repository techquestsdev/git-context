package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/techquestsdev/git-context/internal/config"
	"github.com/techquestsdev/git-context/internal/ui"
)

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the currently active profile",
	Long:  `Display which git configuration profile is currently active globally and effective in the current directory.`,
	RunE:  runCurrent,
}

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

	ui.PrintHeader("Current Profile")

	if cfg.Current == "" {
		ui.PrintWarning("No active profile set")
	} else {
		profile, err := cfg.GetProfile(cfg.Current)
		if err != nil {
			ui.PrintError(fmt.Sprintf("Active profile not found: %v", err))

			return errors.Wrap(err, "failed to get active profile")
		}

		ui.PrintInfo("Default: " + cfg.Current)
		ui.PrintInfo("Name: " + profile.User.Name)
		ui.PrintInfo("Email: " + profile.User.Email)
	}

	if effective := effectiveProfileInCwd(paths, cfg); effective != "" {
		ui.PrintInfo("Effective in " + currentDir() + ": " + effective)
	}

	return nil
}

// effectiveProfileInCwd asks git which file user.email comes from in the
// current working directory. If the answer is one of our per-profile files,
// returns the profile name; otherwise returns "".
func effectiveProfileInCwd(paths *config.Paths, cfg *config.Config) string {
	cmd := exec.Command("git", "config", "--show-origin", "user.email")

	out, err := cmd.Output()
	if err != nil {
		return ""
	}

	line := strings.TrimSpace(string(out))

	parts := strings.SplitN(line, "\t", 2)
	if len(parts) == 0 {
		return ""
	}

	if !strings.HasPrefix(parts[0], "file:") {
		return ""
	}

	origin := resolveSymlinks(strings.TrimPrefix(parts[0], "file:"))

	for name := range cfg.Profiles {
		profilePath := resolveSymlinks(filepath.Join(paths.ProfilesDir, name+".gitconfig"))
		if origin == profilePath {
			return name
		}
	}

	return ""
}

// resolveSymlinks returns the symlink-resolved absolute path. On error
// (e.g. file does not exist) it returns the input unchanged so we can
// still attempt a literal comparison.
func resolveSymlinks(path string) string {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return path
	}

	return resolved
}

func currentDir() string {
	d, err := os.Getwd()
	if err != nil {
		return "."
	}

	return d
}

func init() {
	rootCmd.AddCommand(currentCmd)
}
