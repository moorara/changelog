package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefault(t *testing.T) {
	spec := Default("github.com", "octocat/Hello-World")

	assert.NotNil(t, spec)
	assert.Equal(t, PlatformGitHub, spec.Repo.Platform)
	assert.Equal(t, "octocat/Hello-World", spec.Repo.Path)
	assert.Equal(t, "", spec.Repo.AccessToken)
	assert.Equal(t, "CHANGELOG.md", spec.General.File)
	assert.Equal(t, "HISTORY.md", spec.General.Base)
	assert.Equal(t, false, spec.General.Print)
	assert.Equal(t, false, spec.General.Verbose)
	assert.Equal(t, "", spec.Tags.From)
	assert.Equal(t, "", spec.Tags.To)
	assert.Equal(t, "", spec.Tags.Future)
	assert.Equal(t, []string{}, spec.Tags.Exclude)
	assert.Equal(t, "", spec.Tags.ExcludeRegex)
	assert.Equal(t, SelectionAll, spec.Issues.Selection)
	assert.Nil(t, spec.Issues.IncludeLabels)
	assert.Equal(t, []string{"duplicate", "invalid", "question", "wontfix"}, spec.Issues.ExcludeLabels)
	assert.True(t, spec.Issues.Grouping)
	assert.Equal(t, []string{"summary", "release-summary"}, spec.Issues.SummaryLabels)
	assert.Equal(t, []string{"removed"}, spec.Issues.RemovedLabels)
	assert.Equal(t, []string{"breaking"}, spec.Issues.BreakingLabels)
	assert.Equal(t, []string{"deprecated"}, spec.Issues.DeprecatedLabels)
	assert.Equal(t, []string{"feature"}, spec.Issues.FeatureLabels)
	assert.Equal(t, []string{"enhancement"}, spec.Issues.EnhancementLabels)
	assert.Equal(t, []string{"bug"}, spec.Issues.BugLabels)
	assert.Equal(t, []string{"security"}, spec.Issues.SecurityLabels)
	assert.Equal(t, SelectionAll, spec.Merges.Selection)
	assert.Equal(t, "master", spec.Merges.Branch)
	assert.Nil(t, spec.Merges.IncludeLabels)
	assert.Nil(t, spec.Merges.ExcludeLabels)
	assert.False(t, spec.Merges.Grouping)
	assert.Equal(t, []string{}, spec.Merges.SummaryLabels)
	assert.Equal(t, []string{}, spec.Merges.RemovedLabels)
	assert.Equal(t, []string{}, spec.Merges.BreakingLabels)
	assert.Equal(t, []string{}, spec.Merges.DeprecatedLabels)
	assert.Equal(t, []string{}, spec.Merges.FeatureLabels)
	assert.Equal(t, []string{}, spec.Merges.EnhancementLabels)
	assert.Equal(t, []string{}, spec.Merges.BugLabels)
	assert.Equal(t, []string{}, spec.Merges.SecurityLabels)
	assert.Equal(t, GroupByLabel, spec.Format.GroupBy)
	assert.Equal(t, "", spec.Format.ReleaseURL)
}

