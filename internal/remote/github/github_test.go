package github

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/pkg/log"
	"github.com/moorara/go-github"
)

func TestNewRepo(t *testing.T) {
	tests := []struct {
		name        string
		logger      log.Logger
		repoOwner   string
		repoName    string
		accessToken string
	}{
		{
			name:        "OK",
			logger:      log.New(log.None),
			repoOwner:   "moorara",
			repoName:    "changelog",
			accessToken: "github-access-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRepo(tc.logger, tc.repoOwner, tc.repoName, tc.accessToken)
			assert.NotNil(t, r)

			gr, ok := r.(*repo)
			assert.True(t, ok)

			assert.Equal(t, tc.logger, gr.logger)
			assert.Equal(t, tc.repoOwner, gr.owner)
			assert.Equal(t, tc.repoName, gr.repo)
			assert.NotNil(t, gr.stores.users)
			assert.NotNil(t, gr.stores.commits)
			assert.NotNil(t, gr.services.github)
			assert.NotNil(t, gr.services.users)
			assert.NotNil(t, gr.services.repo)
		})
	}
}

func TestRepo_getUser(t *testing.T) {
	tests := []struct {
		name          string
		usersStore    *store
		usersService  usersService
		ctx           context.Context
		username      string
		expectedUser  github.User
		expectedError string
	}{
		{
			name: "CacheHit",
			usersStore: &store{
				m: map[interface{}]interface{}{
					"octocat": gitHubUser1,
				},
			},
			usersService: nil,
			ctx:          context.Background(),
			username:     "octocat",
			expectedUser: gitHubUser1,
		},
		{
			name: "Error",
			usersStore: &store{
				m: map[interface{}]interface{}{},
			},
			usersService: &MockUsersService{
				GetMocks: []GetUserMock{
					{OutError: errors.New("error on getting github user")},
				},
			},
			ctx:           context.Background(),
			username:      "octocat",
			expectedError: "error on getting github user",
		},
		{
			name: "Success",
			usersStore: &store{
				m: map[interface{}]interface{}{},
			},
			usersService: &MockUsersService{
				GetMocks: []GetUserMock{
					{OutUser: &gitHubUser1, OutResponse: &github.Response{}},
				},
			},
			ctx:          context.Background(),
			username:     "octocat",
			expectedUser: gitHubUser1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{logger: log.New(log.None)}
			r.stores.users = tc.usersStore
			r.services.users = tc.usersService

			user, err := r.getUser(tc.ctx, tc.username)

			if tc.expectedError != "" {
				assert.Empty(t, user)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, user)
			}
		})
	}
}

func TestRepo_getCommit(t *testing.T) {
	tests := []struct {
		name           string
		commitsStore   *store
		repoService    repoService
		ctx            context.Context
		ref            string
		expectedCommit github.Commit
		expectedError  string
	}{
		{
			name: "CacheHit",
			commitsStore: &store{
				m: map[interface{}]interface{}{
					"6dcb09b5b57875f334f61aebed695e2e4193db5e": gitHubCommit1,
				},
			},
			repoService:    nil,
			ctx:            context.Background(),
			ref:            "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedCommit: gitHubCommit1,
		},
		{
			name: "Error",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				CommitMocks: []CommitMock{
					{OutError: errors.New("error on getting github commit")},
				},
			},
			ctx:           context.Background(),
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: "error on getting github commit",
		},
		{
			name: "Success",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				CommitMocks: []CommitMock{
					{OutCommit: &gitHubCommit1, OutResponse: &github.Response{}},
				},
			},
			ctx:            context.Background(),
			ref:            "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedCommit: gitHubCommit1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{logger: log.New(log.None)}
			r.stores.commits = tc.commitsStore
			r.services.repo = tc.repoService

			commit, err := r.getCommit(tc.ctx, tc.ref)

			if tc.expectedError != "" {
				assert.Empty(t, commit)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCommit, commit)
			}
		})
	}
}

