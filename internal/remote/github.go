package remote

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/moorara/changelog/pkg/xhttp"
)

const (
	githubAPIURL      = "https://api.github.com"
	userAgentHeader   = "gelato"
	acceptHeader      = "application/vnd.github.v3+json"
	contentTypeHeader = "application/json"
)

// scope represents a GitHub authorization scope.
// See https://docs.github.com/en/developers/apps/scopes-for-oauth-apps
type scope string

const (
	// scopeRepo grants full access to private and public repositories. It also grants ability to manage user projects.
	scopeRepo scope = "repo"
)

// githubRepo implements the Repo interface for GitHub.
type githubRepo struct {
	logger      *log.Logger
	client      *http.Client
	apiURL      string
	path        string
	accessToken string
}

// NewGitHubRepo creates a new GitHub repository.
func NewGitHubRepo(logger *log.Logger, path, accessToken string) Repo {
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	return &githubRepo{
		logger:      logger,
		client:      client,
		apiURL:      githubAPIURL,
		path:        path,
		accessToken: accessToken,
	}
}

func (r *githubRepo) createRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
	url := r.apiURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "token "+r.accessToken)
	req.Header.Set("User-Agent", userAgentHeader) // See https://docs.github.com/en/rest/overview/resources-in-the-rest-api#user-agent-required
	req.Header.Set("Accept", acceptHeader)        // See https://docs.github.com/en/rest/overview/media-types
	req.Header.Set("Content-Type", contentTypeHeader)

	return req, nil
}

func (r *githubRepo) makeRequest(req *http.Request, expectedStatusCode int) (*http.Response, error) {
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedStatusCode {
		return nil, xhttp.NewClientError(resp)
	}

	return resp, nil
}

func (r *githubRepo) checkScopes(scopes ...scope) error {
	// Call an endpoint to get the OAuth scopes of the access token from the headers
	// See https://docs.github.com/en/developers/apps/scopes-for-oauth-apps

	req, err := r.createRequest(context.Background(), "HEAD", "/user", nil)
	if err != nil {
		return err
	}

	resp, err := r.makeRequest(req, 200)
	if err != nil {
		return err
	}

	// Ensure the access token has all the required OAuth scopes
	oauthScopes := resp.Header.Get("X-OAuth-Scopes")
	for _, s := range scopes {
		if !strings.Contains(oauthScopes, string(s)) {
			return fmt.Errorf("access token does not have the scope: %s", s)
		}
	}

	return nil
}

// FetchClosedIssues
func (r *githubRepo) FetchClosedIssuesAndMerges() ([]Issue, []Merge, error) {
	if err := r.checkScopes(scopeRepo); err != nil {
		return nil, nil, err
	}

	return []Issue{}, []Merge{}, nil
}
