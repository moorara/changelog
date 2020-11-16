package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	publicAPIURL, _    = url.Parse("https://api.github.com")
	publicUploadURL, _ = url.Parse("https://uploads.github.com")
)

const (
	cacheSize = 1024

	headerAuth        = "Authorization"
	headerUserAgent   = "User-Agent"
	headerContentType = "Content-Type"
	headerAccept      = "Accept"
	headerScopes      = "X-OAuth-Scopes"
	headerRetryAfter  = "Retry-After"
)

const (
	// See https://docs.github.com/rest/overview/resources-in-the-rest-api#user-agent-required
	userAgent = "moorara/github"

	// See https://docs.github.com/rest/overview/media-types
	mediaJSON        = "application/json"
	mediaTypeV3      = "application/vnd.github.v3+json"
	mediaTypeV3SHA   = "application/vnd.github.v3.sha"
	mediaTypeV3Diff  = "application/vnd.github.v3.diff"
	mediaTypeV3Patch = "application/vnd.github.v3.patch"
)

// Client is used for making API calls to GitHub API v3.
type Client struct {
	// httpMutex  sync.Mutex
	httpClient *http.Client
	ratesMutex sync.Mutex
	rates      map[rateGroup]Rate

	apiURL      *url.URL
	uploadURL   *url.URL
	accessToken string

	// Services
	Users *UsersService
}

func newHTTPClient() *http.Client {
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	return client
}

// NewClient creates a new client for calling public GitHub API v3.
func NewClient(accessToken string) *Client {
	c := &Client{
		httpClient:  newHTTPClient(),
		rates:       map[rateGroup]Rate{},
		apiURL:      publicAPIURL,
		uploadURL:   publicUploadURL,
		accessToken: accessToken,
	}

	c.Users = &UsersService{
		client: c,
	}

	return c
}

// NewEnterpriseClient creates a new client for calling an enterprise GitHub API v3.
func NewEnterpriseClient(apiURL, uploadURL, accessToken string) (*Client, error) {
	entAPIURL, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	entUploadURL, err := url.Parse(uploadURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		httpClient:  newHTTPClient(),
		rates:       map[rateGroup]Rate{},
		apiURL:      entAPIURL,
		uploadURL:   entUploadURL,
		accessToken: accessToken,
	}

	c.Users = &UsersService{
		client: c,
	}

	return c, nil
}

// NewRequest creates a new HTTP request for a GitHub API v3.
func (c *Client) NewRequest(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	u, err := c.apiURL.Parse(url)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set(headerUserAgent, userAgent)
	req.Header.Set(headerAccept, mediaTypeV3)

	if c.accessToken != "" {
		req.Header.Set(headerAuth, fmt.Sprintf("token %s", c.accessToken))
	}

	if body != nil {
		req.Header.Set(headerContentType, mediaJSON)
	}

	return req, nil
}

// NewPageRequest creates a new HTTP request for a GitHub API v3 with page parameters.
func (c *Client) NewPageRequest(ctx context.Context, method, url string, pageSize, pageNo int, body interface{}) (*http.Request, error) {
	req, err := c.NewRequest(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()

	if pageSize > 0 {
		q.Add("per_page", strconv.Itoa(pageSize))
	}

	if pageNo > 0 {
		q.Add("page", strconv.Itoa(pageNo))
	}

	req.URL.RawQuery = q.Encode()

	return req, nil
}

// NewUploadRequest creates a new HTTP request for uploading a file to GitHub API v3.
func (c *Client) NewUploadRequest(ctx context.Context, url, filepath string) (*http.Request, error) {
	u, err := c.uploadURL.Parse(url)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}

	// Read the first 512 bytes of file to determine the media type of the file
	buff := make([]byte, 512)
	if _, err := f.Read(buff); err != nil {
		return nil, err
	}

	// http.DetectContentType will return "application/octet-stream" if it cannot determine a more specific one
	mediaType := http.DetectContentType(buff)

	// Reset the offset back to the beginning of the file
	if _, err = f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", u.String(), f)
	if err != nil {
		return nil, err
	}

	req.ContentLength = stat.Size()
	req.Header.Set(headerUserAgent, userAgent)
	req.Header.Set(headerAccept, mediaTypeV3)
	req.Header.Set(headerContentType, mediaType)

	if c.accessToken != "" {
		req.Header.Set(headerAuth, fmt.Sprintf("token %s", c.accessToken))
	}

	return req, nil
}

