package github

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/go-github"
)

func TestToUser(t *testing.T) {
	tests := []struct {
		name         string
		u            github.User
		expectedUser remote.User
	}{
		{
			name:         "OK",
			u:            gitHubUser1,
			expectedUser: remoteUser1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user := toUser(tc.u)
			assert.Equal(t, tc.expectedUser, user)
		})
	}
}

func TestToCommit(t *testing.T) {
	tests := []struct {
		name           string
		c              github.Commit
		expectedCommit remote.Commit
	}{
		{
			name:           "OK",
			c:              gitHubCommit1,
			expectedCommit: remoteCommit1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			commit := toCommit(tc.c)
			assert.Equal(t, tc.expectedCommit, commit)
		})
	}
}

func TestToBranch(t *testing.T) {
	tests := []struct {
		name           string
		b              github.Branch
		expectedBranch remote.Branch
	}{
		{
			name:           "OK",
			b:              gitHubBranch,
			expectedBranch: remoteBranch,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			branch := toBranch(tc.b)
			assert.Equal(t, tc.expectedBranch, branch)
		})
	}
}

func TestToTag(t *testing.T) {
	tests := []struct {
		name        string
		t           github.Tag
		c           github.Commit
		owner, repo string
		expectedTag remote.Tag
	}{
		{
			name:        "OK",
			t:           gitHubTag,
			c:           gitHubCommit2,
			owner:       "octocat",
			repo:        "Hello-World",
			expectedTag: remoteTag,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag := toTag(tc.t, tc.c, tc.owner, tc.repo)
			assert.Equal(t, tc.expectedTag, tag)
		})
	}
}

func TestToIssue(t *testing.T) {
	tests := []struct {
		name           string
		i              github.Issue
		e              github.Event
		author, closer github.User
		expectedIssue  remote.Issue
	}{
		{
			name:          "OK",
			i:             gitHubIssue1,
			e:             gitHubEvent1,
			author:        gitHubUser1,
			closer:        gitHubUser1,
			expectedIssue: remoteIssue,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issue := toIssue(tc.i, tc.e, tc.author, tc.closer)
			assert.Equal(t, tc.expectedIssue, issue)
		})
	}
}

func TestToMerge(t *testing.T) {
	tests := []struct {
		name           string
		i              github.Issue
		e              github.Event
		c              github.Commit
		author, merger github.User
		expectedMerge  remote.Merge
	}{
		{
			name:          "OK",
			i:             gitHubIssue2,
			e:             gitHubEvent2,
			c:             gitHubCommit1,
			author:        gitHubUser2,
			merger:        gitHubUser3,
			expectedMerge: remoteMerge,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			merge := toMerge(tc.i, tc.e, tc.c, tc.author, tc.merger)
			assert.Equal(t, tc.expectedMerge, merge)
		})
	}
}

func TestResolveTags(t *testing.T) {
	tests := []struct {
		name          string
		gitHubTags    *store
		gitHubCommits *store
		owner, repo   string
		expectedTags  remote.Tags
	}{
		{
			name: "OK",
			gitHubTags: &store{
				m: map[interface{}]interface{}{
					"v0.1.0": gitHubTag,
				},
			},
			gitHubCommits: &store{
				m: map[interface{}]interface{}{
					"c3d0be41ecbe669545ee3e94d31ed9a4bc91ee3c": gitHubCommit2,
				},
			},
			owner:        "octocat",
			repo:         "Hello-World",
			expectedTags: remote.Tags{remoteTag},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := resolveTags(tc.gitHubTags, tc.gitHubCommits, tc.owner, tc.repo)
			assert.Equal(t, tc.expectedTags, tags)
		})
	}
}

func TestResolveIssuesAndMerges(t *testing.T) {
	tests := []struct {
		name           string
		gitHubIssues   *store
		gitHubEvents   *store
		gitHubCommits  *store
		gitHubUsers    *store
		expectedIssues remote.Issues
		expectedMerges remote.Merges
	}{
		{
			name: "OK",
			gitHubIssues: &store{
				m: map[interface{}]interface{}{
					1001: gitHubIssue1,
					1002: gitHubIssue2,
				},
			},
			gitHubEvents: &store{
				m: map[interface{}]interface{}{
					1001: gitHubEvent1,
					1002: gitHubEvent2,
				},
			},
			gitHubCommits: &store{
				m: map[interface{}]interface{}{
					"6dcb09b5b57875f334f61aebed695e2e4193db5e": gitHubCommit1,
				},
			},
			gitHubUsers: &store{
				m: map[interface{}]interface{}{
					"octocat": gitHubUser1,
					"octodog": gitHubUser2,
					"octofox": gitHubUser3,
				},
			},
			expectedIssues: remote.Issues{remoteIssue},
			expectedMerges: remote.Merges{remoteMerge},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issues, merges := resolveIssuesAndMerges(tc.gitHubIssues, tc.gitHubEvents, tc.gitHubCommits, tc.gitHubUsers)

			assert.Equal(t, tc.expectedIssues, issues)
			assert.Equal(t, tc.expectedMerges, merges)
		})
	}
}
