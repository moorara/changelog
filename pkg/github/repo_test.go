package github

import (
	"context"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	repositoryBody = `{
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

	commitBody1 = `{
		"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		"commit": {
			"author": {
				"name": "The Octocat",
				"email": "octocat@github.com",
				"date": "2020-10-20T19:59:59Z"
			},
			"committer": {
				"name": "The Octocat",
				"email": "octocat@github.com",
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

	commitBody2 = `{
		"sha": "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
		"commit": {
			"author": {
				"name": "The Octocat",
				"email": "octocat@github.com",
				"date": "2020-10-27T23:59:59Z"
			},
			"committer": {
				"name": "The Octocat",
				"email": "octocat@github.com",
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

	commitsBody = `[
		{
			"sha": "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			"commit": {
				"author": {
					"name": "The Octocat",
					"email": "octocat@github.com",
					"date": "2020-10-27T23:59:59Z"
				},
				"committer": {
					"name": "The Octocat",
					"email": "octocat@github.com",
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
		{
			"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			"commit": {
				"author": {
					"name": "The Octocat",
					"email": "octocat@github.com",
					"date": "2020-10-20T19:59:59Z"
				},
				"committer": {
					"name": "The Octocat",
					"email": "octocat@github.com",
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
		}
	]`

	branchBody = `{
		"name": "main",
		"commit": {
			"sha": "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			"commit": {
				"author": {
					"name": "The Octocat",
					"email": "octocat@github.com",
					"date": "2020-10-27T23:59:59Z"
				},
				"committer": {
					"name": "The Octocat",
					"email": "octocat@github.com",
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

	tagsBody = `[
		{
			"name": "v0.1.0",
			"commit": {
				"sha": "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
				"url": "https://api.github.com/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c"
			}
		}
	]`

	issuesBody = `[
		{
			"id": 2,
			"url": "https://api.github.com/repos/octocat/Hello-World/issues/1002",
			"html_url": "https://github.com/octocat/Hello-World/pull/1002",
			"number": 1002,
			"state": "closed",
			"title": "Fixed a bug",
			"body": "I made this to work as expected!",
			"user": {
				"login": "octodog",
				"id": 2,
				"url": "https://api.github.com/users/octodog",
				"html_url": "https://github.com/octodog",
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
		},
		{
			"id": 1,
			"url": "https://api.github.com/repos/octocat/Hello-World/issues/1001",
			"html_url": "https://github.com/octocat/Hello-World/issues/1001",
			"number": 1001,
			"state": "open",
			"title": "Found a bug",
			"body": "This is not working as expected!",
			"user": {
				"login": "octocat",
				"id": 1,
				"url": "https://api.github.com/users/octocat",
				"html_url": "https://github.com/octocat",
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
		}
	]`

	pullBody = `{
		"id": 1,
		"url": "https://api.github.com/repos/octocat/Hello-World/pulls/1002",
		"html_url": "https://github.com/octocat/Hello-World/pull/1002",
		"number": 1002,
		"state": "closed",
		"locked": false,
		"draft": false,
		"title": "Fixed a bug",
		"body": "I made this to work as expected!",
		"user": {
			"login": "octodog",
			"id": 2,
			"url": "https://api.github.com/users/octodog",
			"html_url": "https://github.com/octodog",
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
			"url": "https://api.github.com/users/octofox",
			"html_url": "https://github.com/octofox",
			"type": "User"
		}
	}`

	pullsBody = `[
		{
			"id": 1,
			"url": "https://api.github.com/repos/octocat/Hello-World/pulls/1002",
			"html_url": "https://github.com/octocat/Hello-World/pull/1002",
			"number": 1002,
			"state": "closed",
			"locked": false,
			"draft": false,
			"title": "Fixed a bug",
			"body": "I made this to work as expected!",
			"user": {
				"login": "octodog",
				"id": 2,
				"url": "https://api.github.com/users/octodog",
				"html_url": "https://github.com/octodog",
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
				"url": "https://api.github.com/users/octofox",
				"html_url": "https://github.com/octofox",
				"type": "User"
			}
		}
	]`

	eventsBody = `[
		{
			"id": 2,
			"actor": {
				"login": "octofox",
				"id": 3,
				"url": "https://api.github.com/users/octofox",
				"html_url": "https://github.com/octofox",
				"type": "User"
			},
			"event": "merged",
			"commit_id": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			"created_at": "2020-10-20T20:00:00Z"
		},
		{
			"id": 1,
			"actor": {
				"login": "octocat",
				"id": 1,
				"url": "https://api.github.com/users/octocat",
				"html_url": "https://github.com/octocat",
				"type": "User"
			},
			"event": "closed",
			"commit_id": null,
			"created_at": "2020-10-20T20:00:00Z"
		}
	]`
)

var (
	repository = Repository{
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
		Owner: User{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		CreatedAt: parseGitHubTime("2020-01-20T09:00:00Z"),
		UpdatedAt: parseGitHubTime("2020-10-31T14:00:00Z"),
		PushedAt:  parseGitHubTime("2020-10-31T14:00:00Z"),
	}

	commit1 = Commit{
		SHA: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		Commit: RawCommit{
			Message: "Fix all the bugs",
			Author: Signature{
				Name:  "The Octocat",
				Email: "octocat@github.com",
				Time:  parseGitHubTime("2020-10-20T19:59:59Z"),
			},
			Committer: Signature{
				Name:  "The Octocat",
				Email: "octocat@github.com",
				Time:  parseGitHubTime("2020-10-20T19:59:59Z"),
			},
		},
		Author: User{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		Committer: User{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		Parents: []Hash{
			{
				SHA: "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
				URL: "https://api.github.com/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			},
		},
	}

	commit2 = Commit{
		SHA: "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
		Commit: RawCommit{
			Message: "Release v0.1.0",
			Author: Signature{
				Name:  "The Octocat",
				Email: "octocat@github.com",
				Time:  parseGitHubTime("2020-10-27T23:59:59Z"),
			},
			Committer: Signature{
				Name:  "The Octocat",
				Email: "octocat@github.com",
				Time:  parseGitHubTime("2020-10-27T23:59:59Z"),
			},
		},
		Author: User{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		Committer: User{
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
	}

	branch = Branch{
		Name:      "main",
		Protected: true,
		Commit:    commit2,
	}

	tag = Tag{
		Name: "v0.1.0",
		Commit: Hash{
			SHA: "c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
			URL: "https://api.github.com/repos/octocat/Hello-World/commits/c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c",
		},
	}

	issue1 = Issue{
		ID:     1,
		Number: 1001,
		State:  "open",
		Locked: true,
		Title:  "Found a bug",
		Body:   "This is not working as expected!",
		User: User{
			ID:      1,
			Login:   "octocat",
			Type:    "User",
			URL:     "https://api.github.com/users/octocat",
			HTMLURL: "https://github.com/octocat",
		},
		Labels: []Label{
			{
				ID:      2000,
				Name:    "bug",
				Default: true,
			},
		},
		Milestone: &Milestone{
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

	issue2 = Issue{
		ID:     2,
		Number: 1002,
		State:  "closed",
		Locked: false,
		Title:  "Fixed a bug",
		Body:   "I made this to work as expected!",
		User: User{
			ID:      2,
			Login:   "octodog",
			Type:    "User",
			URL:     "https://api.github.com/users/octodog",
			HTMLURL: "https://github.com/octodog",
		},
		Labels: []Label{
			{
				ID:      2000,
				Name:    "bug",
				Default: true,
			},
		},
		Milestone: &Milestone{
			ID:     3000,
			Number: 1,
			State:  "open",
			Title:  "v1.0",
		},
		URL:     "https://api.github.com/repos/octocat/Hello-World/issues/1002",
		HTMLURL: "https://github.com/octocat/Hello-World/pull/1002",
		PullURLs: &PullURLs{
			URL: "https://api.github.com/repos/octocat/Hello-World/pulls/1002",
		},
		CreatedAt: parseGitHubTime("2020-10-15T15:00:00Z"),
		UpdatedAt: parseGitHubTime("2020-10-22T22:00:00Z"),
		ClosedAt:  parseGitHubTimePtr("2020-10-20T20:00:00Z"),
	}

	pull = Pull{
		ID:     1,
		Number: 1002,
		State:  "closed",
		Draft:  false,
		Locked: false,
		Title:  "Fixed a bug",
		Body:   "I made this to work as expected!",
		User: User{
			ID:      2,
			Login:   "octodog",
			Type:    "User",
			URL:     "https://api.github.com/users/octodog",
			HTMLURL: "https://github.com/octodog",
		},
		Labels: []Label{
			{
				ID:      2000,
				Name:    "bug",
				Default: true,
			},
		},
		Milestone: &Milestone{
			ID:     3000,
			Number: 1,
			State:  "open",
			Title:  "v1.0",
		},
		Base: PullBranch{
			Label: "octodog:master",
			Ref:   "master",
			SHA:   "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		},
		Head: PullBranch{
			Label: "octodog:new-topic",
			Ref:   "new-topic",
			SHA:   "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		},
		Merged:     true,
		Mergeable:  nil,
		Rebaseable: nil,
		MergedBy: &User{
			ID:      3,
			Login:   "octofox",
			Type:    "User",
			URL:     "https://api.github.com/users/octofox",
			HTMLURL: "https://github.com/octofox",
		},
		MergeCommitSHA: "e5bd3914e2e596debea16f433f57875b5b90bcd6",
		URL:            "https://api.github.com/repos/octocat/Hello-World/pulls/1002",
		HTMLURL:        "https://github.com/octocat/Hello-World/pull/1002",
		CreatedAt:      parseGitHubTime("2020-10-15T15:00:00Z"),
		UpdatedAt:      parseGitHubTime("2020-10-22T22:00:00Z"),
		ClosedAt:       parseGitHubTimePtr("2020-10-20T20:00:00Z"),
		MergedAt:       parseGitHubTimePtr("2020-10-20T20:00:00Z"),
	}

	event1 = Event{
		ID:       1,
		Event:    "closed",
		CommitID: "",
		Actor: User{
			ID:      1,
			Login:   "octocat",
			Type:    "User",
			URL:     "https://api.github.com/users/octocat",
			HTMLURL: "https://github.com/octocat",
		},
		CreatedAt: parseGitHubTime("2020-10-20T20:00:00Z"),
	}

	event2 = Event{
		ID:       2,
		Event:    "merged",
		CommitID: "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		Actor: User{
			ID:      3,
			Login:   "octofox",
			Type:    "User",
			URL:     "https://api.github.com/users/octofox",
			HTMLURL: "https://github.com/octofox",
		},
		CreatedAt: parseGitHubTime("2020-10-20T20:00:00Z"),
	}
)

func TestRepoService_Get(t *testing.T) {
	c := &Client{
		httpClient: &http.Client{},
		rates:      map[rateGroup]Rate{},
		apiURL:     publicUploadURL,
	}

	tests := []struct {
		name               string
		mockResponses      []MockResponse
		s                  *RepoService
		ctx                context.Context
		expectedRepository *Repository
		expectedResponse   *Response
		expectedError      string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           nil,
			expectedError: `net/http: nil Context`,
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World", 401, http.Header{}, `{
					"message": "invalid credentials"
				}`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			expectedError: `GET /repos/octocat/Hello-World: 401 invalid credentials`,
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World", 200, http.Header{}, `[`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			expectedError: `unexpected EOF`,
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World", 200, header, repositoryBody},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:                context.Background(),
			expectedRepository: &repository,
			expectedResponse: &Response{
				Pages: expectedPages,
				Rate:  expectedRate,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPTestServer(tc.mockResponses...)
			tc.s.client.apiURL, _ = url.Parse(ts.URL)

			repository, resp, err := tc.s.Get(tc.ctx)

			if tc.expectedError != "" {
				assert.Nil(t, repository)
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedRepository, repository)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
				assert.Equal(t, tc.expectedResponse.Pages, resp.Pages)
				assert.Equal(t, tc.expectedResponse.Rate, resp.Rate)
			}
		})
	}
}

func TestRepoService_Commit(t *testing.T) {
	c := &Client{
		httpClient: &http.Client{},
		rates:      map[rateGroup]Rate{},
		apiURL:     publicUploadURL,
	}

	tests := []struct {
		name             string
		mockResponses    []MockResponse
		s                *RepoService
		ctx              context.Context
		ref              string
		expectedCommit   *Commit
		expectedResponse *Response
		expectedError    string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           nil,
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: `net/http: nil Context`,
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 401, http.Header{}, `{
					"message": "invalid credentials"
				}`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: `GET /repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e: 401 invalid credentials`,
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, http.Header{}, `{`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			ref:           "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedError: `unexpected EOF`,
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits/6dcb09b5b57875f334f61aebed695e2e4193db5e", 200, header, commitBody1},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:            context.Background(),
			ref:            "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			expectedCommit: &commit1,
			expectedResponse: &Response{
				Pages: expectedPages,
				Rate:  expectedRate,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPTestServer(tc.mockResponses...)
			tc.s.client.apiURL, _ = url.Parse(ts.URL)

			commit, resp, err := tc.s.Commit(tc.ctx, tc.ref)

			if tc.expectedError != "" {
				assert.Nil(t, commit)
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCommit, commit)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
				assert.Equal(t, tc.expectedResponse.Pages, resp.Pages)
				assert.Equal(t, tc.expectedResponse.Rate, resp.Rate)
			}
		})
	}
}

func TestRepoService_Commits(t *testing.T) {
	c := &Client{
		httpClient: &http.Client{},
		rates:      map[rateGroup]Rate{},
		apiURL:     publicUploadURL,
	}

	tests := []struct {
		name             string
		mockResponses    []MockResponse
		s                *RepoService
		ctx              context.Context
		pageSize         int
		pageNo           int
		expectedCommits  []Commit
		expectedResponse *Response
		expectedError    string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           nil,
			pageSize:      10,
			pageNo:        1,
			expectedError: `net/http: nil Context`,
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits", 401, http.Header{}, `{
					"message": "invalid credentials"
				}`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			pageSize:      10,
			pageNo:        1,
			expectedError: `GET /repos/octocat/Hello-World/commits: 401 invalid credentials`,
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits", 200, http.Header{}, `[`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			pageSize:      10,
			pageNo:        1,
			expectedError: `unexpected EOF`,
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/commits", 200, header, commitsBody},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:             context.Background(),
			pageSize:        10,
			pageNo:          1,
			expectedCommits: []Commit{commit2, commit1},
			expectedResponse: &Response{
				Pages: expectedPages,
				Rate:  expectedRate,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPTestServer(tc.mockResponses...)
			tc.s.client.apiURL, _ = url.Parse(ts.URL)

			commits, resp, err := tc.s.Commits(tc.ctx, tc.pageSize, tc.pageNo)

			if tc.expectedError != "" {
				assert.Nil(t, commits)
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedCommits, commits)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
				assert.Equal(t, tc.expectedResponse.Pages, resp.Pages)
				assert.Equal(t, tc.expectedResponse.Rate, resp.Rate)
			}
		})
	}
}

func TestRepoService_Branch(t *testing.T) {
	c := &Client{
		httpClient: &http.Client{},
		rates:      map[rateGroup]Rate{},
		apiURL:     publicUploadURL,
	}

	tests := []struct {
		name             string
		mockResponses    []MockResponse
		s                *RepoService
		ctx              context.Context
		branchName       string
		expectedBranch   *Branch
		expectedResponse *Response
		expectedError    string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           nil,
			branchName:    "main",
			expectedError: `net/http: nil Context`,
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/branches/main", 401, http.Header{}, `{
					"message": "invalid credentials"
				}`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			branchName:    "main",
			expectedError: `GET /repos/octocat/Hello-World/branches/main: 401 invalid credentials`,
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/branches/main", 200, http.Header{}, `{`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			branchName:    "main",
			expectedError: `unexpected EOF`,
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/branches/main", 200, header, branchBody},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:            context.Background(),
			branchName:     "main",
			expectedBranch: &branch,
			expectedResponse: &Response{
				Pages: expectedPages,
				Rate:  expectedRate,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPTestServer(tc.mockResponses...)
			tc.s.client.apiURL, _ = url.Parse(ts.URL)

			branch, resp, err := tc.s.Branch(tc.ctx, tc.branchName)

			if tc.expectedError != "" {
				assert.Nil(t, branch)
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedBranch, branch)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
				assert.Equal(t, tc.expectedResponse.Pages, resp.Pages)
				assert.Equal(t, tc.expectedResponse.Rate, resp.Rate)
			}
		})
	}
}

func TestRepoService_Tags(t *testing.T) {
	c := &Client{
		httpClient: &http.Client{},
		rates:      map[rateGroup]Rate{},
		apiURL:     publicUploadURL,
	}

	tests := []struct {
		name             string
		mockResponses    []MockResponse
		s                *RepoService
		ctx              context.Context
		pageSize         int
		pageNo           int
		expectedTags     []Tag
		expectedResponse *Response
		expectedError    string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           nil,
			pageSize:      10,
			pageNo:        1,
			expectedError: `net/http: nil Context`,
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/tags", 401, http.Header{}, `{
					"message": "invalid credentials"
				}`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			pageSize:      10,
			pageNo:        1,
			expectedError: `GET /repos/octocat/Hello-World/tags: 401 invalid credentials`,
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/tags", 200, http.Header{}, `[`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			pageSize:      10,
			pageNo:        1,
			expectedError: `unexpected EOF`,
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/tags", 200, header, tagsBody},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:          context.Background(),
			pageSize:     10,
			pageNo:       1,
			expectedTags: []Tag{tag},
			expectedResponse: &Response{
				Pages: expectedPages,
				Rate:  expectedRate,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPTestServer(tc.mockResponses...)
			tc.s.client.apiURL, _ = url.Parse(ts.URL)

			tags, resp, err := tc.s.Tags(tc.ctx, tc.pageSize, tc.pageNo)

			if tc.expectedError != "" {
				assert.Nil(t, tags)
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedTags, tags)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
				assert.Equal(t, tc.expectedResponse.Pages, resp.Pages)
				assert.Equal(t, tc.expectedResponse.Rate, resp.Rate)
			}
		})
	}
}

func TestRepoService_Issues(t *testing.T) {
	c := &Client{
		httpClient: &http.Client{},
		rates:      map[rateGroup]Rate{},
		apiURL:     publicUploadURL,
	}

	since, _ := time.Parse(time.RFC3339, "2020-10-20T22:30:00-04:00")

	tests := []struct {
		name             string
		mockResponses    []MockResponse
		s                *RepoService
		ctx              context.Context
		pageSize         int
		pageNo           int
		params           IssuesParams
		expectedIssues   []Issue
		expectedResponse *Response
		expectedError    string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:      nil,
			pageSize: 10,
			pageNo:   1,
			params: IssuesParams{
				State: "closed",
				Since: since,
			},
			expectedError: `net/http: nil Context`,
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues", 401, http.Header{}, `{
					"message": "invalid credentials"
				}`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:      context.Background(),
			pageSize: 10,
			pageNo:   1,
			params: IssuesParams{
				State: "closed",
				Since: since,
			},
			expectedError: `GET /repos/octocat/Hello-World/issues: 401 invalid credentials`,
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, `[`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:      context.Background(),
			pageSize: 10,
			pageNo:   1,
			params: IssuesParams{
				State: "closed",
				Since: since,
			},
			expectedError: `unexpected EOF`,
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues", 200, header, issuesBody},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:      context.Background(),
			pageSize: 10,
			pageNo:   1,
			params: IssuesParams{
				State: "closed",
				Since: since,
			},
			expectedIssues: []Issue{issue2, issue1},
			expectedResponse: &Response{
				Pages: expectedPages,
				Rate:  expectedRate,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPTestServer(tc.mockResponses...)
			tc.s.client.apiURL, _ = url.Parse(ts.URL)

			issues, resp, err := tc.s.Issues(tc.ctx, tc.pageSize, tc.pageNo, tc.params)

			if tc.expectedError != "" {
				assert.Nil(t, issues)
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIssues, issues)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
				assert.Equal(t, tc.expectedResponse.Pages, resp.Pages)
				assert.Equal(t, tc.expectedResponse.Rate, resp.Rate)
			}
		})
	}
}

func TestRepoService_Pull(t *testing.T) {
	c := &Client{
		httpClient: &http.Client{},
		rates:      map[rateGroup]Rate{},
		apiURL:     publicUploadURL,
	}

	tests := []struct {
		name             string
		mockResponses    []MockResponse
		s                *RepoService
		ctx              context.Context
		number           int
		expectedPull     *Pull
		expectedResponse *Response
		expectedError    string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           nil,
			number:        1002,
			expectedError: `net/http: nil Context`,
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls/1002", 401, http.Header{}, `{
					"message": "invalid credentials"
				}`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			number:        1002,
			expectedError: `GET /repos/octocat/Hello-World/pulls/1002: 401 invalid credentials`,
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls/1002", 200, http.Header{}, `{`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			number:        1002,
			expectedError: `unexpected EOF`,
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls/1002", 200, header, pullBody},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:          context.Background(),
			number:       1002,
			expectedPull: &pull,
			expectedResponse: &Response{
				Pages: expectedPages,
				Rate:  expectedRate,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPTestServer(tc.mockResponses...)
			tc.s.client.apiURL, _ = url.Parse(ts.URL)

			pull, resp, err := tc.s.Pull(tc.ctx, tc.number)

			if tc.expectedError != "" {
				assert.Nil(t, pull)
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPull, pull)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
				assert.Equal(t, tc.expectedResponse.Pages, resp.Pages)
				assert.Equal(t, tc.expectedResponse.Rate, resp.Rate)
			}
		})
	}
}

func TestRepoService_Pulls(t *testing.T) {
	c := &Client{
		httpClient: &http.Client{},
		rates:      map[rateGroup]Rate{},
		apiURL:     publicUploadURL,
	}

	tests := []struct {
		name             string
		mockResponses    []MockResponse
		s                *RepoService
		ctx              context.Context
		pageSize         int
		pageNo           int
		params           PullsParams
		expectedPulls    []Pull
		expectedResponse *Response
		expectedError    string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:      nil,
			pageSize: 10,
			pageNo:   1,
			params: PullsParams{
				State: "closed",
			},
			expectedError: `net/http: nil Context`,
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls", 401, http.Header{}, `{
					"message": "invalid credentials"
				}`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:      context.Background(),
			pageSize: 10,
			pageNo:   1,
			params: PullsParams{
				State: "closed",
			},
			expectedError: `GET /repos/octocat/Hello-World/pulls: 401 invalid credentials`,
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, `[`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:      context.Background(),
			pageSize: 10,
			pageNo:   1,
			params: PullsParams{
				State: "closed",
			},
			expectedError: `unexpected EOF`,
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/pulls", 200, header, pullsBody},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:      context.Background(),
			pageSize: 10,
			pageNo:   1,
			params: PullsParams{
				State: "closed",
			},
			expectedPulls: []Pull{pull},
			expectedResponse: &Response{
				Pages: expectedPages,
				Rate:  expectedRate,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPTestServer(tc.mockResponses...)
			tc.s.client.apiURL, _ = url.Parse(ts.URL)

			pulls, resp, err := tc.s.Pulls(tc.ctx, tc.pageSize, tc.pageNo, tc.params)

			if tc.expectedError != "" {
				assert.Nil(t, pulls)
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedPulls, pulls)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
				assert.Equal(t, tc.expectedResponse.Pages, resp.Pages)
				assert.Equal(t, tc.expectedResponse.Rate, resp.Rate)
			}
		})
	}
}

func TestRepoService_Events(t *testing.T) {
	c := &Client{
		httpClient: &http.Client{},
		rates:      map[rateGroup]Rate{},
		apiURL:     publicUploadURL,
	}

	tests := []struct {
		name             string
		mockResponses    []MockResponse
		s                *RepoService
		ctx              context.Context
		number           int
		pageSize         int
		pageNo           int
		expectedEvents   []Event
		expectedResponse *Response
		expectedError    string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           nil,
			number:        1001,
			pageSize:      10,
			pageNo:        1,
			expectedError: `net/http: nil Context`,
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 401, http.Header{}, `{
					"message": "invalid credentials"
				}`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			number:        1001,
			pageSize:      10,
			pageNo:        1,
			expectedError: `GET /repos/octocat/Hello-World/issues/1001/events: 401 invalid credentials`,
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, `[`},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:           context.Background(),
			number:        1001,
			pageSize:      10,
			pageNo:        1,
			expectedError: `unexpected EOF`,
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, header, eventsBody},
			},
			s: &RepoService{
				client: c,
				owner:  "octocat",
				repo:   "Hello-World",
			},
			ctx:            context.Background(),
			number:         1001,
			pageSize:       10,
			pageNo:         1,
			expectedEvents: []Event{event2, event1},
			expectedResponse: &Response{
				Pages: expectedPages,
				Rate:  expectedRate,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := newHTTPTestServer(tc.mockResponses...)
			tc.s.client.apiURL, _ = url.Parse(ts.URL)

			events, resp, err := tc.s.Events(tc.ctx, tc.number, tc.pageSize, tc.pageNo)

			if tc.expectedError != "" {
				assert.Nil(t, events)
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedEvents, events)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
				assert.Equal(t, tc.expectedResponse.Pages, resp.Pages)
				assert.Equal(t, tc.expectedResponse.Rate, resp.Rate)
			}
		})
	}
}
