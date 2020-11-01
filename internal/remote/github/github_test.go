package github

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/pkg/log"
)

const (
	mockGitHubUserBody1 = `{
		"login": "octocat",
		"id": 1,
		"url": "https://api.github.com/users/octocat",
		"type": "User",
		"site_admin": false,
		"name": "monalisa octocat",
		"email": "octocat@github.com"
	}`

	mockGitHubUserBody2 = `{
		"login": "octodog",
		"id": 2,
		"url": "https://api.github.com/users/octodog",
		"type": "User",
		"site_admin": false,
		"name": "monalisa octodog",
		"email": "octodog@github.com"
	}`

	mockGitHubUserBody3 = `{
		"login": "octofox",
		"id": 3,
		"url": "https://api.github.com/users/octofox",
		"type": "User",
		"site_admin": false,
		"name": "monalisa octofox",
		"email": "octofox@github.com"
	}`

	mockGitHubRepositoryBody = `{
		"id": 1296269,
		"name": "Hello-World",
		"full_name": "octocat/Hello-World",
		"owner": {
			"login": "octocat",
			"id": 1,
			"type": "User"
		},
		"private": false,
		"description": "This your first repo!",
		"fork": false,
		"default_branch": "main",
		"topics": [
			"octocat",
			"api"
		],
		"archived": false,
		"disabled": false,
		"visibility": "public",
		"pushed_at": "2020-10-31T14:00:00Z",
		"created_at": "2020-01-20T09:00:00Z",
		"updated_at": "2020-10-31T14:00:00Z"
	}`

	mockGitHubCommitBody1 = `{
		"sha": "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
		"commit": {
			"author": {
				"name": "Monalisa Octocat",
				"email": "mona@github.com",
				"date": "2020-10-27T23:59:59Z"
			},
			"committer": {
				"name": "Monalisa Octocat",
				"email": "mona@github.com",
				"date": "2020-10-27T23:59:59Z"
			},
			"message": "Release v0.1.0"
		},
		"author": {
			"login": "octocat",
			"id": 1,
			"type": "User"
		},
		"committer": {
			"login": "octocat",
			"id": 1,
			"type": "User"
		}
	}`

	mockGitHubCommitBody2 = `{
		"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		"commit": {
			"author": {
				"name": "Monalisa Octocat",
				"email": "mona@github.com",
				"date": "2020-10-20T19:59:59Z"
			},
			"committer": {
				"name": "Monalisa Octocat",
				"email": "mona@github.com",
				"date": "2020-10-20T19:59:59Z"
			},
			"message": "Fix all the bugs"
		},
		"author": {
			"login": "octocat",
			"id": 1,
			"type": "User"
		},
		"committer": {
			"login": "octocat",
			"id": 1,
			"type": "User"
		},
		"parents": [
			{
				"url": "https://api.github.com/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
				"sha": "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c"
			}
  	]
	}`

	mockGitHubBranchBody = `{
		"name": "main",
		"commit": {
			"sha": "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			"commit": {
				"author": {
					"name": "Monalisa Octocat",
					"email": "mona@github.com",
					"date": "2020-10-27T23:59:59Z"
				},
				"committer": {
					"name": "Monalisa Octocat",
					"email": "mona@github.com",
					"date": "2020-10-27T23:59:59Z"
				},
				"message": "Release v0.1.0"
			},
			"author": {
				"login": "octocat",
				"id": 1,
				"type": "User"
			},
			"committer": {
				"login": "octocat",
				"id": 1,
				"type": "User"
			}
		},
		"protected": true
	}`

	mockGitHubTagsBody = `[
		{
			"name": "v0.1.0",
			"commit": {
				"sha": "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
				"url": "https://api.github.com/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c"
			}
		}
	]`

	mockGitHubIssuesBody = `[
		{
			"id": 1,
			"number": 1001,
			"state": "open",
			"title": "Found a bug",
			"body": "This is not working as expected!",
			"user": {
				"login": "octocat",
				"id": 1,
				"type": "User"
			},
			"labels": [
				{
					"id": 2000,
					"name": "bug",
					"default": true
				}
			],
			"milestone": {
				"id": 3000,
				"number": 1,
				"state": "open",
				"title": "v1.0"
			},
			"locked": true,
			"pull_request": null,
			"closed_at": null,
			"created_at": "2020-10-10T10:00:00Z",
			"updated_at": "2020-10-20T20:00:00Z"
		},
		{
			"id": 2,
			"number": 1002,
			"state": "closed",
			"title": "Fixed a bug",
			"body": "I made this to work as expected!",
			"user": {
				"login": "octodog",
				"id": 2,
				"type": "User"
			},
			"labels": [
				{
					"id": 2000,
					"name": "bug",
					"default": true
				}
			],
			"milestone": {
				"id": 3000,
				"number": 1,
				"state": "open",
				"title": "v1.0"
			},
			"locked": false,
			"pull_request": {
				"url": "https://api.github.com/repos/octocat/Hello-World/pulls/1002"
			},
			"closed_at": "2020-10-20T20:00:00Z",
			"created_at": "2020-10-15T15:00:00Z",
			"updated_at": "2020-10-22T22:00:00Z"
		}
	]`

	mockGitHubPullsBody = `[
		{
			"id": 1,
			"number": 1002,
			"state": "closed",
			"locked": false,
			"draft": false,
			"title": "Fixed a bug",
			"body": "I made this to work as expected!",
			"user": {
				"login": "octodog",
				"id": 2,
				"type": "User"
			},
			"labels": [
				{
					"id": 2000,
					"name": "bug",
					"default": true
				}
			],
			"milestone": {
				"id": 3000,
				"number": 1,
				"state": "open",
				"title": "v1.0"
			},
			"created_at":  "2020-10-15T15:00:00Z",
			"updated_at": "2020-10-22T22:00:00Z",
			"closed_at": "2020-10-20T20:00:00Z",
			"merged_at": "2020-10-20T20:00:00Z",
			"merge_commit_sha": "e5bd3914e2e596debea16f433f57875b5b90bcd6",
			"head": {
				"label": "octodog:new-topic",
				"ref": "new-topic",
				"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e"
			},
			"base": {
				"label": "octodog:master",
				"ref": "master",
				"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e"
			},
			"merged": true,
			"mergeable": null,
			"rebaseable": null,
			"merged_by": {
				"login": "octofox",
				"id": 3,
				"type": "User"
			}
		}
	]`

	mockGitHubPullBody = `{
		"id": 1,
		"number": 1002,
		"state": "closed",
		"locked": false,
		"draft": false,
		"title": "Fixed a bug",
		"body": "I made this to work as expected!",
		"user": {
			"login": "octodog",
			"id": 2,
			"type": "User"
		},
		"labels": [
			{
				"id": 2000,
				"name": "bug",
				"default": true
			}
		],
		"milestone": {
			"id": 3000,
			"number": 1,
			"state": "open",
			"title": "v1.0"
		},
		"created_at":  "2020-10-15T15:00:00Z",
		"updated_at": "2020-10-22T22:00:00Z",
		"closed_at": "2020-10-20T20:00:00Z",
		"merged_at": "2020-10-20T20:00:00Z",
		"merge_commit_sha": "e5bd3914e2e596debea16f433f57875b5b90bcd6",
		"head": {
			"label": "octodog:new-topic",
			"ref": "new-topic",
			"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e"
		},
		"base": {
			"label": "octodog:master",
			"ref": "master",
			"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e"
		},
		"merged": true,
		"mergeable": null,
		"rebaseable": null,
		"merged_by": {
			"login": "octofox",
			"id": 3,
			"type": "User"
		}
	}`

	mockGitHubEventsBody1 = `[
		{
			"id": 1,
			"actor": {
				"login": "octocat",
				"id": 1,
				"type": "User"
			},
			"event": "closed",
			"commit_id": null,
			"created_at": "2020-10-20T20:00:00Z"
		}
	]`

	mockGitHubEventsBody2 = `[
		{
			"id": 2,
			"actor": {
				"login": "octofox",
				"id": 3,
				"type": "User"
			},
			"event": "merged",
			"commit_id": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			"created_at": "2020-10-20T20:00:00Z"
		}
	]`
)

