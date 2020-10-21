package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/pkg/log"
	"github.com/moorara/changelog/pkg/xhttp"
)

const (
	githubAPIURL      = "https://api.github.com"
	userAgentHeader   = "gelato"
	acceptHeader      = "application/vnd.github.v3+json"
	contentTypeHeader = "application/json"
	pageSize          = 50
)

var (
	relLastRE = regexp.MustCompile(`<https://api.github.com/[\w\?=&-_]+page=(\d+)>; rel="last"`)
)

// scope represents a GitHub authorization scope.
// See https://docs.github.com/en/developers/apps/scopes-for-oauth-apps
type scope string

const (
	// scopeRepo grants full access to private and public repositories. It also grants ability to manage user projects.
	scopeRepo scope = "repo"
)

type (
	user struct {
		ID         int       `json:"id"`
		Login      string    `json:"login"`
		Type       string    `json:"type"`
		Email      string    `json:"email"`
		Name       string    `json:"name"`
		URL        string    `json:"url"`
		HTMLURL    string    `json:"html_url"`
		AvatarURL  string    `json:"avatar_url"`
		GravatarID string    `json:"gravatar_id"`
		CreatedAt  time.Time `json:"created_at"`
		UpdatedAt  time.Time `json:"updated_at"`
	}

	label struct {
		ID          int    `json:"id"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Default     bool   `json:"default"`
		URL         string `json:"url"`
	}

	milestone struct {
		ID           int       `json:"id"`
		Number       int       `json:"number"`
		State        string    `json:"state"`
		Title        string    `json:"title"`
		Description  string    `json:"description"`
		Creator      user      `json:"creator"`
		OpenIssues   int       `json:"open_issues"`
		ClosedIssues int       `json:"closed_issues"`
		DueOn        time.Time `json:"due_on"`
		URL          string    `json:"url"`
		HTMLURL      string    `json:"html_url"`
		LabelsURL    string    `json:"labels_url"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		ClosedAt     time.Time `json:"closed_at"`
	}

	pullRequestURLs struct {
		URL      string `json:"url"`
		HTMLURL  string `json:"html_url"`
		DiffURL  string `json:"diff_url"`
		PatchURL string `json:"patch_url"`
	}

	issue struct {
		ID                int              `json:"id"`
		Number            int              `json:"number"`
		State             string           `json:"state"`
		Title             string           `json:"title"`
		Body              string           `json:"body"`
		URL               string           `json:"url"`
		HTMLURL           string           `json:"html_url"`
		LabelsURL         string           `json:"labels_url"`
		EventsURL         string           `json:"events_url"`
		CommentsURL       string           `json:"comments_url"`
		RepoURL           string           `json:"repository_url"`
		User              user             `json:"user"`
		Assignee          user             `json:"assignee"`
		Assignees         []user           `json:"assignees"`
		Labels            []label          `json:"labels"`
		Milestone         milestone        `json:"milestone"`
		Locked            bool             `json:"locked"`
		LockReason        string           `json:"active_lock_reason"`
		AuthorAssociation string           `json:"author_association"`
		Comments          int              `json:"comments"`
		PullRequest       *pullRequestURLs `json:"pull_request"`
		CreatedAt         time.Time        `json:"created_at"`
		UpdatedAt         time.Time        `json:"updated_at"`
		ClosedAt          time.Time        `json:"closed_at"`
	}
)

// repo implements the remote.Repo interface for GitHub.
type repo struct {
	logger      log.Logger
	client      *http.Client
	apiURL      string
	path        string
	accessToken string
}

// NewRepo creates a new GitHub repository.
func NewRepo(logger log.Logger, path, accessToken string) remote.Repo {
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	return &repo{
		logger:      logger,
		client:      client,
		apiURL:      githubAPIURL,
		path:        path,
		accessToken: accessToken,
	}
}

