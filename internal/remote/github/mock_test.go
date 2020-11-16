package github

import (
	"context"
	"sync"
	"time"

	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/go-github"
)

var (
	gitHubUser1 = github.User{
		ID:      1,
		Login:   "octocat",
		Type:    "User",
		Email:   "octocat@github.com",
		Name:    "The Octocat",
		URL:     "https://api.github.com/users/octocat",
		HTMLURL: "https://github.com/octocat",
	}

	gitHubUser2 = github.User{
		ID:      2,
		Login:   "octodog",
		Type:    "User",
		Email:   "octodog@github.com",
		Name:    "The Octodog",
		URL:     "https://api.github.com/users/octodog",
		HTMLURL: "https://github.com/octodog",
	}

	gitHubUser3 = github.User{
		ID:      3,
		Login:   "octofox",
		Type:    "User",
		Email:   "octofox@github.com",
		Name:    "The Octofox",
		URL:     "https://api.github.com/users/octofox",
		HTMLURL: "https://github.com/octofox",
	}

	gitHubRepository = github.Repository{
		ID:            1296269,
		Name:          "Hello-World",
		FullName:      "octocat/Hello-World",
		Description:   "This your first repo!",
		Topics:        []string{"octocat", "api"},
		Private:       false,
		Fork:          false,
		Archived:      false,
		Disabled:      false,
		DefaultBranch: "main",
		Owner: github.User{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		CreatedAt: parseGitHubTime("2020-01-20T09:00:00Z"),
		UpdatedAt: parseGitHubTime("2020-10-31T14:00:00Z"),
		PushedAt:  parseGitHubTime("2020-10-31T14:00:00Z"),
	}

	gitHubCommit1 = github.Commit{
		SHA: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		Commit: github.RawCommit{
			Message: "Fix all the bugs",
			Author: github.Signature{
				Name:  "The Octocat",
				Email: "octocat@github.com",
				Time:  parseGitHubTime("2020-10-20T19:59:59Z"),
			},
			Committer: github.Signature{
				Name:  "The Octocat",
				Email: "octocat@github.com",
				Time:  parseGitHubTime("2020-10-20T19:59:59Z"),
			},
		},
		Author: github.User{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		Committer: github.User{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
	}

	gitHubCommit2 = github.Commit{
		SHA: "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
		Commit: github.RawCommit{
			Message: "Release v0.1.0",
			Author: github.Signature{
				Name:  "The Octocat",
				Email: "octocat@github.com",
				Time:  parseGitHubTime("2020-10-27T23:59:59Z"),
			},
			Committer: github.Signature{
				Name:  "The Octocat",
				Email: "octocat@github.com",
				Time:  parseGitHubTime("2020-10-27T23:59:59Z"),
			},
		},
		Author: github.User{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		Committer: github.User{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		Parents: []github.Hash{
			{
				SHA: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
				URL: "https://api.github.com/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e",
			},
		},
	}

	gitHubBranch = github.Branch{
		Name:      "main",
		Protected: true,
		Commit:    gitHubCommit2,
	}

	gitHubTag = github.Tag{
		Name: "v0.1.0",
		Commit: github.Hash{
			SHA: "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			URL: "https://api.github.com/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
		},
	}

	gitHubIssue1 = github.Issue{
		ID:     1,
		Number: 1001,
		State:  "open",
		Locked: true,
		Title:  "Found a bug",
		Body:   "This is not working as expected!",
		User: github.User{
			ID:      1,
			Login:   "octocat",
			Type:    "User",
			URL:     "https://api.github.com/users/octocat",
			HTMLURL: "https://github.com/octocat",
		},
		Labels: []github.Label{
			{
				ID:      2000,
				Name:    "bug",
				Default: true,
			},
		},
		Milestone: &github.Milestone{
			ID:     3000,
			Number: 1,
			State:  "open",
			Title:  "v1.0",
		},
		URL:       "https://api.github.com/repos/octocat/Hello-World/issues/1001",
		HTMLURL:   "https://github.com/octocat/Hello-World/issues/1001",
		CreatedAt: parseGitHubTime("2020-10-10T10:00:00Z"),
		UpdatedAt: parseGitHubTime("2020-10-20T20:00:00Z"),
		ClosedAt:  nil,
	}

	gitHubIssue2 = github.Issue{
		ID:     2,
		Number: 1002,
		State:  "closed",
		Locked: false,
		Title:  "Fixed a bug",
		Body:   "I made this to work as expected!",
		User: github.User{
			ID:      2,
			Login:   "octodog",
			Type:    "User",
			URL:     "https://api.github.com/users/octodog",
			HTMLURL: "https://github.com/octodog",
		},
		Labels: []github.Label{
			{
				ID:      2000,
				Name:    "bug",
				Default: true,
			},
		},
		Milestone: &github.Milestone{
			ID:     3000,
			Number: 1,
			State:  "open",
			Title:  "v1.0",
		},
		URL:     "https://api.github.com/repos/octocat/Hello-World/issues/1002",
		HTMLURL: "https://github.com/octocat/Hello-World/pull/1002",
		PullURLs: &github.PullURLs{
			URL: "https://api.github.com/repos/octocat/Hello-World/pulls/1002",
		},
		CreatedAt: parseGitHubTime("2020-10-15T15:00:00Z"),
		UpdatedAt: parseGitHubTime("2020-10-22T22:00:00Z"),
		ClosedAt:  parseGitHubTimePtr("2020-10-20T20:00:00Z"),
	}

	gitHubEvent1 = github.Event{
		ID:       1,
		Event:    "closed",
		CommitID: "",
		Actor: github.User{
			ID:      1,
			Login:   "octocat",
			Type:    "User",
			URL:     "https://api.github.com/users/octocat",
			HTMLURL: "https://github.com/octocat",
		},
		CreatedAt: parseGitHubTime("2020-10-20T20:00:00Z"),
	}

	gitHubEvent2 = github.Event{
		ID:       2,
		Event:    "merged",
		CommitID: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		Actor: github.User{
			ID:      3,
			Login:   "octofox",
			Type:    "User",
			URL:     "https://api.github.com/users/octofox",
			HTMLURL: "https://github.com/octofox",
		},
		CreatedAt: parseGitHubTime("2020-10-20T20:00:00Z"),
	}

	remoteUser1 = remote.User{
		Name:     "The Octocat",
		Email:    "octocat@github.com",
		Username: "octocat",
		WebURL:   "https://github.com/octocat",
	}

	remoteUser2 = remote.User{
		Name:     "The Octodog",
		Email:    "octodog@github.com",
		Username: "octodog",
		WebURL:   "https://github.com/octodog",
	}

	remoteUser3 = remote.User{
		Name:     "The Octofox",
		Email:    "octofox@github.com",
		Username: "octofox",
		WebURL:   "https://github.com/octofox",
	}

	remoteCommit1 = remote.Commit{
		Hash: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		Time: parseGitHubTime("2020-10-20T19:59:59Z"),
	}

	remoteCommit2 = remote.Commit{
		Hash: "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
		Time: parseGitHubTime("2020-10-27T23:59:59Z"),
	}

	remoteBranch = remote.Branch{
		Name:   "main",
		Commit: remoteCommit2,
	}

	remoteTag = remote.Tag{
		Name:   "v0.1.0",
		Time:   parseGitHubTime("2020-10-27T23:59:59Z"),
		Commit: remoteCommit2,
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.1.0",
	}

	remoteIssue = remote.Issue{
		Change: remote.Change{
			Number:    1001,
			Title:     "Found a bug",
			Labels:    []string{"bug"},
			Milestone: "v1.0",
			Time:      time.Time{},
			Author:    remoteUser1,
			WebURL:    "https://github.com/octocat/Hello-World/issues/1001",
		},
		Closer: remoteUser1,
	}

	remoteMerge = remote.Merge{
		Change: remote.Change{
			Number:    1002,
			Title:     "Fixed a bug",
			Labels:    []string{"bug"},
			Milestone: "v1.0",
			Time:      parseGitHubTime("2020-10-20T19:59:59Z"),
			Author:    remoteUser2,
			WebURL:    "https://github.com/octocat/Hello-World/pull/1002",
		},
		Merger: remoteUser3,
		Commit: remoteCommit1,
	}
)

func parseGitHubTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	return t
}

func parseGitHubTimePtr(s string) *time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	return &t
}

type (
	EnsureScopesMock struct {
		InContext context.Context
		InScopes  []github.Scope
		OutError  error
	}

	MockGithubService struct {
		EnsureScopesIndex int
		EnsureScopesMocks []EnsureScopesMock
	}
)

func (m *MockGithubService) EnsureScopes(ctx context.Context, scopes ...github.Scope) error {
	i := m.EnsureScopesIndex
	m.EnsureScopesIndex++
	m.EnsureScopesMocks[i].InContext = ctx
	m.EnsureScopesMocks[i].InScopes = scopes
	return m.EnsureScopesMocks[i].OutError
}

type (
	GetUserMock struct {
		InContext   context.Context
		InUsername  string
		OutUser     *github.User
		OutResponse *github.Response
		OutError    error
	}

	MockUsersService struct {
		GetIndex int
		GetMocks []GetUserMock
	}
)

func (m *MockUsersService) Get(ctx context.Context, username string) (*github.User, *github.Response, error) {
	i := m.GetIndex
	m.GetIndex++
	m.GetMocks[i].InContext = ctx
	m.GetMocks[i].InUsername = username
	return m.GetMocks[i].OutUser, m.GetMocks[i].OutResponse, m.GetMocks[i].OutError
}

type (
	GetRepoMock struct {
		InContext     context.Context
		OutRepository *github.Repository
		OutResponse   *github.Response
		OutError      error
	}

	CommitMock struct {
		InContext   context.Context
		InRef       string
		OutCommit   *github.Commit
		OutResponse *github.Response
		OutError    error
	}

	CommitsMock struct {
		InContext   context.Context
		InPageSize  int
		InPageNo    int
		OutCommits  []github.Commit
		OutResponse *github.Response
		OutError    error
	}

	BranchMock struct {
		InContext   context.Context
		InName      string
		OutBranch   *github.Branch
		OutResponse *github.Response
		OutError    error
	}

	TagsMock struct {
		InContext   context.Context
		InPageSize  int
		InPageNo    int
		OutTags     []github.Tag
		OutResponse *github.Response
		OutError    error
	}

	IssuesMock struct {
		InContext   context.Context
		InPageSize  int
		InPageNo    int
		InParams    github.IssuesParams
		OutIssues   []github.Issue
		OutResponse *github.Response
		OutError    error
	}

	EventsMock struct {
		InContext   context.Context
		InNumber    int
		InPageSize  int
		InPageNo    int
		OutEvents   []github.Event
		OutResponse *github.Response
		OutError    error
	}

	MockRepoService struct {
		GetIndex int
		GetMocks []GetRepoMock

		CommitIndex int
		CommitMocks []CommitMock

		CommitsIndex int
		CommitsMocks []CommitsMock

		BranchIndex int
		BranchMocks []BranchMock

		TagsIndex int
		TagsMocks []TagsMock

		IssuesIndex int
		IssuesMocks []IssuesMock

		EventsMutex sync.Mutex
		EventsIndex int
		EventsMocks []EventsMock
	}
)

func (m *MockRepoService) Get(ctx context.Context) (*github.Repository, *github.Response, error) {
	i := m.GetIndex
	m.GetIndex++
	m.GetMocks[i].InContext = ctx
	return m.GetMocks[i].OutRepository, m.GetMocks[i].OutResponse, m.GetMocks[i].OutError
}

func (m *MockRepoService) Commit(ctx context.Context, ref string) (*github.Commit, *github.Response, error) {
	i := m.CommitIndex
	m.CommitIndex++
	m.CommitMocks[i].InContext = ctx
	m.CommitMocks[i].InRef = ref
	return m.CommitMocks[i].OutCommit, m.CommitMocks[i].OutResponse, m.CommitMocks[i].OutError
}

func (m *MockRepoService) Commits(ctx context.Context, pageSize, pageNo int) ([]github.Commit, *github.Response, error) {
	i := m.CommitsIndex
	m.CommitsIndex++
	m.CommitsMocks[i].InContext = ctx
	m.CommitsMocks[i].InPageSize = pageSize
	m.CommitsMocks[i].InPageNo = pageNo
	return m.CommitsMocks[i].OutCommits, m.CommitsMocks[i].OutResponse, m.CommitsMocks[i].OutError
}

func (m *MockRepoService) Branch(ctx context.Context, name string) (*github.Branch, *github.Response, error) {
	i := m.BranchIndex
	m.BranchIndex++
	m.BranchMocks[i].InContext = ctx
	m.BranchMocks[i].InName = name
	return m.BranchMocks[i].OutBranch, m.BranchMocks[i].OutResponse, m.BranchMocks[i].OutError
}

func (m *MockRepoService) Tags(ctx context.Context, pageSize, pageNo int) ([]github.Tag, *github.Response, error) {
	i := m.TagsIndex
	m.TagsIndex++
	m.TagsMocks[i].InContext = ctx
	m.TagsMocks[i].InPageSize = pageSize
	m.TagsMocks[i].InPageNo = pageNo
	return m.TagsMocks[i].OutTags, m.TagsMocks[i].OutResponse, m.TagsMocks[i].OutError
}

func (m *MockRepoService) Issues(ctx context.Context, pageSize, pageNo int, params github.IssuesParams) ([]github.Issue, *github.Response, error) {
	i := m.IssuesIndex
	m.IssuesIndex++
	m.IssuesMocks[i].InContext = ctx
	m.IssuesMocks[i].InPageSize = pageSize
	m.IssuesMocks[i].InPageNo = pageNo
	m.IssuesMocks[i].InParams = params
	return m.IssuesMocks[i].OutIssues, m.IssuesMocks[i].OutResponse, m.IssuesMocks[i].OutError
}

func (m *MockRepoService) Events(ctx context.Context, number, pageSize, pageNo int) ([]github.Event, *github.Response, error) {
	m.EventsMutex.Lock()
	defer m.EventsMutex.Unlock()

	i := m.EventsIndex
	m.EventsIndex++
	m.EventsMocks[i].InContext = ctx
	m.EventsMocks[i].InNumber = number
	m.EventsMocks[i].InPageSize = pageSize
	m.EventsMocks[i].InPageNo = pageNo
	return m.EventsMocks[i].OutEvents, m.EventsMocks[i].OutResponse, m.EventsMocks[i].OutError
}
