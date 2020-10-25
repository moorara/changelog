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

func TestRepo_FetchClosedIssuesAndMerges(t *testing.T) {
	r := &repo{
		logger: log.New(log.None),
	}

	issues, merges, err := r.FetchClosedIssuesAndMerges(context.Background(), time.Now())

	assert.NoError(t, err)
	assert.NotNil(t, issues)
	assert.NotNil(t, merges)
}
