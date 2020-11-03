package generate

import (
	"testing"
	"time"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/internal/spec"
	"github.com/stretchr/testify/assert"
)

func TestFilterByLabels(t *testing.T) {
	tests := []struct {
		name           string
		issues         remote.Issues
		merges         remote.Merges
		spec           spec.Spec
		expectedIssues remote.Issues
		expectedMerges remote.Merges
	}{
		{
			name:   "None",
			issues: remote.Issues{issue1, issue2},
			merges: remote.Merges{merge1, merge2},
			spec: spec.Spec{
				Issues: spec.Issues{
					Selection: spec.SelectionNone,
				},
				Merges: spec.Merges{
					Selection: spec.SelectionNone,
				},
			},
			expectedIssues: remote.Issues{},
			expectedMerges: remote.Merges{},
		},
		{
			name:   "AllWithIncludeLabels",
			issues: remote.Issues{issue1, issue2},
			merges: remote.Merges{merge1, merge2},
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
			expectedIssues: remote.Issues{issue1},
			expectedMerges: remote.Merges{merge1, merge2},
		},
		{
			name:   "AllWithExcludeLabels",
			issues: remote.Issues{issue1, issue2},
			merges: remote.Merges{merge1, merge2},
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
			expectedIssues: remote.Issues{issue1},
			expectedMerges: remote.Merges{merge2},
		},
		{
			name:   "LabeledWithIncludeLabels",
			issues: remote.Issues{issue1, issue2},
			merges: remote.Merges{merge1, merge2},
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
			expectedIssues: remote.Issues{issue1},
			expectedMerges: remote.Merges{merge1},
		},
		{
			name:   "LabeledWithExcludeLabels",
			issues: remote.Issues{issue1, issue2},
			merges: remote.Merges{merge1, merge2},
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
			expectedIssues: remote.Issues{issue1},
			expectedMerges: remote.Merges{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issues, merges := filterByLabels(tc.issues, tc.merges, tc.spec)

			assert.Equal(t, tc.expectedIssues, issues)
			assert.Equal(t, tc.expectedMerges, merges)
		})
	}
}

func TestResolveIssueMap(t *testing.T) {
	tests := []struct {
		name             string
		issues           remote.Issues
		sortedTags       remote.Tags
		futureTag        remote.Tag
		expectedIssueMap issueMap
	}{
		{
			name:       "OK",
			issues:     remote.Issues{issue1, issue2},
			sortedTags: remote.Tags{tag3, tag2, tag1},
			futureTag: remote.Tag{
				Name: "v0.1.4",
			},
			expectedIssueMap: issueMap{
				"v0.1.4": remote.Issues{issue2},
				"v0.1.3": remote.Issues{issue1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issueMap := resolveIssueMap(tc.issues, tc.sortedTags, tc.futureTag)

			assert.Equal(t, tc.expectedIssueMap, issueMap)
		})
	}
}

func TestResolveMergeMap(t *testing.T) {
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
		commitMap        commitMap
		futureTag        remote.Tag
		expectedMergeMap mergeMap
	}{
		{
			name:      "OK",
			merges:    remote.Merges{merge1, merge2},
			commitMap: cm,
			futureTag: remote.Tag{
				Name: "v0.1.4",
			},
			expectedMergeMap: mergeMap{
				"v0.1.4": remote.Merges{merge2},
				"v0.1.3": remote.Merges{merge1},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mergeMap := resolveMergeMap(tc.merges, tc.commitMap, tc.futureTag)

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

func TestResolveReleases(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name             string
		sortedTags       remote.Tags
		futureTag        remote.Tag
		issueMap         issueMap
		mergeMap         mergeMap
		spec             spec.Spec
		expectedReleases []changelog.Release
	}{
		{
			name:       "WithoutFutureTag",
			sortedTags: remote.Tags{tag3},
			futureTag:  remote.Tag{},
			issueMap: issueMap{
				"v0.1.3": remote.Issues{issue1},
			},
			mergeMap: mergeMap{
				"v0.1.3": remote.Merges{merge1},
			},
			spec: spec.Spec{
				Issues: spec.Issues{
					Grouping:  true,
					BugLabels: []string{"bug"},
				},
				Merges: spec.Merges{
					Grouping:          true,
					EnhancementLabels: []string{"enhancement"},
				},
			},
			expectedReleases: []changelog.Release{
				{
					TagName: "v0.1.3",
					TagURL:  "https://github.com/octocat/Hello-World/tree/v0.1.3",
					TagTime: t3,
					// TODO: CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.2...v0.1.3",
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
			name:       "WithFutureTag",
			sortedTags: remote.Tags{tag3},
			futureTag: remote.Tag{
				Name:   "v0.1.4",
				Time:   now,
				WebURL: "https://github.com/octocat/Hello-World/tree/v0.1.4",
				// TODO: CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.3...v0.1.4",
			},
			issueMap: issueMap{
				"v0.1.4": remote.Issues{issue2},
				"v0.1.3": remote.Issues{issue1},
			},
			mergeMap: mergeMap{
				"v0.1.4": remote.Merges{merge2},
				"v0.1.3": remote.Merges{merge1},
			},
			spec: spec.Spec{
				Issues: spec.Issues{
					Grouping:  true,
					BugLabels: []string{"bug"},
				},
				Merges: spec.Merges{
					Grouping:          true,
					EnhancementLabels: []string{"enhancement"},
				},
			},
			expectedReleases: []changelog.Release{
				{
					TagName: "v0.1.4",
					TagURL:  "https://github.com/octocat/Hello-World/tree/v0.1.4",
					TagTime: now,
					// TODO: CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.3...v0.1.4",
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
					TagName: "v0.1.3",
					TagURL:  "https://github.com/octocat/Hello-World/tree/v0.1.3",
					TagTime: t3,
					// TODO: CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.2...v0.1.3",
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
			releases := resolveReleases(tc.sortedTags, tc.futureTag, tc.issueMap, tc.mergeMap, tc.spec)

			assert.Equal(t, tc.expectedReleases, releases)
		})
	}
}
