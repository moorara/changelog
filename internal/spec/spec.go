package spec

import (
	"os"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/moorara/changelog/internal/git"
)

var specFiles = []string{"changelog.yml", "changelog.yaml"}

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

    -issues                TODO: (values: none|all|labeled) (default: {{.Changes.Issues}})
    -merges                TODO: (values: none|all|labeled) (default: {{.Changes.Merges}})

    -branch                TODO: (default: {{.Release.Branch}})
    -from-tag              TODO: (default: {{.Release.FromTag}})
    -to-tag                TODO: (default: {{.Release.ToTag}})
    -future-tag            TODO: (default: {{.Release.FutureTag}})
    -exclude-tags          TODO: (default: {{Join .Release.ExcludeTags ","}})
    -exclude-tags-regex    TODO: (default: {{.Release.ExcludeTagsRegex}})

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
	Help    bool `yaml:"-" flag:"help"`
	Version bool `yaml:"-" flag:"version"`

	Repo struct {
		Platform    Platform `yaml:"-"`
		Path        string   `yaml:"-"`
		AccessToken string   `yaml:"-" flag:"access-token"`
	} `yaml:"-"`

	General struct {
		File    string `yaml:"file" flag:"file"`
		Base    string `yaml:"base" flag:"base"`
		Print   bool   `yaml:"print" flag:"print"`
		Verbose bool   `yaml:"verbose" flag:"verbose"`
	} `yaml:"general"`

	Changes struct {
		Issues Select `yaml:"issues" flag:"issues"`
		Merges Select `yaml:"merges" flag:"merges"`
	} `yaml:"changes"`

	Release struct {
		Branch           string   `yaml:"branch" flag:"branch"`
		FromTag          string   `yaml:"from-tag" flag:"from-tag"`
		ToTag            string   `yaml:"to-tag" flag:"to-tag"`
		FutureTag        string   `yaml:"future-tag" flag:"future-tag"`
		ExcludeTags      []string `yaml:"exclude-tags" flag:"exclude-tags"`
		ExcludeTagsRegex string   `yaml:"exclude-tags-regex" flag:"exclude-tags-regex"`
	} `yaml:"release"`

	Format struct {
		GroupBy    GroupBy `yaml:"group-by" flag:"group-by"`
		ReleaseURL string  `yaml:"release-url" flag:"release-url"`
	} `yaml:"format"`

	Labels struct {
		Include     []string `yaml:"include" flag:"include-labels"`
		Exclude     []string `yaml:"exclude" flag:"exclude-labels"`
		Summary     []string `yaml:"summary" flag:"summary-labels"`
		Removed     []string `yaml:"removed" flag:"removed-labels"`
		Breaking    []string `yaml:"breaking" flag:"breaking-labels"`
		Deprecated  []string `yaml:"deprecated" flag:"deprecated-labels"`
		Feature     []string `yaml:"feature" flag:"feature-labels"`
		Enhancement []string `yaml:"enhancement" flag:"enhancement-labels"`
		Bug         []string `yaml:"bug" flag:"bug-labels"`
		Security    []string `yaml:"security" flag:"security-labels"`
	} `yaml:"labels"`
}

// Default returns specfications with default values.
// The default access token will be read from the CHANGELOG_ACCESS_TOKEN environment variable (if set).
func Default(gitRepo *git.Repo) (Spec, error) {
	accessToken := os.Getenv("CHANGELOG_ACCESS_TOKEN")

	domain, path, err := gitRepo.GetRemoteInfo()
	if err != nil {
		return Spec{}, err
	}

	spec := Spec{
		Help:    false,
		Version: false,
	}

	spec.Repo.Platform = Platform(domain)
	spec.Repo.Path = path
	spec.Repo.AccessToken = accessToken

	spec.General.File = "CHANGELOG.md"
	spec.General.Base = "HISTORY.md"
	spec.General.Print = false
	spec.General.Verbose = false

	spec.Changes.Issues = SelectAll
	spec.Changes.Merges = SelectAll

	spec.Release.Branch = "master"
	spec.Release.FromTag = ""
	spec.Release.ToTag = ""
	spec.Release.FutureTag = ""
	spec.Release.ExcludeTags = []string{}
	spec.Release.ExcludeTagsRegex = ""

	spec.Format.GroupBy = GroupByLabel
	spec.Format.ReleaseURL = ""

	spec.Labels.Include = []string{} // All labels
	spec.Labels.Exclude = []string{"duplicate", "invalid", "question", "wontfix"}
	spec.Labels.Summary = []string{"summary", "release-summary"}
	spec.Labels.Removed = []string{"removed"}
	spec.Labels.Breaking = []string{"breaking"}
	spec.Labels.Deprecated = []string{"deprecated"}
	spec.Labels.Feature = []string{"feature"}
	spec.Labels.Enhancement = []string{"enhancement"}
	spec.Labels.Bug = []string{"bug"}
	spec.Labels.Security = []string{"security"}

	return spec, nil
}

// FromFile updates a base spec from a spec file if it exists.
func FromFile(s Spec) (Spec, error) {
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

// PrintHelp prints the help text.
func (s Spec) PrintHelp() error {
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