func TestFromFile(t *testing.T) {
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
			spec:         Default("github.com", "octocat/Hello-World"),
			expectedSpec: Default("github.com", "octocat/Hello-World"),
		},
		{
			name:          "EmptySpecFile",
			specFiles:     []string{"test/empty.yaml"},
			spec:          Default("github.com", "octocat/Hello-World"),
			expectedError: "EOF",
		},
		{
			name:          "InvalidSpecFile",
			specFiles:     []string{"test/invalid.yaml"},
			spec:          Default("github.com", "octocat/Hello-World"),
			expectedError: "yaml: unmarshal errors",
		},
		{
			name:      "MinimumSpecFile",
			specFiles: []string{"test/min.yaml"},
			spec:      Default("github.com", "octocat/Hello-World"),
			expectedSpec: Spec{
				Help:    false,
				Version: false,
				Repo: Repo{
					Platform:    Platform("github.com"),
					Path:        "octocat/Hello-World",
					AccessToken: "",
				},
				General: General{
					File:    "CHANGELOG.md",
					Base:    "HISTORY.md",
					Print:   true,
					Verbose: false,
				},
				Tags: Tags{
					From:         "",
					To:           "",
					Future:       "",
					Exclude:      []string{},
					ExcludeRegex: "",
				},
				Issues: Issues{
					Selection:         SelectionLabeled,
					IncludeLabels:     nil,
					ExcludeLabels:     []string{"duplicate", "invalid", "question", "wontfix"},
					Grouping:          true,
					SummaryLabels:     []string{"summary", "release-summary"},
					RemovedLabels:     []string{"removed"},
					BreakingLabels:    []string{"breaking"},
					DeprecatedLabels:  []string{"deprecated"},
					FeatureLabels:     []string{"feature"},
					EnhancementLabels: []string{"enhancement"},
					BugLabels:         []string{"bug"},
					SecurityLabels:    []string{"security"},
				},
				Merges: Merges{
					Selection:         SelectionAll,
					Branch:            "production",
					IncludeLabels:     nil,
					ExcludeLabels:     nil,
					Grouping:          false,
					SummaryLabels:     []string{},
					RemovedLabels:     []string{},
					BreakingLabels:    []string{},
					DeprecatedLabels:  []string{},
					FeatureLabels:     []string{},
					EnhancementLabels: []string{},
					BugLabels:         []string{},
					SecurityLabels:    []string{},
				},
				Format: Format{
					GroupBy:    GroupByMilestone,
					ReleaseURL: "",
				},
			},
		},
		{
			name:      "MaximumSpecFile",
			specFiles: []string{"test/max.yaml"},
			spec:      Default("github.com", "octocat/Hello-World"),
			expectedSpec: Spec{
				Help:    false,
				Version: false,
				Repo: Repo{
					Platform:    Platform("github.com"),
					Path:        "octocat/Hello-World",
					AccessToken: "",
				},
				General: General{
					File:    "RELEASE-NOTES.md",
					Base:    "SUMMARY-NOTES.md",
					Print:   true,
					Verbose: true,
				},
				Tags: Tags{
					From:         "v0.1.0",
					To:           "v0.2.0",
					Future:       "v0.3.0",
					Exclude:      []string{"staging"},
					ExcludeRegex: `(.*)-(alpha|beta)`,
				},
				Issues: Issues{
					Selection:         SelectionLabeled,
					IncludeLabels:     []string{"breaking", "bug", "defect", "deprecated", "dropped", "enhancement", "feature", "highlight", "improvement", "incompatible", "new", "not-supported", "privacy", "removed", "security", "summary"},
					ExcludeLabels:     []string{"documentation", "duplicate", "invalid", "question", "wontfix"},
					Grouping:          true,
					SummaryLabels:     []string{"summary", "highlight"},
					RemovedLabels:     []string{"removed", "dropped"},
					BreakingLabels:    []string{"breaking", "incompatible"},
					DeprecatedLabels:  []string{"deprecated", "not-supported"},
					FeatureLabels:     []string{"feature", "new"},
					EnhancementLabels: []string{"enhancement", "improvement"},
					BugLabels:         []string{"bug", "defect"},
					SecurityLabels:    []string{"security", "privacy"},
				},
				Merges: Merges{
					Selection:         SelectionLabeled,
					Branch:            "production",
					IncludeLabels:     []string{"breaking", "bug", "defect", "deprecated", "dropped", "enhancement", "feature", "highlight", "improvement", "incompatible", "new", "not-supported", "privacy", "removed", "security", "summary"},
					ExcludeLabels:     []string{"documentation", "duplicate", "invalid", "question", "wontfix"},
					Grouping:          true,
					SummaryLabels:     []string{"summary", "highlight"},
					RemovedLabels:     []string{"removed", "dropped"},
					BreakingLabels:    []string{"breaking", "incompatible"},
					DeprecatedLabels:  []string{"deprecated", "not-supported"},
					FeatureLabels:     []string{"feature", "new"},
					EnhancementLabels: []string{"enhancement", "improvement"},
					BugLabels:         []string{"bug", "defect"},
					SecurityLabels:    []string{"security", "privacy"},
				},
				Format: Format{
					GroupBy:    GroupByMilestone,
					ReleaseURL: "https://storage.artifactory.com/project/releases/{tag}",
				},
			},
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

func TestSpec_PrintHelp(t *testing.T) {
	s := new(Spec)
	err := s.PrintHelp()

	assert.NoError(t, err)
}

func TestSpec_String(t *testing.T) {
	s := new(Spec)
	str := s.String()

	assert.NotEmpty(t, str)
}
