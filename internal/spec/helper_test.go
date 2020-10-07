package spec

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

func TestGetGitRemoteInfo(t *testing.T) {
	repo, err := git.PlainOpen("../..")
	assert.NoError(t, err)

	tests := []struct {
		name           string
		repo           *git.Repository
		expectedDomain string
		expectedPath   string
		expectedError  string
	}{
		{
			name:           "OK",
			repo:           repo,
			expectedDomain: "github.com",
			expectedPath:   "moorara/changelog",
			expectedError:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			domain, path, err := getGitRemoteInfo(tc.repo)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDomain, domain)
				assert.Equal(t, tc.expectedPath, path)
			} else {
				assert.Empty(t, domain)
				assert.Empty(t, path)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
