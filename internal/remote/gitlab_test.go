package remote

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGitLabRepo(t *testing.T) {
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
			accessToken: "gitlab-access-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := NewGitLabRepo(tc.logger, tc.path, tc.accessToken)
			assert.NotNil(t, repo)

			r, ok := repo.(*gitlabRepo)
			assert.True(t, ok)

			assert.Equal(t, tc.logger, r.logger)
			assert.NotNil(t, r.client)
			assert.Equal(t, gitlabAPIURL, r.apiURL)
			assert.Equal(t, tc.path, r.path)
			assert.Equal(t, tc.accessToken, r.accessToken)
		})
	}
}
