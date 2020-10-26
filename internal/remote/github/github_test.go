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
			"created_at":  "2020-10-15T15:00:00Z",
			"updated_at": "2020-10-22T22:00:00Z",
			"closed_at": "2020-10-20T20:00:00Z",
			"merged_at": "2020-10-20T20:00:00Z",
			"merge_commit_sha": "e5bd3914e2e596debea16f433f57875b5b90bcd6",
			"head": {
				"label": "octocat:new-topic",
				"ref": "new-topic",
				"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e"
			},
			"base": {
				"label": "octocat:master",
				"ref": "master",
				"sha": "6dcb09b5b57875f334f61aebed695e2e4193db5e"
			}
		}
	]`

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
				"login": "octocat",
				"id": 1,
				"type": "User"
			},
			"event": "merged",
			"commit_id": "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			"created_at": "2020-10-20T20:00:00Z"
		}
	]`
)

var (
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
		PullRequest: nil,
		CreatedAt:   parseGitHubTime("2020-10-10T10:00:00Z"),
		UpdatedAt:   parseGitHubTime("2020-10-20T20:00:00Z"),
		ClosedAt:    nil,
	}

	gitHubIssue2 = issue{
		ID:     2,
		Number: 1002,
		State:  "closed",
		Locked: false,
		Title:  "Fixed a bug",
		Body:   "I made this to work as expected!",
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
		PullRequest: &pullRequestURLs{
			URL: "https://api.github.com/repos/octocat/Hello-World/pulls/1002",
		},
		CreatedAt: parseGitHubTime("2020-10-15T15:00:00Z"),
		UpdatedAt: parseGitHubTime("2020-10-22T22:00:00Z"),
		ClosedAt:  parseGitHubTimePtr("2020-10-20T20:00:00Z"),
	}

	gitHubPull1 = pullRequest{
		ID:     1,
		Number: 1002,
		State:  "closed",
		Draft:  false,
		Locked: false,
		Title:  "Fixed a bug",
		Body:   "I made this to work as expected!",
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
		Base: reference{
			Label: "octocat:master",
			Ref:   "master",
			Sha:   "6dcb09b5b57875f334f61aebed695e2e4193db5e",
		},
		Head: reference{
			Label: "octocat:new-topic",
			Ref:   "new-topic",
			Sha:   "6dcb09b5b57875f334f61aebed695e2e4193db5e",
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
			ID:    1,
			Login: "octocat",
			Type:  "User",
		},
		CreatedAt: parseGitHubTime("2020-10-20T20:00:00Z"),
	}

	remoteIssue = remote.Change{
		Number:    1001,
		Title:     "Found a bug",
		Labels:    []string{"bug"},
		Milestone: "v1.0",
		Timestamp: time.Time{},
	}

	remoteMerge = remote.Change{
		Number:    1002,
		Title:     "Fixed a bug",
		Labels:    []string{"bug"},
		Milestone: "v1.0",
		Timestamp: parseGitHubTime("2020-10-20T20:00:00Z"),
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
		})
	}
}

