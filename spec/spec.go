package spec

import (
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/moorara/color"
	"gopkg.in/yaml.v3"
)

const envVarName = "CHANGELOG_ACCESS_TOKEN"

var specFiles = []string{"changelog.yml", "changelog.yaml"}

const helpTemplate = `
  changelog is a simple command-line tool for generating changelogs based on issues and pull/merge requests.
  It assumes the remote repository name is origin.

  You can also have a changelog.yaml file in your repository for configuring how changelogs are generated.
  For more information, please see https://github.com/moorara/changelog

  Supported Remote Repositories:

    • GitHub (github.com)

  Usage: changelog [flags]

  Flags:

    -help                         Show the help text
    -version                      Print the version number

    {{ yellow "-access-token                 The OAuth access token for making API calls" }}
    {{ yellow "                              The default value is read from the CHANGELOG_ACCESS_TOKEN environment variable" }}

    -file                         The output file for the generated changelog (default: {{.General.File}})
    -base                         An optional file for appending the generated changelog to it {{if .General.Base}}(default: {{.General.Base}}){{end}}
                                  This option can only be used when generating the changelog for the first time
    -print                        Print the generated changelong to STDOUT (default: {{.General.Print}})
                                  If this option is enabled, all logs will be disabled
    -verbose                      Show the vervbosity logs (default: {{.General.Verbose}})

    -from-tag                     Changelog will be generated for all changes after this tag (default: last tag on changelog)
    -to-tag                       Changelog will be generated for all changes before this tag (default: last git tag)
    -future-tag                   A future tag for all unreleased changes (changes after the last git tag) {{if .Tags.Future}}(default: {{.Tags.Future ","}}){{end}}
    -exclude-tags                 These tags will be excluded from changelog {{if .Tags.Exclude}}(default: {{Join .Tags.Exclude ","}}){{end}}
    -exclude-tags-regex           A POSIX-compliant regex for excluding certain tags from changelog {{if .Tags.ExcludeRegex}}(default: {{.Tags.ExcludeRegex}}){{end}}

    -issues-selection             Include closed issues in changelog (values: none|all|labeled) (default: {{.Issues.Selection}})
    -issues-include-labels        Include issues with these labels {{if .Issues.IncludeLabels}}(default: {{Join .Issues.IncludeLabels ","}}){{end}}
    -issues-exclude-labels        Exclude issues with these labels {{if .Issues.ExcludeLabels}}(default: {{Join .Issues.ExcludeLabels ","}}){{end}}
    -issues-grouping              Grouping style for issues (values: simple|milestone|label) (default: {{.Issues.Grouping}})
    -issues-summary-labels        Labels for summary group {{if .Issues.SummaryLabels}}(default: {{Join .Issues.SummaryLabels ","}}){{end}}
    -issues-removed-labels        Labels for removed group {{if .Issues.RemovedLabels}}(default: {{Join .Issues.RemovedLabels ","}}){{end}}
    -issues-breaking-labels       Labels for breaking group {{if .Issues.BreakingLabels}}(default: {{Join .Issues.BreakingLabels ","}}){{end}}
    -issues-deprecated-labels     Labels for deprecated group {{if .Issues.DeprecatedLabels}}(default: {{Join .Issues.DeprecatedLabels ","}}){{end}}
    -issues-feature-labels        Labels for feature group {{if .Issues.FeatureLabels}}(default: {{Join .Issues.FeatureLabels ","}}){{end}}
    -issues-enhancement-labels    Labels for enhancement group {{if .Issues.EnhancementLabels}}(default: {{Join .Issues.EnhancementLabels ","}}){{end}}
    -issues-bug-labels            Labels for bug group {{if .Issues.BugLabels}}(default: {{Join .Issues.BugLabels ","}}){{end}}
    -issues-security-labels       Labels for security group {{if .Issues.SecurityLabels}}(default: {{Join .Issues.SecurityLabels ","}}){{end}}

    -merges-selection             Include merged pull/merge requests in changelog (values: none|all|labeled) (default: {{.Merges.Selection}})
    -merges-branch                Include pull/merge requests merged into this branch (default: default remote branch)
    -merges-include-labels        Include merges with these labels {{if .Merges.IncludeLabels}}(default: {{Join .Merges.IncludeLabels ","}}){{end}}
    -merges-exclude-labels        Exclude merges with these labels {{if .Merges.ExcludeLabels}}(default: {{Join .Merges.ExcludeLabels ","}}){{end}}
    -merges-grouping              Grouping style for pull/merge requests (values: simple|milestone|label) (default: {{.Merges.Grouping}})
    -merges-summary-labels        Labels for summary group {{if .Merges.SummaryLabels}}(default: {{Join .Merges.SummaryLabels ","}}){{end}}
    -merges-removed-labels        Labels for removed group {{if .Merges.RemovedLabels}}(default: {{Join .Merges.RemovedLabels ","}}){{end}}
    -merges-breaking-labels       Labels for breaking group {{if .Merges.BreakingLabels}}(default: {{Join .Merges.BreakingLabels ","}}){{end}}
    -merges-deprecated-labels     Labels for deprecated group {{if .Merges.DeprecatedLabels}}(default: {{Join .Merges.DeprecatedLabels ","}}){{end}}
    -merges-feature-labels        Labels for feature group {{if .Merges.FeatureLabels}}(default: {{Join .Merges.FeatureLabels ","}}){{end}}
    -merges-enhancement-labels    Labels for enhancement group {{if .Merges.EnhancementLabels}}(default: {{Join .Merges.EnhancementLabels ","}}){{end}}
    -merges-bug-labels            Labels for bug group {{if .Merges.BugLabels}}(default: {{Join .Merges.BugLabels ","}}){{end}}
    -merges-security-labels       Labels for security group {{if .Merges.SecurityLabels}}(default: {{Join .Merges.SecurityLabels ","}}){{end}}

    -release-url                  An external release URL with the '{tag}' placeholder for the release tag

  Examples:

    changelog
    changelog -access-token=<your-access-token>
    changelog -access-token=<your-access-token> -base=HISTORY.md
    changelog -access-token=<your-access-token> -future-tag=v0.1.0

`

