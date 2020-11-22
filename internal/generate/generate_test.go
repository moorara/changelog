package generate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/internal/spec"
	"github.com/moorara/changelog/pkg/log"
)

var (
	t1 = parseGitHubTime("2020-10-02T02:00:00-04:00")
	t2 = parseGitHubTime("2020-10-12T12:00:00-04:00")
	t3 = parseGitHubTime("2020-10-22T22:00:00-04:00")
	t4 = parseGitHubTime("2020-10-31T23:00:00-04:00")

	user1 = remote.User{
		Name:     "The Octocat",
		Email:    "octocat@github.com",
		Username: "octocat",
		WebURL:   "https://github.com/octocat",
	}

	user2 = remote.User{
		Name:     "The Octodog",
		Email:    "octodog@github.com",
		Username: "octodog",
		WebURL:   "https://github.com/octodog",
	}

	commit1 = remote.Commit{
		Hash: "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		Time: t1,
	}

	commit2 = remote.Commit{
		Hash: "0251a422d2038967eeaaaa5c8aa76c7067fdef05",
		Time: t2,
	}

	commit3 = remote.Commit{
		Hash: "c414d1004154c6c324bd78c69d10ee101e676059",
		Time: t3,
	}

	commit4 = remote.Commit{
		Hash: "20c5414eccaa147f2d6644de4ca36f35293fa43e",
		Time: t4,
	}

	branch = remote.Branch{
		Name:   "main",
		Commit: commit3,
	}

	tag1 = remote.Tag{
		Name:   "v0.1.1",
		Time:   t1,
		Commit: commit1,
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.1.1",
	}

	tag2 = remote.Tag{
		Name:   "v0.1.2",
		Time:   t2,
		Commit: commit2,
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.1.2",
	}

	tag3 = remote.Tag{
		Name:   "v0.1.3",
		Time:   t3,
		Commit: commit3,
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.1.3",
	}

	issue1 = remote.Issue{
		Change: remote.Change{
			Number:    1001,
			Title:     "Found a bug",
			Labels:    []string{"bug"},
			Milestone: "v1.0",
			Time:      t3,
			Author:    user1,
			WebURL:    "https://github.com/octocat/Hello-World/issues/1001",
		},
		Closer: user1,
	}

	issue2 = remote.Issue{
		Change: remote.Change{
			Number:    1002,
			Title:     "Discovered a vulnerability",
			Labels:    []string{"invalid"},
			Milestone: "",
			Time:      t4, // Unrleased change for future tag
			Author:    user1,
			WebURL:    "https://github.com/octocat/Hello-World/issues/1002",
		},
		Closer: user2,
	}

	merge1 = remote.Merge{
		Change: remote.Change{
			Number:    1003,
			Title:     "Added a feature",
			Labels:    []string{"enhancement"},
			Milestone: "v1.0",
			Time:      t3,
			Author:    user1,
			WebURL:    "https://github.com/octocat/Hello-World/pull/1003",
		},
		Merger: user1,
		Commit: commit3,
	}

	merge2 = remote.Merge{
		Change: remote.Change{
			Number:    1004,
			Title:     "Refactored code",
			Labels:    nil,
			Milestone: "",
			Time:      t4, // Unrleased change for future tag
			Author:    user1,
			WebURL:    "https://github.com/octocat/Hello-World/pull/1004",
		},
		Merger: user2,
		Commit: commit4,
	}

	changelogIssue1 = changelog.Issue{
		Number: 1001,
		Title:  "Found a bug",
		URL:    "https://github.com/octocat/Hello-World/issues/1001",
		OpenedBy: changelog.User{
			Name:     "The Octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
		ClosedBy: changelog.User{
			Name:     "The Octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
	}

	changelogIssue2 = changelog.Issue{
		Number: 1002,
		Title:  "Discovered a vulnerability",
		URL:    "https://github.com/octocat/Hello-World/issues/1002",
		OpenedBy: changelog.User{
			Name:     "The Octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
		ClosedBy: changelog.User{
			Name:     "The Octodog",
			Username: "octodog",
			URL:      "https://github.com/octodog",
		},
	}

	changelogMerge1 = changelog.Merge{
		Number: 1003,
		Title:  "Added a feature",
		URL:    "https://github.com/octocat/Hello-World/pull/1003",
		OpenedBy: changelog.User{
			Name:     "The Octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
		MergedBy: changelog.User{
			Name:     "The Octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
	}

	changelogMerge2 = changelog.Merge{
		Number: 1004,
		Title:  "Refactored code",
		URL:    "https://github.com/octocat/Hello-World/pull/1004",
		OpenedBy: changelog.User{
			Name:     "The Octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
		MergedBy: changelog.User{
			Name:     "The Octodog",
			Username: "octodog",
			URL:      "https://github.com/octodog",
		},
	}
)

func parseGitHubTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}

	return t
}

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		s             spec.Spec
		logger        log.Logger
		expectedError string
	}{
		{
			name: "InvalidSpec",
			s: spec.Spec{
				Repo: spec.Repo{
					Platform: spec.PlatformGitHub,
					Path:     "octocat/invalid/Hello-World",
				},
			},
			logger:        nil,
			expectedError: "unexpected GitHub repository: cannot parse owner and repo",
		},
		{
			name: "GitHub",
			s: spec.Spec{
				Repo: spec.Repo{
					Platform: spec.PlatformGitHub,
					Path:     "octocat/Hello-World",
				},
			},
			logger:        log.New(log.None),
			expectedError: "",
		},
		{
			name: "GitLab",
			s: spec.Spec{
				Repo: spec.Repo{
					Platform: spec.PlatformGitLab,
				},
			},
			logger:        log.New(log.None),
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g, err := New(tc.s, tc.logger)

			if tc.expectedError != "" {
				assert.Nil(t, g)
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NotNil(t, g)
				assert.Equal(t, tc.logger, g.logger)
				assert.NotNil(t, g.remoteRepo)
				assert.NotNil(t, g.processor)
			}
		})
	}
}

func TestGenerator_resolveTags(t *testing.T) {
	futureTag1 := remote.Tag{
		Name:   "v0.1.0",
		Time:   parseGitHubTime("2020-11-02T16:00:00-04:00"),
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.1.0",
	}

	futureTag2 := remote.Tag{
		Name:   "v0.2.0",
		Time:   parseGitHubTime("2020-11-03T18:00:00-04:00"),
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.2.0",
	}

	futureTag3 := remote.Tag{
		Name:   "v0.3.0",
		Time:   parseGitHubTime("2020-11-04T20:00:00-04:00"),
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.3.0",
	}

	futureTag4 := remote.Tag{
		Name:   "v0.4.0",
		Time:   parseGitHubTime("2020-11-05T22:00:00-04:00"),
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.4.0",
	}

	tests := []struct {
		name          string
		g             *Generator
		sortedTags    remote.Tags
		chlog         *changelog.Changelog
		expectedTags  remote.Tags
		expectedError error
	}{
		{
			name: "NoGitTag_NoChangelogTag_NoFutureTag",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			sortedTags:    remote.Tags{},
			chlog:         &changelog.Changelog{},
			expectedTags:  remote.Tags{},
			expectedError: nil,
		},
		{
			name: "NoGitTag_NoChangelogTag_FutureTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						Future: "v0.1.0",
					},
				},
				logger: log.New(log.None),
				remoteRepo: &MockRemoteRepo{
					FutureTagMocks: []FutureTagMock{
						{OutTag: futureTag1},
					},
				},
			},
			sortedTags:    remote.Tags{},
			chlog:         &changelog.Changelog{},
			expectedTags:  remote.Tags{futureTag1},
			expectedError: nil,
		},
		{
			name: "GitTag_NoChangelogTag_NoFutureTag",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			sortedTags:    remote.Tags{tag1},
			chlog:         &changelog.Changelog{},
			expectedTags:  remote.Tags{tag1},
			expectedError: nil,
		},
		{
			name: "GitTag_NoChangelogTag_FutureTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						Future: "v0.2.0",
					},
				},
				logger: log.New(log.None),
				remoteRepo: &MockRemoteRepo{
					FutureTagMocks: []FutureTagMock{
						{OutTag: futureTag2},
					},
				},
			},
			sortedTags:    remote.Tags{tag1},
			chlog:         &changelog.Changelog{},
			expectedTags:  remote.Tags{futureTag2, tag1},
			expectedError: nil,
		},
		{
			name: "SameGitTag_SameChangelogTag_NoFutureTag",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			sortedTags: remote.Tags{tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  remote.Tags{},
			expectedError: nil,
		},
		{
			name: "SameGitTag_SameChangelogTag_FutureTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						Future: "v0.2.0",
					},
				},
				logger: log.New(log.None),
				remoteRepo: &MockRemoteRepo{
					FutureTagMocks: []FutureTagMock{
						{OutTag: futureTag2},
					},
				},
			},
			sortedTags: remote.Tags{tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  remote.Tags{futureTag2},
			expectedError: nil,
		},
		{
			name: "NewGitTag_ChangelogTag_NoFutureTag",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			sortedTags: remote.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  remote.Tags{tag2},
			expectedError: nil,
		},
		{
			name: "NewGitTag_ChangelogTag_FutureTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						Future: "v0.3.0",
					},
				},
				logger: log.New(log.None),
				remoteRepo: &MockRemoteRepo{
					FutureTagMocks: []FutureTagMock{
						{OutTag: futureTag3},
					},
				},
			},
			sortedTags: remote.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  remote.Tags{futureTag3, tag2},
			expectedError: nil,
		},
		{
			name: "InvalidFromTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From: "invalid",
					},
				},
				logger: log.New(log.None),
			},
			sortedTags: remote.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  nil,
			expectedError: errors.New("from-tag can be one of [v0.1.2]"),
		},
		{
			name: "InvalidToTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						To: "invalid",
					},
				},
				logger: log.New(log.None),
			},
			sortedTags: remote.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  nil,
			expectedError: errors.New("to-tag can be one of [v0.1.2]"),
		},
		{
			name: "FromTag_Before_NewTags",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From: "v0.1.1",
					},
				},
				logger: log.New(log.None),
			},
			sortedTags: remote.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  nil,
			expectedError: errors.New("from-tag can be one of [v0.1.2]"),
		},
		{
			name: "FromTag_After_ToTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From: "v0.1.3",
						To:   "v0.1.2",
					},
				},
				logger: log.New(log.None),
			},
			sortedTags: remote.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  nil,
			expectedError: errors.New("to-tag can be one of [v0.1.3]"),
		},
		{
			name: "FromTag_Equal_ToTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From: "v0.1.2",
						To:   "v0.1.2",
					},
				},
				logger: log.New(log.None),
			},
			sortedTags: remote.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  remote.Tags{tag2},
			expectedError: nil,
		},
		{
			name: "FromTag_Before_ToTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From: "v0.1.2",
						To:   "v0.1.3",
					},
				},
				logger: log.New(log.None),
			},
			sortedTags: remote.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  remote.Tags{tag3, tag2},
			expectedError: nil,
		},
		{
			name: "InvalidFutureTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From:   "v0.1.2",
						To:     "v0.1.3",
						Future: "v0.1.1",
					},
				},
				logger: log.New(log.None),
				remoteRepo: &MockRemoteRepo{
					FutureTagMocks: []FutureTagMock{
						{OutTag: futureTag4},
					},
				},
			},
			sortedTags: remote.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  nil,
			expectedError: errors.New("future tag cannot be same as an existing tag: v0.1.1"),
		},
		{
			name: "FromTag_ToTag_FutureTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From:   "v0.1.2",
						To:     "v0.1.3",
						Future: "v0.1.4",
					},
				},
				logger: log.New(log.None),
				remoteRepo: &MockRemoteRepo{
					FutureTagMocks: []FutureTagMock{
						{OutTag: futureTag4},
					},
				},
			},
			sortedTags: remote.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Existing: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			expectedTags:  remote.Tags{futureTag4, tag3, tag2},
			expectedError: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags, err := tc.g.resolveTags(tc.sortedTags, tc.chlog)

			assert.Equal(t, tc.expectedTags, tags)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestGenerator_resolveCommitMap(t *testing.T) {
	tests := []struct {
		name              string
		g                 *Generator
		ctx               context.Context
		branch            remote.Branch
		sortedTags        remote.Tags
		expectedError     string
		expectedCommitMap commitMap
	}{
		{
			name: "FetchParentCommitsFails_Branch",
			g: &Generator{
				logger: log.New(log.None),
				remoteRepo: &MockRemoteRepo{
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutError: errors.New("error on fetching parent commits for branch")},
					},
				},
			},
			ctx:           context.Background(),
			branch:        branch,
			sortedTags:    remote.Tags{tag2, tag1},
			expectedError: "error on fetching parent commits for branch",
		},
		{
			name: "FetchParentCommitsFails_FirstTag",
			g: &Generator{
				logger: log.New(log.None),
				remoteRepo: &MockRemoteRepo{
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutError: errors.New("error on fetching parent commits for tag")},
					},
				},
			},
			ctx:           context.Background(),
			branch:        branch,
			sortedTags:    remote.Tags{tag2, tag1},
			expectedError: "error on fetching parent commits for tag",
		},
		{
			name: "FetchParentCommitsFails_SecondTag",
			g: &Generator{
				logger: log.New(log.None),
				remoteRepo: &MockRemoteRepo{
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
						{OutError: errors.New("error on fetching parent commits for tag")},
					},
				},
			},
			ctx:           context.Background(),
			branch:        branch,
			sortedTags:    remote.Tags{tag2, tag1},
			expectedError: "error on fetching parent commits for tag",
		},
		{
			name: "Success",
			g: &Generator{
				logger: log.New(log.None),
				remoteRepo: &MockRemoteRepo{
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
						{OutCommits: remote.Commits{commit1}},
					},
				},
			},
			ctx:        context.Background(),
			branch:     branch,
			sortedTags: remote.Tags{tag2, tag1},
			expectedCommitMap: commitMap{
				"c414d1004154c6c324bd78c69d10ee101e676059": &revisions{
					Branch: "main",
				},
				"0251a422d2038967eeaaaa5c8aa76c7067fdef05": &revisions{
					Branch: "main",
					Tags:   []string{"v0.1.2"},
				},
				"25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378": &revisions{
					Branch: "main",
					Tags:   []string{"v0.1.2", "v0.1.1"},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			commitMap, err := tc.g.resolveCommitMap(tc.ctx, tc.branch, tc.sortedTags)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, commitMap, tc.expectedCommitMap)
			} else {
				assert.Nil(t, commitMap)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestGenerator_resolveReleases(t *testing.T) {
	now := time.Now()

	futureTag := remote.Tag{
		Name:   "v0.1.4",
		Time:   now,
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.1.4",
	}

	tests := []struct {
		name             string
		g                *Generator
		ctx              context.Context
		sortedTags       remote.Tags
		baseRev          string
		issueMap         issueMap
		mergeMap         mergeMap
		expectedReleases []changelog.Release
	}{
		{
			name: "WithoutFutureTag_GroupingMilestone",
			g: &Generator{
				logger: log.New(log.None),
				spec: spec.Spec{
					Issues: spec.Issues{
						Grouping: spec.GroupingMilestone,
					},
					Merges: spec.Merges{
						Grouping: spec.GroupingMilestone,
					},
					Content: spec.Content{
						ReleaseURL: "https://storage.artifactory.com/project/releases/{tag}",
					},
				},
				remoteRepo: &MockRemoteRepo{
					CompareURLMocks: []CompareURLMock{
						{OutString: "https://github.com/octocat/Hello-World/compare/v0.1.2...v0.1.3"},
					},
				},
			},
			ctx:        context.Background(),
			sortedTags: remote.Tags{tag3},
			baseRev:    "v0.1.2",
			issueMap: issueMap{
				"v0.1.3": remote.Issues{issue1},
			},
			mergeMap: mergeMap{
				"v0.1.3": remote.Merges{merge1},
			},
			expectedReleases: []changelog.Release{
				{
					TagName:    "v0.1.3",
					TagURL:     "https://github.com/octocat/Hello-World/tree/v0.1.3",
					TagTime:    t3,
					ReleaseURL: "https://storage.artifactory.com/project/releases/v0.1.3",
					CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.2...v0.1.3",
					IssueGroups: []changelog.IssueGroup{
						{
							Title:  "Milestone v1.0",
							Issues: []changelog.Issue{changelogIssue1},
						},
					},
					MergeGroups: []changelog.MergeGroup{
						{
							Title:  "Milestone v1.0",
							Merges: []changelog.Merge{changelogMerge1},
						},
					},
				},
			},
		},
		{
			name: "WithoutFutureTag_GroupingLabel",
			g: &Generator{
				logger: log.New(log.None),
				spec: spec.Spec{
					Issues: spec.Issues{
						Grouping:  spec.GroupingLabel,
						BugLabels: []string{"bug"},
					},
					Merges: spec.Merges{
						Grouping:          spec.GroupingLabel,
						EnhancementLabels: []string{"enhancement"},
					},
					Content: spec.Content{
						ReleaseURL: "https://storage.artifactory.com/project/releases/{tag}",
					},
				},
				remoteRepo: &MockRemoteRepo{
					CompareURLMocks: []CompareURLMock{
						{OutString: "https://github.com/octocat/Hello-World/compare/v0.1.2...v0.1.3"},
					},
				},
			},
			ctx:        context.Background(),
			sortedTags: remote.Tags{tag3},
			baseRev:    "v0.1.2",
			issueMap: issueMap{
				"v0.1.3": remote.Issues{issue1},
			},
			mergeMap: mergeMap{
				"v0.1.3": remote.Merges{merge1},
			},
			expectedReleases: []changelog.Release{
				{
					TagName:    "v0.1.3",
					TagURL:     "https://github.com/octocat/Hello-World/tree/v0.1.3",
					TagTime:    t3,
					ReleaseURL: "https://storage.artifactory.com/project/releases/v0.1.3",
					CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.2...v0.1.3",
					IssueGroups: []changelog.IssueGroup{
						{
							Title:  "Fixed Bugs",
							Issues: []changelog.Issue{changelogIssue1},
						},
					},
					MergeGroups: []changelog.MergeGroup{
						{
							Title:  "Enhancements",
							Merges: []changelog.Merge{changelogMerge1},
						},
					},
				},
			},
		},
		{
			name: "WithFutureTag_GroupingMilestone",
			g: &Generator{
				logger: log.New(log.None),
				spec: spec.Spec{
					Issues: spec.Issues{
						Grouping: spec.GroupingMilestone,
					},
					Merges: spec.Merges{
						Grouping: spec.GroupingMilestone,
					},
					Content: spec.Content{
						ReleaseURL: "https://storage.artifactory.com/project/releases/{tag}",
					},
				},
				remoteRepo: &MockRemoteRepo{
					CompareURLMocks: []CompareURLMock{
						{OutString: "https://github.com/octocat/Hello-World/compare/v0.1.3...v0.1.4"},
						{OutString: "https://github.com/octocat/Hello-World/compare/v0.1.2...v0.1.3"},
					},
				},
			},
			ctx:        context.Background(),
			sortedTags: remote.Tags{futureTag, tag3},
			baseRev:    "v0.1.2",
			issueMap: issueMap{
				"v0.1.4": remote.Issues{issue2},
				"v0.1.3": remote.Issues{issue1},
			},
			mergeMap: mergeMap{
				"v0.1.4": remote.Merges{merge2},
				"v0.1.3": remote.Merges{merge1},
			},
			expectedReleases: []changelog.Release{
				{
					TagName:    "v0.1.4",
					TagURL:     "https://github.com/octocat/Hello-World/tree/v0.1.4",
					TagTime:    now,
					ReleaseURL: "https://storage.artifactory.com/project/releases/v0.1.4",
					CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.3...v0.1.4",
					IssueGroups: []changelog.IssueGroup{
						{
							Title:  "Closed Issues",
							Issues: []changelog.Issue{changelogIssue2},
						},
					},
					MergeGroups: []changelog.MergeGroup{
						{
							Title:  "Merged Changes",
							Merges: []changelog.Merge{changelogMerge2},
						},
					},
				},
				{
					TagName:    "v0.1.3",
					TagURL:     "https://github.com/octocat/Hello-World/tree/v0.1.3",
					TagTime:    t3,
					ReleaseURL: "https://storage.artifactory.com/project/releases/v0.1.3",
					CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.2...v0.1.3",
					IssueGroups: []changelog.IssueGroup{
						{
							Title:  "Milestone v1.0",
							Issues: []changelog.Issue{changelogIssue1},
						},
					},
					MergeGroups: []changelog.MergeGroup{
						{
							Title:  "Milestone v1.0",
							Merges: []changelog.Merge{changelogMerge1},
						},
					},
				},
			},
		},
		{
			name: "WithFutureTag_GroupingLabel",
			g: &Generator{
				logger: log.New(log.None),
				spec: spec.Spec{
					Issues: spec.Issues{
						Grouping:  spec.GroupingLabel,
						BugLabels: []string{"bug"},
					},
					Merges: spec.Merges{
						Grouping:          spec.GroupingLabel,
						EnhancementLabels: []string{"enhancement"},
					},
					Content: spec.Content{
						ReleaseURL: "https://storage.artifactory.com/project/releases/{tag}",
					},
				},
				remoteRepo: &MockRemoteRepo{
					CompareURLMocks: []CompareURLMock{
						{OutString: "https://github.com/octocat/Hello-World/compare/v0.1.3...v0.1.4"},
						{OutString: "https://github.com/octocat/Hello-World/compare/v0.1.2...v0.1.3"},
					},
				},
			},
			ctx:        context.Background(),
			sortedTags: remote.Tags{futureTag, tag3},
			baseRev:    "v0.1.2",
			issueMap: issueMap{
				"v0.1.4": remote.Issues{issue2},
				"v0.1.3": remote.Issues{issue1},
			},
			mergeMap: mergeMap{
				"v0.1.4": remote.Merges{merge2},
				"v0.1.3": remote.Merges{merge1},
			},
			expectedReleases: []changelog.Release{
				{
					TagName:    "v0.1.4",
					TagURL:     "https://github.com/octocat/Hello-World/tree/v0.1.4",
					TagTime:    now,
					ReleaseURL: "https://storage.artifactory.com/project/releases/v0.1.4",
					CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.3...v0.1.4",
					IssueGroups: []changelog.IssueGroup{
						{
							Title:  "Closed Issues",
							Issues: []changelog.Issue{changelogIssue2},
						},
					},
					MergeGroups: []changelog.MergeGroup{
						{
							Title:  "Merged Changes",
							Merges: []changelog.Merge{changelogMerge2},
						},
					},
				},
				{
					TagName:    "v0.1.3",
					TagURL:     "https://github.com/octocat/Hello-World/tree/v0.1.3",
					TagTime:    t3,
					ReleaseURL: "https://storage.artifactory.com/project/releases/v0.1.3",
					CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.2...v0.1.3",
					IssueGroups: []changelog.IssueGroup{
						{
							Title:  "Fixed Bugs",
							Issues: []changelog.Issue{changelogIssue1},
						},
					},
					MergeGroups: []changelog.MergeGroup{
						{
							Title:  "Enhancements",
							Merges: []changelog.Merge{changelogMerge1},
						},
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			releases := tc.g.resolveReleases(tc.ctx, tc.sortedTags, tc.baseRev, tc.issueMap, tc.mergeMap)

			assert.Equal(t, tc.expectedReleases, releases)
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name          string
		g             *Generator
		ctx           context.Context
		expectedError string
	}{
		{
			name: "ParseFails",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutError: errors.New("error on parsing the changelog file")},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on parsing the changelog file",
		},
		{
			name: "CheckPermissionsFails",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: errors.New("error on checking permissions")},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on checking permissions",
		},
		{
			name: "FetchBranchFails",
			g: &Generator{
				spec: spec.Spec{
					Merges: spec.Merges{
						Branch: "main",
					},
				},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchBranchMocks: []FetchBranchMock{
						{OutError: errors.New("error on getting remote branch")},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on getting remote branch",
		},
		{
			name: "FetchDefaultBranchFails",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutError: errors.New("error on getting default remote branch")},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on getting default remote branch",
		},
		{
			name: "FetchTagsFails",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutError: errors.New("error on getting remote tags")},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on getting remote tags",
		},
		{
			name: "NoNewTag",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{}},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "",
		},
		{
			name: "FetchFirstCommitFails",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
					},
					FetchFirstCommitMocks: []FetchFirstCommitMock{
						{OutError: errors.New("error on fetching first commit")},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on fetching first commit",
		},
		{
			name: "FetchParentCommitsFails_Branch",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
					},
					FetchFirstCommitMocks: []FetchFirstCommitMock{
						{OutCommit: commit1},
					},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutError: errors.New("error on fetching parent commits for branch")},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on fetching parent commits for branch",
		},
		{
			name: "FetchParentCommitsFails_Tag",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
					},
					FetchFirstCommitMocks: []FetchFirstCommitMock{
						{OutCommit: commit1},
					},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutError: errors.New("error on fetching parent commits for tag")},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on fetching parent commits for tag",
		},
		{
			name: "FetchIssuesAndMergesFails",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
					},
					FetchFirstCommitMocks: []FetchFirstCommitMock{
						{OutCommit: commit1},
					},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{OutError: errors.New("error on fetching issues and merges")},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on fetching issues and merges",
		},
		{
			name: "RenderFails",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
					RenderMocks: []RenderMock{
						{OutError: errors.New("error on rendering changelog")},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
					},
					FetchFirstCommitMocks: []FetchFirstCommitMock{
						{OutCommit: commit1},
					},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Issues{},
							OutMerges: remote.Merges{},
						},
					},
					CompareURLMocks: []CompareURLMock{
						{OutString: "https://github.com/octocat/Hello-World/compare/25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378...v0.1.1"},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on rendering changelog",
		},
		{
			name: "Success_ToTag",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
					RenderMocks: []RenderMock{
						{OutContent: "changelog"},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
					},
					FetchFirstCommitMocks: []FetchFirstCommitMock{
						{OutCommit: commit1},
					},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Issues{},
							OutMerges: remote.Merges{},
						},
					},
					CompareURLMocks: []CompareURLMock{
						{OutString: "https://github.com/octocat/Hello-World/compare/25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378...v0.1.1"},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "",
		},
		{
			name: "Success_FromAndToTags",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{
							OutChangelog: &changelog.Changelog{
								Existing: []changelog.Release{
									{TagName: "v0.1.1"},
								},
							},
						},
					},
					RenderMocks: []RenderMock{
						{OutContent: "changelog"},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag2, tag1}},
					},
					FetchFirstCommitMocks: []FetchFirstCommitMock{},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
						{OutCommits: remote.Commits{commit1}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Issues{},
							OutMerges: remote.Merges{},
						},
					},
					CompareURLMocks: []CompareURLMock{
						{OutString: "https://github.com/octocat/Hello-World/compare/25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378...v0.1.1"},
						{OutString: "https://github.com/octocat/Hello-World/compare/v0.1.1...v0.1.2"},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "",
		},
		{
			name: "Success_FutureTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						Future: "v0.1.0",
					},
				},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
					RenderMocks: []RenderMock{
						{OutContent: "changelog"},
					},
				},
				remoteRepo: &MockRemoteRepo{
					CheckPermissionsMocks: []CheckPermissionsMock{
						{OutError: nil},
					},
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{}},
					},
					FutureTagMocks: []FutureTagMock{
						{
							OutTag: remote.Tag{
								Name:   "v0.1.0",
								Time:   time.Now(),
								WebURL: "https://github.com/octocat/Hello-World/tree/v0.1.0",
							},
						},
					},
					FetchFirstCommitMocks: []FetchFirstCommitMock{
						{OutCommit: commit1},
					},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Issues{},
							OutMerges: remote.Merges{},
						},
					},
					CompareURLMocks: []CompareURLMock{
						{OutString: "https://github.com/octocat/Hello-World/compare/25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378...v0.1.0"},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.g.Generate(tc.ctx)

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
