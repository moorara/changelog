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

func (r *repo) createRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
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

	url := fmt.Sprintf("%s/user", r.apiURL)
	req, err := r.createRequest(ctx, "HEAD", url, nil)
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

func (r *repo) fetchUser(ctx context.Context, username string) (user, error) {
	// See https://docs.github.com/en/rest/reference/users#get-a-user

	r.logger.Debugf("Fetching GitHub user %s ...", username)

	url := fmt.Sprintf("%s/users/%s", r.apiURL, username)
	req, err := r.createRequest(ctx, "GET", url, nil)
	if err != nil {
		return user{}, err
	}

	resp, err := r.makeRequest(req, 200)
	if err != nil {
		return user{}, err
	}

	u := user{}
	if err = json.NewDecoder(resp.Body).Decode(&u); err != nil {
		return user{}, err
	}

	r.logger.Debugf("Fetched GitHub user %s", username)

	return u, nil
}

func (r *repo) fetchCommit(ctx context.Context, ref string) (commit, error) {
	// See https://docs.github.com/en/rest/reference/repos#get-a-commit

	r.logger.Debugf("Fetching GitHub commit %s ...", ref)

	url := fmt.Sprintf("%s/repos/%s/commits/%s", r.apiURL, r.path, ref)
	req, err := r.createRequest(ctx, "GET", url, nil)
	if err != nil {
		return commit{}, err
	}

	resp, err := r.makeRequest(req, 200)
	if err != nil {
		return commit{}, err
	}

	c := commit{}
	if err = json.NewDecoder(resp.Body).Decode(&c); err != nil {
		return commit{}, err
	}

	r.logger.Debugf("Fetched GitHub commit %s", ref)

	return c, nil
}

func (r *repo) fetchPull(ctx context.Context, number int) (pull, error) {
	// See https://docs.github.com/en/rest/reference/pulls#get-a-pull-request

	r.logger.Debugf("Fetching GitHub pull %d ...", number)

	url := fmt.Sprintf("%s/repos/%s/pulls/%d", r.apiURL, r.path, number)
	req, err := r.createRequest(ctx, "GET", url, nil)
	if err != nil {
		return pull{}, err
	}

	resp, err := r.makeRequest(req, 200)
	if err != nil {
		return pull{}, err
	}

	p := pull{}
	if err = json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return pull{}, err
	}

	r.logger.Debugf("Fetched GitHub pull %d", number)

	return p, nil
}