var (
	gitHubUser1 = user{
		ID:    1,
		Login: "octocat",
		Type:  "User",
		Email: "octocat@github.com",
		Name:  "monalisa octocat",
		URL:   "https://api.github.com/users/octocat",
	}

	gitHubUser2 = user{
		ID:    2,
		Login: "octodog",
		Type:  "User",
		Email: "octodog@github.com",
		Name:  "monalisa octodog",
		URL:   "https://api.github.com/users/octodog",
	}

	gitHubUser3 = user{
		ID:    3,
		Login: "octofox",
		Type:  "User",
		Email: "octofox@github.com",
		Name:  "monalisa octofox",
		URL:   "https://api.github.com/users/octofox",
	}

	gitHubRepository = repository{
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
		Owner: user{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		CreatedAt: parseGitHubTime("2020-01-20T09:00:00Z"),
		UpdatedAt: parseGitHubTime("2020-10-31T14:00:00Z"),
		PushedAt:  parseGitHubTime("2020-10-31T14:00:00Z"),
	}

	gitHubCommit1 = commit{
		SHA: "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
		Commit: rawCommit{
			Message: "Release v0.1.0",
			Author: signature{
				Name:  "Monalisa Octocat",
				Email: "mona@github.com",
				Time:  parseGitHubTime("2020-10-27T23:59:59Z"),
			},
			Committer: signature{
				Name:  "Monalisa Octocat",
				Email: "mona@github.com",
				Time:  parseGitHubTime("2020-10-27T23:59:59Z"),
			},
		},
		Author: user{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		Committer: user{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
	}

	gitHubCommit2 = commit{
		SHA: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		Commit: rawCommit{
			Message: "Fix all the bugs",
			Author: signature{
				Name:  "Monalisa Octocat",
				Email: "mona@github.com",
				Time:  parseGitHubTime("2020-10-20T19:59:59Z"),
			},
			Committer: signature{
				Name:  "Monalisa Octocat",
				Email: "mona@github.com",
				Time:  parseGitHubTime("2020-10-20T19:59:59Z"),
			},
		},
		Author: user{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		Committer: user{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		Parents: []hash{
			{
				SHA: "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
				URL: "https://api.github.com/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			},
		},
	}

	gitHubBranch = branch{
		Name:      "main",
		Protected: true,
		Commit:    gitHubCommit1,
	}

	gitHubTag1 = tag{
		Name: "v0.1.0",
		Commit: hash{
			SHA: "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			URL: "https://api.github.com/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
		},
	}

	gitHubIssue1 = issue{
		ID:     1,
		Number: 1001,
		State:  "open",
		Locked: true,
		Title:  "Found a bug",
		Body:   "This is not working as expected!",
		User: user{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		Labels: []label{
			{
				ID:      2000,
				Name:    "bug",
				Default: true,
			},
		},
		Milestone: &milestone{
			ID:     3000,
			Number: 1,
			State:  "open",
			Title:  "v1.0",
		},
		CreatedAt: parseGitHubTime("2020-10-10T10:00:00Z"),
		UpdatedAt: parseGitHubTime("2020-10-20T20:00:00Z"),
		ClosedAt:  nil,
	}

	gitHubIssue2 = issue{
		ID:     2,
		Number: 1002,
		State:  "closed",
		Locked: false,
		Title:  "Fixed a bug",
		Body:   "I made this to work as expected!",
		User: user{
			ID:    2,
			Login: "octodog",
			Type:  "User",
		},
		Labels: []label{
			{
				ID:      2000,
				Name:    "bug",
				Default: true,
			},
		},
		Milestone: &milestone{
			ID:     3000,
			Number: 1,
			State:  "open",
			Title:  "v1.0",
		},
		PullRequest: &pullURLs{
			URL: "https://api.github.com/repos/octocat/Hello-World/pulls/1002",
		},
		CreatedAt: parseGitHubTime("2020-10-15T15:00:00Z"),
		UpdatedAt: parseGitHubTime("2020-10-22T22:00:00Z"),
		ClosedAt:  parseGitHubTimePtr("2020-10-20T20:00:00Z"),
	}

	gitHubPull1 = pull{
		ID:     1,
		Number: 1002,
		State:  "closed",
		Draft:  false,
		Locked: false,
		Title:  "Fixed a bug",
		Body:   "I made this to work as expected!",
		User: user{
			ID:    2,
			Login: "octodog",
			Type:  "User",
		},
		Labels: []label{
			{
				ID:      2000,
				Name:    "bug",
				Default: true,
			},
		},
		Milestone: &milestone{
			ID:     3000,
			Number: 1,
			State:  "open",
			Title:  "v1.0",
		},
		Base: reference{
			Label: "octodog:master",
			Ref:   "master",
			SHA:   "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		},
		Head: reference{
			Label: "octodog:new-topic",
			Ref:   "new-topic",
			SHA:   "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		},
		Merged:     true,
		Mergeable:  nil,
		Rebaseable: nil,
		MergedBy: &user{
			ID:    3,
			Login: "octofox",
			Type:  "User",
		},
		MergeCommitSHA: "e5bd3914e2e596debea16f433f57875b5b90bcd6",
		CreatedAt:      parseGitHubTime("2020-10-15T15:00:00Z"),
		UpdatedAt:      parseGitHubTime("2020-10-22T22:00:00Z"),
		ClosedAt:       parseGitHubTimePtr("2020-10-20T20:00:00Z"),
		MergedAt:       parseGitHubTimePtr("2020-10-20T20:00:00Z"),
	}

	gitHubEvent1 = event{
		ID:       1,
		Event:    "closed",
		CommitID: "",
		Actor: user{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		CreatedAt: parseGitHubTime("2020-10-20T20:00:00Z"),
	}

	gitHubEvent2 = event{
		ID:       2,
		Event:    "merged",
		CommitID: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		Actor: user{
			ID:    3,
			Login: "octofox",
			Type:  "User",
		},
		CreatedAt: parseGitHubTime("2020-10-20T20:00:00Z"),
	}

	remoteCommit1 = remote.Commit{
		Hash: "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
		Time: parseGitHubTime("2020-10-27T23:59:59Z"),
	}

	remoteCommit2 = remote.Commit{
		Hash: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		Time: parseGitHubTime("2020-10-20T19:59:59Z"),
	}

	remoteBranch = remote.Branch{
		Name:   "main",
		Commit: remoteCommit1,
	}

	remoteTag = remote.Tag{
		Name:   "v0.1.0",
		Time:   parseGitHubTime("2020-10-27T23:59:59Z"),
		Commit: remoteCommit1,
	}

	remoteIssue = remote.Issue{
		Change: remote.Change{
			Number:    1001,
			Title:     "Found a bug",
			Labels:    []string{"bug"},
			Milestone: "v1.0",
			Time:      time.Time{},
			Creator: remote.User{
				Name:     "monalisa octocat",
				Email:    "octocat@github.com",
				Username: "octocat",
				URL:      "https://api.github.com/users/octocat",
			},
		},
		Closer: remote.User{
			Name:     "monalisa octocat",
			Email:    "octocat@github.com",
			Username: "octocat",
			URL:      "https://api.github.com/users/octocat",
		},
	}

	remoteMerge = remote.Merge{
		Change: remote.Change{
			Number:    1002,
			Title:     "Fixed a bug",
			Labels:    []string{"bug"},
			Milestone: "v1.0",
			Time:      parseGitHubTime("2020-10-20T19:59:59Z"),
			Creator: remote.User{
				Name:     "monalisa octodog",
				Email:    "octodog@github.com",
				Username: "octodog",
				URL:      "https://api.github.com/users/octodog",
			},
		},
		Merger: remote.User{
			Name:     "monalisa octofox",
			Email:    "octofox@github.com",
			Username: "octofox",
			URL:      "https://api.github.com/users/octofox",
		},
		Commit: remoteCommit2,
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

type MockResponse struct {
	Method             string
	Path               string
	ResponseStatusCode int
	ResponseHeader     http.Header
	ResponseBody       string
}

func createMockHTTPServer(mocks ...MockResponse) *httptest.Server {
	r := mux.NewRouter()
	for _, m := range mocks {
		m := m
		r.Methods(m.Method).Path(m.Path).HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			for k, vals := range m.ResponseHeader {
				for _, v := range vals {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(m.ResponseStatusCode)
			_, _ = io.WriteString(w, m.ResponseBody)
		})
	}

	return httptest.NewServer(r)
}

func TestNotFoundError(t *testing.T) {
	tests := []struct {
		name          string
		err           notFoundError
		expectedError string
	}{
		{
			name: "OK",
			err: notFoundError{
				message: "event not found",
			},
			expectedError: "event not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedError, tc.err.Error())
		})
	}
}

func TestNewRepo(t *testing.T) {
	tests := []struct {
		name        string
		logger      log.Logger
		path        string
		accessToken string
	}{
		{
			name:        "OK",
			logger:      log.New(log.None),
			path:        "moorara/changelog",
			accessToken: "github-access-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRepo(tc.logger, tc.path, tc.accessToken)
			assert.NotNil(t, r)

			gr, ok := r.(*repo)
			assert.True(t, ok)

			assert.Equal(t, tc.logger, gr.logger)
			assert.NotNil(t, gr.client)
			assert.Equal(t, githubAPIURL, gr.apiURL)
			assert.Equal(t, tc.path, gr.path)
			assert.Equal(t, tc.accessToken, gr.accessToken)
			assert.NotNil(t, gr.users)
			assert.NotNil(t, gr.commits)
		})
	}
}

func TestRepo_createRequest(t *testing.T) {
	tests := []struct {
		name          string
		accessToken   string
		ctx           context.Context
		method        string
		url           string
		body          io.Reader
		expectedError string
	}{
		{
			name:          "RequestError",
			accessToken:   "github-token",
			ctx:           nil,
			method:        "",
			url:           "",
			body:          nil,
			expectedError: "net/http: nil Context",
		},
		{
			name:          "Success",
			accessToken:   "github-token",
			ctx:           context.Background(),
			method:        "GET",
			url:           "/users/octocat",
			body:          nil,
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				accessToken: tc.accessToken,
			}

			req, err := r.createRequest(tc.ctx, tc.method, tc.url, tc.body)

			if tc.expectedError != "" {
				assert.Nil(t, req)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, req)
				assert.Equal(t, "token "+tc.accessToken, req.Header.Get("Authorization"))
				assert.Equal(t, userAgentHeader, req.Header.Get("User-Agent"))
				assert.Equal(t, acceptHeader, req.Header.Get("Accept"))
				assert.Equal(t, contentTypeHeader, req.Header.Get("Content-Type"))
			}
		})
	}
}

func TestRepo_makeRequest(t *testing.T) {
	tests := []struct {
		name               string
		mockResponses      []MockResponse
		method             string
		endpoint           string
		body               io.Reader
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:               "ClientError",
			mockResponses:      []MockResponse{},
			method:             "GET",
			endpoint:           "",
			body:               nil,
			expectedStatusCode: 200,
			expectedError:      `Get "": unsupported protocol scheme ""`,
		},
		{
			name: "UnexpectedStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/users/{username}", 400, nil, `bad request`},
			},
			method:             "GET",
			endpoint:           "/users/octocat",
			body:               nil,
			expectedStatusCode: 200,
			expectedError:      "GET /users/octocat 400: bad request",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{
					"GET", "/users/{username}", 200, nil, `{
						"id": 1,
						"type": "User",
						"login": "octocat",
						"email": "octocat@github.com",
						"name": "monalisa octocat"
					}`,
				},
			},
			method:             "GET",
			endpoint:           "/users/octocat",
			body:               nil,
			expectedStatusCode: 200,
			expectedError:      "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				client: new(http.Client),
			}

			var url string
			if len(tc.mockResponses) > 0 {
				ts := createMockHTTPServer(tc.mockResponses...)
				defer ts.Close()
				url = ts.URL + tc.endpoint
			}

			req, err := http.NewRequest(tc.method, url, tc.body)
			assert.NoError(t, err)

			resp, err := r.makeRequest(req, tc.expectedStatusCode)

			if tc.expectedError != "" {
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestRepo_checkScopes(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		scopes        []scope
		expectedError string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			scopes:        []scope{scopeRepo},
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			scopes:        []scope{scopeRepo},
			expectedError: "HEAD /user 401: ",
		},
		{
			name: "MissingScope",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, nil, ``},
			},
			ctx:           context.Background(),
			scopes:        []scope{scopeRepo},
			expectedError: "access token does not have the scope: repo",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
			},
			ctx:           context.Background(),
			scopes:        []scope{scopeRepo},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			err := r.checkScopes(tc.ctx, tc.scopes...)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchUser(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		users         *userStore
		ctx           context.Context
		username      string
		expectedError string
		expectedUser  user
	}{
		{
			name:          "CacheHit",
			mockResponses: []MockResponse{},
			users: &userStore{
				m: map[string]user{
					"octocat": gitHubUser1,
				},
			},
			ctx:          context.Background(),
			username:     "octocat",
			expectedUser: gitHubUser1,
		},
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			users:         newUserStore(),
			ctx:           nil,
			username:      "octocat",
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/users/octocat", 401, nil, `bad credentials`},
			},
			users:         newUserStore(),
			ctx:           context.Background(),
			username:      "octocat",
			expectedError: "GET /users/octocat 401: bad credentials",
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/users/octocat", 200, nil, `[`},
			},
			users:         newUserStore(),
			ctx:           context.Background(),
			username:      "octocat",
			expectedError: "unexpected EOF",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/users/octocat", 200, nil, mockGitHubUserBody1},
			},
			users:        newUserStore(),
			ctx:          context.Background(),
			username:     "octocat",
			expectedUser: gitHubUser1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				users:  tc.users,
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			user, err := r.fetchUser(tc.ctx, tc.username)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, user)
			} else {
				assert.Empty(t, user)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchRepository(t *testing.T) {
	tests := []struct {
		name               string
		mockResponses      []MockResponse
		ctx                context.Context
		expectedError      string
		expectedRepository repository
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			expectedError: "GET /repos/octocat/Hello-World 401: bad credentials",
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World", 200, nil, `[`},
			},
			ctx:           context.Background(),
			expectedError: "unexpected EOF",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World", 200, nil, mockGitHubRepositoryBody},
			},
			ctx:                context.Background(),
			expectedRepository: gitHubRepository,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			repository, err := r.fetchRepository(tc.ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRepository, repository)
			} else {
				assert.Empty(t, repository)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchCommit(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  []MockResponse
		commits        *commitStore
		ctx            context.Context
		ref            string
		expectedError  string
		expectedCommit commit
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			commits: &commitStore{
				m: map[string]commit{
					"6dcb09b5b57875f334f61aebed695e2e4193db5e": gitHubCommit2,
				},
			},
			ctx:            context.Background(),
			ref:            "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedCommit: gitHubCommit2,
		},
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			commits:       newCommitStore(),
			ctx:           nil,
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 401, nil, `bad credentials`},
			},
			commits:       newCommitStore(),
			ctx:           context.Background(),
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: "GET /repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e 401: bad credentials",
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, `[`},
			},
			commits:       newCommitStore(),
			ctx:           context.Background(),
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: "unexpected EOF",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, mockGitHubCommitBody2},
			},
			commits:        newCommitStore(),
			ctx:            context.Background(),
			ref:            "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedCommit: gitHubCommit2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger:  log.New(log.None),
				client:  new(http.Client),
				path:    "octocat/Hello-World",
				commits: tc.commits,
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			commit, err := r.fetchCommit(tc.ctx, tc.ref)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCommit, commit)
			} else {
				assert.Empty(t, commit)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchParentCommits(t *testing.T) {
	tests := []struct {
		name            string
		mockResponses   []MockResponse
		ctx             context.Context
		ref             string
		expectedError   string
		expectedCommits remote.Commits
	}{
		{
			name:          "FetchCommitsFails#1",
			mockResponses: []MockResponse{},
			ctx:           context.Background(),
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: "GET /repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e 404: 404 page not found\n",
		},
		{
			name: "FetchCommitsFails#2",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, mockGitHubCommitBody2},
			},
			ctx:           context.Background(),
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: "GET /repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c 404: 404 page not found\n",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, mockGitHubCommitBody2},
				{"GET", "/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c", 200, nil, mockGitHubCommitBody1},
			},
			ctx:             context.Background(),
			ref:             "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedCommits: remote.Commits{remoteCommit2, remoteCommit1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger:  log.New(log.None),
				client:  new(http.Client),
				path:    "octocat/Hello-World",
				commits: newCommitStore(),
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			commits, err := r.fetchParentCommits(tc.ctx, tc.ref)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCommits, commits)
			} else {
				assert.Empty(t, commits)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchBranch(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  []MockResponse
		ctx            context.Context
		branchName     string
		expectedError  string
		expectedBranch branch
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			branchName:    "main",
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/branches/main", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			branchName:    "main",
			expectedError: "GET /repos/octocat/Hello-World/branches/main 401: bad credentials",
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/branches/main", 200, nil, `[`},
			},
			ctx:           context.Background(),
			branchName:    "main",
			expectedError: "unexpected EOF",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/branches/main", 200, nil, mockGitHubBranchBody},
			},
			ctx:            context.Background(),
			branchName:     "main",
			expectedBranch: gitHubBranch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			branch, err := r.fetchBranch(tc.ctx, tc.branchName)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBranch, branch)
			} else {
				assert.Empty(t, branch)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchPull(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		number        int
		expectedError string
		expectedPull  pull
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			number:        1002,
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls/1002", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			number:        1002,
			expectedError: "GET /repos/octocat/Hello-World/pulls/1002 401: bad credentials",
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls/1002", 200, http.Header{}, `[`},
			},
			ctx:           context.Background(),
			number:        1002,
			expectedError: "unexpected EOF",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls/1002", 200, http.Header{}, mockGitHubPullBody},
			},
			ctx:          context.Background(),
			number:       1002,
			expectedPull: gitHubPull1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			pull, err := r.fetchPull(tc.ctx, tc.number)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPull, pull)
			} else {
				assert.Empty(t, pull)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_findEvent(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		issueNumber   int
		eventName     string
		expectedError string
		expectedEvent event
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			issueNumber:   1001,
			eventName:     "closed",
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			issueNumber:   1001,
			eventName:     "closed",
			expectedError: "GET /repos/octocat/Hello-World/issues/1001/events 401: bad credentials",
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, `[`},
			},
			ctx:           context.Background(),
			issueNumber:   1001,
			eventName:     "closed",
			expectedError: "unexpected EOF",
		},
		{
			name: "Found_FirstPage",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, mockGitHubEventsBody1},
			},
			ctx:           context.Background(),
			issueNumber:   1001,
			eventName:     "closed",
			expectedEvent: gitHubEvent1,
		},
		{
			name: "NotFound_NoLinkHeader",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, `[]`},
			},
			ctx:           context.Background(),
			issueNumber:   1001,
			eventName:     "closed",
			expectedError: "GitHub closed event for issue 1001 not found",
		},
		{
			name: "NotFound_InvalidLinkHeader",
			mockResponses: []MockResponse{
				{
					"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{
						"Link": []string{`<>; rel="next", <>; rel="last"`},
					}, `[]`,
				},
			},
			ctx:           context.Background(),
			issueNumber:   1001,
			eventName:     "closed",
			expectedError: "GitHub closed event for issue 1001 not found",
		},
		/* {
			name: "Found_SecondPage",
			mockResponses: []MockResponse{
				{
					"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{
						"Link": []string{`<https://api.github.com/repositories/1/issues/events?page=2>; rel="next", <>; rel="last"`},
					}, mockGitHubEventsBody1,
				},
			},
			ctx:           context.Background(),
			issueNumber:   1001,
			eventName:     "closed",
			expectedEvent: gitHubEvent1,
		}, */
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			event, err := r.findEvent(tc.ctx, tc.issueNumber, tc.eventName)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedEvent, event)
			} else {
				assert.Empty(t, event)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchTagsPageCount(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		expectedCount int
		expectedError string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"HEAD", "/repos/octocat/Hello-World/tags", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			expectedError: "HEAD /repos/octocat/Hello-World/tags 401: ",
		},
		{
			name: "Success_NoLinkHeader",
			mockResponses: []MockResponse{
				{"HEAD", "/repos/octocat/Hello-World/tags", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			expectedCount: 1,
		},
		{
			name: "InvalidLinkHeader",
			mockResponses: []MockResponse{
				{
					"HEAD", "/repos/octocat/Hello-World/tags", 200, http.Header{
						"Link": []string{`<>; rel="next", <>; rel="last"`},
					}, ``,
				},
			},
			ctx:           context.Background(),
			expectedError: `invalid Link header received from GitHub: <>; rel="next", <>; rel="last"`,
		},
		{
			name: "Success_LinkHeader",
			mockResponses: []MockResponse{
				{
					"HEAD", "/repos/octocat/Hello-World/tags", 200, http.Header{
						"Link": []string{`<https://api.github.com/repositories/1/tags?state=closed&page=2>; rel="next", <https://api.github.com/repositories/1/tags?state=closed&page=3>; rel="last"`},
					}, ``,
				},
			},
			ctx:           context.Background(),
			expectedCount: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			count, err := r.fetchTagsPageCount(tc.ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, count)
			} else {
				assert.Equal(t, -1, count)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchTags(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		pageNo        int
		expectedError string
		expectedTags  []tag
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			pageNo:        1,
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/tags", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			pageNo:        1,
			expectedError: "GET /repos/octocat/Hello-World/tags 401: bad credentials",
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/tags", 200, nil, `[`},
			},
			ctx:           context.Background(),
			pageNo:        1,
			expectedError: "unexpected EOF",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/tags", 200, nil, mockGitHubTagsBody},
			},
			ctx:          context.Background(),
			pageNo:       1,
			expectedTags: []tag{gitHubTag1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			tags, err := r.fetchTags(tc.ctx, tc.pageNo)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTags, tags)
			} else {
				assert.Nil(t, tags)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchIssuesPageCount(t *testing.T) {
	since, _ := time.Parse(time.RFC3339, "2020-10-20T22:30:00-04:00")

	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		since         time.Time
		expectedCount int
		expectedError string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			since:         since,
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"HEAD", "/repos/octocat/Hello-World/issues", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "HEAD /repos/octocat/Hello-World/issues 401: ",
		},
		{
			name: "Success_NoLinkHeader",
			mockResponses: []MockResponse{
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			since:         since,
			expectedCount: 1,
		},
		{
			name: "InvalidLinkHeader",
			mockResponses: []MockResponse{
				{
					"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{
						"Link": []string{`<>; rel="next", <>; rel="last"`},
					}, ``,
				},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: `invalid Link header received from GitHub: <>; rel="next", <>; rel="last"`,
		},
		{
			name: "Success_LinkHeader",
			mockResponses: []MockResponse{
				{
					"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{
						"Link": []string{`<https://api.github.com/repositories/1/issues?state=closed&page=2>; rel="next", <https://api.github.com/repositories/1/issues?state=closed&page=7>; rel="last"`},
					}, ``,
				},
			},
			ctx:           context.Background(),
			since:         since,
			expectedCount: 7,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			count, err := r.fetchIssuesPageCount(tc.ctx, tc.since)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, count)
			} else {
				assert.Equal(t, -1, count)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchIssues(t *testing.T) {
	since, _ := time.Parse(time.RFC3339, "2020-10-20T22:30:00-04:00")

	tests := []struct {
		name           string
		mockResponses  []MockResponse
		ctx            context.Context
		since          time.Time
		pageNo         int
		expectedError  string
		expectedIssues []issue
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			since:         since,
			pageNo:        1,
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			since:         since,
			pageNo:        1,
			expectedError: "GET /repos/octocat/Hello-World/issues 401: bad credentials",
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues", 200, nil, `[`},
			},
			ctx:           context.Background(),
			since:         since,
			pageNo:        1,
			expectedError: "unexpected EOF",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues", 200, nil, mockGitHubIssuesBody},
			},
			ctx:            context.Background(),
			since:          since,
			pageNo:         1,
			expectedIssues: []issue{gitHubIssue1, gitHubIssue2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			issues, err := r.fetchIssues(tc.ctx, tc.since, tc.pageNo)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIssues, issues)
			} else {
				assert.Nil(t, issues)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchPullsPageCount(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		expectedCount int
		expectedError string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"HEAD", "/repos/octocat/Hello-World/pulls", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			expectedError: "HEAD /repos/octocat/Hello-World/pulls 401: ",
		},
		{
			name: "Success_NoLinkHeader",
			mockResponses: []MockResponse{
				{"HEAD", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			expectedCount: 1,
		},
		{
			name: "InvalidLinkHeader",
			mockResponses: []MockResponse{
				{
					"HEAD", "/repos/octocat/Hello-World/pulls", 200, http.Header{
						"Link": []string{`<>; rel="next", <>; rel="last"`},
					}, ``,
				},
			},
			ctx:           context.Background(),
			expectedError: `invalid Link header received from GitHub: <>; rel="next", <>; rel="last"`,
		},
		{
			name: "Success_LinkHeader",
			mockResponses: []MockResponse{
				{
					"HEAD", "/repos/octocat/Hello-World/pulls", 200, http.Header{
						"Link": []string{`<https://api.github.com/repositories/1/pulls?state=closed&page=2>; rel="next", <https://api.github.com/repositories/1/pulls?state=closed&page=5>; rel="last"`},
					}, ``,
				},
			},
			ctx:           context.Background(),
			expectedCount: 5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			count, err := r.fetchPullsPageCount(tc.ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCount, count)
			} else {
				assert.Equal(t, -1, count)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_fetchPulls(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		pageNo        int
		expectedError string
		expectedPulls []pull
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			ctx:           nil,
			pageNo:        1,
			expectedError: "net/http: nil Context",
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls", 401, nil, `bad credentials`},
			},
			ctx:           context.Background(),
			pageNo:        1,
			expectedError: "GET /repos/octocat/Hello-World/pulls 401: bad credentials",
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls", 200, nil, `[`},
			},
			ctx:           context.Background(),
			pageNo:        1,
			expectedError: "unexpected EOF",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls", 200, nil, mockGitHubPullsBody},
			},
			ctx:           context.Background(),
			pageNo:        1,
			expectedPulls: []pull{gitHubPull1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			pulls, err := r.fetchPulls(tc.ctx, tc.pageNo)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPulls, pulls)
			} else {
				assert.Nil(t, pulls)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_FetchBranch(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  []MockResponse
		ctx            context.Context
		branchName     string
		expectedError  string
		expectedBranch remote.Branch
	}{
		{
			name: "CheckScopesFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			branchName:    "main",
			expectedError: "access token does not have the scope: repo",
		},
		{
			name: "FetchBranchFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
			},
			ctx:           context.Background(),
			branchName:    "main",
			expectedError: "GET /repos/octocat/Hello-World/branches/main 404: 404 page not found\n",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"GET", "/repos/octocat/Hello-World/branches/main", 200, http.Header{}, mockGitHubBranchBody},
			},
			ctx:            context.Background(),
			branchName:     "main",
			expectedBranch: remoteBranch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger:  log.New(log.None),
				client:  new(http.Client),
				path:    "octocat/Hello-World",
				commits: newCommitStore(),
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			branch, err := r.FetchBranch(tc.ctx, tc.branchName)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBranch, branch)
			} else {
				assert.Empty(t, branch)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_FetchDefaultBranch(t *testing.T) {
	tests := []struct {
		name           string
		mockResponses  []MockResponse
		ctx            context.Context
		expectedError  string
		expectedBranch remote.Branch
	}{
		{
			name: "CheckScopesFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			expectedError: "access token does not have the scope: repo",
		},
		{
			name: "FetchRepositoryFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
			},
			ctx:           context.Background(),
			expectedError: "GET /repos/octocat/Hello-World 404: 404 page not found\n",
		},
		{
			name: "FetchBranchFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"GET", "/repos/octocat/Hello-World", 200, nil, mockGitHubRepositoryBody},
			},
			ctx:           context.Background(),
			expectedError: "GET /repos/octocat/Hello-World/branches/main 404: 404 page not found\n",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"GET", "/repos/octocat/Hello-World", 200, nil, mockGitHubRepositoryBody},
				{"GET", "/repos/octocat/Hello-World/branches/main", 200, http.Header{}, mockGitHubBranchBody},
			},
			ctx:            context.Background(),
			expectedBranch: remoteBranch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger:  log.New(log.None),
				client:  new(http.Client),
				path:    "octocat/Hello-World",
				commits: newCommitStore(),
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			branch, err := r.FetchDefaultBranch(tc.ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBranch, branch)
			} else {
				assert.Empty(t, branch)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_FetchTags(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		expectedError string
		expectedTags  remote.Tags
	}{
		{
			name: "CheckScopesFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			expectedError: "access token does not have the scope: repo",
		},
		{
			name: "FetchTagsPageCountFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
			},
			ctx:           context.Background(),
			expectedError: "HEAD /repos/octocat/Hello-World/tags 404: ",
		},
		{
			name: "FetchTagsFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/tags", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			expectedError: "GET /repos/octocat/Hello-World/tags 405: ",
		},
		{
			name: "FetchCommitFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/tags", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/tags", 200, http.Header{}, mockGitHubTagsBody},
			},
			ctx:           context.Background(),
			expectedError: "GET /repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c 404: 404 page not found\n",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/tags", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/tags", 200, http.Header{}, mockGitHubTagsBody},
				{"GET", "/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c", 200, http.Header{}, mockGitHubCommitBody1},
			},
			ctx:          context.Background(),
			expectedTags: remote.Tags{remoteTag},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger:  log.New(log.None),
				client:  new(http.Client),
				path:    "octocat/Hello-World",
				commits: newCommitStore(),
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			tags, err := r.FetchTags(tc.ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTags, tags)
			} else {
				assert.Nil(t, tags)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_FetchIssuesAndMerges(t *testing.T) {
	since, _ := time.Parse(time.RFC3339, "2020-10-20T22:30:00-04:00")

	tests := []struct {
		name           string
		mockResponses  []MockResponse
		ctx            context.Context
		since          time.Time
		expectedError  string
		expectedIssues remote.Issues
		expectedMerges remote.Merges
	}{
		{
			name: "CheckScopesFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			since:         time.Time{},
			expectedError: "access token does not have the scope: repo",
		},
		{
			name: "FetchIssuesPageCountFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
			},
			ctx:           context.Background(),
			since:         time.Time{},
			expectedError: "HEAD /repos/octocat/Hello-World/issues 404: ",
		},
		{
			name: "FetchIssuesFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "GET /repos/octocat/Hello-World/issues 405: ",
		},
		{
			name: "FindEventFails#1001",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/issues/1002/events", 200, http.Header{}, mockGitHubEventsBody2},
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, mockGitHubCommitBody2},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "GET /repos/octocat/Hello-World/issues/1001/events 404: 404 page not found\n",
		},
		{
			name: "FindEventFails#1002",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, mockGitHubEventsBody1},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "GET /repos/octocat/Hello-World/issues/1002/events 404: 404 page not found\n",
		},
		{
			name: "MergeEventNotFound",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, mockGitHubEventsBody1},
				{"GET", "/repos/octocat/Hello-World/issues/1002/events", 200, http.Header{}, `[]`},
				{"GET", "/users/octocat", 200, nil, mockGitHubUserBody1},
			},
			ctx:            context.Background(),
			since:          since,
			expectedIssues: remote.Issues{remoteIssue},
			expectedMerges: remote.Merges{},
		},
		{
			name: "FetchCommitFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, mockGitHubEventsBody1},
				{"GET", "/repos/octocat/Hello-World/issues/1002/events", 200, http.Header{}, mockGitHubEventsBody2},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "GET /repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e 404: 404 page not found\n",
		},
		{
			name: "FetchUserFails#1",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, mockGitHubEventsBody1},
				{"GET", "/repos/octocat/Hello-World/issues/1002/events", 200, http.Header{}, mockGitHubEventsBody2},
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, mockGitHubCommitBody2},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "GET /users/octocat 404: 404 page not found\n",
		},
		{
			name: "FetchUserFails#2",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, mockGitHubEventsBody1},
				{"GET", "/repos/octocat/Hello-World/issues/1002/events", 200, http.Header{}, mockGitHubEventsBody2},
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, mockGitHubCommitBody2},
				{"GET", "/users/octocat", 200, nil, mockGitHubUserBody1},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "GET /users/octodog 404: 404 page not found\n",
		},
		{
			name: "FetchUserFails#3",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, mockGitHubEventsBody1},
				{"GET", "/repos/octocat/Hello-World/issues/1002/events", 200, http.Header{}, mockGitHubEventsBody2},
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, mockGitHubCommitBody2},
				{"GET", "/users/octocat", 200, nil, mockGitHubUserBody1},
				{"GET", "/users/octodog", 200, nil, mockGitHubUserBody2},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "GET /users/octofox 404: 404 page not found\n",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, mockGitHubEventsBody1},
				{"GET", "/repos/octocat/Hello-World/issues/1002/events", 200, http.Header{}, mockGitHubEventsBody2},
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, mockGitHubCommitBody2},
				{"GET", "/users/octocat", 200, nil, mockGitHubUserBody1},
				{"GET", "/users/octodog", 200, nil, mockGitHubUserBody2},
				{"GET", "/users/octofox", 200, nil, mockGitHubUserBody3},
			},
			ctx:            context.Background(),
			since:          since,
			expectedIssues: remote.Issues{remoteIssue},
			expectedMerges: remote.Merges{remoteMerge},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger:  log.New(log.None),
				client:  new(http.Client),
				path:    "octocat/Hello-World",
				users:   newUserStore(),
				commits: newCommitStore(),
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			issues, merges, err := r.FetchIssuesAndMerges(tc.ctx, tc.since)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIssues, issues)
				assert.Equal(t, tc.expectedMerges, merges)
			} else {
				assert.Nil(t, issues)
				assert.Nil(t, merges)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_FetchParentCommits(t *testing.T) {
	tests := []struct {
		name            string
		mockResponses   []MockResponse
		ctx             context.Context
		ref             string
		expectedError   string
		expectedCommits remote.Commits
	}{
		{
			name: "CheckScopesFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: "access token does not have the scope: repo",
		},
		{
			name: "FetchCommitsFails#1",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
			},
			ctx:           context.Background(),
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: "GET /repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e 404: 404 page not found\n",
		},
		{
			name: "FetchCommitsFails#2",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, mockGitHubCommitBody2},
			},
			ctx:           context.Background(),
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: "GET /repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c 404: 404 page not found\n",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, nil, mockGitHubCommitBody2},
				{"GET", "/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c", 200, nil, mockGitHubCommitBody1},
			},
			ctx:             context.Background(),
			ref:             "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedCommits: remote.Commits{remoteCommit2, remoteCommit1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger:  log.New(log.None),
				client:  new(http.Client),
				path:    "octocat/Hello-World",
				commits: newCommitStore(),
			}

			ts := createMockHTTPServer(tc.mockResponses...)
			defer ts.Close()
			r.apiURL = ts.URL

			commits, err := r.FetchParentCommits(tc.ctx, tc.ref)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCommits, commits)
			} else {
				assert.Empty(t, commits)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
