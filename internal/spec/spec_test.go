package spec

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	repo, err := git.PlainOpen("../..")
	assert.NoError(t, err)

	spec, err := Default(repo)
	assert.NoError(t, err)

	assert.NotNil(t, spec)
	assert.Equal(t, PlatformGitHub, spec.Repo.Platform)
	assert.Equal(t, "moorara/changelog", spec.Repo.Path)
	assert.Equal(t, "", spec.Repo.AccessToken)
	assert.Equal(t, "CHANGELOG.md", spec.General.File)
	assert.Equal(t, "HISTORY.md", spec.General.Base)
	assert.Equal(t, false, spec.General.Print)
	assert.Equal(t, false, spec.General.Verbose)
	assert.Equal(t, "master", spec.Release.Branch)
	assert.Equal(t, "", spec.Release.FromTag)
	assert.Equal(t, "", spec.Release.ToTag)
	assert.Equal(t, "", spec.Release.FutureTag)
	assert.Equal(t, []string{}, spec.Release.ExcludeTags)
	assert.Equal(t, "", spec.Release.ExcludeTagsRegex)
	assert.Equal(t, SelectAll, spec.Changes.Issues)
	assert.Equal(t, SelectAll, spec.Changes.Merges)
	assert.Equal(t, GroupByLabel, spec.Format.GroupBy)
	assert.Equal(t, "", spec.Format.ReleaseURL)
	assert.Equal(t, []string{}, spec.Labels.Include)
	assert.Equal(t, []string{"duplicate", "invalid", "question", "wontfix"}, spec.Labels.Exclude)
	assert.Equal(t, []string{"summary", "release-summary"}, spec.Labels.Summary)
	assert.Equal(t, []string{"removed"}, spec.Labels.Removed)
	assert.Equal(t, []string{"breaking", "backward-incompatible", "backwards-incompatible"}, spec.Labels.Breaking)
	assert.Equal(t, []string{"deprecated"}, spec.Labels.Deprecated)
	assert.Equal(t, []string{"feature"}, spec.Labels.Feature)
	assert.Equal(t, []string{"enhancement"}, spec.Labels.Enhancement)
	assert.Equal(t, []string{"bug"}, spec.Labels.Bug)
	assert.Equal(t, []string{"security"}, spec.Labels.Security)
}

func TestSpecPrintHelp(t *testing.T) {
	s := new(Spec)
	err := s.PrintHelp()

	assert.NoError(t, err)
}
