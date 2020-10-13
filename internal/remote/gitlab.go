package remote

import (
	"log"
)

// gitlabRepo implements the Repo interface for GitLab.
type gitlabRepo struct {
	logger      *log.Logger
	path        string
	accessToken string
}

// NewGitLabRepo creates a new GitLab repository.
func NewGitLabRepo(logger *log.Logger, path, accessToken string) Repo {
	return &gitlabRepo{
		logger:      logger,
		path:        path,
		accessToken: accessToken,
	}
}
