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
	t1, _ = time.Parse(time.RFC3339, "2020-10-02T02:00:00-04:00")
	t2, _ = time.Parse(time.RFC3339, "2020-10-12T09:00:00-04:00")
	t3, _ = time.Parse(time.RFC3339, "2020-10-22T16:00:00-04:00")

	JohnDoe = git.Signature{
		Name:  "John Doe",
		Email: "john@doe.com",
		Time:  t1,
	}

	JaneDoe = git.Signature{
		Name:  "Jane Doe",
		Email: "jane@doe.com",
		Time:  t2,
	}

	JimDoe = git.Signature{
		Name:  "Jim Doe",
		Email: "jim@doe.com",
		Time:  t3,
	}

	commit1 = git.Commit{
		Hash:      "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		Author:    JohnDoe,
		Committer: JohnDoe,
		Message:   "foo",
	}

	commit2 = git.Commit{
		Hash:      "0251a422d2038967eeaaaa5c8aa76c7067fdef05",
		Author:    JaneDoe,
		Committer: JaneDoe,
		Message:   "bar",
	}

	commit3 = git.Commit{
		Hash:      "c414d1004154c6c324bd78c69d10ee101e676059",
		Author:    JimDoe,
		Committer: JimDoe,
		Message:   "baz",
	}

	tag1 = git.Tag{
		Type:   git.Lightweight,
		Hash:   "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		Name:   "v0.1.1",
		Commit: commit1,
	}

	tag2 = git.Tag{
		Type:   git.Lightweight,
		Hash:   "0251a422d2038967eeaaaa5c8aa76c7067fdef05",
		Name:   "v0.1.2",
		Commit: commit2,
	}

	tag3 = git.Tag{
		Type:   git.Lightweight,
		Hash:   "c414d1004154c6c324bd78c69d10ee101e676059",
		Name:   "v0.1.3",
		Commit: commit3,
	}

	change1 = remote.Change{
		Number:    1001,
		Title:     "Found a bug",
		Labels:    []string{"bug"},
		Milestone: "v1.0",
	}

	change2 = remote.Change{
		Number:    1002,
		Title:     "Discovered a vulnerability",
		Labels:    []string{"invalid"},
		Milestone: "",
	}

	change3 = remote.Change{
		Number:    1003,
		Title:     "Added a feature",
		Labels:    []string{"enhancement"},
		Milestone: "v1.0",
	}

	change4 = remote.Change{
		Number:    1004,
		Title:     "Refactored code",
		Labels:    nil,
		Milestone: "v1.0",
	}
)

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