func TestRepo_getParentCommits(t *testing.T) {
	tests := []struct {
		name            string
		commitsStore    *store
		repoService     repoService
		ctx             context.Context
		ref             string
		expectedCommits remote.Commits
		expectedError   string
	}{
		{
			name: "CommitFails_1",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				CommitMocks: []CommitMock{
					{OutError: errors.New("error on getting github commit")},
				},
			},
			ctx:           context.Background(),
			ref:           "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			expectedError: "error on getting github commit",
		},
		{
			name: "CommitFails_2",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				CommitMocks: []CommitMock{
					{OutCommit: &gitHubCommit2, OutResponse: &github.Response{}},
					{OutError: errors.New("error on getting github commit")},
				},
			},
			ctx:           context.Background(),
			ref:           "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			expectedError: "error on getting github commit",
		},
		{
			name: "Success",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				CommitMocks: []CommitMock{
					{OutCommit: &gitHubCommit2, OutResponse: &github.Response{}},
					{OutCommit: &gitHubCommit1, OutResponse: &github.Response{}},
				},
			},
			ctx:             context.Background(),
			ref:             "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			expectedCommits: remote.Commits{remoteCommit2, remoteCommit1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{logger: log.New(log.None)}
			r.stores.commits = tc.commitsStore
			r.services.repo = tc.repoService

			commits, err := r.getParentCommits(tc.ctx, tc.ref)

			if tc.expectedError != "" {
				assert.Nil(t, commits)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCommits, commits)
			}
		})
	}
}

func TestRepo_findEvent(t *testing.T) {
	tests := []struct {
		name          string
		repoService   repoService
		ctx           context.Context
		num           int
		eventName     string
		expectedEvent github.Event
		expectedError string
	}{
		{
			name: "Error",
			repoService: &MockRepoService{
				EventsMocks: []EventsMock{
					{OutError: errors.New("error on getting github events")},
				},
			},
			ctx:           context.Background(),
			num:           1001,
			eventName:     "closed",
			expectedError: "error on getting github events",
		},
		{
			name: "Found_FirstPage",
			repoService: &MockRepoService{
				EventsMocks: []EventsMock{
					{
						OutEvents: []github.Event{gitHubEvent1},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
				},
			},
			ctx:           context.Background(),
			num:           1001,
			eventName:     "closed",
			expectedEvent: gitHubEvent1,
		},
		{
			name: "Found_SecondPage",
			repoService: &MockRepoService{
				EventsMocks: []EventsMock{
					{
						OutEvents: []github.Event{},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{
						OutEvents: []github.Event{gitHubEvent1},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 1, Next: 0, Last: 0},
						},
					},
				},
			},
			ctx:           context.Background(),
			num:           1001,
			eventName:     "closed",
			expectedEvent: gitHubEvent1,
		},
		{
			name: "NotFound",
			repoService: &MockRepoService{
				EventsMocks: []EventsMock{
					{
						OutEvents: []github.Event{},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 0, Next: 2, Last: 2},
						},
					},
					{
						OutEvents: []github.Event{},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 1, Next: 0, Last: 2},
						},
					},
				},
			},
			ctx:           context.Background(),
			num:           1001,
			eventName:     "closed",
			expectedEvent: github.Event{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{logger: log.New(log.None)}
			r.services.repo = tc.repoService

			event, err := r.findEvent(tc.ctx, tc.num, tc.eventName)

			if tc.expectedError != "" {
				assert.Empty(t, event)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedEvent, event)
			}
		})
	}
}

func TestRepo_FutureTag(t *testing.T) {
	tests := []struct {
		name            string
		owner           string
		repo            string
		tagName         string
		expectedTagName string
		expectedTagURL  string
	}{
		{
			name:            "OK",
			owner:           "octocat",
			repo:            "Hello-World",
			tagName:         "v0.1.1",
			expectedTagName: "v0.1.1",
			expectedTagURL:  "https://github.com/octocat/Hello-World/tree/v0.1.1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				owner:  tc.owner,
				repo:   tc.repo,
			}

			tag := r.FutureTag(tc.tagName)

			assert.NotEmpty(t, tag)
			assert.NotZero(t, tag.Time)
			assert.Equal(t, tc.expectedTagName, tag.Name)
			assert.Equal(t, tc.expectedTagURL, tag.WebURL)
		})
	}
}

