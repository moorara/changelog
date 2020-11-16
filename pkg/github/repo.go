package github

import (
	"context"
	"fmt"
	"time"
)

// RepoService provides GitHub APIs for a specific repository.
// See https://docs.github.com/en/rest/reference/repos
type RepoService struct {
	client      *Client
	owner, repo string
}

// Repository is a GitHub repository object.
type Repository struct {
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
	Owner         User      `json:"owner"`
	URL           string    `json:"url"`
	HTMLURL       string    `json:"html_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	PushedAt      time.Time `json:"pushed_at"`
}

type (
	// Hash is a GitHub hash object.
	Hash struct {
		SHA string `json:"sha"`
		URL string `json:"url"`
	}

	// Signature is a GitHub signature object.
	Signature struct {
		Name  string    `json:"name"`
		Email string    `json:"email"`
		Time  time.Time `json:"date"`
	}

	// RawCommit is a GitHub raw commit object.
	RawCommit struct {
		Message   string    `json:"message"`
		Author    Signature `json:"author"`
		Committer Signature `json:"committer"`
		Tree      Hash      `json:"tree"`
		URL       string    `json:"url"`
	}

	// Commit is a GitHub repository commit object.
	Commit struct {
		SHA       string    `json:"sha"`
		Commit    RawCommit `json:"commit"`
		Author    User      `json:"author"`
		Committer User      `json:"committer"`
		Parents   []Hash    `json:"parents"`
		URL       string    `json:"url"`
		HTMLURL   string    `json:"html_url"`
	}
)

// Branch is a GitHub branch object.
type Branch struct {
	Name      string `json:"name"`
	Protected bool   `json:"protected"`
	Commit    Commit `json:"commit"`
}

// Tag is a GitHib tag object.
type Tag struct {
	Name   string `json:"name"`
	Commit Hash   `json:"commit"`
}

// Label is a GitHub label object.
type Label struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Default     bool   `json:"default"`
	URL         string `json:"url"`
}

// Milestone is a GitHub milestone object.
type Milestone struct {
	ID           int        `json:"id"`
	Number       int        `json:"number"`
	State        string     `json:"state"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Creator      User       `json:"creator"`
	OpenIssues   int        `json:"open_issues"`
	ClosedIssues int        `json:"closed_issues"`
	DueOn        *time.Time `json:"due_on"`
	URL          string     `json:"url"`
	HTMLURL      string     `json:"html_url"`
	LabelsURL    string     `json:"labels_url"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ClosedAt     *time.Time `json:"closed_at"`
}

type (
	// PullURLs is an object added to an issue representing a pull request.
	PullURLs struct {
		URL      string `json:"url"`
		HTMLURL  string `json:"html_url"`
		DiffURL  string `json:"diff_url"`
		PatchURL string `json:"patch_url"`
	}

	// Issue is a GitHub issue object.
	Issue struct {
		ID        int        `json:"id"`
		Number    int        `json:"number"`
		State     string     `json:"state"`
		Locked    bool       `json:"locked"`
		Title     string     `json:"title"`
		Body      string     `json:"body"`
		User      User       `json:"user"`
		Labels    []Label    `json:"labels"`
		Milestone *Milestone `json:"milestone"`
		URL       string     `json:"url"`
		HTMLURL   string     `json:"html_url"`
		LabelsURL string     `json:"labels_url"`
		PullURLs  *PullURLs  `json:"pull_request"`
		CreatedAt time.Time  `json:"created_at"`
		UpdatedAt time.Time  `json:"updated_at"`
		ClosedAt  *time.Time `json:"closed_at"`
	}
)

type (
	// PullBranch represents a base or head object in a Pull object.
	PullBranch struct {
		Label string     `json:"label"`
		Ref   string     `json:"ref"`
		SHA   string     `json:"sha"`
		User  User       `json:"user"`
		Repo  Repository `json:"repo"`
	}

	// Pull is a GitHub pull request object.
	Pull struct {
		ID             int        `json:"id"`
		Number         int        `json:"number"`
		State          string     `json:"state"`
		Draft          bool       `json:"draft"`
		Locked         bool       `json:"locked"`
		Title          string     `json:"title"`
		Body           string     `json:"body"`
		User           User       `json:"user"`
		Labels         []Label    `json:"labels"`
		Milestone      *Milestone `json:"milestone"`
		Base           PullBranch `json:"base"`
		Head           PullBranch `json:"head"`
		Merged         bool       `json:"merged"`
		Mergeable      *bool      `json:"mergeable"`
		Rebaseable     *bool      `json:"rebaseable"`
		MergedBy       *User      `json:"merged_by"`
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
)

// Event is a GitHub event object.
type Event struct {
	ID        int       `json:"id"`
	Event     string    `json:"event"`
	CommitID  string    `json:"commit_id"`
	Actor     User      `json:"actor"`
	URL       string    `json:"url"`
	CommitURL string    `json:"commit_url"`
	CreatedAt time.Time `json:"created_at"`
}

// Get retrieves a repository by its name.
// See https://docs.github.com/rest/reference/repos#get-a-repository
func (s *RepoService) Get(ctx context.Context) (*Repository, *Response, error) {
	url := fmt.Sprintf("/repos/%s/%s", s.owner, s.repo)
	req, err := s.client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	repository := new(Repository)

	resp, err := s.client.Do(req, repository)
	if err != nil {
		return nil, nil, err
	}

	return repository, resp, nil
}

// Commit retrieves a commit for a given repository by its reference.
// See https://docs.github.com/rest/reference/repos#get-a-commit
func (s *RepoService) Commit(ctx context.Context, ref string) (*Commit, *Response, error) {
	url := fmt.Sprintf("/repos/%s/%s/commits/%s", s.owner, s.repo, ref)
	req, err := s.client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	commit := new(Commit)

	resp, err := s.client.Do(req, commit)
	if err != nil {
		return nil, nil, err
	}

	return commit, resp, nil
}

// Commits retrieves all commits for a given repository page by page.
// See https://docs.github.com/rest/reference/repos#list-commits
func (s *RepoService) Commits(ctx context.Context, pageSize, pageNo int) ([]Commit, *Response, error) {
	url := fmt.Sprintf("/repos/%s/%s/commits", s.owner, s.repo)
	req, err := s.client.NewPageRequest(ctx, "GET", url, pageSize, pageNo, nil)
	if err != nil {
		return nil, nil, err
	}

	commits := []Commit{}

	resp, err := s.client.Do(req, &commits)
	if err != nil {
		return nil, nil, err
	}

	return commits, resp, nil
}

// Branch retrieves a branch for a given repository by its name.
// See https://docs.github.com/rest/reference/repos#get-a-branch
func (s *RepoService) Branch(ctx context.Context, name string) (*Branch, *Response, error) {
	url := fmt.Sprintf("/repos/%s/%s/branches/%s", s.owner, s.repo, name)
	req, err := s.client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	branch := new(Branch)

	resp, err := s.client.Do(req, branch)
	if err != nil {
		return nil, nil, err
	}

	return branch, resp, nil
}

// Tags retrieves all tags for a given repository page by page.
// This GitHub API is not officially documented.
// For the closest documentation see https://docs.github.com/rest/reference/git#get-a-tag
func (s *RepoService) Tags(ctx context.Context, pageSize, pageNo int) ([]Tag, *Response, error) {
	url := fmt.Sprintf("/repos/%s/%s/tags", s.owner, s.repo)
	req, err := s.client.NewPageRequest(ctx, "GET", url, pageSize, pageNo, nil)
	if err != nil {
		return nil, nil, err
	}

	tags := []Tag{}

	resp, err := s.client.Do(req, &tags)
	if err != nil {
		return nil, nil, err
	}

	return tags, resp, nil
}

// IssuesParams are optional parameters for Issues.
type IssuesParams struct {
	State string
	Since time.Time
}

// Issues retrieves all issues for a given repository page by page.
// See https://docs.github.com/rest/reference/issues#list-repository-issues
func (s *RepoService) Issues(ctx context.Context, pageSize, pageNo int, params IssuesParams) ([]Issue, *Response, error) {
	url := fmt.Sprintf("/repos/%s/%s/issues", s.owner, s.repo)
	req, err := s.client.NewPageRequest(ctx, "GET", url, pageSize, pageNo, nil)
	if err != nil {
		return nil, nil, err
	}

	q := req.URL.Query()

	if params.State != "" {
		q.Add("state", params.State)
	}

	if !params.Since.IsZero() {
		q.Add("since", params.Since.Format(time.RFC3339))
	}

	req.URL.RawQuery = q.Encode()

	issues := []Issue{}

	resp, err := s.client.Do(req, &issues)
	if err != nil {
		return nil, nil, err
	}

	return issues, resp, nil
}

// Pull retrieves a pull request for a given repository by its number.
// See https://docs.github.com/rest/reference/pulls#get-a-pull-request
func (s *RepoService) Pull(ctx context.Context, number int) (*Pull, *Response, error) {
	url := fmt.Sprintf("/repos/%s/%s/pulls/%d", s.owner, s.repo, number)
	req, err := s.client.NewRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	pull := new(Pull)

	resp, err := s.client.Do(req, pull)
	if err != nil {
		return nil, nil, err
	}

	return pull, resp, nil
}

// PullsParams are optional parameters for Pulls.
type PullsParams struct {
	State string
}

// Pulls retrieves all pull requests for a given repository page by page.
// See https://docs.github.com/rest/reference/pulls#list-pull-requests
func (s *RepoService) Pulls(ctx context.Context, pageSize, pageNo int, params PullsParams) ([]Pull, *Response, error) {
	url := fmt.Sprintf("/repos/%s/%s/pulls", s.owner, s.repo)
	req, err := s.client.NewPageRequest(ctx, "GET", url, pageSize, pageNo, nil)
	if err != nil {
		return nil, nil, err
	}

	q := req.URL.Query()

	if params.State != "" {
		q.Add("state", params.State)
	}

	req.URL.RawQuery = q.Encode()

	pulls := []Pull{}

	resp, err := s.client.Do(req, &pulls)
	if err != nil {
		return nil, nil, err
	}

	return pulls, resp, nil
}

// Events retrieves all events for a given repository and an issue page by page.
// See https://docs.github.com/rest/reference/issues#list-issue-events
func (s *RepoService) Events(ctx context.Context, number, pageSize, pageNo int) ([]Event, *Response, error) {
	url := fmt.Sprintf("/repos/%s/%s/issues/%d/events", s.owner, s.repo, number)
	req, err := s.client.NewPageRequest(ctx, "GET", url, pageSize, pageNo, nil)
	if err != nil {
		return nil, nil, err
	}

	events := []Event{}

	resp, err := s.client.Do(req, &events)
	if err != nil {
		return nil, nil, err
	}

	return events, resp, nil
}
