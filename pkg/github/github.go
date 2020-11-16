package github

import (
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	relFirstRE = regexp.MustCompile(`<[\w\.:?&=/-]+page=(\d+)[\w\.:?&=/-]*>; rel="first"`)
	relPrevRE  = regexp.MustCompile(`<[\w\.:?&=/-]+page=(\d+)[\w\.:?&=/-]*>; rel="prev"`)
	relNextRE  = regexp.MustCompile(`<[\w\.:?&=/-]+page=(\d+)[\w\.:?&=/-]*>; rel="next"`)
	relLastRE  = regexp.MustCompile(`<[\w\.:?&=/-]+page=(\d+)[\w\.:?&=/-]*>; rel="last"`)
)

const (
	headerLink          = "Link"
	headerRateLimit     = "X-RateLimit-Limit"
	headerRateUsed      = "X-RateLimit-Used"
	headerRateRemaining = "X-RateLimit-Remaining"
	headerRateReset     = "X-RateLimit-Reset"
)

// Scope represents a GitHub authorization scope.
// See https://docs.github.com/developers/apps/scopes-for-oauth-apps
type Scope string

const (
	// ScopeRepo grants full access to private and public repositories. It also grants ability to manage user projects.
	ScopeRepo Scope = "repo"
	// ScopeRepoStatus grants read/write access to public and private repository commit statuses.
	ScopeRepoStatus Scope = "repo:status"
	// ScopeRepoDeployment grants access to deployment statuses for public and private repositories.
	ScopeRepoDeployment Scope = "repo_deployment"
	// ScopePublicRepo grants access only to public repositories.
	ScopePublicRepo Scope = "public_repo"
	// ScopeRepoInvite grants accept/decline abilities for invitations to collaborate on a repository.
	ScopeRepoInvite Scope = "repo:invite"
	// ScopeSecurityEvents grants read and write access to security events in the code scanning API.
	ScopeSecurityEvents Scope = "security_events"

	// ScopeWritePackages grants access to upload or publish a package in GitHub Packages.
	ScopeWritePackages Scope = "write:packages"
	// ScopeReadPackages grants access to download or install packages from GitHub Packages.
	ScopeReadPackages Scope = "read:packages"
	// ScopeDeletePackages grants access to delete packages from GitHub Packages.
	ScopeDeletePackages Scope = "delete:packages"

	// ScopeAdminOrg grants access to fully manage the organization and its teams, projects, and memberships.
	ScopeAdminOrg Scope = "admin:org"
	// ScopeWriteOrg grants read and write access to organization membership, organization projects, and team membership.
	ScopeWriteOrg Scope = "write:org"
	// ScopeReadOrg grants read-only access to organization membership, organization projects, and team membership.
	ScopeReadOrg Scope = "read:org"

	// ScopeAdminPublicKey grants access to fully manage public keys.
	ScopeAdminPublicKey Scope = "admin:public_key"
	// ScopeWritePublicKey grants access to create, list, and view details for public keys.
	ScopeWritePublicKey Scope = "write:public_key"
	// ScopeReadPublicKey grants access to list and view details for public keys.
	ScopeReadPublicKey Scope = "read:public_key"

	// ScopeAdminRepoHook grants read, write, ping, and delete access to repository hooks in public and private repositories.
	ScopeAdminRepoHook Scope = "admin:repo_hook"
	// ScopeWriteRepoHook grants read, write, and ping access to hooks in public or private repositories.
	ScopeWriteRepoHook Scope = "write:repo_hook"
	// ScopeReadRepoHook grants read and ping access to hooks in public or private repositories.
	ScopeReadRepoHook Scope = "read:repo_hook"

	// ScopeAdminOrgHook grants read, write, ping, and delete access to organization hooks.
	ScopeAdminOrgHook Scope = "admin:org_hook"
	// ScopeGist grants write access to gists.
	ScopeGist Scope = "gist"
	// ScopeNotifications grants read access to a user's notifications and misc.
	ScopeNotifications Scope = "notifications"

	// ScopeUser grants read/write access to profile info only.
	ScopeUser Scope = "user"
	// ScopeReadUser grants access to read a user's profile data.
	ScopeReadUser Scope = "read:user"
	// ScopeUserEmail grants read access to a user's email addresses.
	ScopeUserEmail Scope = "user:email"
	// ScopeUserFollow grants access to follow or unfollow other users.
	ScopeUserFollow Scope = "user:follow"

	// ScopeDeleteRepo grants access to delete adminable repositories.
	ScopeDeleteRepo Scope = "delete_repo"

	// ScopeWriteDiscussion allows read and write access for team discussions.
	ScopeWriteDiscussion Scope = "write:discussion"
	// ScopeReadDiscussion allows read access for team discussions.
	ScopeReadDiscussion Scope = "read:discussion"

	// ScopeAdminGPGKey grants access to fully manage GPG keys.
	ScopeAdminGPGKey Scope = "admin:gpg_key"
	// ScopeWriteGPGKey grants access to create, list, and view details for GPG keys.
	ScopeWriteGPGKey Scope = "write:gpg_key"
	// ScopeReadGPGKey grants access to list and view details for GPG keys.
	ScopeReadGPGKey Scope = "read:gpg_key"

	// ScopeWorkflow grants the ability to add and update GitHub Actions workflow files.
	ScopeWorkflow Scope = "workflow"
)

