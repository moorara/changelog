package changelog

import (
	"fmt"
	"time"
)

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
	Releases []Release
}

// Release represents a single release of a repository in a changelog.
type Release struct {
	GitTag      string
	URL         string
	Timestamp   time.Time
	IssueGroups []IssueGroup
	MergeGroups []MergeGroup
}

func (r Release) String() string {
	return fmt.Sprintf("%s %s %s\n%v\n%v",
		r.GitTag,
		r.URL,
		r.Timestamp.Format(time.RFC3339),
		r.IssueGroups, r.MergeGroups,
	)
}

// IssueGroup represents a group of issues.
type IssueGroup struct {
	Title  string
	Issues []Issue
}

// Issue represents a single issue.
type Issue struct {
	Number   uint
	Title    string
	Author   User
	ClosedBy User
}

// MergeGroup represents a group of pull/merge/change requests.
type MergeGroup struct {
	Title  string
	Merges []Merge
}

// Merge represents a single pull/merge/change request.
type Merge struct {
	Number   uint
	Title    string
	Author   User
	MergedBy User
}

// User represents a user.
type User struct {
	Username string
	Name     string
}

// NewChangelog creates a new empty default changelog.
func NewChangelog() *Changelog {
	return &Changelog{
		Title: "Changelog",
	}
}
