package generate

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/git"
	"github.com/moorara/changelog/internal/spec"
	"github.com/moorara/changelog/pkg/log"
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
		gitRepo *git.Repo
	}{
		{
			name:    "GitHub",
			s:       specGitHub,
			logger:  log.New(log.None),
			gitRepo: &git.Repo{},
		},
		{
			name:    "GitLab",
			s:       specGitLab,
			logger:  log.New(log.None),
			gitRepo: &git.Repo{},
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

func TestGenerator_ResolveTags(t *testing.T) {
	specWithDefaults := spec.Spec{}

	specWithInvalidFromTag := spec.Spec{}
	specWithInvalidFromTag.Release.FromTag = "0.1.0"

	specWithInvalidToTag := spec.Spec{}
	specWithInvalidToTag.Release.ToTag = "v0.2.0"

	specWithBeforeFromTag := spec.Spec{}
	specWithBeforeFromTag.Release.FromTag = "v0.1.0"

	specWithToTagBeforeFromTag := spec.Spec{}
	specWithToTagBeforeFromTag.Release.FromTag = "v0.1.1"
	specWithToTagBeforeFromTag.Release.ToTag = "v0.1.0"

	specWithSameFromAndToTags := spec.Spec{}
	specWithSameFromAndToTags.Release.FromTag = "v0.1.1"
	specWithSameFromAndToTags.Release.ToTag = "v0.1.1"

	specWithValidFromAndToTags := spec.Spec{}
	specWithValidFromAndToTags.Release.FromTag = "v0.1.1"
	specWithValidFromAndToTags.Release.ToTag = "v0.1.2"

	ts1, _ := time.Parse(time.RFC3339, "2020-10-10T10:00:00-04:00")
	ts2, _ := time.Parse(time.RFC3339, "2020-10-11T11:00:00-04:00")
	ts3, _ := time.Parse(time.RFC3339, "2020-10-12T12:00:00-04:00")

	tag1 := git.Tag{
		Name: "v0.1.0",
		Tagger: git.Tagger{
			Timestamp: ts1,
		},
	}

	tag2 := git.Tag{
		Name: "v0.1.1",
		Tagger: git.Tagger{
			Timestamp: ts2,
		},
	}

	tag3 := git.Tag{
		Name: "v0.1.2",
		Tagger: git.Tagger{
			Timestamp: ts3,
		},
	}

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
				spec:   specWithDefaults,
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
				spec:   specWithDefaults,
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
				spec:   specWithDefaults,
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
				spec:   specWithDefaults,
				logger: log.New(log.None),
			},
			tags: git.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.0"},
				},
			},
			expectedFromTag: tag1,
			expectedToTag:   tag2,
			expectedError:   nil,
		},
		{
			name: "InvalidFromTag",
			g: &Generator{
				spec:   specWithInvalidFromTag,
				logger: log.New(log.None),
			},
			tags: git.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.0"},
				},
			},
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedError:   errors.New("from-tag not found: 0.1.0"),
		},
		{
			name: "InvalidToTag",
			g: &Generator{
				spec:   specWithInvalidToTag,
				logger: log.New(log.None),
			},
			tags: git.Tags{tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.0"},
				},
			},
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedError:   errors.New("to-tag not found: v0.2.0"),
		},
		{
			name: "FromTagBeforeLastChangelogTag",
			g: &Generator{
				spec:   specWithBeforeFromTag,
				logger: log.New(log.None),
			},
			tags: git.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.1"},
					{GitTag: "v0.1.0"},
				},
			},
			expectedFromTag: tag2,
			expectedToTag:   tag3,
			expectedError:   nil,
		},
		{
			name: "ToTagBeforeFromTag",
			g: &Generator{
				spec:   specWithToTagBeforeFromTag,
				logger: log.New(log.None),
			},
			tags: git.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.1"},
					{GitTag: "v0.1.0"},
				},
			},
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedError:   errors.New("to-tag should be after the from-tag"),
		},
		{
			name: "SameFromAndToTags",
			g: &Generator{
				spec:   specWithSameFromAndToTags,
				logger: log.New(log.None),
			},
			tags: git.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.1"},
					{GitTag: "v0.1.0"},
				},
			},
			expectedFromTag: git.Tag{},
			expectedToTag:   git.Tag{},
			expectedError:   errors.New("to-tag should be after the from-tag"),
		},
		{
			name: "ValidFromAndToTags",
			g: &Generator{
				spec:   specWithValidFromAndToTags,
				logger: log.New(log.None),
			},
			tags: git.Tags{tag3, tag2, tag1},
			chlog: &changelog.Changelog{
				Releases: []changelog.Release{
					{GitTag: "v0.1.1"},
					{GitTag: "v0.1.0"},
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