const format = `
Specifications
Repo:
  Platform:           %s
  Path:               %s
  AccessToken:        %s
General:
  File:               %s
  Base:               %s
  Print:              %t
  Verbose:            %t
Tags:
  From:               %s
  To:                 %s
  Future:             %s
  Exclude:            %s
  ExcludeRegex:       %s
Issues:
  Selection:          %s
  IncludeLabels:      %s
  ExcludeLabels:      %s
  Grouping:           %s
  SummaryLabels:      %s
  RemovedLabels:      %s
  BreakingLabels:     %s
  DeprecatedLabels:   %s
  FeatureLabels:      %s
  EnhancementLabels:  %s
  BugLabels:          %s
  SecurityLabels:     %s
Merges:
  Selection:          %s
  Branch:             %s
  IncludeLabels:      %s
  ExcludeLabels:      %s
  Grouping:           %s
  SummaryLabels:      %s
  RemovedLabels:      %s
  BreakingLabels:     %s
  DeprecatedLabels:   %s
  FeatureLabels:      %s
  EnhancementLabels:  %s
  BugLabels:          %s
  SecurityLabels:     %s
Content:
  ReleaseURL:         %s
`

// Platform is the platform for managing a Git remote repository.
type Platform string

const (
	// PlatformGitHub represents the GitHub platform.
	PlatformGitHub Platform = "github.com"
	// PlatformGitLab represents the GitLab platform.
	PlatformGitLab Platform = "gitlab.com"
)