func (r *repo) findEvent(ctx context.Context, number int, name string) (event, error) {
	// See https://docs.github.com/en/rest/reference/issues#list-issue-events

	r.logger.Debugf("Finding %s event for issue %d ...", name, number)

	url := fmt.Sprintf("%s/repos/%s/issues/%d/events", r.apiURL, r.path, number)
	for {
		req, err := r.createRequest(ctx, "GET", url, nil)
		if err != nil {
			return event{}, err
		}

		q := req.URL.Query()
		q.Add("per_page", strconv.Itoa(pageSize))
		req.URL.RawQuery = q.Encode()

		resp, err := r.makeRequest(req, 200)
		if err != nil {
			return event{}, err
		}

		events := []event{}
		if err = json.NewDecoder(resp.Body).Decode(&events); err != nil {
			return event{}, err
		}

		for _, e := range events {
			if e.Event == name {
				r.logger.Debugf("Found %s event for issue %d ...", name, number)
				return e, nil
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

	return event{}, fmt.Errorf("GitHub %s event for issue %d not found", name, number)
}

func (r *repo) fetchTagsPageCount(ctx context.Context) (int, error) {
	// This GitHub API is not officially documented.
	// For the closest documentation see https://docs.github.com/en/rest/reference/git#get-a-tag

	r.logger.Debug("Fetching the total number of pages for GitHub tags ...")

	url := fmt.Sprintf("%s/repos/%s/tags", r.apiURL, r.path)
	req, err := r.createRequest(ctx, "HEAD", url, nil)
	if err != nil {
		return -1, err
	}

	q := req.URL.Query()
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

	r.logger.Debugf("Fetched the total number of pages for GitHub tags: %d", count)

	return count, nil
}

func (r *repo) fetchTags(ctx context.Context, pageNo int) ([]tag, error) {
	// This GitHub API is not officially documented.
	// For the closest documentation see https://docs.github.com/en/rest/reference/git#get-a-tag

	r.logger.Debugf("Fetching GitHub tags page %d ...", pageNo)

	url := fmt.Sprintf("%s/repos/%s/tags", r.apiURL, r.path)
	req, err := r.createRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("per_page", strconv.Itoa(pageSize))
	q.Add("page", strconv.Itoa(pageNo))
	req.URL.RawQuery = q.Encode()

	resp, err := r.makeRequest(req, 200)
	if err != nil {
		return nil, err
	}

	tags := []tag{}
	if err = json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return nil, err
	}

	r.logger.Debugf("Fetched GitHub tags page %d: %d", pageNo, len(tags))

	return tags, nil
}

func (r *repo) fetchIssuesPageCount(ctx context.Context, since time.Time) (int, error) {
	// See https://docs.github.com/en/rest/reference/issues#list-repository-issues

	r.logger.Debug("Fetching the total number of pages for GitHub issues ...")

	url := fmt.Sprintf("%s/repos/%s/issues", r.apiURL, r.path)
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

	r.logger.Debugf("Fetched the total number of pages for GitHub issues: %d", count)

	return count, nil
}

func (r *repo) fetchIssues(ctx context.Context, since time.Time, pageNo int) ([]issue, error) {
	// See https://docs.github.com/en/rest/reference/issues#list-repository-issues

	r.logger.Debugf("Fetching GitHub issues page %d ...", pageNo)

	url := fmt.Sprintf("%s/repos/%s/issues", r.apiURL, r.path)
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
	// See https://docs.github.com/en/rest/reference/pulls#list-pull-requests

	r.logger.Debug("Fetching the total number of pages for GitHub pull requests ...")

	url := fmt.Sprintf("%s/repos/%s/pulls", r.apiURL, r.path)
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

	r.logger.Debugf("Fetched the total number of pages for GitHub pull requests: %d", count)

	return count, nil
}

func (r *repo) fetchPulls(ctx context.Context, pageNo int) ([]pull, error) {
	// See https://docs.github.com/en/rest/reference/pulls#list-pull-requests

	r.logger.Debugf("Fetching GitHub pulls page %d ...", pageNo)

	url := fmt.Sprintf("%s/repos/%s/pulls", r.apiURL, r.path)
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

	pulls := []pull{}
	if err = json.NewDecoder(resp.Body).Decode(&pulls); err != nil {
		return nil, err
	}

	r.logger.Debugf("Fetched GitHub pulls page %d: %d", pageNo, len(pulls))

	return pulls, nil
}

// FetchAllTags retrieves all tags for a GitHub repository.
func (r *repo) FetchAllTags(ctx context.Context) (remote.Tags, error) {
	if err := r.checkScopes(ctx, scopeRepo); err != nil {
		return nil, err
	}

	// ==============================> FETCH TAGS <==============================

	r.logger.Debug("Fetching GitHub tags ...")

	g1, ctx1 := errgroup.WithContext(ctx)
	gitHubTags := newTagStore()

	tagPages, err := r.fetchTagsPageCount(ctx)
	if err != nil {
		return nil, err
	}

	// Fetch tags
	for i := 1; i <= tagPages; i++ {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		g1.Go(func() error {
			tags, err := r.fetchTags(ctx1, i)
			if err != nil {
				return err
			}
			for _, tag := range tags {
				gitHubTags.Save(tag.Name, tag)
			}
			return nil
		})
	}

	if err := g1.Wait(); err != nil {
		return nil, err
	}

	// ==============================> FETCH COMMITS <==============================

	r.logger.Debug("Fetching GitHub commits for tags ...")

	g2, ctx2 := errgroup.WithContext(ctx)
	gitHubCommits := newCommitStore()

	// Fetch closed pull requests
	_ = gitHubTags.ForEach(func(name string, t tag) error {
		g2.Go(func() error {
			commit, err := r.fetchCommit(ctx2, t.Commit.SHA)
			if err != nil {
				return err
			}
			gitHubCommits.Save(commit.SHA, commit)
			return nil
		})
		return nil
	})

	if err := g2.Wait(); err != nil {
		return nil, err
	}

	// ==============================> JOINING TAGS & COMMITS <==============================

	tags := resolveTags(gitHubTags, gitHubCommits)

	r.logger.Debugf("Resolved and sorted GitHub tags: %d", len(tags))
	r.logger.Info("GitHub tags are fetched")

	return tags, nil
}

// FetchIssuesAndMerges retrieves all closed issues and merged pull requests for a GitHub repository.
func (r *repo) FetchIssuesAndMerges(ctx context.Context, since time.Time) (remote.Changes, remote.Changes, error) {
	if err := r.checkScopes(ctx, scopeRepo); err != nil {
		return nil, nil, err
	}

	// ==============================> FETCH ISSUES <==============================

	if since.IsZero() {
		r.logger.Info("Fetching GitHub issues since the beginning ...")
	} else {
		r.logger.Infof("Fetching GitHub issues since %s ...", since.Format(time.RFC3339))
	}

	g1, ctx1 := errgroup.WithContext(ctx)
	gitHubIssues := newIssueStore()

	issuePages, err := r.fetchIssuesPageCount(ctx, since)
	if err != nil {
		return nil, nil, err
	}

	// Fetch closed issues
	for i := 1; i <= issuePages; i++ {
		i := i // https://golang.org/doc/faq#closures_and_goroutines
		g1.Go(func() error {
			issues, err := r.fetchIssues(ctx1, since, i)
			if err != nil {
				return err
			}
			for _, issue := range issues {
				gitHubIssues.Save(issue.Number, issue)
			}
			return nil
		})
	}

	if err := g1.Wait(); err != nil {
		return nil, nil, err
	}

	// ==============================> FETCH PULL REQUESTS <==============================

	r.logger.Info("Fetching GitHub pull requests for issues ...")

	g2, ctx2 := errgroup.WithContext(ctx)
	gitHubPulls := newPullStore()

	// Fetch closed pull requests
	_ = gitHubIssues.ForEach(func(num int, i issue) error {
		if i.PullRequest != nil {
			g2.Go(func() error {
				pull, err := r.fetchPull(ctx2, num)
				if err != nil {
					return err
				}
				if pull.Merged {
					gitHubPulls.Save(num, pull)
				}
				return nil
			})
		}
		return nil
	})

	if err := g2.Wait(); err != nil {
		return nil, nil, err
	}

	// ==============================> FETCH EVENTS & COMMITS <==============================

	r.logger.Info("Fetching GitHub events and commits for issues and pull requests ...")

	g3, ctx3 := errgroup.WithContext(ctx)
	gitHubEvents := newEventStore()
	gitHubCommits := newCommitStore()

	// Fetch and search events
	_ = gitHubIssues.ForEach(func(num int, i issue) error {

		if i.PullRequest == nil {
			g3.Go(func() error {
				e, err := r.findEvent(ctx3, num, "closed")
				if err != nil {
					return err
				}
				gitHubEvents.Save(num, e)
				return nil
			})
		}

		// Fetch the merged event only if the issue represents a merged pull request
		if _, ok := gitHubPulls.Load(num); ok {
			// All Pulls expected to be merged at this point
			g3.Go(func() error {
				e, err := r.findEvent(ctx3, num, "merged")
				if err != nil {
					return err
				}
				gitHubEvents.Save(num, e)

				c, err := r.fetchCommit(ctx3, e.CommitID)
				if err != nil {
					return err
				}
				gitHubCommits.Save(e.CommitID, c)

				return nil
			})
		}

		return nil
	})

	if err := g3.Wait(); err != nil {
		return nil, nil, err
	}

	// ==============================> FETCH USERS <==============================

	r.logger.Info("Fetching GitHub users for issues and pull requests ...")

	gitHubUsers := newUserStore()

	// Fetch author users
	err = gitHubIssues.ForEach(func(num int, i issue) error {
		// Fetch author users for issues
		if i.PullRequest == nil {
			// Fetch the user if it is not already fetched
			if _, ok := gitHubUsers.Load(i.User.Login); !ok {
				u, err := r.fetchUser(ctx, i.User.Login)
				if err != nil {
					return err
				}
				gitHubUsers.Save(i.User.Login, u)
			}
		}

		// Fetch author users for pull requests
		if p, ok := gitHubPulls.Load(num); ok {
			// All Pulls expected to be merged at this point
			// Fetch the user if it is not already fetched
			if _, ok := gitHubUsers.Load(p.User.Login); !ok {
				u, err := r.fetchUser(ctx, p.User.Login)
				if err != nil {
					return err
				}
				gitHubUsers.Save(p.User.Login, u)
			}
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// Fetch closer/merger users
	err = gitHubEvents.ForEach(func(num int, e event) error {
		// Fetch the user if it is not already fetched
		if _, ok := gitHubUsers.Load(e.Actor.Login); !ok {
			u, err := r.fetchUser(ctx, e.Actor.Login)
			if err != nil {
				return err
			}
			gitHubUsers.Save(e.Actor.Login, u)
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	// ==============================> JOINING ISSUES, PULLS, EVENTS, COMMITS, & USERS <==============================

	issues, merges := resolveIssuesAndMerges(gitHubIssues, gitHubPulls, gitHubEvents, gitHubCommits, gitHubUsers)

	r.logger.Debugf("Resolved and sorted GitHub issues and pull requests: %d, %d", len(issues), len(merges))
	r.logger.Infof("All GitHub issues and pull requests are fetched: %d, %d", len(issues), len(merges))

	return issues, merges, nil
}
