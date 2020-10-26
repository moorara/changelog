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
	relNextRE = regexp.MustCompile(`<(https://api.github.com/[\w\?=&-_]+page=\d+)>; rel="next"`)
	relLastRE = regexp.MustCompile(`<https://api.github.com/[\w\?=&-_]+page=(\d+)>; rel="last"`)
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

func (r *repo) fetchIssuesPageCount(ctx context.Context, since time.Time) (int, error) {
	// See https://docs.github.com/en/rest/reference/issues#list-repository-issues

	r.logger.Debug("Fetching the total number of pages of GitHub issues ...")

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

	if link := resp.Header.Get("Link"); link != "" {
		sm := relLastRE.FindStringSubmatch(link)
		if len(sm) != 2 {
			return -1, fmt.Errorf("invalid Link header received from GitHub: %s", link)
		}

		// sm[1] is guaranteed to be a number at this point
		count, _ = strconv.Atoi(sm[1])
	}

	r.logger.Debugf("Fetched the total number of pages of GitHub issues: %d", count)

	return count, nil
}

func (r *repo) fetchIssues(ctx context.Context, since time.Time, pageNo int) ([]issue, error) {
	// See https://docs.github.com/en/rest/reference/issues#list-repository-issues

	r.logger.Debugf("Fetching GitHub issues page %d ...", pageNo)

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

	r.logger.Debugf("Fetched GitHub issues page %d: %d", pageNo, len(issues))

	return issues, nil
}

func (r *repo) fetchPullsPageCount(ctx context.Context) (int, error) {
	// See https://docs.github.com/en/free-pro-team@latest/rest/reference/pulls#list-pull-requests

	r.logger.Debug("Fetching the total number of pages of GitHub pull requests ...")

	url := fmt.Sprintf("/repos/%s/pulls", r.path)
	req, err := r.createRequest(ctx, "HEAD", url, nil)
	if err != nil {
		return -1, err
	}

	q := req.URL.Query()
	q.Add("state", "closed")
	q.Add("per_page", strconv.Itoa(pageSize))
	req.URL.RawQuery = q.Encode()

	resp, err := r.makeRequest(req, 200)
	if err != nil {
		return -1, err
	}

	count := 1

	if link := resp.Header.Get("Link"); link != "" {
		sm := relLastRE.FindStringSubmatch(link)
		if len(sm) != 2 {
			return -1, fmt.Errorf("invalid Link header received from GitHub: %s", link)
		}

		// sm[1] is guaranteed to be a number at this point
		count, _ = strconv.Atoi(sm[1])
	}

	r.logger.Debugf("Fetched the total number of pages of GitHub pull requests: %d", count)

	return count, nil
}

func (r *repo) fetchPulls(ctx context.Context, pageNo int) ([]pullRequest, error) {
	// See https://docs.github.com/en/free-pro-team@latest/rest/reference/pulls#list-pull-requests

	r.logger.Debugf("Fetching GitHub pulls page %d ...", pageNo)

	url := fmt.Sprintf("/repos/%s/pulls", r.path)
	req, err := r.createRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("state", "closed")
	q.Add("per_page", strconv.Itoa(pageSize))
	q.Add("page", strconv.Itoa(pageNo))
	req.URL.RawQuery = q.Encode()

	resp, err := r.makeRequest(req, 200)
	if err != nil {
		return nil, err
	}

	pulls := []pullRequest{}
	if err = json.NewDecoder(resp.Body).Decode(&pulls); err != nil {
		return nil, err
	}

	r.logger.Debugf("Fetched GitHub pulls page %d: %d", pageNo, len(pulls))

	return pulls, nil
}