func TestRepo_CompareURL(t *testing.T) {
	tests := []struct {
		name        string
		owner       string
		repo        string
		base        string
		head        string
		expectedURL string
	}{
		{
			name:        "OK",
			owner:       "octocat",
			repo:        "Hello-World",
			base:        "v0.1.1",
			head:        "v0.1.2",
			expectedURL: "https://github.com/octocat/Hello-World/compare/v0.1.1...v0.1.2",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				owner:  tc.owner,
				repo:   tc.repo,
			}

			url := r.CompareURL(tc.base, tc.head)

			assert.Equal(t, tc.expectedURL, url)
		})
	}
}

func TestRepo_CheckPermissions(t *testing.T) {
	tests := []struct {
		name          string
		githubService githubService
		ctx           context.Context
		expectedError string
	}{
		{
			name: "Error",
			githubService: &MockGithubService{
				EnsureScopesMocks: []EnsureScopesMock{
					{OutError: errors.New("error on checking github scopes")},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on checking github scopes",
		},
		{
			name: "Success",
			githubService: &MockGithubService{
				EnsureScopesMocks: []EnsureScopesMock{
					{OutError: nil},
				},
			},
			ctx:           context.Background(),
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{logger: log.New(log.None)}
			r.services.github = tc.githubService

			err := r.CheckPermissions(tc.ctx)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRepo_FetchFirstCommit(t *testing.T) {
	tests := []struct {
		name           string
		commitsStore   *store
		repoService    repoService
		ctx            context.Context
		expectedCommit remote.Commit
		expectedError  string
	}{
		{
			name: "Error",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				CommitsMocks: []CommitsMock{
					{OutError: errors.New("error on getting github commits")},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on getting github commits",
		},
		{
			name: "Success_OnePage",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				CommitsMocks: []CommitsMock{
					{
						OutCommits:  []github.Commit{gitHubCommit1},
						OutResponse: &github.Response{},
					},
				},
			},
			ctx:            context.Background(),
			expectedCommit: remoteCommit1,
		},
		{
			name: "Success_TwoPages",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				CommitsMocks: []CommitsMock{
					{
						OutCommits: []github.Commit{},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{
						OutCommits: []github.Commit{gitHubCommit1},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 1, Next: 0, Last: 0},
						},
					},
				},
			},
			ctx:            context.Background(),
			expectedCommit: remoteCommit1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{logger: log.New(log.None)}
			r.stores.commits = tc.commitsStore
			r.services.repo = tc.repoService

			commit, err := r.FetchFirstCommit(tc.ctx)

			if tc.expectedError != "" {
				assert.Empty(t, commit)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCommit, commit)
			}
		})
	}
}

func TestRepo_FetchBranch(t *testing.T) {
	tests := []struct {
		name           string
		repoService    repoService
		ctx            context.Context
		branchName     string
		expectedBranch remote.Branch
		expectedError  string
	}{
		{
			name: "Error",
			repoService: &MockRepoService{
				BranchMocks: []BranchMock{
					{OutError: errors.New("error on getting github branch")},
				},
			},
			ctx:           context.Background(),
			branchName:    "main",
			expectedError: "error on getting github branch",
		},
		{
			name: "Success",
			repoService: &MockRepoService{
				BranchMocks: []BranchMock{
					{OutBranch: &gitHubBranch, OutResponse: &github.Response{}},
				},
			},
			ctx:            context.Background(),
			branchName:     "main",
			expectedBranch: remoteBranch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{logger: log.New(log.None)}
			r.services.repo = tc.repoService

			branch, err := r.FetchBranch(tc.ctx, tc.branchName)

			if tc.expectedError != "" {
				assert.Empty(t, branch)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBranch, branch)
			}
		})
	}
}

