package github

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	userBody1 = `{
		"login": "octocat",
		"id": 1,
		"url": "https://api.github.com/users/octocat",
		"html_url": "https://github.com/octocat",
		"type": "User",
		"site_admin": false,
		"name": "The Octocat",
		"email": "octocat@github.com"
	}`
)

var (
	user1 = User{
		ID:      1,
		Login:   "octocat",
		Type:    "User",
		Email:   "octocat@github.com",
		Name:    "The Octocat",
		URL:     "https://api.github.com/users/octocat",
		HTMLURL: "https://github.com/octocat",
	}
)

func TestUserService_Get(t *testing.T) {
	c := &Client{
		httpClient: &http.Client{},
		rates:      map[rateGroup]Rate{},
		apiURL:     publicUploadURL,
	}

	tests := []struct {
		name             string
		mockResponses    []MockResponse
		s                *UsersService
		ctx              context.Context
		username         string
		expectedUser     *User
		expectedResponse *Response
		expectedError    string
	}{
		{
			name:          "NilContext",
			mockResponses: []MockResponse{},
			s: &UsersService{
				client: c,
			},
			ctx:           nil,
			username:      "octocat",
			expectedError: `net/http: nil Context`,
		},
		{
			name: "InvalidStatusCode",
			mockResponses: []MockResponse{
				{"GET", "/users/octocat", 401, http.Header{}, `{
					"message": "invalid credentials"
				}`},
			},
			s: &UsersService{
				client: c,
			},
			ctx:           context.Background(),
			username:      "octocat",
			expectedError: `GET /users/octocat: 401 invalid credentials`,
		},
		{
			name: "ّInvalidResponse",
			mockResponses: []MockResponse{
				{"GET", "/users/octocat", 200, http.Header{}, `{`},
			},
			s: &UsersService{
				client: c,
			},
			ctx:           context.Background(),
			username:      "octocat",
			expectedError: `unexpected EOF`,
		},
		{
			name: "Success",
			mockResponses: []MockResponse{
				{"GET", "/users/octocat", 200, header, userBody1},
			},
			s: &UsersService{
				client: c,
			},
			ctx:          context.Background(),
			username:     "octocat",
			expectedUser: &user1,
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

			user, resp, err := tc.s.Get(tc.ctx, tc.username)

			if tc.expectedError != "" {
				assert.Nil(t, user)
				assert.Nil(t, resp)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedUser, user)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Response)
				assert.Equal(t, tc.expectedResponse.Pages, resp.Pages)
				assert.Equal(t, tc.expectedResponse.Rate, resp.Rate)
			}
		})
	}
}
