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

	repository struct {
		ID            int       `json:"id"`
		Name          string    `json:"name"`
		FullName      string    `json:"full_name"`
		Description   string    `json:"description"`
		Topics        []string  `json:"topics"`
		Private       bool      `json:"private"`
		Fork          bool      `json:"fork"`
		Archived      bool      `json:"archived"`
		Disabled      bool      `json:"disabled"`
		DefaultBranch string    `json:"default_branch"`
		Owner         user      `json:"owner"`
		URL           string    `json:"url"`
		HTMLURL       string    `json:"html_url"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
		PushedAt      time.Time `json:"pushed_at"`
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

	hash struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	}

	signature struct {
		Name  string    `json:"name"`
		Email string    `json:"email"`
		Time  time.Time `json:"date"`
	}

	rawCommit struct {
		Message   string    `json:"message"`
		Author    signature `json:"author"`
		Committer signature `json:"committer"`
		Tree      hash      `json:"tree"`
		URL       string    `json:"url"`
	}

	commit struct {
		SHA       string    `json:"sha"`
		Commit    rawCommit `json:"commit"`
		Author    user      `json:"author"`
		Committer user      `json:"committer"`
		Parents   []hash    `json:"parents"`
		URL       string    `json:"url"`
		HTMLURL   string    `json:"html_url"`
	}

	tag struct {
		Name   string `json:"name"`
		Commit hash   `json:"commit"`
	}

	pullURLs struct {
		URL      string `json:"url"`
		HTMLURL  string `json:"html_url"`
		DiffURL  string `json:"diff_url"`
		PatchURL string `json:"patch_url"`
	}

	issue struct {
		ID          int        `json:"id"`
		Number      int        `json:"number"`
		State       string     `json:"state"`
		Locked      bool       `json:"locked"`
		Title       string     `json:"title"`
		Body        string     `json:"body"`
		User        user       `json:"user"`
		Labels      []label    `json:"labels"`
		Milestone   *milestone `json:"milestone"`
		URL         string     `json:"url"`
		HTMLURL     string     `json:"html_url"`
		LabelsURL   string     `json:"labels_url"`
		PullRequest *pullURLs  `json:"pull_request"`
		CreatedAt   time.Time  `json:"created_at"`
		UpdatedAt   time.Time  `json:"updated_at"`
		ClosedAt    *time.Time `json:"closed_at"`
	}

	reference struct {
		Label string     `json:"label"`
		Ref   string     `json:"ref"`
		SHA   string     `json:"sha"`
		User  user       `json:"user"`
		Repo  repository `json:"repo"`
	}

	pull struct {
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
)

func toTag(t tag, c commit) remote.Tag {
	return remote.Tag{
		Name:       t.Name,
		CommitHash: c.SHA,
		Time:       c.Commit.Committer.Time,
	}
}

func toIssueChange(i issue, e event, creator, closer user) remote.Change {
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

func toPullChange(p pull, e event, c commit, creator, merger user) remote.Change {
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

func resolveTags(gitHubTags *tagStore, gitHubCommits *commitStore) remote.Tags {
	tags := remote.Tags{}

	_ = gitHubTags.ForEach(func(name string, t tag) error {
		if c, ok := gitHubCommits.Load(t.Commit.SHA); ok {
			tags = append(tags, toTag(t, c))
		}
		return nil
	})

	tags = tags.Sort()

	return tags
}

func resolveIssuesAndMerges(
	gitHubIssues *issueStore, gitHubPulls *pullStore,
	gitHubEvents *eventStore, gitHubCommits *commitStore,
	gitHubUsers *userStore,
) (remote.Changes, remote.Changes) {
	issues := remote.Changes{}
	merges := remote.Changes{}

	_ = gitHubIssues.ForEach(func(num int, i issue) error {
		e, _ := gitHubEvents.Load(num)
		creator, _ := gitHubUsers.Load(i.User.Login)

		if i.PullRequest == nil {
			closer, _ := gitHubUsers.Load(e.Actor.Login)
			issues = append(issues, toIssueChange(i, e, creator, closer))
		} else {
			// Every GitHub pull request is also an issue (see https://docs.github.com/en/rest/reference/issues#list-repository-issues)
			// For every closed issue that is a pull request, we also need to make sure it is merged.
			if p, ok := gitHubPulls.Load(num); ok {
				// All Pulls expected to be merged at this point
				c, _ := gitHubCommits.Load(e.CommitID)
				merger, _ := gitHubUsers.Load(e.Actor.Login)
				merges = append(merges, toPullChange(p, e, c, creator, merger))
			}
		}

		return nil
	})

	issues = issues.Sort()
	merges = merges.Sort()

	return issues, merges
}