func (r *repo) createRequest(ctx context.Context, method, endpoint string, body io.Reader) (*http.Request, error) {
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

func (r *repo) makeRequest(req *http.Request, expectedStatusCode int) (*http.Response, error) {
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != expectedStatusCode {
		return nil, xhttp.NewClientError(resp)
	}

	return resp, nil
}

func (r *repo) checkScopes(ctx context.Context, scopes ...scope) error {
	// Call an endpoint to get the OAuth scopes of the access token from the headers
	// See https://docs.github.com/en/developers/apps/scopes-for-oauth-apps

	r.logger.Debugf("Checking GitHub token scopes: %s", scopes)

	req, err := r.createRequest(ctx, "HEAD", "/user", nil)
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

	r.logger.Infof("GitHub token scopes verified: %s", scopes)

	return nil
}

func (r *repo) fetchPageCount(ctx context.Context, since time.Time) (int, error) {
	// See https://docs.github.com/en/rest/reference/issues#list-repository-issues

	r.logger.Debug("Fetching the total number of issue pages ...")

	url := fmt.Sprintf("/repos/%s/issues", r.path)
	req, err := r.createRequest(ctx, "HEAD", url, nil)
	if err != nil {
		return -1, err
	}

	q := req.URL.Query()
	q.Add("state", "closed")
	q.Add("per_page", strconv.Itoa(pageSize))
	if !since.IsZero() {
		q.Add("since", since.Format(time.RFC3339))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := r.makeRequest(req, 200)
	if err != nil {
		return -1, err
	}

	count := 1

	if val := resp.Header.Get("Link"); val != "" {
		sm := relLastRE.FindStringSubmatch(val)
		if len(sm) != 2 {
			return -1, fmt.Errorf("invalid Link header received from GitHub: %s", val)
		}

		// sm[1] is guaranteed to be a number at this point
		count, _ = strconv.Atoi(sm[1])
	}

	r.logger.Debugf("Fetched the total number of pages: %d", count)

	return count, nil
}

func (r *repo) fetchIssues(ctx context.Context, since time.Time, pageNo int) ([]issue, error) {
	// See https://docs.github.com/en/rest/reference/issues#list-repository-issues

	r.logger.Debugf("Fetching issues page %d ...", pageNo)

	url := fmt.Sprintf("/repos/%s/issues", r.path)
	req, err := r.createRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("state", "closed")
	q.Add("per_page", strconv.Itoa(pageSize))
	q.Add("page", strconv.Itoa(pageNo))
	if !since.IsZero() {
		q.Add("since", since.Format(time.RFC3339))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := r.makeRequest(req, 200)
	if err != nil {
		return nil, err
	}

	issues := []issue{}
	if err = json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
	}

	r.logger.Debugf("Fetched issues page %d: %d", pageNo, len(issues))

	return issues, nil
}

// FetchClosedIssuesAndMerges retrieves all issues and merges requests for a GitHub repository.
// See https://docs.github.com/en/rest/reference/issues#list-repository-issues
func (r *repo) FetchClosedIssuesAndMerges(ctx context.Context, since time.Time) ([]remote.Issue, []remote.Merge, error) {
	if err := r.checkScopes(ctx, scopeRepo); err != nil {
		return nil, nil, err
	}

	count, err := r.fetchPageCount(ctx, since)
	if err != nil {
		return nil, nil, err
	}

	group, ctx := errgroup.WithContext(ctx)
	gitHubIssues := make([]issue, 0)

	for i := 1; i <= count; i++ {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		group.Go(func() error {
			res, err := r.fetchIssues(ctx, since, i)
			if err != nil {
				return err
			}
			gitHubIssues = append(gitHubIssues, res...)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, nil, err
	}

	r.logger.Debugf("All issues are fetched: %d", len(gitHubIssues))

	// TODO:
	issues, merges := []remote.Issue{}, []remote.Merge{}

	return issues, merges, nil
}
