package remote

import (
	"context"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGitHubRepo(t *testing.T) {
	tests := []struct {
		name        string
		logger      *log.Logger
		path        string
		accessToken string
	}{
		{
			name:        "OK",
			logger:      &log.Logger{},
			path:        "moorara/changelog",
			accessToken: "github-access-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewGitHubRepo(tc.logger, tc.path, tc.accessToken)
			assert.NotNil(t, repo)

			r, ok := repo.(*githubRepo)
			assert.True(t, ok)

			assert.Equal(t, tc.logger, r.logger)
			assert.NotNil(t, r.client)
			assert.Equal(t, githubAPIURL, r.apiURL)
			assert.Equal(t, tc.path, r.path)
			assert.Equal(t, tc.accessToken, r.accessToken)
		})
	}
}

func TestGitHubRepo_CreateRequest(t *testing.T) {
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
			r := &githubRepo{
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

func TestGitHubRepo_MakeRequest(t *testing.T) {
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
				{"GET", "/users/{username}", 400, `bad request`},
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
					"GET", "/users/{username}", 200, `{
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
			r := &githubRepo{
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

func TestGitHub_CheckScopes(t *testing.T) {
	tests := []struct {
		name          string
		mockResponses []MockResponse
		accessToken   string
		scopes        []scope
		expectedError string
	}{
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 401, `bad credentials`},
			},
			accessToken:   "github-token",
			scopes:        []scope{scopeRepo},
			expectedError: "HEAD /user 401: ",
		},
		{
			name: "MissingScope",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, ``},
			},
			accessToken:   "github-token",
			scopes:        []scope{scopeRepo},
			expectedError: "access token does not have the scope: repo",
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"HEAD", "/user", 200, ``},
			},
			accessToken:   "github-token",
			scopes:        []scope{},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &githubRepo{
				client:      new(http.Client),
				accessToken: tc.accessToken,
			}

			if len(tc.mockResponses) > 0 {
				ts := createMockHTTPServer(tc.mockResponses...)
				defer ts.Close()
				r.apiURL = ts.URL
			}

			err := r.checkScopes(tc.scopes...)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