func TestGenerator_ResolveGitTags(t *testing.T) {
	tests := []struct {
		name              string
		g                 *Generator
		tags              git.Tags
		chlog             *changelog.Changelog
		expectedFromTag   git.Tag
		expectedToTag     git.Tag
		expectedFutureTag git.Tag
		expectedError     error
	}{
		{
			name: "NoTagAndNoChangelog",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			tags:              git.Tags{},
			chlog:             &changelog.Changelog{},
			expectedFromTag:   git.Tag{},
			expectedToTag:     git.Tag{},
			expectedFutureTag: git.Tag{},
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
			tags:            git.Tags{},
			chlog:           &changelog.Changelog{},
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedFutureTag: git.Tag{
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
			tags:              git.Tags{tag1},
			chlog:             &changelog.Changelog{},
			expectedFromTag:   git.Tag{},
			expectedToTag:     tag1,
			expectedFutureTag: git.Tag{},
			expectedError:     nil,
		},
		{
			name: "TagNotInChangelog",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			tags: git.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "0.1.0"},
				},
			},
			expectedFromTag:   git.Tag{},
			expectedToTag:     git.Tag{},
			expectedFutureTag: git.Tag{},
			expectedError:     errors.New("changelog tag not found: 0.1.0"),
		},
		{
			name: "SecondRelease",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			tags: git.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.1"},
				},
			},
			expectedFromTag:   tag1,
			expectedToTag:     tag2,
			expectedFutureTag: git.Tag{},
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
			tags: git.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.1"},
				},
			},
			expectedFromTag:   git.Tag{},
			expectedToTag:     git.Tag{},
			expectedFutureTag: git.Tag{},
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
			tags: git.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.1"},
				},
			},
			expectedFromTag:   git.Tag{},
			expectedToTag:     git.Tag{},
			expectedFutureTag: git.Tag{},
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
			tags: git.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.2"},
					{GitTag: "v0.1.1"},
				},
			},
			expectedFromTag:   tag2,
			expectedToTag:     tag3,
			expectedFutureTag: git.Tag{},
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
			tags: git.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.2"},
					{GitTag: "v0.1.1"},
				},
			},
			expectedFromTag:   git.Tag{},
			expectedToTag:     git.Tag{},
			expectedFutureTag: git.Tag{},
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
			tags: git.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.2"},
					{GitTag: "v0.1.1"},
				},
			},
			expectedFromTag:   git.Tag{},
			expectedToTag:     git.Tag{},
			expectedFutureTag: git.Tag{},
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
			tags: git.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.2"},
					{GitTag: "v0.1.1"},
				},
			},
			expectedFromTag:   tag2,
			expectedToTag:     tag3,
			expectedFutureTag: git.Tag{},
			expectedError:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fromTag, toTag, futureTag, err := tc.g.resolveGitTags(tc.tags, tc.chlog)

			assert.Equal(t, tc.expectedFromTag, fromTag)
			assert.Equal(t, tc.expectedToTag, toTag)
			assert.Equal(t, tc.expectedFutureTag, futureTag)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestGenerator_ResolveRemoteTags(t *testing.T) {
	t1, _ = time.Parse(time.RFC3339, "2020-10-02T02:00:00-04:00")
	t2, _ = time.Parse(time.RFC3339, "2020-10-12T09:00:00-04:00")
	t3, _ = time.Parse(time.RFC3339, "2020-10-22T16:00:00-04:00")

	tag1 := remote.Tag{
		Name: "v0.1.1",
		SHA:  "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		Time: t1,
	}

	tag2 := remote.Tag{
		Name: "v0.1.2",
		SHA:  "0251a422d2038967eeaaaa5c8aa76c7067fdef05",
		Time: t2,
	}

	tag3 := remote.Tag{
		Name: "v0.1.3",
		SHA:  "c414d1004154c6c324bd78c69d10ee101e676059",
		Time: t3,
	}

	tests := []struct {
		name              string
		g                 *Generator
		tags              remote.Tags
		chlog             *changelog.Changelog
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
			tags:              remote.Tags{},
			chlog:             &changelog.Changelog{},
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
			tags:            remote.Tags{},
			chlog:           &changelog.Changelog{},
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
			tags:              remote.Tags{tag1},
			chlog:             &changelog.Changelog{},
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
			tags: remote.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "0.1.0"},
				},
			},
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
			tags: remote.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.1"},
				},
			},
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
			tags: remote.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.1"},
				},
			},
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
			tags: remote.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.1"},
				},
			},
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
			tags: remote.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.2"},
					{GitTag: "v0.1.1"},
				},
			},
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
			tags: remote.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.2"},
					{GitTag: "v0.1.1"},
				},
			},
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
			tags: remote.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.2"},
					{GitTag: "v0.1.1"},
				},
			},
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
			tags: remote.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.2"},
					{GitTag: "v0.1.1"},
				},
			},
			expectedFromTag:   tag2,
			expectedToTag:     tag3,
			expectedFutureTag: remote.Tag{},
			expectedError:     nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fromTag, toTag, futureTag, err := tc.g.resolveRemoteTags(tc.tags, tc.chlog)

			assert.Equal(t, tc.expectedFromTag, fromTag)
			assert.Equal(t, tc.expectedToTag, toTag)
			assert.Equal(t, tc.expectedFutureTag, futureTag)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestGenerator_FilterByLabels(t *testing.T) {
	tests := []struct {
		name           string
		g              *Generator
		issues         remote.Changes
		merges         remote.Changes
		expectedIssues remote.Changes
		expectedMerges remote.Changes
	}{
		{
			name: "None",
			g: &Generator{
				spec: spec.Spec{
					Issues: spec.Issues{
						Selection: spec.SelectionNone,
					},
					Merges: spec.Merges{
						Selection: spec.SelectionNone,
					},
				},
				logger: log.New(log.None),
			},
			issues:         remote.Changes{change1, change2},
			merges:         remote.Changes{change3, change4},
			expectedIssues: remote.Changes{},
			expectedMerges: remote.Changes{},
		},
		{
			name: "AllWithIncludeLabels",
			g: &Generator{
				spec: spec.Spec{
					Issues: spec.Issues{
						Selection:     spec.SelectionAll,
						IncludeLabels: []string{"bug"},
					},
					Merges: spec.Merges{
						Selection:     spec.SelectionAll,
						IncludeLabels: []string{"enhancement"},
					},
				},
				logger: log.New(log.None),
			},
			issues:         remote.Changes{change1, change2},
			merges:         remote.Changes{change3, change4},
			expectedIssues: remote.Changes{change1},
			expectedMerges: remote.Changes{change3, change4},
		},
		{
			name: "AllWithExcludeLabels",
			g: &Generator{
				spec: spec.Spec{
					Issues: spec.Issues{
						Selection:     spec.SelectionAll,
						ExcludeLabels: []string{"invalid"},
					},
					Merges: spec.Merges{
						Selection:     spec.SelectionAll,
						ExcludeLabels: []string{"enhancement"},
					},
				},
				logger: log.New(log.None),
			},
			issues:         remote.Changes{change1, change2},
			merges:         remote.Changes{change3, change4},
			expectedIssues: remote.Changes{change1},
			expectedMerges: remote.Changes{change4},
		},
		{
			name: "LabeledWithIncludeLabels",
			g: &Generator{
				spec: spec.Spec{
					Issues: spec.Issues{
						Selection:     spec.SelectionLabeled,
						IncludeLabels: []string{"bug"},
					},
					Merges: spec.Merges{
						Selection:     spec.SelectionLabeled,
						IncludeLabels: []string{"enhancement"},
					},
				},
				logger: log.New(log.None),
			},
			issues:         remote.Changes{change1, change2},
			merges:         remote.Changes{change3, change4},
			expectedIssues: remote.Changes{change1},
			expectedMerges: remote.Changes{change3},
		},
		{
			name: "LabeledWithExcludeLabels",
			g: &Generator{
				spec: spec.Spec{
					Issues: spec.Issues{
						Selection:     spec.SelectionLabeled,
						ExcludeLabels: []string{"invalid"},
					},
					Merges: spec.Merges{
						Selection:     spec.SelectionLabeled,
						ExcludeLabels: []string{"enhancement"},
					},
				},
				logger: log.New(log.None),
			},
			issues:         remote.Changes{change1, change2},
			merges:         remote.Changes{change3, change4},
			expectedIssues: remote.Changes{change1},
			expectedMerges: remote.Changes{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issues, merges := tc.g.filterByLabels(tc.issues, tc.merges)

			assert.Equal(t, tc.expectedIssues, issues)
			assert.Equal(t, tc.expectedMerges, merges)
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
			name: "TagsFails",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
				processor: &MockChangelogProcessor{
					ParseMocks: []ParseMock{
						{OutChangelog: &changelog.Changelog{}},
					},
				},
				gitRepo: &MockGitRepo{
					TagsMocks: []TagsMock{
						{OutError: errors.New("error on getting git tags")},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on getting git tags",
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
				gitRepo: &MockGitRepo{
					TagsMocks: []TagsMock{
						{OutTags: git.Tags{}},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "",
		},
		{
			name: "FetchIssuesAndMergesFails",
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
				},
				gitRepo: &MockGitRepo{
					TagsMocks: []TagsMock{
						{OutTags: git.Tags{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
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
						{OutError: errors.New("error on rendering changelog")},
					},
				},
				gitRepo: &MockGitRepo{
					TagsMocks: []TagsMock{
						{OutTags: git.Tags{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Changes{},
							OutMerges: remote.Changes{},
						},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "error on rendering changelog",
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
				gitRepo: &MockGitRepo{
					TagsMocks: []TagsMock{
						{OutTags: git.Tags{}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Changes{},
							OutMerges: remote.Changes{},
						},
					},
				},
			},
			ctx:           context.Background(),
			expectedError: "",
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
				gitRepo: &MockGitRepo{
					TagsMocks: []TagsMock{
						{OutTags: git.Tags{tag1}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Changes{},
							OutMerges: remote.Changes{},
						},
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
									{GitTag: "v0.1.1"},
								},
							},
						},
					},
					RenderMocks: []RenderMock{
						{OutContent: "changelog"},
					},
				},
				gitRepo: &MockGitRepo{
					TagsMocks: []TagsMock{
						{OutTags: git.Tags{tag2, tag1}},
					},
				},
				remoteRepo: &MockRemoteRepo{
					FetchIssuesAndMergesMocks: []FetchIssuesAndMergesMock{
						{
							OutIssues: remote.Changes{},
							OutMerges: remote.Changes{},
						},
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
