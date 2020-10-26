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

type (
	user struct {
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

	label struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Default     bool   `json:"default"`
		URL         string `json:"url"`
	}

	milestone struct {
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

	pullRequestURLs struct {
		URL      string `json:"url"`
		HTMLURL  string `json:"html_url"`
		DiffURL  string `json:"diff_url"`
		PatchURL string `json:"patch_url"`
	}

	issue struct {
		ID          int              `json:"id"`
		Number      int              `json:"number"`
		State       string           `json:"state"`
		Locked      bool             `json:"locked"`
		Title       string           `json:"title"`
		Body        string           `json:"body"`
		User        user             `json:"user"`
		Labels      []label          `json:"labels"`
		Milestone   *milestone       `json:"milestone"`
		URL         string           `json:"url"`
		HTMLURL     string           `json:"html_url"`
		LabelsURL   string           `json:"labels_url"`
		PullRequest *pullRequestURLs `json:"pull_request"`
		CreatedAt   time.Time        `json:"created_at"`
		UpdatedAt   time.Time        `json:"updated_at"`
		ClosedAt    *time.Time       `json:"closed_at"`
	}

	reference struct {
		Label string `json:"label"`
		Ref   string `json:"ref"`
		Sha   string `json:"sha"`
	}

	pullRequest struct {
		ID             int        `json:"id"`
		Number         int        `json:"number"`
		State          string     `json:"state"`
		Draft          bool       `json:"draft"`
		Locked         bool       `json:"locked"`
		Title          string     `json:"title"`
		Body           string     `json:"body"`
		User           user       `json:"user"`
		Labels         []label    `json:"labels"`
		Milestone      *milestone `json:"milestone"`
		Base           reference  `json:"base"`
		Head           reference  `json:"head"`
		MergeCommitSHA string     `json:"merge_commit_sha"`
		URL            string     `json:"url"`
		HTMLURL        string     `json:"html_url"`
		DiffURL        string     `json:"diff_url"`
		PatchURL       string     `json:"patch_url"`
		IssueURL       string     `json:"issue_url"`
		CommitsURL     string     `json:"commits_url"`
		StatusesURL    string     `json:"statuses_url"`
		CreatedAt      time.Time  `json:"created_at"`
		UpdatedAt      time.Time  `json:"updated_at"`
		ClosedAt       *time.Time `json:"closed_at"`
		MergedAt       *time.Time `json:"merged_at"`
	}

	event struct {
		ID        int       `json:"id"`
		Event     string    `json:"event"`
		CommitID  string    `json:"commit_id"`
		Actor     user      `json:"actor"`
		URL       string    `json:"url"`
		CommitURL string    `json:"commit_url"`
		CreatedAt time.Time `json:"created_at"`
	}
)

func issueToChange(i issue) remote.Change {
	labels := make([]string, len(i.Labels))
	for i, l := range i.Labels {
		labels[i] = l.Name
	}

	var milestone string
	if i.Milestone != nil {
		milestone = i.Milestone.Title
	}

	var timestamp time.Time
	if i.ClosedAt != nil {
		timestamp = *i.ClosedAt
	}

	return remote.Change{
		Number:    i.Number,
		Title:     i.Title,
		Labels:    labels,
		Milestone: milestone,
		Timestamp: timestamp,
	}
}

func pullToChange(p pullRequest) remote.Change {
	labels := make([]string, len(p.Labels))
	for i, l := range p.Labels {
		labels[i] = l.Name
	}

	var milestone string
	if p.Milestone != nil {
		milestone = p.Milestone.Title
	}

	var timestamp time.Time
	if p.MergedAt != nil {
		timestamp = *p.MergedAt
	}

	return remote.Change{
		Number:    p.Number,
		Title:     p.Title,
		Labels:    labels,
		Milestone: milestone,
		Timestamp: timestamp,
	}
}

func partitionIssuesAndMerges(gitHubIssues map[int]issue, gitHubPulls map[int]pullRequest) (remote.Changes, remote.Changes) {
	issues := remote.Changes{}
	merges := remote.Changes{}

	for no, i := range gitHubIssues {
		if i.PullRequest == nil {
			issues = append(issues, issueToChange(i))
		} else {
			// Every GitHub pull request is also an issue (see https://docs.github.com/en/rest/reference/issues#list-repository-issues)
			// For every closed issue that is a pull request, we also need to make sure it is merged.
			if p, ok := gitHubPulls[no]; ok {
				if p.MergedAt != nil {
					merges = append(merges, pullToChange(p))
				}
			}
		}
	}

	issues = issues.Sort()
	merges = merges.Sort()

	return issues, merges
}
