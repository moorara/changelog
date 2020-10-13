package remote

import (
	"log"
)

// githubRepo implements the Repo interface for GitHub.
type githubRepo struct {
	logger      *log.Logger
	path        string
	accessToken string
}

// NewGitHubRepo creates a new GitHub repository.
func NewGitHubRepo(logger *log.Logger, path, accessToken string) Repo {
	return &githubRepo{
		logger:      logger,
		path:        path,
		accessToken: accessToken,
	}
}
