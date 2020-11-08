package github

import (
	"fmt"
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

	branch struct {
		Name      string `json:"name"`
		Protected bool   `json:"protected"`
		Commit    commit `json:"commit"`
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

func toUser(u user) remote.User {
	return remote.User{
		Name:     u.Name,
		Email:    u.Email,
		Username: u.Login,
		WebURL:   u.HTMLURL,
	}
}

func toCommit(c commit) remote.Commit {
	return remote.Commit{
		Hash: c.SHA,
		Time: c.Commit.Committer.Time,
	}
}

func toBranch(b branch) remote.Branch {
	return remote.Branch{
		Name:   b.Name,
		Commit: toCommit(b.Commit),
	}
}

func toTag(t tag, c commit, repoPath string) remote.Tag {
	return remote.Tag{
		Name:   t.Name,
		Time:   c.Commit.Committer.Time,
		Commit: toCommit(c),
		WebURL: fmt.Sprintf("https://github.com/%s/tree/%s", repoPath, t.Name),
	}
}

func toIssue(i issue, e event, author, closer user) remote.Issue {
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

	return remote.Issue{
		Change: remote.Change{
			Number:    i.Number,
			Title:     i.Title,
			Labels:    labels,
			Milestone: milestone,
			Time:      time,
			Author:    toUser(author),
			WebURL:    i.HTMLURL,
		},
		Closer: toUser(closer),
	}
}

func toMerge(i issue, e event, c commit, author, merger user) remote.Merge {
	// e is the merged event of the pull request

	labels := make([]string, len(i.Labels))
	for i, l := range i.Labels {
		labels[i] = l.Name
	}

	var milestone string
	if i.Milestone != nil {
		milestone = i.Milestone.Title
	}

	// p.MergedAt and e.CreatedAt are the same times
	// c.Commit.Committer.Time is the actual time of merge
	time := c.Commit.Committer.Time

	return remote.Merge{
		Change: remote.Change{
			Number:    i.Number,
			Title:     i.Title,
			Labels:    labels,
			Milestone: milestone,
			Time:      time,
			Author:    toUser(author),
			WebURL:    i.HTMLURL,
		},
		Merger: toUser(merger),
		Commit: toCommit(c),
	}
}

func resolveTags(gitHubTags *tagStore, gitHubCommits *commitStore, repoPath string) remote.Tags {
	tags := remote.Tags{}

	_ = gitHubTags.ForEach(func(name string, t tag) error {
		if c, ok := gitHubCommits.Load(t.Commit.SHA); ok {
			tags = append(tags, toTag(t, c, repoPath))
		}
		return nil
	})

	return tags
}

func resolveIssuesAndMerges(gitHubIssues *issueStore, gitHubEvents *eventStore, gitHubCommits *commitStore, gitHubUsers *userStore) (remote.Issues, remote.Merges) {
	issues := remote.Issues{}
	merges := remote.Merges{}

	_ = gitHubIssues.ForEach(func(num int, i issue) error {
		if i.PullRequest == nil { // Issue
			e, _ := gitHubEvents.Load(num)
			author, _ := gitHubUsers.Load(i.User.Login)
			closer, _ := gitHubUsers.Load(e.Actor.Login)
			issues = append(issues, toIssue(i, e, author, closer))
		} else { // Pull request
			// If no event found, the pull request is closed without being merged
			if e, ok := gitHubEvents.Load(num); ok {
				c, _ := gitHubCommits.Load(e.CommitID)
				author, _ := gitHubUsers.Load(i.User.Login)
				merger, _ := gitHubUsers.Load(e.Actor.Login)
				merges = append(merges, toMerge(i, e, c, author, merger))
			}
		}

		return nil
	})

	issues = issues.Sort()
	merges = merges.Sort()

	return issues, merges
}
