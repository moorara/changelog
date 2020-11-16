package github

import (
	"fmt"
	"net/http"
	"time"
)

// ResponseError is a generic error for HTTP calls to GitHub API v3.
// See https://docs.github.com/en/free-pro-team@latest/rest/overview/resources-in-the-rest-api#client-errors
type ResponseError struct {
	Response         *http.Response
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url,omitempty"`
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("%s %s: %d %s",
		e.Response.Request.Method, e.Response.Request.URL.Path,
		e.Response.StatusCode, e.Message,
	)
}

// AuthError occurs when there is an authentication problem.
type AuthError struct {
	err *ResponseError
}

func (e *AuthError) Error() string {
	if e.err == nil {
		return "requires authentication"
	}

	return e.err.Error()
}

func (e *AuthError) Unwrap() error {
	return e.err
}

// RateLimitError occurs when there is no remaining call in the current hour for the authenticated user.
// See https://docs.github.com/rest/overview/resources-in-the-rest-api#rate-limiting
type RateLimitError struct {
	err     *ResponseError
	Request *http.Request
	Rate    Rate
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("%s %s: rate limit %d used: rate limit will reset at %s",
		e.Request.Method, e.Request.URL.Path, e.Rate.Limit, e.Rate.Reset,
	)
}

func (e *RateLimitError) Unwrap() error {
	return e.err
}

// RateLimitAbuseError occurs when best practices for using the legitimate rate limit are not observed.
// See https://docs.github.com/rest/overview/resources-in-the-rest-api#abuse-rate-limits
type RateLimitAbuseError struct {
	err        *ResponseError
	Rate       Rate
	RetryAfter time.Duration
}

func (e *RateLimitAbuseError) Error() string {
	if e.err == nil {
		return "rate limit is abused"
	}

	return e.err.Error()
}

func (e *RateLimitAbuseError) Unwrap() error {
	return e.err
}

// NotFoundError occurs when a resource is not found.
type NotFoundError struct {
	err *ResponseError
}

func (e *NotFoundError) Error() string {
	if e.err == nil {
		return "resource not found"
	}

	return e.err.Error()
}

func (e *NotFoundError) Unwrap() error {
	return e.err
}
