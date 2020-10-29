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

// FetchAllTags retrieves all tags for a GitLab repository.
func (r *repo) FetchAllTags(ctx context.Context) (remote.Tags, error) {
	return remote.Tags{}, nil
}

// FetchIssuesAndMerges retrieves all closed issues and merged merge requests for a GitLab repository.
func (r *repo) FetchIssuesAndMerges(ctx context.Context, since time.Time) (remote.Changes, remote.Changes, error) {
	return remote.Changes{}, remote.Changes{}, nil
}
