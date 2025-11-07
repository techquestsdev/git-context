package config

import (
	"slices"
	"testing"
)

//nolint:gocognit
func TestGetSection(t *testing.T) {
	t.Parallel()

	// Factory function to create a fresh profile for each test
	newTestProfile := func() *Profile {
		return &Profile{
			HTTP: map[string]any{
				"postBuffer": "524288000",
			},
			Core: map[string]any{
				"editor": "vim",
			},
			Interactive: map[string]any{
				"diffFilter": "delta --color-only",
			},
			Add: map[string]any{
				"interactive": map[string]any{
					"useBuiltin": true,
				},
			},
			Delta: map[string]any{
				"navigate": true,
			},
			Push: map[string]any{
				"default": "current",
			},
			Merge: map[string]any{
				"conflictStyle": "diff3",
			},
			Commit: map[string]any{
				"gpgSign": true,
			},
			GPG: map[string]any{
				"format": "ssh",
			},
			Pull: map[string]any{
				"rebase": true,
			},
			Rerere: map[string]any{
				"enabled": true,
			},
			Column: map[string]any{
				"ui": "auto",
			},
			Branch: map[string]any{
				"sort": "-committerdate",
			},
			Init: map[string]any{
				"defaultBranch": "main",
			},
			Diff: map[string]any{
				"algorithm": "histogram",
			},
			Fetch: map[string]any{
				"prune": true,
			},
			Rebase: map[string]any{
				"autoStash": true,
			},
			Maintenance: map[string]any{
				"auto": false,
			},
			Feature: map[string]any{
				"manyFiles": true,
			},
			Alias: map[string]any{
				"st": "status",
			},
			Tag: map[string]any{
				"sort": "-version:refname",
			},
		}
	}

	tests := []struct {
		name     string
		section  string
		expected map[string]any
	}{
		{"HTTP section", "http", map[string]any{"postBuffer": "524288000"}},
		{"Core section", "core", map[string]any{"editor": "vim"}},
		{"Interactive section", "interactive", map[string]any{"diffFilter": "delta --color-only"}},
		{"Add section", "add", map[string]any{"interactive": map[string]any{"useBuiltin": true}}},
		{"Delta section", "delta", map[string]any{"navigate": true}},
		{"Push section", "push", map[string]any{"default": "current"}},
		{"Merge section", "merge", map[string]any{"conflictStyle": "diff3"}},
		{"Commit section", "commit", map[string]any{"gpgSign": true}},
		{"GPG section", "gpg", map[string]any{"format": "ssh"}},
		{"Pull section", "pull", map[string]any{"rebase": true}},
		{"Rerere section", "rerere", map[string]any{"enabled": true}},
		{"Column section", "column", map[string]any{"ui": "auto"}},
		{"Branch section", "branch", map[string]any{"sort": "-committerdate"}},
		{"Init section", "init", map[string]any{"defaultBranch": "main"}},
		{"Diff section", "diff", map[string]any{"algorithm": "histogram"}},
		{"Fetch section", "fetch", map[string]any{"prune": true}},
		{"Rebase section", "rebase", map[string]any{"autoStash": true}},
		{"Maintenance section", "maintenance", map[string]any{"auto": false}},
		{"Feature section", "feature", map[string]any{"manyFiles": true}},
		{"Alias section", "alias", map[string]any{"st": "status"}},
		{"Tag section", "tag", map[string]any{"sort": "-version:refname"}},
		{"Unknown section", "unknown", nil},
		{"Empty section", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			profile := newTestProfile()
			result := profile.GetSection(tt.section)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("GetSection(%q) = %v, want nil", tt.section, result)
				}

				return
			}

			if result == nil {
				t.Errorf("GetSection(%q) = nil, want %v", tt.section, tt.expected)

				return
			}

			// Deep compare maps
			if len(result) != len(tt.expected) {
				t.Errorf("GetSection(%q) length = %d, want %d", tt.section, len(result), len(tt.expected))

				return
			}

			for key, expectedValue := range tt.expected {
				actualValue, ok := result[key]
				if !ok {
					t.Errorf("GetSection(%q) missing key %q", tt.section, key)

					continue
				}

				// Handle nested maps
				if expectedMap, ok := expectedValue.(map[string]any); ok {
					actualMap, ok := actualValue.(map[string]any)
					if !ok {
						t.Errorf("GetSection(%q)[%q] type mismatch, got %T, want map[string]any", tt.section, key, actualValue)

						continue
					}

					for k, v := range expectedMap {
						if actualMap[k] != v {
							t.Errorf("GetSection(%q)[%q][%q] = %v, want %v", tt.section, key, k, actualMap[k], v)
						}
					}
				} else if actualValue != expectedValue {
					t.Errorf("GetSection(%q)[%q] = %v, want %v", tt.section, key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestSetSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		section string
		values  map[string]any
	}{
		{"Set HTTP", "http", map[string]any{"postBuffer": "524288000"}},
		{"Set Core", "core", map[string]any{"editor": "nvim"}},
		{"Set Interactive", "interactive", map[string]any{"diffFilter": "delta"}},
		{"Set Add", "add", map[string]any{"interactive.useBuiltin": true}},
		{"Set Delta", "delta", map[string]any{"navigate": true}},
		{"Set Push", "push", map[string]any{"default": "simple"}},
		{"Set Merge", "merge", map[string]any{"conflictStyle": "zdiff3"}},
		{"Set Commit", "commit", map[string]any{"gpgSign": false}},
		{"Set GPG", "gpg", map[string]any{"format": "openpgp"}},
		{"Set Pull", "pull", map[string]any{"ff": "only"}},
		{"Set Rerere", "rerere", map[string]any{"enabled": false}},
		{"Set Column", "column", map[string]any{"ui": "never"}},
		{"Set Branch", "branch", map[string]any{"autoSetupRebase": "always"}},
		{"Set Init", "init", map[string]any{"defaultBranch": "develop"}},
		{"Set Diff", "diff", map[string]any{"algorithm": "patience"}},
		{"Set Fetch", "fetch", map[string]any{"prune": false}},
		{"Set Rebase", "rebase", map[string]any{"autoSquash": true}},
		{"Set Maintenance", "maintenance", map[string]any{"auto": true}},
		{"Set Feature", "feature", map[string]any{"experimental": true}},
		{"Set Alias", "alias", map[string]any{"co": "checkout"}},
		{"Set Tag", "tag", map[string]any{"gpgSign": true}},
		{"Set unknown section", "unknown", map[string]any{"key": "value"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			profile := &Profile{}
			profile.SetSection(tt.section, tt.values)

			result := profile.GetSection(tt.section)

			// For unknown sections, result should be nil
			if tt.section == "unknown" {
				if result != nil {
					t.Errorf("SetSection(%q) should not set unknown section, got %v", tt.section, result)
				}

				return
			}

			if result == nil {
				t.Errorf("SetSection(%q) did not set section", tt.section)

				return
			}

			if len(result) != len(tt.values) {
				t.Errorf("SetSection(%q) length = %d, want %d", tt.section, len(result), len(tt.values))

				return
			}

			for key, expectedValue := range tt.values {
				actualValue, ok := result[key]
				if !ok {
					t.Errorf("SetSection(%q) missing key %q", tt.section, key)

					continue
				}

				if actualValue != expectedValue {
					t.Errorf("SetSection(%q)[%q] = %v, want %v", tt.section, key, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestConfigSections(t *testing.T) {
	t.Parallel()

	// Verify ConfigSections contains expected sections
	expectedSections := map[string]bool{
		"http":        true,
		"core":        true,
		"interactive": true,
		"add":         true,
		"delta":       true,
		"push":        true,
		"merge":       true,
		"commit":      true,
		"gpg":         true,
		"pull":        true,
		"rerere":      true,
		"column":      true,
		"branch":      true,
		"init":        true,
		"diff":        true,
		"fetch":       true,
		"rebase":      true,
		"maintenance": true,
		"feature":     true,
		"alias":       true,
		"tag":         true,
	}

	if len(ConfigSections) != len(expectedSections) {
		t.Errorf("ConfigSections length = %d, want %d", len(ConfigSections), len(expectedSections))
	}

	for _, section := range ConfigSections {
		if !expectedSections[section] {
			t.Errorf("ConfigSections contains unexpected section: %q", section)
		}
	}

	for section := range expectedSections {
		found := false

		if slices.Contains(ConfigSections, section) {
			found = true
		}

		if !found {
			t.Errorf("ConfigSections missing expected section: %q", section)
		}
	}
}

func TestGetSetSectionRoundTrip(t *testing.T) {
	t.Parallel()

	testValues := map[string]any{
		"key1": "value1",
		"key2": true,
		"key3": 123,
	}

	for _, section := range ConfigSections {
		t.Run(section, func(t *testing.T) {
			t.Parallel()

			profile := &Profile{}

			// Set the section
			profile.SetSection(section, testValues) // Get the section back
			result := profile.GetSection(section)

			if result == nil {
				t.Fatalf("GetSection(%q) returned nil after SetSection", section)
			}

			// Verify values match
			for key, expectedValue := range testValues {
				actualValue, ok := result[key]
				if !ok {
					t.Errorf("Missing key %q in section %q", key, section)

					continue
				}

				if actualValue != expectedValue {
					t.Errorf("Section %q key %q = %v, want %v", section, key, actualValue, expectedValue)
				}
			}
		})
	}
}
