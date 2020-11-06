package gitlab

import (
	"context"
	"net/http"
	"time"

	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/pkg/log"
)

const (
	gitlabAPIURL = "https://gitlab.com/api/v4/"
)

// repo implements the remote.Repo interface for GitLab.
type repo struct {
	logger      log.Logger
	client      *http.Client
	apiURL      string
	path        string
	accessToken string
}

// NewRepo creates a new GitLab repository.
func NewRepo(logger log.Logger, path, accessToken string) remote.Repo {
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	return &repo{
		logger:      logger,
		client:      client,
		apiURL:      gitlabAPIURL,
		path:        path,
		accessToken: accessToken,
	}
}

// FutureTag returns a tag that does not exist yet for a GitLab repository.
func (r *repo) FutureTag(name string) remote.Tag {
	return remote.Tag{}
}

// FetchFirstCommit retrieves the firist/initial commit for a GitLab repository.
func (r *repo) FetchFirstCommit(ctx context.Context) (remote.Commit, error) {
	return remote.Commit{}, nil
}

// FetchBranch retrieves a branch by name for a GitLab repository.
func (r *repo) FetchBranch(ctx context.Context, name string) (remote.Branch, error) {
	return remote.Branch{}, nil
}

// FetchDefaultBranch retrieves the default branch for a GitLab repository.
func (r *repo) FetchDefaultBranch(ctx context.Context) (remote.Branch, error) {
	return remote.Branch{}, nil
}

// FetchTags retrieves all tags for a GitLab repository.
func (r *repo) FetchTags(ctx context.Context) (remote.Tags, error) {
	return remote.Tags{}, nil
}

// FetchIssuesAndMerges retrieves all closed issues and merged merge requests for a GitLab repository.
func (r *repo) FetchIssuesAndMerges(ctx context.Context, since time.Time) (remote.Issues, remote.Merges, error) {
	return remote.Issues{}, remote.Merges{}, nil
}

// FetchParentCommits retrieves all parent commits of a given commit hash for a GitLab repository.
func (r *repo) FetchParentCommits(ctx context.Context, hash string) (remote.Commits, error) {
	return remote.Commits{}, nil
}
