package git

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"

	"github.com/moorara/changelog/pkg/log"
)

func TestNewRepo(t *testing.T) {
	tests := []struct {
		name   string
		logger log.Logger
		path   string
	}{
		{
			name:   "OK",
			logger: log.New(log.None),
			path:   ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewRepo(tc.logger, tc.path)

			rp, ok := r.(*repo)
			assert.True(t, ok)

			assert.NoError(t, err)
			assert.NotNil(t, rp)
			assert.NotNil(t, rp.git)
		})
	}
}

func TestRepo_GetRemoteInfo(t *testing.T) {
	g, err := git.PlainOpen("../..")
	assert.NoError(t, err)

	tests := []struct {
		name           string
		expectedDomain string
		expectedPath   string
		expectedError  string
	}{
		{
			name:           "OK",
			expectedDomain: "github.com",
			expectedPath:   "moorara/changelog",
			expectedError:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				git:    g,
			}

			domain, path, err := r.GetRemoteInfo()

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

func TestRepo_Tags(t *testing.T) {
	g, err := git.PlainOpen("../..")
	assert.NoError(t, err)

	tests := []struct {
		name          string
		expectedError string
	}{
		{
			name:          "OK",
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &repo{
				logger: log.New(log.None),
				git:    g,
			}

			tags, err := r.Tags()

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.NotNil(t, tags)
			} else {
				assert.Nil(t, tags)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
