package github

import (
	"testing"

	"github.com/moorara/changelog/internal/remote"
	"github.com/stretchr/testify/assert"
)

func TestIssueToChange(t *testing.T) {
	tests := []struct {
		name           string
		i              issue
		expectedChange remote.Change
	}{
		{
			name:           "Issue",
			i:              gitHubIssue1,
			expectedChange: remoteIssue,
		},
		{
			name:           "Merge",
			i:              gitHubIssue2,
			expectedChange: remoteMerge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			change := issueToChange(tc.i)

			assert.Equal(t, tc.expectedChange, change)
		})
	}
}

func TestPullToChange(t *testing.T) {
	tests := []struct {
		name           string
		p              pullRequest
		expectedChange remote.Change
	}{
		{
			name:           "Merge",
			p:              gitHubPull1,
			expectedChange: remoteMerge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			change := pullToChange(tc.p)

			assert.Equal(t, tc.expectedChange, change)
		})
	}
}

func TestPartitionIssuesAndMerges(t *testing.T) {
	tests := []struct {
		name           string
		gitHubIssues   map[int]issue
		gitHubPulls    map[int]pullRequest
		expectedIssues remote.Changes
		expectedMerges remote.Changes
	}{
		{
			name: "OK",
			gitHubIssues: map[int]issue{
				1001: gitHubIssue1,
				1002: gitHubIssue2,
			},
			gitHubPulls: map[int]pullRequest{
				1002: gitHubPull1,
			},
			expectedIssues: remote.Changes{remoteIssue},
			expectedMerges: remote.Changes{remoteMerge},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issues, merges := partitionIssuesAndMerges(tc.gitHubIssues, tc.gitHubPulls)

			assert.Equal(t, tc.expectedIssues, issues)
			assert.Equal(t, tc.expectedMerges, merges)
		})
	}
}