func TestRepo_FetchDefaultBranch(t *testing.T) {
	tests := []struct {
		name           string
		repoService    repoService
		ctx            context.Context
		expectedBranch remote.Branch
		expectedError  string
	}{
		{
			name: "RepoGetError",
			repoService: &MockRepoService{
				GetMocks: []GetRepoMock{
					{OutError: errors.New("error on getting github repository")},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on getting github repository",
		},
		{
			name: "RepoBranchError",
			repoService: &MockRepoService{
				GetMocks: []GetRepoMock{
					{OutRepository: &gitHubRepository, OutResponse: &github.Response{}},
				},
				BranchMocks: []BranchMock{
					{OutError: errors.New("error on getting github branch")},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on getting github branch",
		},
		{
			name: "Success",
			repoService: &MockRepoService{
				GetMocks: []GetRepoMock{
					{OutRepository: &gitHubRepository, OutResponse: &github.Response{}},
				},
				BranchMocks: []BranchMock{
					{OutBranch: &gitHubBranch, OutResponse: &github.Response{}},
				},
			},
			ctx:            context.Background(),
			expectedBranch: remoteBranch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{logger: log.New(log.None)}
			r.services.repo = tc.repoService

			branch, err := r.FetchDefaultBranch(tc.ctx)

			if tc.expectedError != "" {
				assert.Empty(t, branch)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBranch, branch)
			}
		})
	}
}

func TestRepo_FetchTags(t *testing.T) {
	tests := []struct {
		name          string
		owner         string
		repo          string
		commitsStore  *store
		repoService   repoService
		ctx           context.Context
		expectedTags  remote.Tags
		expectedError string
	}{
		{
			name:  "TagsFails_FirstPage",
			owner: "octocat",
			repo:  "Hello-World",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				TagsMocks: []TagsMock{
					{OutError: errors.New("error on getting github tags")},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on getting github tags",
		},
		{
			name:  "TagsFails_SecondPage",
			owner: "octocat",
			repo:  "Hello-World",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				TagsMocks: []TagsMock{
					{
						OutTags: []github.Tag{gitHubTag},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{OutError: errors.New("error on getting github tags")},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on getting github tags",
		},
		{
			name:  "CommitFails",
			owner: "octocat",
			repo:  "Hello-World",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				TagsMocks: []TagsMock{
					{
						OutTags: []github.Tag{},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{
						OutTags: []github.Tag{gitHubTag},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 1, Next: 0, Last: 0},
						},
					},
				},
				CommitMocks: []CommitMock{
					{OutError: errors.New("error on getting github commits")},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on getting github commits",
		},
		{
			name:  "Success",
			owner: "octocat",
			repo:  "Hello-World",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				TagsMocks: []TagsMock{
					{
						OutTags: []github.Tag{},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{
						OutTags: []github.Tag{gitHubTag},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 1, Next: 0, Last: 0},
						},
					},
				},
				CommitMocks: []CommitMock{
					{OutCommit: &gitHubCommit2, OutResponse: &github.Response{}},
				},
			},
			ctx:          context.Background(),
			expectedTags: remote.Tags{remoteTag},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				owner:  tc.owner,
				repo:   tc.repo,
			}

			r.stores.commits = tc.commitsStore
			r.services.repo = tc.repoService

			tags, err := r.FetchTags(tc.ctx)

			if tc.expectedError != "" {
				assert.Empty(t, tags)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTags, tags)
			}
		})
	}
}