func (r *repo) findEvent(ctx context.Context, number int, name string) (*event, error) {
	// See https://docs.github.com/en/free-pro-team@latest/rest/reference/issues#list-issue-events

	r.logger.Debugf("Finding %s event for issue %d ...", name, number)

	url := fmt.Sprintf("/repos/%s/issues/%d/events", r.path, number)
	for {
		req, err := r.createRequest(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("per_page", strconv.Itoa(pageSize))
		req.URL.RawQuery = q.Encode()

		resp, err := r.makeRequest(req, 200)
		if err != nil {
			return nil, err
		}

		events := []event{}
		if err = json.NewDecoder(resp.Body).Decode(&events); err != nil {
			return nil, err
		}

		for _, e := range events {
			if e.Event == name {
				r.logger.Debugf("Found %s event for issue %d ...", name, number)
				return &e, nil
			}
		}

		// Navigate to the next page

		link := resp.Header.Get("Link")
		if link == "" {
			break
		}

		sm := relNextRE.FindStringSubmatch(link)
		if len(sm) != 2 {
			break
		}

		// Update the url for fetching the next page
		url = sm[1]
	}

	return nil, fmt.Errorf("GitHub %s event for issue %d not found", name, number)
}

// FetchIssuesAndMerges retrieves all closed issues and merged pull requests for a GitHub repository.
// See https://docs.github.com/en/rest/reference/issues#list-repository-issues
func (r *repo) FetchIssuesAndMerges(ctx context.Context, since time.Time) (remote.Changes, remote.Changes, error) {
	if since.IsZero() {
		r.logger.Info("Fetching GitHub issues and pull requests since the beginning ...")
	} else {
		r.logger.Infof("Fetching GitHub issues and pull requests since %s ...", since.Format(time.RFC3339))
	}

	if err := r.checkScopes(ctx, scopeRepo); err != nil {
		return nil, nil, err
	}

	issuePages, err := r.fetchIssuesPageCount(ctx, since)
	if err != nil {
		return nil, nil, err
	}

	pullPages, err := r.fetchPullsPageCount(ctx)
	if err != nil {
		return nil, nil, err
	}

	g1, ctx1 := errgroup.WithContext(ctx)
	gitHubIssues := make(map[int]issue)
	gitHubPulls := make(map[int]pullRequest)

	// Fetch closed issues
	for i := 1; i <= issuePages; i++ {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		g1.Go(func() error {
			issues, err := r.fetchIssues(ctx1, since, i)
			if err != nil {
				return err
			}
			for _, issue := range issues {
				gitHubIssues[issue.Number] = issue
			}
			return nil
		})
	}

	// Fetch closed pull requests
	for i := 1; i <= pullPages; i++ {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		g1.Go(func() error {
			pulls, err := r.fetchPulls(ctx1, i)
			if err != nil {
				return err
			}
			for _, pull := range pulls {
				gitHubPulls[pull.Number] = pull
			}
			return nil
		})
	}

	if err := g1.Wait(); err != nil {
		return nil, nil, err
	}

	issues, merges := partitionIssuesAndMerges(gitHubIssues, gitHubPulls)
	r.logger.Debugf("Partitioned and sorted GitHub issues and pull requests: %d, %d", len(issues), len(merges))

	r.logger.Info("Fetching GitHub events for issues and pull requests ...")

	g2, ctx2 := errgroup.WithContext(ctx)

	// Fetch and search events for issues
	for _, issue := range issues {
		no := issue.Number // https://golang.org/doc/faq#closures_and_goroutines
		g2.Go(func() error {
			// TODO:
			_, err := r.findEvent(ctx2, no, "closed")
			return err
		})
	}

	// Fetch and search events for pull requests
	for _, merge := range merges {
		no := merge.Number // https://golang.org/doc/faq#closures_and_goroutines
		g2.Go(func() error {
			// TODO:
			_, err := r.findEvent(ctx2, no, "merged")
			return err
		})
	}

	if err := g2.Wait(); err != nil {
		return nil, nil, err
	}

	r.logger.Infof("All GitHub issues and pull requests are fetched: %d, %d", len(issues), len(merges))

	return issues, merges, nil
}
