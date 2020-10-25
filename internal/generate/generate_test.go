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
	ts1, _ = time.Parse(time.RFC3339, "2020-10-02T02:00:00-04:00")
	ts2, _ = time.Parse(time.RFC3339, "2020-10-12T09:00:00-04:00")
	ts3, _ = time.Parse(time.RFC3339, "2020-10-22T16:00:00-04:00")

	JohnDoe = git.Signature{
		Name:      "John Doe",
		Email:     "john@doe.com",
		Timestamp: ts1,
	}

	JaneDoe = git.Signature{
		Name:      "Jane Doe",
		Email:     "jane@doe.com",
		Timestamp: ts2,
	}

	JimDoe = git.Signature{
		Name:      "Jim Doe",
		Email:     "jim@doe.com",
		Timestamp: ts3,
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

func TestGenerator_ResolveTags(t *testing.T) {
	tests := []struct {
		name            string
		g               *Generator
		tags            git.Tags
		chlog           *changelog.Changelog
		expectedFromTag git.Tag
		expectedToTag   git.Tag
		expectedError   error
	}{
		{
			name: "NoTagAndNoChangelog",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			tags:            git.Tags{},
			chlog:           &changelog.Changelog{},
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedError:   nil,
		},
		{
			name: "FirstRelease",
			g: &Generator{
				spec:   spec.Spec{},
				logger: log.New(log.None),
			},
			tags:            git.Tags{tag1},
			chlog:           &changelog.Changelog{},
			expectedFromTag: git.Tag{},
			expectedToTag:   tag1,
			expectedError:   nil,
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
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedError:   errors.New("changelog tag not found: 0.1.0"),
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
			expectedFromTag: tag1,
			expectedToTag:   tag2,
			expectedError:   nil,
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
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedError:   errors.New("from-tag not found: invalid"),
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
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedError:   errors.New("to-tag not found: invalid"),
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
			expectedFromTag: tag2,
			expectedToTag:   tag3,
			expectedError:   nil,
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
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedError:   errors.New("to-tag should be after the from-tag"),
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
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedError:   errors.New("to-tag should be after the from-tag"),
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
			expectedFromTag: tag2,
			expectedToTag:   tag3,
			expectedError:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fromTag, toTag, err := tc.g.resolveTags(tc.tags, tc.chlog)

			assert.Equal(t, tc.expectedFromTag, fromTag)
			assert.Equal(t, tc.expectedToTag, toTag)
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
	}{}

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
