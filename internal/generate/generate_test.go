package generate

import (
	"log"
	"testing"

	"github.com/moorara/changelog/internal/git"
	"github.com/moorara/changelog/internal/spec"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	specGitHub := spec.Spec{}
	specGitHub.Repo.Platform = spec.PlatformGitHub

	specGitLab := spec.Spec{}
	specGitLab.Repo.Platform = spec.PlatformGitLab

	tests := []struct {
		name    string
		logger  *log.Logger
		gitRepo *git.Repo
		s       spec.Spec
	}{
		{
			name:    "GitHub",
			logger:  &log.Logger{},
			gitRepo: &git.Repo{},
			s:       specGitHub,
		},
		{
			name:    "GitLab",
			logger:  &log.Logger{},
			gitRepo: &git.Repo{},
			s:       specGitLab,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := New(tc.logger, tc.gitRepo, tc.s)

			assert.NotNil(t, g)
			assert.Equal(t, tc.logger, g.logger)
			assert.Equal(t, tc.gitRepo, g.gitRepo)
			assert.NotNil(t, g.remoteRepo)
			assert.NotNil(t, g.processor)
		})
	}
}

func TestGeneratorGenerate(t *testing.T) {
	tests := []struct {
		name          string
		g             *Generator
		expectedError string
	}{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.g.Generate()

			if tc.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
