package gitlab

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/moorara/changelog/pkg/log"
)

func TestNewRepo(t *testing.T) {
	tests := []struct {
		name        string
		logger      log.Logger
		path        string
		accessToken string
	}{
		{
			name:        "OK",
			logger:      log.New(log.None),
			path:        "moorara/changelog",
			accessToken: "gitlab-access-token",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := NewRepo(tc.logger, tc.path, tc.accessToken)
			assert.NotNil(t, r)

			gr, ok := r.(*repo)
			assert.True(t, ok)

			assert.Equal(t, tc.logger, gr.logger)
			assert.NotNil(t, gr.client)
			assert.Equal(t, gitlabAPIURL, gr.apiURL)
			assert.Equal(t, tc.path, gr.path)
			assert.Equal(t, tc.accessToken, gr.accessToken)
		})
	}
}

func TestRepo_FutureTag(t *testing.T) {
	r := &repo{
		logger: log.New(log.None),
	}

	tag := r.FutureTag("v0.1.0")

	assert.Empty(t, tag)
}

func TestRepo_FetchBranch(t *testing.T) {
	r := &repo{
		logger: log.New(log.None),
	}

	branch, err := r.FetchBranch(context.Background(), "main")

	assert.NoError(t, err)
	assert.Empty(t, branch)
}

func TestRepo_FetchDefaultBranch(t *testing.T) {
	r := &repo{
		logger: log.New(log.None),
	}

	branch, err := r.FetchDefaultBranch(context.Background())

	assert.NoError(t, err)
	assert.Empty(t, branch)
}

func TestRepo_FetchTags(t *testing.T) {
	r := &repo{
		logger: log.New(log.None),
	}

	tags, err := r.FetchTags(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, tags)
}

func TestRepo_FetchIssuesAndMerges(t *testing.T) {
	r := &repo{
		logger: log.New(log.None),
	}

	issues, merges, err := r.FetchIssuesAndMerges(context.Background(), time.Now())

	assert.NoError(t, err)
	assert.NotNil(t, issues)
	assert.NotNil(t, merges)
}

func TestRepo_FetchParentCommits(t *testing.T) {
	r := &repo{
		logger: log.New(log.None),
	}

	commits, err := r.FetchParentCommits(context.Background(), "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378")

	assert.NoError(t, err)
	assert.NotNil(t, commits)
}