// Do makes an HTTP request and returns the API response.
// If body implements the io.Writer interface, the raw response body will be copied to.
// Otherwise, the response body will be JOSN-decoded into it.
func (c *Client) Do(req *http.Request, body interface{}) (*Response, error) {
	// ====================> CHECK RATE LIMITS <====================

	g := getRateGroup(req.URL)

	c.ratesMutex.Lock()
	rate, ok := c.rates[g]
	c.ratesMutex.Unlock()

	if ok && rate.Remaining == 0 && time.Now().Before(rate.Reset.Time()) {
		return nil, &RateLimitError{
			Request: req,
			Rate:    rate,
		}
	}

	// ====================> MAKE THE REQUEST <====================

	r, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		// Ensure we fully read and close the response body, so the underlying TCP connection can be reused.
		// If it errors, the TCP connection will not be reused anyway.
		_, _ = io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}()

	resp := newResponse(r)

	// Update rate limits
	c.ratesMutex.Lock()
	c.rates[g] = resp.Rate
	c.ratesMutex.Unlock()

	// ====================> CHECK THE RESPONSE <====================

	isSuccess := func(statusCode int) bool {
		return statusCode == http.StatusOK ||
			statusCode == http.StatusCreated ||
			statusCode == http.StatusNoContent
	}

	if !isSuccess(r.StatusCode) {
		respErr := &ResponseError{
			Response: r,
		}

		b, err := ioutil.ReadAll(resp.Body)
		if err == nil && b != nil {
			_ = json.Unmarshal(b, respErr)
		}

		// Restore response body
		// r.Body = ioutil.NopCloser(bytes.NewBuffer(b))

		switch r.StatusCode {
		case http.StatusBadRequest:
			return nil, respErr

		case http.StatusUnauthorized:
			return nil, &AuthError{
				err: respErr,
			}

		case http.StatusForbidden:
			if r.Header.Get(headerRateRemaining) == "0" {
				return nil, &RateLimitError{
					err:     respErr,
					Request: req,
					Rate:    resp.Rate,
				}
			} else if strings.HasSuffix(respErr.DocumentationURL, "#abuse-rate-limits") {
				retryAfter, _ := time.ParseDuration(r.Header.Get(headerRetryAfter) + "s")
				return nil, &RateLimitAbuseError{
					err:        respErr,
					Rate:       resp.Rate,
					RetryAfter: retryAfter,
				}
			}

		case http.StatusNotFound:
			return nil, &NotFoundError{
				err: respErr,
			}

		default:
			return nil, respErr
		}
	}

	// ====================> READ THE BODY <====================

	if body != nil {
		if w, ok := body.(io.Writer); ok {
			if _, err := io.Copy(w, r.Body); err != nil {
				return nil, err
			}
		} else {
			if err := json.NewDecoder(r.Body).Decode(body); err != nil && err != io.EOF {
				return nil, err
			}
		}
	}

	return resp, nil
}

// EnsureScopes makes sure the client and the access token have the given scopes.
// See https://docs.github.com/developers/apps/scopes-for-oauth-apps
func (c *Client) EnsureScopes(ctx context.Context, scopes ...Scope) error {
	// Call an endpoint to get the OAuth scopes of the access token from the headers
	req, err := c.NewRequest(ctx, "HEAD", "/user", nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(req, nil)
	if err != nil {
		return err
	}

	// Ensure the access token has all the required OAuth scopes
	oauthScopes := resp.Header.Get(headerScopes)
	for _, scope := range scopes {
		if !strings.Contains(oauthScopes, string(scope)) {
			return fmt.Errorf("access token does not have the scope: %s", scope)
		}
	}

	return nil
}

// Repo returns a service providing GitHub APIs for a specific repository.
func (c *Client) Repo(owner, repo string) *RepoService {
	return &RepoService{
		client: c,
		owner:  owner,
		repo:   repo,
	}
}
