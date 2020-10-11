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
	assert.Equal(t, SelectAll, spec.Changes.Issues)
	assert.Equal(t, SelectAll, spec.Changes.Merges)
	assert.Equal(t, "master", spec.Release.Branch)
	assert.Equal(t, "", spec.Release.FromTag)
	assert.Equal(t, "", spec.Release.ToTag)
	assert.Equal(t, "", spec.Release.FutureTag)
	assert.Equal(t, []string{}, spec.Release.ExcludeTags)
	assert.Equal(t, "", spec.Release.ExcludeTagsRegex)
	assert.Equal(t, GroupByLabel, spec.Format.GroupBy)
	assert.Equal(t, "", spec.Format.ReleaseURL)
	assert.Equal(t, []string{}, spec.Labels.Include)
	assert.Equal(t, []string{"duplicate", "invalid", "question", "wontfix"}, spec.Labels.Exclude)
	assert.Equal(t, []string{"summary", "release-summary"}, spec.Labels.Summary)
	assert.Equal(t, []string{"removed"}, spec.Labels.Removed)
	assert.Equal(t, []string{"breaking"}, spec.Labels.Breaking)
	assert.Equal(t, []string{"deprecated"}, spec.Labels.Deprecated)
	assert.Equal(t, []string{"feature"}, spec.Labels.Feature)
	assert.Equal(t, []string{"enhancement"}, spec.Labels.Enhancement)
	assert.Equal(t, []string{"bug"}, spec.Labels.Bug)
	assert.Equal(t, []string{"security"}, spec.Labels.Security)
}

func TestSpecUpdateFromFile(t *testing.T) {
	repo, err := git.PlainOpen("../..")
	assert.NoError(t, err)

	spec, err := Default(repo)
	assert.NoError(t, err)

	expectedMinSpec := spec
	expectedMinSpec.General.Print = true
	expectedMinSpec.Release.Branch = "release"
	expectedMinSpec.Format.GroupBy = GroupBy("milestone")
	expectedMinSpec.Labels.Summary = []string{"summary", "highlight"}
	expectedMinSpec.Labels.Removed = []string{"removed", "dropped"}
	expectedMinSpec.Labels.Breaking = []string{"breaking", "incompatible"}
	expectedMinSpec.Labels.Enhancement = []string{"enhancement", "improvement"}

	expectedMaxSpec := spec
	expectedMaxSpec.General.File = "RELEASE-NOTES.md"
	expectedMaxSpec.General.Base = "OLD-NOTES.md"
	expectedMaxSpec.General.Print = true
	expectedMaxSpec.General.Verbose = true
	expectedMaxSpec.Changes.Issues = Select("labeled")
	expectedMaxSpec.Changes.Merges = Select("labeled")
	expectedMaxSpec.Release.Branch = "release"
	expectedMaxSpec.Release.FromTag = "v0.1.0"
	expectedMaxSpec.Release.ToTag = "v0.5.0"
	expectedMaxSpec.Release.FutureTag = "v1.0.0"
	expectedMaxSpec.Release.ExcludeTags = []string{"staging"}
	expectedMaxSpec.Release.ExcludeTagsRegex = "(.*)-(alpha|beta)"
	expectedMaxSpec.Format.GroupBy = GroupBy("milestone")
	expectedMaxSpec.Format.ReleaseURL = "https://storage.artifactory.com/project/releases/{tag}"
	expectedMaxSpec.Labels.Include = []string{"breaking", "bug", "defect", "deprecated", "dropped", "enhancement", "fault", "feature", "highlight", "improvement", "incompatible", "new", "not-supported", "privacy", "removed", "security", "summary"}
	expectedMaxSpec.Labels.Exclude = []string{"documentation", "duplicate", "invalid", "question", "wontfix"}
	expectedMaxSpec.Labels.Summary = []string{"summary", "highlight"}
	expectedMaxSpec.Labels.Removed = []string{"removed", "dropped"}
	expectedMaxSpec.Labels.Breaking = []string{"breaking", "incompatible"}
	expectedMaxSpec.Labels.Deprecated = []string{"deprecated", "not-supported"}
	expectedMaxSpec.Labels.Feature = []string{"feature", "new"}
	expectedMaxSpec.Labels.Enhancement = []string{"enhancement", "improvement"}
	expectedMaxSpec.Labels.Bug = []string{"bug", "defect", "fault"}
	expectedMaxSpec.Labels.Security = []string{"security", "privacy"}

	tests := []struct {
		name          string
		specFiles     []string
		spec          Spec
		expectedSpec  Spec
		expectedError string
	}{
		{
			name:         "NoSpecFile",
			specFiles:    []string{"test/null"},
			spec:         spec,
			expectedSpec: spec,
		},
		{
			name:          "EmptySpecFile",
			specFiles:     []string{"test/empty.yaml"},
			spec:          spec,
			expectedError: "EOF",
		},
		{
			name:          "InvalidSpecFile",
			specFiles:     []string{"test/invalid.yaml"},
			spec:          spec,
			expectedError: "yaml: unmarshal errors",
		},
		{
			name:         "MinimumSpecFile",
			specFiles:    []string{"test/min.yaml"},
			spec:         spec,
			expectedSpec: expectedMinSpec,
		},
		{
			name:         "MaximumSpecFile",
			specFiles:    []string{"test/max.yaml"},
			spec:         spec,
			expectedSpec: expectedMaxSpec,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			specFiles = tc.specFiles
			spec, err := FromFile(tc.spec)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedSpec, spec)
			} else {
				assert.Empty(t, spec)
				assert.Contains(t, err.Error(), tc.expectedError)
			}
		})
	}
}

func TestSpecPrintHelp(t *testing.T) {
	s := new(Spec)
	err := s.PrintHelp()

	assert.NoError(t, err)
}
