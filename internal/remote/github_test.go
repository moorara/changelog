package remote

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGitHubRepo(t *testing.T) {
	tests := []struct {
		name        string
		logger      *log.Logger
		path        string
		accessToken string
	}{
		{
			name:        "OK",
			logger:      &log.Logger{},
			path:        "moorara/changelog",
			accessToken: "github-access-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewGitHubRepo(tc.logger, tc.path, tc.accessToken)
			assert.NotNil(t, repo)

			r, ok := repo.(*githubRepo)
			assert.True(t, ok)

			assert.Equal(t, tc.logger, r.logger)
			assert.Equal(t, tc.path, r.path)
			assert.Equal(t, tc.accessToken, r.accessToken)
		})
	}
}