// Repo has the specifications for a git repository.
type Repo struct {
	Platform    Platform `yaml:"-"`
	Path        string   `yaml:"-"`
	AccessToken string   `yaml:"-" flag:"access-token"`
}

// General has the general specifications.
type General struct {
	File    string `yaml:"file" flag:"file"`
	Base    string `yaml:"base" flag:"base"`
	Print   bool   `yaml:"print" flag:"print"`
	Verbose bool   `yaml:"verbose" flag:"verbose"`
}

// Tags has the specifications for identifying git tags.
type Tags struct {
	From         string   `yaml:"-" flag:"from-tag"`
	To           string   `yaml:"-" flag:"to-tag"`
	Future       string   `yaml:"-" flag:"future-tag"`
	Exclude      []string `yaml:"exclude" flag:"exclude-tags"`
	ExcludeRegex string   `yaml:"exclude-regex" flag:"exclude-tags-regex"`
}

// Selection determines how changes should be selected for a changelog.
type Selection string

const (
	// SelectionNone does not select any changes.
	SelectionNone = Selection("none")
	// SelectionAll selects all changes.
	SelectionAll = Selection("all")
	// SelectionLabeled selects only those changes that have are labeled.
	SelectionLabeled = Selection("labeled")
)

// Grouping determnies how changes are grouped together.
type Grouping string

const (
	// GroupingSimple groups changes in simple lists.
	GroupingSimple = Grouping("simple")
	// GroupingMilestone groups changes by milestones.
	GroupingMilestone = Grouping("milestone")
	// GroupingLabel groups changes by labels.
	GroupingLabel = Grouping("label")
)

// LabelGroup represents a group of issues or merges characterized by a set of labels.
type LabelGroup struct {
	Title  string
	Labels []string
}

// Issues has the specifications for fetching, flitering, and grouping issues.
type Issues struct {
	Selection         Selection `yaml:"selection" flag:"issues-selection"`
	IncludeLabels     []string  `yaml:"include-labels" flag:"issues-include-labels"`
	ExcludeLabels     []string  `yaml:"exclude-labels" flag:"issues-exclude-labels"`
	Grouping          Grouping  `yaml:"grouping" flag:"issues-grouping"`
	SummaryLabels     []string  `yaml:"summary-labels" flag:"issues-summary-labels"`
	RemovedLabels     []string  `yaml:"removed-labels" flag:"issues-removed-labels"`
	BreakingLabels    []string  `yaml:"breaking-labels" flag:"issues-breaking-labels"`
	DeprecatedLabels  []string  `yaml:"deprecated-labels" flag:"issues-deprecated-labels"`
	FeatureLabels     []string  `yaml:"feature-labels" flag:"issues-feature-labels"`
	EnhancementLabels []string  `yaml:"enhancement-labels" flag:"issues-enhancement-labels"`
	BugLabels         []string  `yaml:"bug-labels" flag:"issues-bug-labels"`
	SecurityLabels    []string  `yaml:"security-labels" flag:"issues-security-labels"`
}

// LabelGroups returns the label groups for issues.
func (i Issues) LabelGroups() []LabelGroup {
	groups := []LabelGroup{}

	if len(i.SummaryLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Release Summary",
			Labels: i.SummaryLabels,
		})
	}

	if len(i.RemovedLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Removed",
			Labels: i.RemovedLabels,
		})
	}

	if len(i.BreakingLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Breaking Changes",
			Labels: i.BreakingLabels,
		})
	}

	if len(i.DeprecatedLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Deprecated",
			Labels: i.DeprecatedLabels,
		})
	}

	if len(i.FeatureLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "New Features",
			Labels: i.FeatureLabels,
		})
	}

	if len(i.EnhancementLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Enhancements",
			Labels: i.EnhancementLabels,
		})
	}

	if len(i.BugLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Fixed Bugs",
			Labels: i.BugLabels,
		})
	}

	if len(i.SecurityLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Security Fixes",
			Labels: i.SecurityLabels,
		})
	}

	return groups
}

