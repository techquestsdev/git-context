package git

import (
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
)

// Git handles git config operations.
type Git struct {
	globalConfigPath string
}

// NewGit creates a new Git instance.
func NewGit(globalConfigPath string) *Git {
	return &Git{
		globalConfigPath: globalConfigPath,
	}
}

// WriteConfig writes configuration to git global config.
func (g *Git) WriteConfig(config map[string]any) error {
	content := buildGitConfig(config)
	if err := os.WriteFile(g.globalConfigPath, []byte(content), 0o644); err != nil {
		return errors.Wrap(err, "failed to write git config")
	}

	return nil
}

// BackupConfig creates a backup of the git config.
func (g *Git) BackupConfig(backupPath string) error {
	data, err := os.ReadFile(g.globalConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No config to backup
		}

		return errors.Wrap(err, "failed to read git config for backup")
	}

	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return errors.Wrap(err, "failed to create backup")
	}

	return nil
}

// buildGitConfig builds git config format from a map of key-value pairs.
// It handles both regular dotted notation and quoted subsections.
// Returns a formatted git config string with sections and key-value pairs.
//
//nolint:nestif
func buildGitConfig(config map[string]any) string {
	var content strings.Builder

	sectionMap := make(map[string]map[string]any)

	// Group keys by section
	for key, value := range config {
		// Handle quoted subsections like url."pattern".insteadof
		var section, keyName string

		if strings.Contains(key, "\"") {
			// This is a quoted subsection like url."ssh://git@gitlab.molops.io/".insteadof
			// Extract the section with the quoted part intact
			lastDotIdx := strings.LastIndex(key, ".")
			if lastDotIdx > 0 {
				section = key[:lastDotIdx]
				keyName = key[lastDotIdx+1:]
			} else {
				continue
			}
		} else {
			// Regular dot notation
			parts := strings.Split(key, ".")
			if len(parts) >= 2 {
				section = strings.Join(parts[:len(parts)-1], ".")
				keyName = parts[len(parts)-1]
			} else {
				continue
			}
		}

		if sectionMap[section] == nil {
			sectionMap[section] = make(map[string]any)
		}

		sectionMap[section][keyName] = value
	}

	// Write sections
	for section, values := range sectionMap {
		content.WriteString(fmt.Sprintf("[%s]\n", section))

		for k, v := range values {
			content.WriteString(fmt.Sprintf("\t%s = %v\n", k, v))
		}

		content.WriteString("\n")
	}

	return content.String()
}
