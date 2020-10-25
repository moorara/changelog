package github

import (
	"testing"

	"github.com/moorara/changelog/internal/remote"
	"github.com/stretchr/testify/assert"
)

func TestToChange(t *testing.T) {
	tests := []struct {
		name           string
		gitHubIssue    issue
		expectedChange remote.Change
	}{
		{
			name:           "Issue",
			gitHubIssue:    gitHubIssue1,
			expectedChange: remoteIssue,
		},
		{
			name:           "Merge",
			gitHubIssue:    gitHubIssue2,
			expectedChange: remoteMerge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			change := toChange(tc.gitHubIssue)

			assert.Equal(t, tc.expectedChange, change)
		})
	}
}

func TestPartitionIssuesAndMerges(t *testing.T) {
	tests := []struct {
		name           string
		gitHubIssues   []issue
		expectedIssues []remote.Change
		expectedMerges []remote.Change
	}{
		{
			name:           "OK",
			gitHubIssues:   []issue{gitHubIssue1, gitHubIssue2},
			expectedIssues: []remote.Change{remoteIssue},
			expectedMerges: []remote.Change{remoteMerge},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issues, merges := partitionIssuesAndMerges(tc.gitHubIssues)

			assert.Equal(t, tc.expectedIssues, issues)
			assert.Equal(t, tc.expectedMerges, merges)
		})
	}
}
