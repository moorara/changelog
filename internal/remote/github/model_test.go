package github

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moorara/changelog/internal/remote"
)

func TestToTag(t *testing.T) {
	tests := []struct {
		name        string
		t           tag
		c           commit
		expectedTag remote.Tag
	}{
		{
			name:        "Tag",
			t:           gitHubTag1,
			c:           gitHubCommit1,
			expectedTag: remoteTag,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag := toTag(tc.t, tc.c)

			assert.Equal(t, tc.expectedTag, tag)
		})
	}
}

func TestToIssueChange(t *testing.T) {
	tests := []struct {
		name            string
		i               issue
		e               event
		creator, closer user
		expectedChange  remote.Change
	}{
		{
			name:           "Issue",
			i:              gitHubIssue1,
			e:              gitHubEvent1,
			creator:        gitHubUser1,
			closer:         gitHubUser1,
			expectedChange: remoteIssue,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			change := toIssueChange(tc.i, tc.e, tc.creator, tc.closer)

			assert.Equal(t, tc.expectedChange, change)
		})
	}
}

func TestToPullChange(t *testing.T) {
	tests := []struct {
		name            string
		p               pull
		e               event
		c               commit
		creator, merger user
		expectedChange  remote.Change
	}{
		{
			name:           "Merge",
			p:              gitHubPull1,
			e:              gitHubEvent2,
			c:              gitHubCommit2,
			creator:        gitHubUser2,
			merger:         gitHubUser3,
			expectedChange: remoteMerge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			change := toPullChange(tc.p, tc.e, tc.c, tc.creator, tc.merger)

			assert.Equal(t, tc.expectedChange, change)
		})
	}
}

func TestResolveTags(t *testing.T) {
	tests := []struct {
		name          string
		gitHubTags    *tagStore
		gitHubCommits *commitStore
		expectedTags  remote.Tags
	}{
		{
			name: "OK",
			gitHubTags: &tagStore{
				m: map[string]tag{
					"v0.1.0": gitHubTag1,
				},
			},
			gitHubCommits: &commitStore{
				m: map[string]commit{
					"c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c": gitHubCommit1,
				},
			},
			expectedTags: remote.Tags{remoteTag},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := resolveTags(tc.gitHubTags, tc.gitHubCommits)

			assert.Equal(t, tc.expectedTags, tags)
		})
	}
}

func TestResolveIssuesAndMerges(t *testing.T) {
	tests := []struct {
		name           string
		gitHubIssues   *issueStore
		gitHubPulls    *pullStore
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
			gitHubPulls: &pullStore{
				m: map[int]pull{
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
					"6dcb09b5b57875f334f61aebed695e2e4193db5e": gitHubCommit2,
				},
			},
			gitHubUsers: &userStore{
				m: map[string]user{
					"octocat": gitHubUser1,
					"octodog": gitHubUser2,
					"octofox": gitHubUser3,
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
