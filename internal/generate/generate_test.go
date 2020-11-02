package generate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/git"
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
		Name:     "monalisa octocat",
		Email:    "octocat@github.com",
		Username: "octocat",
		URL:      "https://github.com/octocat",
	}

	user2 = remote.User{
		Name:     "monalisa octodog",
		Email:    "octodog@github.com",
		Username: "octodog",
		URL:      "https://github.com/octodog",
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
	}

	tag2 = remote.Tag{
		Name:   "v0.1.2",
		Time:   t2,
		Commit: commit2,
	}

	tag3 = remote.Tag{
		Name:   "v0.1.3",
		Time:   t3,
		Commit: commit3,
	}

	issue1 = remote.Issue{
		Change: remote.Change{
			Number:    1001,
			Title:     "Found a bug",
			Labels:    []string{"bug"},
			Milestone: "v1.0",
			Time:      t3,
			Creator:   user1,
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
			Creator:   user1,
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
			Creator:   user1,
		},
		Merger: user1,
		Commit: commit3,
	}

	merge2 = remote.Merge{
		Change: remote.Change{
			Number:    1004,
			Title:     "Refactored code",
			Labels:    nil,
			Milestone: "v1.0",
			Time:      t4, // Unrleased change for future tag
			Creator:   user1,
		},
		Merger: user2,
		Commit: commit4,
	}

	changelogIssue1 = changelog.Issue{
		Number: 1001,
		Title:  "Found a bug",
		OpenedBy: changelog.User{
			Name:     "monalisa octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
		ClosedBy: changelog.User{
			Name:     "monalisa octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
	}

	changelogIssue2 = changelog.Issue{
		Number: 1002,
		Title:  "Discovered a vulnerability",
		OpenedBy: changelog.User{
			Name:     "monalisa octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
		ClosedBy: changelog.User{
			Name:     "monalisa octodog",
			Username: "octodog",
			URL:      "https://github.com/octodog",
		},
	}

	changelogMerge1 = changelog.Merge{
		Number: 1003,
		Title:  "Added a feature",
		OpenedBy: changelog.User{
			Name:     "monalisa octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
		MergedBy: changelog.User{
			Name:     "monalisa octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
	}

	changelogMerge2 = changelog.Merge{
		Number: 1004,
		Title:  "Refactored code",
		OpenedBy: changelog.User{
			Name:     "monalisa octocat",
			Username: "octocat",
			URL:      "https://github.com/octocat",
		},
		MergedBy: changelog.User{
			Name:     "monalisa octodog",
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
	specGitHub := spec.Spec{}
	specGitHub.Repo.Platform = spec.PlatformGitHub

	specGitLab := spec.Spec{}
	specGitLab.Repo.Platform = spec.PlatformGitLab

	tests := []struct {
		name    string
		s       spec.Spec
		logger  log.Logger
		gitRepo git.Repo
	}{
		{
			name:    "GitHub",
			s:       specGitHub,
			logger:  log.New(log.None),
			gitRepo: &MockGitRepo{},
		},
		{
			name:    "GitLab",
			s:       specGitLab,
			logger:  log.New(log.None),
			gitRepo: &MockGitRepo{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := New(tc.s, tc.logger, tc.gitRepo)

			assert.NotNil(t, g)
			assert.Equal(t, tc.logger, g.logger)
			assert.Equal(t, tc.gitRepo, g.gitRepo)
			assert.NotNil(t, g.remoteRepo)
			assert.NotNil(t, g.processor)
		})
	}
}

func TestGenerator_resolveTags(t *testing.T) {
	tests := []struct {
		name              string
		g                 *Generator
		chlog             *changelog.Changelog
		sortedTags        remote.Tags
		expectedFromTag   remote.Tag
		expectedToTag     remote.Tag
		expectedFutureTag remote.Tag
		expectedError     error
	}{
		{
			name: "NoTagAndNoChangelog",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			chlog:             &changelog.Changelog{},
			sortedTags:        remote.Tags{},
			expectedFromTag:   remote.Tag{},
			expectedToTag:     remote.Tag{},
			expectedFutureTag: remote.Tag{},
			expectedError:     nil,
		},
		{
			name: "FutureRelease",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						Future: "v0.1.0",
					},
				},
				logger: log.New(log.None),
			},
			chlog:           &changelog.Changelog{},
			sortedTags:      remote.Tags{},
			expectedFromTag: remote.Tag{},
			expectedToTag:   remote.Tag{},
			expectedFutureTag: remote.Tag{
				Name: "v0.1.0",
			},
			expectedError: nil,
		},
		{
			name: "FirstRelease",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			chlog:             &changelog.Changelog{},
			sortedTags:        remote.Tags{tag1},
			expectedFromTag:   remote.Tag{},
			expectedToTag:     tag1,
			expectedFutureTag: remote.Tag{},
			expectedError:     nil,
		},
		{
			name: "TagNotInChangelog",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{TagName: "0.1.0"},
				},
			},
			sortedTags:        remote.Tags{tag2, tag1},
			expectedFromTag:   remote.Tag{},
			expectedToTag:     remote.Tag{},
			expectedFutureTag: remote.Tag{},
			expectedError:     errors.New("changelog tag not found: 0.1.0"),
		},
		{
			name: "SecondRelease",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			sortedTags:        remote.Tags{tag2, tag1},
			expectedFromTag:   tag1,
			expectedToTag:     tag2,
			expectedFutureTag: remote.Tag{},
			expectedError:     nil,
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
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			sortedTags:        remote.Tags{tag2, tag1},
			expectedFromTag:   remote.Tag{},
			expectedToTag:     remote.Tag{},
			expectedFutureTag: remote.Tag{},
			expectedError:     errors.New("from-tag not found: invalid"),
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
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{TagName: "v0.1.1"},
				},
			},
			sortedTags:        remote.Tags{tag2, tag1},
			expectedFromTag:   remote.Tag{},
			expectedToTag:     remote.Tag{},
			expectedFutureTag: remote.Tag{},
			expectedError:     errors.New("to-tag not found: invalid"),
		},
		{
			name: "FromTagBeforeLastChangelogTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From: "v0.1.1",
					},
				},
				logger: log.New(log.None),
			},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{TagName: "v0.1.2"},
					{TagName: "v0.1.1"},
				},
			},
			sortedTags:        remote.Tags{tag3, tag2, tag1},
			expectedFromTag:   tag2,
			expectedToTag:     tag3,
			expectedFutureTag: remote.Tag{},
			expectedError:     nil,
		},
		{
			name: "ToTagBeforeFromTag",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From: "v0.1.2",
						To:   "v0.1.1",
					},
				},
				logger: log.New(log.None),
			},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{TagName: "v0.1.2"},
					{TagName: "v0.1.1"},
				},
			},
			sortedTags:        remote.Tags{tag3, tag2, tag1},
			expectedFromTag:   remote.Tag{},
			expectedToTag:     remote.Tag{},
			expectedFutureTag: remote.Tag{},
			expectedError:     errors.New("to-tag should be after the from-tag"),
		},
		{
			name: "SameFromAndToTags",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From: "v0.1.2",
						To:   "v0.1.2",
					},
				},
				logger: log.New(log.None),
			},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{TagName: "v0.1.2"},
					{TagName: "v0.1.1"},
				},
			},
			sortedTags:        remote.Tags{tag3, tag2, tag1},
			expectedFromTag:   remote.Tag{},
			expectedToTag:     remote.Tag{},
			expectedFutureTag: remote.Tag{},
			expectedError:     errors.New("to-tag should be after the from-tag"),
		},
		{
			name: "ValidFromAndToTags",
			g: &Generator{
				spec: spec.Spec{
					Tags: spec.Tags{
						From: "v0.1.2",
						To:   "v0.1.3",
					},
				},
				logger: log.New(log.None),
			},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{TagName: "v0.1.2"},
					{TagName: "v0.1.1"},
				},
			},
			sortedTags:        remote.Tags{tag3, tag2, tag1},
			expectedFromTag:   tag2,
			expectedToTag:     tag3,
			expectedFutureTag: remote.Tag{},
			expectedError:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fromTag, toTag, futureTag, err := tc.g.resolveTags(tc.chlog, tc.sortedTags)

			assert.Equal(t, tc.expectedFromTag, fromTag)
			assert.Equal(t, tc.expectedToTag, toTag)
			assert.Equal(t, tc.expectedFutureTag, futureTag)
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
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
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
			name: "FetchParentCommitsFails_Branch",
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
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Issues{},
							OutMerges: remote.Merges{},
						},
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
					RenderMocks: []RenderMock{
						{OutError: errors.New("error on rendering changelog")},
					},
				},
				remoteRepo: &MockRemoteRepo{
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Issues{},
							OutMerges: remote.Merges{},
						},
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
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Issues{},
							OutMerges: remote.Merges{},
						},
					},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
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
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag1}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Issues{},
							OutMerges: remote.Merges{},
						},
					},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
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
								Releases: []changelog.Release{
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
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{tag2, tag1}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Issues{},
							OutMerges: remote.Merges{},
						},
					},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
						{OutCommits: remote.Commits{commit1}},
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
					FetchDefaultBranchMocks: []FetchDefaultBranchMock{
						{OutBranch: branch},
					},
					FetchTagsMocks: []FetchTagsMock{
						{OutTags: remote.Tags{}},
					},
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Issues{},
							OutMerges: remote.Merges{},
						},
					},
					FetchParentCommitsMocks: []FetchParentCommitsMock{
						{OutCommits: remote.Commits{commit3, commit2, commit1}},
						{OutCommits: remote.Commits{commit2, commit1}},
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
