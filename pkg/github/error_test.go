package github

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResponseError(t *testing.T) {
	req, _ := http.NewRequest("PATCH", "/user", nil)

	tests := []struct {
		name          string
		err           *ResponseError
		expectedError string
	}{
		{
			name: "OK",
			err: &ResponseError{
				Response: &http.Response{
					StatusCode: 400,
					Request:    req,
				},
				Message:          "Problems parsing JSON",
				DocumentationURL: "https://docs.github.com/rest/reference/users#update-the-authenticated-user",
			},
			expectedError: "PATCH /user: 400 Problems parsing JSON",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.EqualError(t, tc.err, tc.expectedError)
		})
	}
}

func TestAuthError(t *testing.T) {
	req, _ := http.NewRequest("GET", "/user", nil)

	tests := []struct {
		name          string
		err           *AuthError
		expectedError string
	}{
		{
			name:          "WithoutResponseError",
			err:           &AuthError{},
			expectedError: "requires authentication",
		},
		{
			name: "WithResponseError",
			err: &AuthError{
				err: &ResponseError{
					Response: &http.Response{
						StatusCode: 401,
						Request:    req,
					},
					Message:          "Requires authentication",
					DocumentationURL: "https://docs.github.com/rest/reference/users#get-the-authenticated-user",
				},
			},
			expectedError: "GET /user: 401 Requires authentication",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.EqualError(t, tc.err, tc.expectedError)
			assert.Equal(t, tc.err.err, tc.err.Unwrap())
		})
	}
}

func TestRateLimitError(t *testing.T) {
	req, _ := http.NewRequest("GET", "/user", nil)

	tests := []struct {
		name          string
		err           *RateLimitError
		expectedError string
	}{
		{
			name: "WithoutResponseError",
			err: &RateLimitError{
				Request: req,
				Rate: Rate{
					Limit:     5000,
					Used:      5000,
					Remaining: 0,
					Reset:     Epoch(1605125898),
				},
			},
			expectedError: "GET /user: rate limit 5000 used: rate limit will reset at " + time.Unix(1605125898, 0).Format("15:04:05"),
		},
		{
			name: "WithResponseError",
			err: &RateLimitError{
				err: &ResponseError{
					Response: &http.Response{
						StatusCode: 403,
						Request:    req,
					},
					Message:          "API rate limit exceeded",
					DocumentationURL: "https://developer.github.com/v3/#rate-limiting",
				},
				Request: req,
				Rate: Rate{
					Limit:     5000,
					Used:      5000,
					Remaining: 0,
					Reset:     Epoch(1605125898),
				},
			},
			expectedError: "GET /user: rate limit 5000 used: rate limit will reset at " + time.Unix(1605125898, 0).Format("15:04:05"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.EqualError(t, tc.err, tc.expectedError)
			assert.Equal(t, tc.err.err, tc.err.Unwrap())
		})
	}
}

func TestRateLimitAbuseError(t *testing.T) {
	req, _ := http.NewRequest("GET", "/user", nil)

	tests := []struct {
		name          string
		err           *RateLimitAbuseError
		expectedError string
	}{
		{
			name: "WithoutResponseError",
			err: &RateLimitAbuseError{
				Rate: Rate{
					Limit:     5000,
					Used:      1000,
					Remaining: 4000,
					Reset:     Epoch(1605125898),
				},
				RetryAfter: 30 * time.Second,
			},
			expectedError: "rate limit is abused",
		},
		{
			name: "WithResponseError",
			err: &RateLimitAbuseError{
				err: &ResponseError{
					Response: &http.Response{
						StatusCode: 403,
						Request:    req,
					},
					Message:          "You have triggered an abuse detection mechanism",
					DocumentationURL: "https://developer.github.com/v3/#abuse-rate-limits",
				},
				Rate: Rate{
					Limit:     5000,
					Used:      1000,
					Remaining: 4000,
					Reset:     Epoch(1605125898),
				},
				RetryAfter: 30 * time.Second,
			},
			expectedError: "GET /user: 403 You have triggered an abuse detection mechanism",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.EqualError(t, tc.err, tc.expectedError)
			assert.Equal(t, tc.err.err, tc.err.Unwrap())
		})
	}
}

func TestNotFoundError(t *testing.T) {
	req, _ := http.NewRequest("GET", "/users/octocat", nil)

	tests := []struct {
		name          string
		err           *NotFoundError
		expectedError string
	}{
		{
			name:          "WithoutResponseError",
			err:           &NotFoundError{},
			expectedError: "resource not found",
		},
		{
			name: "WithResponseError",
			err: &NotFoundError{
				err: &ResponseError{
					Response: &http.Response{
						StatusCode: 404,
						Request:    req,
					},
					Message:          "Not Found",
					DocumentationURL: "https://docs.github.com/rest",
				},
			},
			expectedError: "GET /users/octocat: 404 Not Found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.EqualError(t, tc.err, tc.expectedError)
			assert.Equal(t, tc.err.err, tc.err.Unwrap())
		})
	}
}
