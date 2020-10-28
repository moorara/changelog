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
		SHA   string `json:"sha"`
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
		Merged         bool       `json:"merged"`
		Mergeable      *bool      `json:"mergeable"`
		Rebaseable     *bool      `json:"rebaseable"`
		MergedBy       *user      `json:"merged_by"`
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

	gitSignature struct {
		Name  string    `json:"name"`
		Email string    `json:"email"`
		Time  time.Time `json:"date"`
	}

	gitCommit struct {
		Message   string       `json:"message"`
		Author    gitSignature `json:"author"`
		Committer gitSignature `json:"committer"`
		URL       string       `json:"url"`
	}

	commit struct {
		SHA       string    `json:"sha"`
		Commit    gitCommit `json:"commit"`
		Author    user      `json:"author"`
		Committer user      `json:"committer"`
		URL       string    `json:"url"`
		HTMLURL   string    `json:"html_url"`
	}
)

func issueToChange(i issue, e event, creator, closer user) remote.Change {
	// e is the closed event of the issue

	labels := make([]string, len(i.Labels))
	for i, l := range i.Labels {
		labels[i] = l.Name
	}

	var milestone string
	if i.Milestone != nil {
		milestone = i.Milestone.Title
	}

	// *i.ClosedAt and e.CreatedAt are the same times
	var time time.Time
	if i.ClosedAt != nil {
		time = *i.ClosedAt
	}

	return remote.Change{
		Number:    i.Number,
		Title:     i.Title,
		Labels:    labels,
		Milestone: milestone,
		Time:      time,
		Creator: remote.User{
			Name:     creator.Name,
			Email:    creator.Email,
			Username: creator.Login,
			URL:      creator.URL,
		},
		CloserOrMerger: remote.User{
			Name:     closer.Name,
			Email:    closer.Email,
			Username: closer.Login,
			URL:      closer.URL,
		},
	}
}

func pullToChange(p pullRequest, e event, c commit, creator, merger user) remote.Change {
	// e is the merged event of the pull request

	labels := make([]string, len(p.Labels))
	for i, l := range p.Labels {
		labels[i] = l.Name
	}

	var milestone string
	if p.Milestone != nil {
		milestone = p.Milestone.Title
	}

	// *p.MergedAt and e.CreatedAt are the same times
	// c.Commit.Author.Time is the actual time of merge
	time := c.Commit.Author.Time

	return remote.Change{
		Number:    p.Number,
		Title:     p.Title,
		Labels:    labels,
		Milestone: milestone,
		Time:      time,
		Creator: remote.User{
			Name:     creator.Name,
			Email:    creator.Email,
			Username: creator.Login,
			URL:      creator.URL,
		},
		CloserOrMerger: remote.User{
			Name:     merger.Name,
			Email:    merger.Email,
			Username: merger.Login,
			URL:      merger.URL,
		},
	}
}

func resolveIssuesAndMerges(
	gitHubIssues *issueStore,
	gitHubPulls *pullRequestStore,
	gitHubEvents *eventStore,
	gitHubCommits *commitStore,
	gitHubUsers *userStore,
) (remote.Changes, remote.Changes) {
	issues := remote.Changes{}
	merges := remote.Changes{}

	_ = gitHubIssues.ForEach(func(num int, i issue) error {
		e, _ := gitHubEvents.Load(num)
		creator, _ := gitHubUsers.Load(i.User.Login)

		if i.PullRequest == nil {
			closer, _ := gitHubUsers.Load(e.Actor.Login)
			issues = append(issues, issueToChange(i, e, creator, closer))
		} else {
			// Every GitHub pull request is also an issue (see https://docs.github.com/en/rest/reference/issues#list-repository-issues)
			// For every closed issue that is a pull request, we also need to make sure it is merged.
			if p, ok := gitHubPulls.Load(num); ok {
				// All Pulls expected to be merged at this point
				c, _ := gitHubCommits.Load(e.CommitID)
				merger, _ := gitHubUsers.Load(e.Actor.Login)
				merges = append(merges, pullToChange(p, e, c, creator, merger))
			}
		}

		return nil
	})

	issues = issues.Sort()
	merges = merges.Sort()

	return issues, merges
}