func TestRepo_FetchIssuesAndMerges(t *testing.T) {
	since, _ := time.Parse(time.RFC3339, "2020-10-20T22:30:00-04:00")

	tests := []struct {
		name           string
		usersStore     *store
		commitsStore   *store
		usersService   usersService
		repoService    repoService
		ctx            context.Context
		since          time.Time
		expectedIssues remote.Issues
		expectedMerges remote.Merges
		expectedError  string
	}{
		{
			name: "IssuesFails_FirstPage",
			usersStore: &store{
				m: map[interface{}]interface{}{},
			},
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			usersService: &MockUsersService{},
			repoService: &MockRepoService{
				IssuesMocks: []IssuesMock{
					{OutError: errors.New("error on getting github issues")},
				},
			},
			ctx:           context.Background(),
			since:         time.Time{},
			expectedError: "error on getting github issues",
		},
		{
			name: "IssuesFails_Second",
			usersStore: &store{
				m: map[interface{}]interface{}{},
			},
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			usersService: &MockUsersService{},
			repoService: &MockRepoService{
				IssuesMocks: []IssuesMock{
					{
						OutIssues: []github.Issue{gitHubIssue1},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{OutError: errors.New("error on getting github issues")},
				},
			},
			ctx:           context.Background(),
			since:         time.Time{},
			expectedError: "error on getting github issues",
		},
		{
			name: "EventsFail",
			usersStore: &store{
				m: map[interface{}]interface{}{},
			},
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			usersService: &MockUsersService{},
			repoService: &MockRepoService{
				IssuesMocks: []IssuesMock{
					{
						OutIssues: []github.Issue{gitHubIssue1},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{
						OutIssues: []github.Issue{gitHubIssue2},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 1, Next: 0, Last: 0},
						},
					},
				},
				EventsMocks: []EventsMock{
					{OutError: errors.New("error on getting github events")},
					{OutError: errors.New("error on getting github events")},
				},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "error on getting github events",
		},
		{
			name: "CommitFails",
			usersStore: &store{
				m: map[interface{}]interface{}{},
			},
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			usersService: &MockUsersService{},
			repoService: &MockRepoService{
				IssuesMocks: []IssuesMock{
					{
						OutIssues: []github.Issue{gitHubIssue1},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{
						OutIssues: []github.Issue{gitHubIssue2},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 1, Next: 0, Last: 0},
						},
					},
				},
				EventsMocks: []EventsMock{
					// TODO: In lack of proper mocking, we need to return both events, so the findEvent method will not fail
					{OutEvents: []github.Event{gitHubEvent1, gitHubEvent2}, OutResponse: &github.Response{}},
					{OutEvents: []github.Event{gitHubEvent2, gitHubEvent1}, OutResponse: &github.Response{}},
				},
				CommitMocks: []CommitMock{
					{OutError: errors.New("error on getting github commit")},
				},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "error on getting github commit",
		},
		{
			name: "UserFails_Author",
			usersStore: &store{
				m: map[interface{}]interface{}{},
			},
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			usersService: &MockUsersService{
				GetMocks: []GetUserMock{
					{OutError: errors.New("error on getting github user")},
				},
			},
			repoService: &MockRepoService{
				IssuesMocks: []IssuesMock{
					{
						OutIssues: []github.Issue{gitHubIssue1},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{
						OutIssues: []github.Issue{gitHubIssue2},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 1, Next: 0, Last: 0},
						},
					},
				},
				EventsMocks: []EventsMock{
					// TODO: In lack of proper mocking, we need to return both events, so the findEvent method will not fail
					{OutEvents: []github.Event{gitHubEvent1, gitHubEvent2}, OutResponse: &github.Response{}},
					{OutEvents: []github.Event{gitHubEvent2, gitHubEvent1}, OutResponse: &github.Response{}},
				},
				CommitMocks: []CommitMock{
					{OutCommit: &gitHubCommit1, OutResponse: &github.Response{}},
				},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "error on getting github user",
		},
		{
			name: "UserFails_Merger",
			usersStore: &store{
				m: map[interface{}]interface{}{
					"octocat": gitHubUser1,
					"octodog": gitHubUser2,
				},
			},
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			usersService: &MockUsersService{
				GetMocks: []GetUserMock{
					{OutError: errors.New("error on getting github user")},
				},
			},
			repoService: &MockRepoService{
				IssuesMocks: []IssuesMock{
					{
						OutIssues: []github.Issue{gitHubIssue1},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{
						OutIssues: []github.Issue{gitHubIssue2},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 1, Next: 0, Last: 0},
						},
					},
				},
				EventsMocks: []EventsMock{
					// TODO: In lack of proper mocking, we need to return both events, so the findEvent method will not fail
					{OutEvents: []github.Event{gitHubEvent1, gitHubEvent2}, OutResponse: &github.Response{}},
					{OutEvents: []github.Event{gitHubEvent2, gitHubEvent1}, OutResponse: &github.Response{}},
				},
				CommitMocks: []CommitMock{
					{OutCommit: &gitHubCommit1, OutResponse: &github.Response{}},
				},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "error on getting github user",
		},
		{
			name: "Success",
			usersStore: &store{
				m: map[interface{}]interface{}{
					"octocat": gitHubUser1,
					"octodog": gitHubUser2,
					"octofox": gitHubUser3,
				},
			},
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			usersService: &MockUsersService{},
			repoService: &MockRepoService{
				IssuesMocks: []IssuesMock{
					{
						OutIssues: []github.Issue{gitHubIssue1},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 0, Prev: 0, Next: 2, Last: 2},
						},
					},
					{
						OutIssues: []github.Issue{gitHubIssue2},
						OutResponse: &github.Response{
							Pages: github.Pages{First: 1, Prev: 1, Next: 0, Last: 0},
						},
					},
				},
				EventsMocks: []EventsMock{
					// TODO: In lack of proper mocking, we need to return both events, so the findEvent method will not fail
					{OutEvents: []github.Event{gitHubEvent1, gitHubEvent2}, OutResponse: &github.Response{}},
					{OutEvents: []github.Event{gitHubEvent2, gitHubEvent1}, OutResponse: &github.Response{}},
				},
				CommitMocks: []CommitMock{
					{OutCommit: &gitHubCommit1, OutResponse: &github.Response{}},
				},
			},
			ctx:            context.Background(),
			since:          since,
			expectedIssues: remote.Issues{remoteIssue},
			expectedMerges: remote.Merges{remoteMerge},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{logger: log.New(log.None)}
			r.stores.users = tc.usersStore
			r.stores.commits = tc.commitsStore
			r.services.users = tc.usersService
			r.services.repo = tc.repoService

			issues, merges, err := r.FetchIssuesAndMerges(tc.ctx, tc.since)

			if tc.expectedError != "" {
				assert.Nil(t, issues)
				assert.Nil(t, merges)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIssues, issues)
				assert.Equal(t, tc.expectedMerges, merges)
			}
		})
	}
}

