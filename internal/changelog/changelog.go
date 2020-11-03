package changelog

import "time"

// Processor is an abstraction for reading and writing changelogs.
type Processor interface {
	Parse(ParseOptions) (*Changelog, error)
	Render(*Changelog) (string, error)
}

// ParseOptions determines how a changelog file should be parsed.
type ParseOptions struct{}

// Changelog represents the entire changelog of a repository.
type Changelog struct {
	Title    string
	New      []Release
	Existing []Release
}

// Release represents a single release of a repository in a changelog.
type Release struct {
	TagName     string
	TagURL      string
	TagTime     time.Time
	CompareURL  string
	IssueGroups []IssueGroup
	MergeGroups []MergeGroup
}

// IssueGroup represents a group of issues.
type IssueGroup struct {
	Title  string
	Issues []Issue
}

// Issue represents a single issue.
type Issue struct {
	Number   int
	Title    string
	URL      string
	OpenedBy User
	ClosedBy User
}

// MergeGroup represents a group of pull/merge requests.
type MergeGroup struct {
	Title  string
	Merges []Merge
}

// Merge represents a single pull/merge request.
type Merge struct {
	Number   int
	Title    string
	URL      string
	OpenedBy User
	MergedBy User
}

// User represents a user.
type User struct {
	Name     string
	Username string
	URL      string
}

// NewChangelog creates a new empty default changelog.
func NewChangelog() *Changelog {
	return &Changelog{
		Title: "Changelog",
	}
}
