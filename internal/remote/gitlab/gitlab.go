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

// FetchClosedIssuesAndMerges retrieves all issues and merge requests for a GitLab repository.
func (r *repo) FetchClosedIssuesAndMerges(ctx context.Context, since time.Time) ([]remote.Change, []remote.Change, error) {
	return []remote.Change{}, []remote.Change{}, nil
}
