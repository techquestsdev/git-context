package cmd

import (
	"fmt"

	"github.com/aanogueira/git-context/internal/config"
	"github.com/aanogueira/git-context/internal/ui"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "git-context",
	Short: "Manage multiple git configuration profiles",
	Long: `Git Context is a CLI tool that helps you manage multiple git configuration profiles.

Switch between different git identities (work, personal, school, etc.) with a single command.
Profiles are stored in ~/.config/git-context/config.yaml`,
	Version: "1.0.0",
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize git-context configuration",
	Long:  `Create the configuration directory and initialize an empty config file.`,
	RunE:  runInit,
}

// runInit handles the 'init' command to initialize the configuration.
// It creates the config directory and file if they don't exist.
func runInit(cmd *cobra.Command, args []string) error {
	paths, err := config.NewPaths()
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to initialize paths: %v", err))

		return errors.Wrap(err, "failed to initialize paths")
	}

	// Check if config already exists
	cfg, err := config.LoadConfig(paths.ConfigFile)
	if err != nil {
		ui.PrintError(fmt.Sprintf("Failed to load existing config: %v", err))

		return errors.Wrap(err, "failed to load config")
	}

	// If config already has profiles, don't overwrite
	if len(cfg.Profiles) > 0 {
		ui.PrintSuccess(
			fmt.Sprintf(
				"Configuration already exists at %s with %d profile(s)",
				paths.ConfigFile,
				len(cfg.Profiles),
			),
		)

		return nil
	}

	// Save the config (this will create directories if needed)
	if err := cfg.SaveConfig(paths.ConfigFile); err != nil {
		ui.PrintError(fmt.Sprintf("Failed to save config: %v", err))

		return errors.Wrap(err, "failed to save config")
	}

	ui.PrintSuccess("Initialized git-context configuration at " + paths.ConfigFile)

	return nil
}

// Execute is the main entry point for the CLI.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return errors.Wrap(err, "failed to execute root command")
	}

	return nil
}

func init() {
	rootCmd.AddCommand(initCmd)
}
