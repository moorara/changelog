package spec

import (
	"os"
	"strings"
	"text/template"

	"github.com/go-git/go-git/v5"
)

const helpTemplate = `
  It assumes the remote repository name is origin.

  Supported Remote Repositories:

    • GitHub (github.com)
    • GitLab (gitlab.com)

  Usage: changelog [flags]

  Flags:

    -help                  Show the help text
    -version               Print the version number

    -access-token          The OAuth access token for making API calls
                           The default value is read from the CHANGELOG_ACCESS_TOKEN environment variable

    -file                  The output file for the generated changelog (default: {{.General.File}})
    -base                  An optional file for appending the generated changelog to it (default: {{.General.Base}})
    -print                 Print the generated changelong to STDOUT (default: {{.General.Print}})
    -verbose               Show the vervbosity logs (default: {{.General.Verbose}})

    -branch                TODO: (default: {{.Release.Branch}})
    -from-tag              TODO: (default: {{.Release.FromTag}})
    -to-tag                TODO: (default: {{.Release.ToTag}})
    -future-tag            TODO: (default: {{.Release.FutureTag}})
    -exclude-tags          TODO: (default: {{Join .Release.ExcludeTags ","}})
    -exclude-tags-regex    TODO: (default: {{.Release.ExcludeTagsRegex}})

    -issues                TODO: (values: none|all|labeled) (default: {{.Changes.Issues}})
    -merges                TODO: (values: none|all|labeled) (default: {{.Changes.Merges}})

    -group-by              TODO: (values: simple|label|milestone) (default: {{.Format.GroupBy}})
    -release-url           TODO: (default: {{.Format.ReleaseURL}})

    -include-labels        TODO: (default: {{Join .Labels.Include ","}})
    -exclude-labels        TODO: (default: {{Join .Labels.Exclude ","}})
    -summary-labels        TODO: (default: {{Join .Labels.Summary ","}})
    -removed-labels        TODO: (default: {{Join .Labels.Removed ","}})
    -breaking-labels       TODO: (default: {{Join .Labels.Breaking ","}})
    -deprecated-labels     TODO: (default: {{Join .Labels.Deprecated ","}})
    -feature-labels        TODO: (default: {{Join .Labels.Feature ","}})
    -enhancement-labels    TODO: (default: {{Join .Labels.Enhancement ","}})
    -bug-labels            TODO: (default: {{Join .Labels.Bug ","}})
    -security-labels       TODO: (default: {{Join .Labels.Security ","}})

  Examples:

    changelog
    changelog -access-token=<your-access-token>

`

// Platform is the platform for managing a Git remote repository.
type Platform string

const (
	// PlatformGitHub represents the GitHub platform.
	PlatformGitHub Platform = "github.com"
	// PlatformGitLab represents the GitLab platform.
	PlatformGitLab Platform = "gitlab.com"
)

// Select determines how changes should be selected for a changelog.
type Select string

const (
	// SelectNone does not select any changes.
	SelectNone = Select("none")
	// SelectAll selects all changes.
	SelectAll = Select("all")
	// SelectLabeled selects only those changes that have are labeled.
	SelectLabeled = Select("labeled")
)

// GroupBy determnies how changes are grouped together.
type GroupBy string

const (
	// GroupBySimple groups changes in simple lists.
	GroupBySimple = GroupBy("simple")
	// GroupByLabel groups changes by labels.
	GroupByLabel = GroupBy("label")
	// GroupByMilestone groups changes by milestones.
	GroupByMilestone = GroupBy("milestone")
)

// Spec has all the specifications required for generating a changelog.
type Spec struct {
	Help    bool `flag:"help"`
	Version bool `flag:"version"`

	Repo struct {
		Platform    Platform
		Path        string
		AccessToken string `flag:"access-token"`
	}

	General struct {
		File    string `flag:"file"`
		Base    string `flag:"base"`
		Print   bool   `flag:"print"`
		Verbose bool   `flag:"verbose"`
	}

	Release struct {
		Branch           string   `flag:"branch"`
		FromTag          string   `flag:"from-tag"`
		ToTag            string   `flag:"to-tag"`
		FutureTag        string   `flag:"future-tag"`
		ExcludeTags      []string `flag:"exclude-tags"`
		ExcludeTagsRegex string   `flag:"exclude-tag-regex"`
	}

	Changes struct {
		Issues Select `flag:"issues"`
		Merges Select `flag:"merges"`
	}

	Format struct {
		GroupBy    GroupBy `flag:"group-by"`
		ReleaseURL string  `flag:"release-url"`
	}

	Labels struct {
		Include     []string `flag:"include-labels"`
		Exclude     []string `flag:"exclude-labels"`
		Summary     []string `flag:"summary-labels"`
		Removed     []string `flag:"removed-labels"`
		Breaking    []string `flag:"breaking-labels"`
		Deprecated  []string `flag:"deprecated-labels"`
		Feature     []string `flag:"feature-labels"`
		Enhancement []string `flag:"enhancement-labels"`
		Bug         []string `flag:"bug-labels"`
		Security    []string `flag:"security-labels"`
	}
}

// Default returns specfications with default values.
// The default access token will be read from the CHANGELOG_ACCESS_TOKEN environment variable (if set).
func Default(repo *git.Repository) (*Spec, error) {
	accessToken := os.Getenv("CHANGELOG_ACCESS_TOKEN")

	repoDomain, repoPath, err := getGitRemoteInfo(repo)
	if err != nil {
		return nil, err
	}

	spec := &Spec{
		Help:    false,
		Version: false,
	}

	spec.Repo.Platform = Platform(repoDomain)
	spec.Repo.Path = repoPath
	spec.Repo.AccessToken = accessToken

	spec.General.File = "CHANGELOG.md"
	spec.General.Base = "HISTORY.md"
	spec.General.Print = false
	spec.General.Verbose = false

	spec.Release.Branch = "master"
	spec.Release.FromTag = ""
	spec.Release.ToTag = ""
	spec.Release.FutureTag = ""
	spec.Release.ExcludeTags = []string{}
	spec.Release.ExcludeTagsRegex = ""

	spec.Changes.Issues = SelectAll
	spec.Changes.Merges = SelectAll

	spec.Format.GroupBy = GroupByLabel
	spec.Format.ReleaseURL = ""

	spec.Labels.Include = []string{} // All labels
	spec.Labels.Exclude = []string{"duplicate", "invalid", "question", "wontfix"}
	spec.Labels.Summary = []string{"summary", "release-summary"}
	spec.Labels.Removed = []string{"removed"}
	spec.Labels.Breaking = []string{"breaking", "backward-incompatible", "backwards-incompatible"}
	spec.Labels.Deprecated = []string{"deprecated"}
	spec.Labels.Feature = []string{"feature"}
	spec.Labels.Enhancement = []string{"enhancement"}
	spec.Labels.Bug = []string{"bug"}
	spec.Labels.Security = []string{"security"}

	return spec, nil
}

// PrintHelp prints the help text.
func (s *Spec) PrintHelp() error {
	tmpl := template.New("help")
	tmpl = tmpl.Funcs(template.FuncMap{
		"Join": strings.Join,
	})

	tmpl, err := tmpl.Parse(helpTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(os.Stdout, s)
}
