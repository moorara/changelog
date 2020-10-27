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
		e              event
		u              user
		expectedChange remote.Change
	}{
		{
			name:           "Issue",
			i:              gitHubIssue1,
			e:              gitHubEvent1,
			u:              gitHubUser,
			expectedChange: remoteIssue,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			change := issueToChange(tc.i, tc.e, tc.u)

			assert.Equal(t, tc.expectedChange, change)
		})
	}
}

func TestPullToChange(t *testing.T) {
	tests := []struct {
		name           string
		p              pullRequest
		e              event
		c              commit
		u              user
		expectedChange remote.Change
	}{
		{
			name:           "Merge",
			p:              gitHubPull1,
			e:              gitHubEvent2,
			c:              gitHubCommit,
			u:              gitHubUser,
			expectedChange: remoteMerge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			change := pullToChange(tc.p, tc.e, tc.c, tc.u)

			assert.Equal(t, tc.expectedChange, change)
		})
	}
}

func TestResolveIssuesAndMerges(t *testing.T) {
	tests := []struct {
		name           string
		gitHubIssues   *issueStore
		gitHubPulls    *pullRequestStore
		gitHubEvents   *eventStore
		gitHubCommits  *commitStore
		gitHubUsers    *userStore
		expectedIssues remote.Changes
		expectedMerges remote.Changes
	}{
		{
			name: "OK",
			gitHubIssues: &issueStore{
				m: map[int]issue{
					1001: gitHubIssue1,
					1002: gitHubIssue2,
				},
			},
			gitHubPulls: &pullRequestStore{
				m: map[int]pullRequest{
					1002: gitHubPull1,
				},
			},
			gitHubEvents: &eventStore{
				m: map[int]event{
					1001: gitHubEvent1,
					1002: gitHubEvent2,
				},
			},
			gitHubCommits: &commitStore{
				m: map[string]commit{
					"6dcb09b5b57875f334f61aebed695e2e4193db5e": gitHubCommit,
				},
			},
			gitHubUsers: &userStore{
				m: map[string]user{
					"octocat": gitHubUser,
				},
			},
			expectedIssues: remote.Changes{remoteIssue},
			expectedMerges: remote.Changes{remoteMerge},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issues, merges := resolveIssuesAndMerges(tc.gitHubIssues, tc.gitHubPulls, tc.gitHubEvents, tc.gitHubCommits, tc.gitHubUsers)

			assert.Equal(t, tc.expectedIssues, issues)
			assert.Equal(t, tc.expectedMerges, merges)
		})
	}
}
