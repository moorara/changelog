package remote

import (
	"log"
	"net/http"
)

const (
	gitlabAPIURL = "https://gitlab.com/api/v4/"
)

// gitlabRepo implements the Repo interface for GitLab.
type gitlabRepo struct {
	logger      *log.Logger
	client      *http.Client
	apiURL      string
	path        string
	accessToken string
}

// NewGitLabRepo creates a new GitLab repository.
func NewGitLabRepo(logger *log.Logger, path, accessToken string) Repo {
	transport := &http.Transport{}
	client := &http.Client{
		Transport: transport,
	}

	return &gitlabRepo{
		logger:      logger,
		client:      client,
		apiURL:      gitlabAPIURL,
		path:        path,
		accessToken: accessToken,
	}
}

// FetchClosedIssues
func (r *gitlabRepo) FetchClosedIssuesAndMerges() ([]Issue, []Merge, error) {
	return []Issue{}, []Merge{}, nil
}
