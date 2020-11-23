package generate

import (
	"testing"
	"time"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/spec"
	"github.com/stretchr/testify/assert"
)

func TestFilterByLabels(t *testing.T) {
	tests := []struct {
		name           string
		s              spec.Spec
		issues         remote.Issues
		merges         remote.Merges
		expectedIssues remote.Issues
		expectedMerges remote.Merges
	}{
		{
			name: "None",
			s: spec.Spec{
				Issues: spec.Issues{
					Selection: spec.SelectionNone,
				},
				Merges: spec.Merges{
					Selection: spec.SelectionNone,
				},
			},
			issues:         remote.Issues{issue1, issue2},
			merges:         remote.Merges{merge1, merge2},
			expectedIssues: remote.Issues{},
			expectedMerges: remote.Merges{},
		},
		{
			name: "AllWithIncludeLabels",
			s: spec.Spec{
				Issues: spec.Issues{
					Selection:     spec.SelectionAll,
					IncludeLabels: []string{"bug"},
				},
				Merges: spec.Merges{
					Selection:     spec.SelectionAll,
					IncludeLabels: []string{"enhancement"},
				},
			},
			issues:         remote.Issues{issue1, issue2},
			merges:         remote.Merges{merge1, merge2},
			expectedIssues: remote.Issues{issue1},
			expectedMerges: remote.Merges{merge1, merge2},
		},
		{
			name: "AllWithExcludeLabels",
			s: spec.Spec{
				Issues: spec.Issues{
					Selection:     spec.SelectionAll,
					ExcludeLabels: []string{"invalid"},
				},
				Merges: spec.Merges{
					Selection:     spec.SelectionAll,
					ExcludeLabels: []string{"enhancement"},
				},
			},
			issues:         remote.Issues{issue1, issue2},
			merges:         remote.Merges{merge1, merge2},
			expectedIssues: remote.Issues{issue1},
			expectedMerges: remote.Merges{merge2},
		},
		{
			name: "LabeledWithIncludeLabels",
			s: spec.Spec{
				Issues: spec.Issues{
					Selection:     spec.SelectionLabeled,
					IncludeLabels: []string{"bug"},
				},
				Merges: spec.Merges{
					Selection:     spec.SelectionLabeled,
					IncludeLabels: []string{"enhancement"},
				},
			},
			issues:         remote.Issues{issue1, issue2},
			merges:         remote.Merges{merge1, merge2},
			expectedIssues: remote.Issues{issue1},
			expectedMerges: remote.Merges{merge1},
		},
		{
			name: "LabeledWithExcludeLabels",
			s: spec.Spec{
				Issues: spec.Issues{
					Selection:     spec.SelectionLabeled,
					ExcludeLabels: []string{"invalid"},
				},
				Merges: spec.Merges{
					Selection:     spec.SelectionLabeled,
					ExcludeLabels: []string{"enhancement"},
				},
			},
			issues:         remote.Issues{issue1, issue2},
			merges:         remote.Merges{merge1, merge2},
			expectedIssues: remote.Issues{issue1},
			expectedMerges: remote.Merges{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issues, merges := filterByLabels(tc.s, tc.issues, tc.merges)

			assert.Equal(t, tc.expectedIssues, issues)
			assert.Equal(t, tc.expectedMerges, merges)
		})
	}
}

func TestResolveIssueMap(t *testing.T) {
	futureTag := remote.Tag{
		Name: "v0.1.4",
		Time: time.Now(),
	}

	tests := []struct {
		name             string
		issues           remote.Issues
		sortedTags       remote.Tags
		expectedIssueMap issueMap
	}{
		{
			name:       "OK",
			issues:     remote.Issues{issue1, issue2},
			sortedTags: remote.Tags{futureTag, tag3, tag2, tag1},
			expectedIssueMap: issueMap{
				"v0.1.4": remote.Issues{issue2},
				"v0.1.3": remote.Issues{issue1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issueMap := resolveIssueMap(tc.issues, tc.sortedTags)

			assert.Equal(t, tc.expectedIssueMap, issueMap)
		})
	}
}

func TestResolveMergeMap(t *testing.T) {
	futureTag := remote.Tag{
		Name: "v0.1.4",
	}

	cm := commitMap{
		"20c5414eccaa147f2d6644de4ca36f35293fa43e": &revisions{
			Branch: "main",
		},
		"c414d1004154c6c324bd78c69d10ee101e676059": &revisions{
			Branch: "main",
			Tags:   []string{"v0.1.3"},
		},
		"0251a422d2038967eeaaaa5c8aa76c7067fdef05": &revisions{
			Branch: "main",
			Tags:   []string{"v0.1.3", "v0.1.2"},
		},
		"25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378": &revisions{
			Branch: "main",
			Tags:   []string{"v0.1.3", "v0.1.2", "v0.1.1"},
		},
	}

	tests := []struct {
		name             string
		merges           remote.Merges
		sortedTags       remote.Tags
		commitMap        commitMap
		expectedMergeMap mergeMap
	}{
		{
			name:       "OK",
			merges:     remote.Merges{merge1, merge2},
			sortedTags: remote.Tags{futureTag},
			commitMap:  cm,
			expectedMergeMap: mergeMap{
				"v0.1.4": remote.Merges{merge2},
				"v0.1.3": remote.Merges{merge1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mergeMap := resolveMergeMap(tc.merges, tc.sortedTags, tc.commitMap)

			assert.Equal(t, tc.expectedMergeMap, mergeMap)
		})
	}
}

func TestToIssueGroup(t *testing.T) {
	tests := []struct {
		name               string
		title              string
		issues             remote.Issues
		expectedIssueGroup changelog.IssueGroup
	}{
		{
			name:   "OK",
			title:  "Enhancements",
			issues: remote.Issues{issue1, issue2},
			expectedIssueGroup: changelog.IssueGroup{
				Title:  "Enhancements",
				Issues: []changelog.Issue{changelogIssue1, changelogIssue2},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issueGroup := toIssueGroup(tc.title, tc.issues)

			assert.Equal(t, tc.expectedIssueGroup, issueGroup)
		})
	}
}

func TestToMergeGroup(t *testing.T) {
	tests := []struct {
		name               string
		title              string
		merges             remote.Merges
		expectedMergeGroup changelog.MergeGroup
	}{
		{
			name:   "OK",
			title:  "Enhancements",
			merges: remote.Merges{merge1, merge2},
			expectedMergeGroup: changelog.MergeGroup{
				Title:  "Enhancements",
				Merges: []changelog.Merge{changelogMerge1, changelogMerge2},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mergeGroup := toMergeGroup(tc.title, tc.merges)

			assert.Equal(t, tc.expectedMergeGroup, mergeGroup)
		})
	}
}