func TestRepo_FetchParentCommits(t *testing.T) {
	tests := []struct {
		name            string
		commitsStore    *store
		repoService     repoService
		ctx             context.Context
		ref             string
		expectedCommits remote.Commits
		expectedError   string
	}{
		{
			name: "CommitFails",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				CommitMocks: []CommitMock{
					{OutError: errors.New("error on getting github commit")},
				},
			},
			ctx:           context.Background(),
			ref:           "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			expectedError: "error on getting github commit",
		},
		{
			name: "Success",
			commitsStore: &store{
				m: map[interface{}]interface{}{},
			},
			repoService: &MockRepoService{
				CommitMocks: []CommitMock{
					{OutCommit: &gitHubCommit2, OutResponse: &github.Response{}},
					{OutCommit: &gitHubCommit1, OutResponse: &github.Response{}},
				},
			},
			ctx:             context.Background(),
			ref:             "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			expectedCommits: remote.Commits{remoteCommit2, remoteCommit1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{logger: log.New(log.None)}
			r.stores.commits = tc.commitsStore
			r.services.repo = tc.repoService

			commits, err := r.FetchParentCommits(tc.ctx, tc.ref)

			if tc.expectedError != "" {
				assert.Nil(t, commits)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCommits, commits)
			}
		})
	}
}
