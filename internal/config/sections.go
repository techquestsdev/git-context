package config

// ConfigSections defines all supported git config sections.
// This makes it easy to add new sections without modifying multiple places.
var ConfigSections = []string{
	"add",
	"alias",
	"branch",
	"column",
	"commit",
	"core",
	"delta",
	"diff",
	"feature",
	"fetch",
	"gpg",
	"http",
	"init",
	"interactive",
	"maintenance",
	"merge",
	"pull",
	"push",
	"rebase",
	"rerere",
	"tag",
}

// GetSection returns the section map from a profile by name.
// This allows dynamic access to profile sections.
func (p *Profile) GetSection(name string) map[string]any {
	switch name {
	case "add":
		return p.Add
	case "alias":
		return p.Alias
	case "branch":
		return p.Branch
	case "column":
		return p.Column
	case "commit":
		return p.Commit
	case "core":
		return p.Core
	case "delta":
		return p.Delta
	case "diff":
		return p.Diff
	case "feature":
		return p.Feature
	case "fetch":
		return p.Fetch
	case "gpg":
		return p.GPG
	case "http":
		return p.HTTP
	case "init":
		return p.Init
	case "interactive":
		return p.Interactive
	case "maintenance":
		return p.Maintenance
	case "merge":
		return p.Merge
	case "pull":
		return p.Pull
	case "push":
		return p.Push
	case "rebase":
		return p.Rebase
	case "rerere":
		return p.Rerere
	case "tag":
		return p.Tag
	default:
		return nil
	}
}

// SetSection sets the section map in a profile by name.
// This allows dynamic modification of profile sections.
func (p *Profile) SetSection(name string, values map[string]any) {
	switch name {
	case "add":
		p.Add = values
	case "alias":
		p.Alias = values
	case "branch":
		p.Branch = values
	case "column":
		p.Column = values
	case "commit":
		p.Commit = values
	case "core":
		p.Core = values
	case "delta":
		p.Delta = values
	case "diff":
		p.Diff = values
	case "feature":
		p.Feature = values
	case "fetch":
		p.Fetch = values
	case "gpg":
		p.GPG = values
	case "http":
		p.HTTP = values
	case "init":
		p.Init = values
	case "interactive":
		p.Interactive = values
	case "maintenance":
		p.Maintenance = values
	case "merge":
		p.Merge = values
	case "pull":
		p.Pull = values
	case "push":
		p.Push = values
	case "rebase":
		p.Rebase = values
	case "rerere":
		p.Rerere = values
	case "tag":
		p.Tag = values
	}
}