// Pages represents the pagination information for GitHub API v3.
type Pages struct {
	First int
	Prev  int
	Next  int
	Last  int
}

// Epoch is a Unix timestamp.
type Epoch int64

// Time returns the Time representation of an epoch timestamp.
func (e Epoch) Time() time.Time {
	return time.Unix(int64(e), 0)
}

// String returns string representation of an epoch timestamp.
func (e Epoch) String() string {
	return e.Time().Format("15:04:05")
}

// Rate represents the rate limit status for the authenticated user.
type Rate struct {
	// The number of requests per hour.
	Limit int `json:"limit"`

	// The number of requests used in the current hour.
	Used int `json:"used,omitempty"`

	// The number of requests remaining in the current hour.
	Remaining int `json:"remaining"`

	// The time at which the current rate will reset.
	Reset Epoch `json:"reset"`
}

// Response represents an HTTP response for GitHub API v3.
type Response struct {
	*http.Response

	Pages Pages
	Rate  Rate
}

func newResponse(resp *http.Response) *Response {
	r := &Response{
		Response: resp,
	}

	h := resp.Header

	if link := h.Get(headerLink); link != "" {
		if m := relFirstRE.FindStringSubmatch(link); len(m) == 2 {
			r.Pages.First, _ = strconv.Atoi(m[1])
		}

		if m := relPrevRE.FindStringSubmatch(link); len(m) == 2 {
			r.Pages.Prev, _ = strconv.Atoi(m[1])
		}

		if m := relNextRE.FindStringSubmatch(link); len(m) == 2 {
			r.Pages.Next, _ = strconv.Atoi(m[1])
		}

		if m := relLastRE.FindStringSubmatch(link); len(m) == 2 {
			r.Pages.Last, _ = strconv.Atoi(m[1])
		}
	}

	if limit := h.Get(headerRateLimit); limit != "" {
		r.Rate.Limit, _ = strconv.Atoi(limit)
	}

	if used := h.Get(headerRateUsed); used != "" {
		r.Rate.Used, _ = strconv.Atoi(used)
	}

	if remaining := h.Get(headerRateRemaining); remaining != "" {
		r.Rate.Remaining, _ = strconv.Atoi(remaining)
	}

	if reset := h.Get(headerRateReset); reset != "" {
		i64, _ := strconv.ParseInt(reset, 10, 64)
		r.Rate.Reset = Epoch(i64)
	}

	return r
}

// rateGroup determines the rate limit group for GitHub API v3.
type rateGroup string

const (
	rateGroupCore    = rateGroup("core")
	rateGroupSearch  = rateGroup("search")
	rateGroupGraphQL = rateGroup("graphql")
)

func getRateGroup(u *url.URL) rateGroup {
	switch {
	case strings.HasPrefix(u.Path, "/search"):
		return rateGroupSearch
	case strings.HasPrefix(u.Path, "/graphql"):
		return rateGroupGraphQL
	default:
		return rateGroupCore
	}
}