// Merges has the specifications for fetching, flitering, and grouping pull/merge requests.
type Merges struct {
	Selection         Selection `yaml:"selection" flag:"merges-selection"`
	Branch            string    `yaml:"branch" flag:"merges-branch"`
	IncludeLabels     []string  `yaml:"include-labels" flag:"merges-include-labels"`
	ExcludeLabels     []string  `yaml:"exclude-labels" flag:"merges-exclude-labels"`
	Grouping          Grouping  `yaml:"grouping" flag:"merges-grouping"`
	SummaryLabels     []string  `yaml:"summary-labels" flag:"merges-summary-labels"`
	RemovedLabels     []string  `yaml:"removed-labels" flag:"merges-removed-labels"`
	BreakingLabels    []string  `yaml:"breaking-labels" flag:"merges-breaking-labels"`
	DeprecatedLabels  []string  `yaml:"deprecated-labels" flag:"merges-deprecated-labels"`
	FeatureLabels     []string  `yaml:"feature-labels" flag:"merges-feature-labels"`
	EnhancementLabels []string  `yaml:"enhancement-labels" flag:"merges-enhancement-labels"`
	BugLabels         []string  `yaml:"bug-labels" flag:"merges-bug-labels"`
	SecurityLabels    []string  `yaml:"security-labels" flag:"merges-security-labels"`
}

// LabelGroups returns the label groups for merges.
func (m Merges) LabelGroups() []LabelGroup {
	groups := []LabelGroup{}

	if len(m.SummaryLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Release Summary",
			Labels: m.SummaryLabels,
		})
	}

	if len(m.RemovedLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Removed",
			Labels: m.RemovedLabels,
		})
	}

	if len(m.BreakingLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Breaking Changes",
			Labels: m.BreakingLabels,
		})
	}

	if len(m.DeprecatedLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Deprecated",
			Labels: m.DeprecatedLabels,
		})
	}

	if len(m.FeatureLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "New Features",
			Labels: m.FeatureLabels,
		})
	}

	if len(m.EnhancementLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Enhancements",
			Labels: m.EnhancementLabels,
		})
	}

	if len(m.BugLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Fixed Bugs",
			Labels: m.BugLabels,
		})
	}

	if len(m.SecurityLabels) > 0 {
		groups = append(groups, LabelGroup{
			Title:  "Security Fixes",
			Labels: m.SecurityLabels,
		})
	}

	return groups
}

// Content has the specifications for the content of changelogs.
type Content struct {
	ReleaseURL string `yaml:"release-url" flag:"release-url"`
}

// GetReleaseURL returns the actual release url for a tag/release.
func (c Content) GetReleaseURL(tag string) string {
	return strings.Replace(c.ReleaseURL, "{tag}", tag, 1)
}

// Spec has all the specifications required for generating a changelog.
type Spec struct {
	Help    bool    `yaml:"-" flag:"help"`
	Version bool    `yaml:"-" flag:"version"`
	Repo    Repo    `yaml:"-"`
	General General `yaml:"general"`
	Tags    Tags    `yaml:"tags"`
	Issues  Issues  `yaml:"issues"`
	Merges  Merges  `yaml:"merges"`
	Content Content `yaml:"content"`
}