func TestRepo_CreateRequest(t *testing.T) {
	tests := []struct {
		name          string
		accessToken   string
		ctx           context.Context
		method        string
		endpoint      string
		body          io.Reader
		expectedError string
	}{
		{
			name:          "RequestError",
			accessToken:   "github-token",
			ctx:           nil,
			method:        "",
			endpoint:      "",
			body:          nil,
			expectedError: "net/http: nil Context",
		},
		{
			name:          "Success",
			accessToken:   "github-token",
			ctx:           context.Background(),
			method:        "GET",
			endpoint:      "/users/octocat",
			body:          nil,
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				accessToken: tc.accessToken,
			}

			req, err := r.createRequest(tc.ctx, tc.method, tc.endpoint, tc.body)

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

func TestRepo_MakeRequest(t *testing.T) {
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

func TestRepo_CheckScopes(t *testing.T) {
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

			if len(tc.mockResponses) > 0 {
				ts := createMockHTTPServer(tc.mockResponses...)
				defer ts.Close()
				r.apiURL = ts.URL
			}

			err := r.checkScopes(tc.ctx, tc.scopes...)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_FetchIssuesPageCount(t *testing.T) {
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

			if len(tc.mockResponses) > 0 {
				ts := createMockHTTPServer(tc.mockResponses...)
				defer ts.Close()
				r.apiURL = ts.URL
			}

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

func TestRepo_FetchIssues(t *testing.T) {
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

			if len(tc.mockResponses) > 0 {
				ts := createMockHTTPServer(tc.mockResponses...)
				defer ts.Close()
				r.apiURL = ts.URL
			}

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

func TestRepo_FetchPullsPageCount(t *testing.T) {
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

			if len(tc.mockResponses) > 0 {
				ts := createMockHTTPServer(tc.mockResponses...)
				defer ts.Close()
				r.apiURL = ts.URL
			}

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

func TestRepo_FetchPulls(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		pageNo        int
		expectedError string
		expectedPulls []pullRequest
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
			expectedPulls: []pullRequest{gitHubPull1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			if len(tc.mockResponses) > 0 {
				ts := createMockHTTPServer(tc.mockResponses...)
				defer ts.Close()
				r.apiURL = ts.URL
			}

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

func TestRepo_FindEvent(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		ctx           context.Context
		issueNumber   int
		eventName     string
		expectedError string
		expectedEvent *event
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
			expectedEvent: &gitHubEvent1,
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
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			if len(tc.mockResponses) > 0 {
				ts := createMockHTTPServer(tc.mockResponses...)
				defer ts.Close()
				r.apiURL = ts.URL
			}

			event, err := r.findEvent(tc.ctx, tc.issueNumber, tc.eventName)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedEvent, event)
			} else {
				assert.Nil(t, event)
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
		expectedIssues remote.Changes
		expectedMerges remote.Changes
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
			since:         since,
			expectedError: "HEAD /repos/octocat/Hello-World/issues 404: ",
		},
		{
			name: "FetchPullsPageCountFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "HEAD /repos/octocat/Hello-World/pulls 404: ",
		},
		{
			name: "FetchIssuesFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"HEAD", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, mockGitHubPullsBody},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "GET /repos/octocat/Hello-World/issues 405: ",
		},
		{
			name: "FetchPullsFails",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"HEAD", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "GET /repos/octocat/Hello-World/pulls 405: ",
		},
		{
			name: "FindEventFails#1001",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"HEAD", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, mockGitHubPullsBody},
				{"GET", "/repos/octocat/Hello-World/issues/1002/events", 200, http.Header{}, mockGitHubEventsBody2},
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
				{"HEAD", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, mockGitHubPullsBody},
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, mockGitHubEventsBody1},
			},
			ctx:           context.Background(),
			since:         since,
			expectedError: "GET /repos/octocat/Hello-World/issues/1002/events 404: 404 page not found\n",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, http.Header{"X-OAuth-Scopes": []string{"repo"}}, ``},
				{"HEAD", "/repos/octocat/Hello-World/issues", 200, http.Header{}, ``},
				{"HEAD", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, ``},
				{"GET", "/repos/octocat/Hello-World/issues", 200, http.Header{}, mockGitHubIssuesBody},
				{"GET", "/repos/octocat/Hello-World/pulls", 200, http.Header{}, mockGitHubPullsBody},
				{"GET", "/repos/octocat/Hello-World/issues/1001/events", 200, http.Header{}, mockGitHubEventsBody1},
				{"GET", "/repos/octocat/Hello-World/issues/1002/events", 200, http.Header{}, mockGitHubEventsBody2},
			},
			ctx:            context.Background(),
			since:          since,
			expectedIssues: remote.Changes{remoteIssue},
			expectedMerges: remote.Changes{remoteMerge},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				client: new(http.Client),
				path:   "octocat/Hello-World",
			}

			if len(tc.mockResponses) > 0 {
				ts := createMockHTTPServer(tc.mockResponses...)
				defer ts.Close()
				r.apiURL = ts.URL
			}

			issues, merges, err := r.FetchIssuesAndMerges(tc.ctx, tc.since)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedIssues, issues)
				assert.Equal(t, tc.expectedMerges, merges)
			} else {
				assert.Nil(t, issues)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
