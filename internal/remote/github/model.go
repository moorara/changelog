package github

import (
	"time"

	"github.com/moorara/changelog/internal/remote"
)

// scope represents a GitHub authorization scope.
// See https://docs.github.com/en/developers/apps/scopes-for-oauth-apps
type scope string

const (
	// scopeRepo grants full access to private and public repositories. It also grants ability to manage user projects.
	scopeRepo scope = "repo"
)

type user struct {
	ID         int       `json:"id"`
	Login      string    `json:"login"`
	Type       string    `json:"type"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	HTMLURL    string    `json:"html_url"`
	OrgsURL    string    `json:"organizations_url"`
	AvatarURL  string    `json:"avatar_url"`
	GravatarID string    `json:"gravatar_id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type label struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Default     bool   `json:"default"`
	URL         string `json:"url"`
}

type milestone struct {
	ID           int       `json:"id"`
	Number       int       `json:"number"`
	State        string    `json:"state"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Creator      user      `json:"creator"`
	OpenIssues   int       `json:"open_issues"`
	ClosedIssues int       `json:"closed_issues"`
	DueOn        time.Time `json:"due_on"`
	URL          string    `json:"url"`
	HTMLURL      string    `json:"html_url"`
	LabelsURL    string    `json:"labels_url"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	ClosedAt     time.Time `json:"closed_at"`
}

type pullRequestURLs struct {
	URL      string `json:"url"`
	HTMLURL  string `json:"html_url"`
	DiffURL  string `json:"diff_url"`
	PatchURL string `json:"patch_url"`
}

type issue struct {
	ID          int              `json:"id"`
	Number      int              `json:"number"`
	State       string           `json:"state"`
	Title       string           `json:"title"`
	Body        string           `json:"body"`
	Locked      bool             `json:"locked"`
	User        user             `json:"user"`
	Labels      []label          `json:"labels"`
	Milestone   milestone        `json:"milestone"`
	URL         string           `json:"url"`
	HTMLURL     string           `json:"html_url"`
	LabelsURL   string           `json:"labels_url"`
	PullRequest *pullRequestURLs `json:"pull_request"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	ClosedAt    *time.Time       `json:"closed_at"`
}

type reference struct {
	Label string `json:"label"`
	Ref   string `json:"ref"`
	Sha   string `json:"sha"`
}

type pullRequest struct {
	ID          int       `json:"id"`
	Number      int       `json:"number"`
	State       string    `json:"state"`
	Title       string    `json:"title"`
	Body        string    `json:"body"`
	User        user      `json:"user"`
	Draft       bool      `json:"draft"`
	Locked      bool      `json:"locked"`
	Labels      []label   `json:"labels"`
	Milestone   milestone `json:"milestone"`
	Head        reference `json:"head"`
	Base        reference `json:"base"`
	URL         string    `json:"url"`
	HTMLURL     string    `json:"html_url"`
	DiffURL     string    `json:"diff_url"`
	PatchURL    string    `json:"patch_url"`
	IssueURL    string    `json:"issue_url"`
	CommitsURL  string    `json:"commits_url"`
	StatusesURL string    `json:"statuses_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	ClosedAt    time.Time `json:"closed_at"`
	MergedAt    time.Time `json:"merged_at"`
}

func toChange(gitHubIssue issue) remote.Change {
	labels := make([]string, len(gitHubIssue.Labels))
	for i, l := range gitHubIssue.Labels {
		labels[i] = l.Name
	}

	return remote.Change{
		Number:    gitHubIssue.Number,
		Title:     gitHubIssue.Title,
		Labels:    labels,
		Milestone: gitHubIssue.Milestone.Title,
	}
}

func partitionIssuesAndMerges(gitHubIssues []issue) ([]remote.Change, []remote.Change) {
	issues := []remote.Change{}
	merges := []remote.Change{}

	for _, i := range gitHubIssues {
		if i.PullRequest == nil {
			issues = append(issues, toChange(i))
		} else {
			merges = append(merges, toChange(i))
		}
	}

	return issues, merges
}