// Default returns specfications with default values.
func Default() Spec {
	return Spec{
		Help:    false,
		Version: false,
		Repo:    Repo{},
		General: General{
			File:    "CHANGELOG.md",
			Base:    "",
			Print:   false,
			Verbose: false,
		},
		Tags: Tags{
			From:         "",
			To:           "",
			Future:       "",
			Exclude:      []string{},
			ExcludeRegex: "",
		},
		Issues: Issues{
			Selection:         SelectionAll,
			IncludeLabels:     nil, // All labels included
			ExcludeLabels:     []string{"duplicate", "invalid", "question", "wontfix"},
			Grouping:          GroupingLabel,
			SummaryLabels:     []string{"summary", "release-summary"},
			RemovedLabels:     []string{"removed"},
			BreakingLabels:    []string{"breaking", "backward-incompatible"},
			DeprecatedLabels:  []string{"deprecated"},
			FeatureLabels:     []string{"feature"},
			EnhancementLabels: []string{"enhancement"},
			BugLabels:         []string{"bug"},
			SecurityLabels:    []string{"security"},
		},
		Merges: Merges{
			Selection:         SelectionAll,
			Branch:            "",  // Default branch
			IncludeLabels:     nil, // All labels
			ExcludeLabels:     nil, // No label excluded
			Grouping:          GroupingSimple,
			SummaryLabels:     []string{},
			RemovedLabels:     []string{},
			BreakingLabels:    []string{},
			DeprecatedLabels:  []string{},
			FeatureLabels:     []string{},
			EnhancementLabels: []string{},
			BugLabels:         []string{},
			SecurityLabels:    []string{},
		},
		Content: Content{
			ReleaseURL: "",
		},
	}
}

// FromFile updates a base spec from a spec file if it exists.
func (s Spec) FromFile() (Spec, error) {
	for _, filename := range specFiles {
		f, err := os.Open(filename)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Spec{}, err
		}
		defer f.Close()

		if err = yaml.NewDecoder(f).Decode(&s); err != nil {
			return Spec{}, err
		}

		return s, nil
	}

	return s, nil
}

// WithRepo adds populates Repo sepcs and returns a new spec object.
func (s Spec) WithRepo(domain, path string) Spec {
	s.Repo = Repo{
		Platform:    Platform(domain),
		Path:        path,
		AccessToken: os.Getenv(envVarName),
	}

	return s
}

// PrintHelp prints the help text.
func (s Spec) PrintHelp() error {
	blue := color.New(color.FgBlue)
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	magenta := color.New(color.FgMagenta)
	red := color.New(color.FgRed)
	white := color.New(color.FgWhite)
	yellow := color.New(color.FgYellow)

	tmpl := template.New("help")
	tmpl = tmpl.Funcs(template.FuncMap{
		"Join":    strings.Join,
		"blue":    blue.Sprintf,
		"green":   green.Sprintf,
		"cyan":    cyan.Sprintf,
		"magenta": magenta.Sprintf,
		"red":     red.Sprintf,
		"white":   white.Sprintf,
		"yellow":  yellow.Sprintf,
	})

	tmpl, err := tmpl.Parse(helpTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, s)
}

func (s Spec) String() string {
	return fmt.Sprintf(format,
		s.Repo.Platform, s.Repo.Path, strings.Repeat("*", len(s.Repo.AccessToken)),
		s.General.File, s.General.Base, s.General.Print, s.General.Verbose,
		s.Tags.From, s.Tags.To, s.Tags.Future, s.Tags.Exclude, s.Tags.ExcludeRegex,
		s.Issues.Selection, s.Issues.IncludeLabels, s.Issues.ExcludeLabels,
		s.Issues.Grouping, s.Issues.SummaryLabels, s.Issues.RemovedLabels, s.Issues.BreakingLabels, s.Issues.DeprecatedLabels, s.Issues.FeatureLabels, s.Issues.EnhancementLabels, s.Issues.BugLabels, s.Issues.SecurityLabels,
		s.Merges.Selection, s.Merges.Branch, s.Merges.IncludeLabels, s.Merges.ExcludeLabels,
		s.Merges.Grouping, s.Merges.SummaryLabels, s.Merges.RemovedLabels, s.Merges.BreakingLabels, s.Merges.DeprecatedLabels, s.Merges.FeatureLabels, s.Merges.EnhancementLabels, s.Merges.BugLabels, s.Merges.SecurityLabels,
		s.Content.ReleaseURL,
	)
}
