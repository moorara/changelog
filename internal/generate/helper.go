package generate

import (
	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/internal/spec"
)

// revisions refers to a branch name and list of tags sorted from the most recent to the least recent.
type revisions struct {
	Branch string
	Tags   []string
}

// commitMap is a map of commit hashes to revisions (branch name and tags).
// It allows us to know the branch and tag names for each commit.
type commitMap map[string]*revisions

// issueMap is a map of tag names to issues.
// It allows us to look up all issues for a tatg.
type issueMap map[string]remote.Issues

// mergeMap is a map of tag names to issues.
// It allows us to look up all merges for a tatg.
type mergeMap map[string]remote.Merges

func filterByLabels(issues remote.Issues, merges remote.Merges, s spec.Spec) (remote.Issues, remote.Merges) {
	switch s.Issues.Selection {
	case spec.SelectionNone:
		issues = remote.Issues{}

	case spec.SelectionAll:
		// All issues without labels or if any labels, they should include one of the given labels
		if len(s.Issues.IncludeLabels) > 0 {
			issues = issues.Select(func(c remote.Issue) bool {
				return len(c.Labels) == 0 || c.Labels.Any(s.Issues.IncludeLabels...)
			})
		}
		if len(s.Issues.ExcludeLabels) > 0 {
			issues = issues.Select(func(c remote.Issue) bool {
				return len(c.Labels) == 0 || !c.Labels.Any(s.Issues.ExcludeLabels...)
			})
		}

	case spec.SelectionLabeled:
		// Select only labeled issues
		issues = issues.Select(func(c remote.Issue) bool {
			return len(c.Labels) > 0
		})
		if len(s.Issues.IncludeLabels) > 0 {
			issues = issues.Select(func(c remote.Issue) bool {
				return c.Labels.Any(s.Issues.IncludeLabels...)
			})
		}
		if len(s.Issues.ExcludeLabels) > 0 {
			issues = issues.Select(func(c remote.Issue) bool {
				return !c.Labels.Any(s.Issues.ExcludeLabels...)
			})
		}
	}

	switch s.Merges.Selection {
	case spec.SelectionNone:
		merges = remote.Merges{}

	case spec.SelectionAll:
		// All merges without labels or if any labels, they should include one of the given labels
		if len(s.Merges.IncludeLabels) > 0 {
			merges = merges.Select(func(c remote.Merge) bool {
				return len(c.Labels) == 0 || c.Labels.Any(s.Merges.IncludeLabels...)
			})
		}
		if len(s.Merges.ExcludeLabels) > 0 {
			merges = merges.Select(func(c remote.Merge) bool {
				return len(c.Labels) == 0 || !c.Labels.Any(s.Merges.ExcludeLabels...)
			})
		}

	case spec.SelectionLabeled:
		// Select only labeled merges
		merges = merges.Select(func(c remote.Merge) bool {
			return len(c.Labels) > 0
		})
		if len(s.Merges.IncludeLabels) > 0 {
			merges = merges.Select(func(c remote.Merge) bool {
				return c.Labels.Any(s.Merges.IncludeLabels...)
			})
		}
		if len(s.Merges.ExcludeLabels) > 0 {
			merges = merges.Select(func(c remote.Merge) bool {
				return !c.Labels.Any(s.Merges.ExcludeLabels...)
			})
		}
	}

	return issues, merges
}

func resolveIssueMap(issues remote.Issues, sortedTags remote.Tags, futureTag remote.Tag) issueMap {
	im := issueMap{}

	for _, i := range issues {
		// sortedTags are sorted from the most recent to the least recent
		tag, ok := sortedTags.Last(func(tag remote.Tag) bool {
			// If issue was closed before or at the time of tag
			return i.Time.Before(tag.Time) || i.Time.Equal(tag.Time)
		})

		var tagName string
		if ok {
			tagName = tag.Name
		} else {
			tagName = futureTag.Name
		}

		im[tagName] = append(im[tagName], i)
	}

	return im
}

func resolveMergeMap(merges remote.Merges, cm commitMap, futureTag remote.Tag) mergeMap {
	mm := mergeMap{}

	for _, m := range merges {
		if rev, ok := cm[m.Commit.Hash]; ok {
			var tagName string
			if len(rev.Tags) == 0 {
				tagName = futureTag.Name
			} else {
				tagName = rev.Tags[len(rev.Tags)-1]
			}

			mm[tagName] = append(mm[tagName], m)
		}
	}

	return mm
}

func toIssueGroup(title string, issues remote.Issues) changelog.IssueGroup {
	issueGroup := changelog.IssueGroup{
		Title: title,
	}

	for _, i := range issues {
		issueGroup.Issues = append(issueGroup.Issues, changelog.Issue{
			Number: i.Number,
			Title:  i.Title,
			URL:    i.WebURL,
			OpenedBy: changelog.User{
				Name:     i.Author.Name,
				Username: i.Author.Username,
				URL:      i.Author.WebURL,
			},
			ClosedBy: changelog.User{
				Name:     i.Closer.Name,
				Username: i.Closer.Username,
				URL:      i.Closer.WebURL,
			},
		})
	}

	return issueGroup
}

func toMergeGroup(title string, merges remote.Merges) changelog.MergeGroup {
	mergeGroup := changelog.MergeGroup{
		Title: title,
	}

	for _, m := range merges {
		mergeGroup.Merges = append(mergeGroup.Merges, changelog.Merge{
			Number: m.Number,
			Title:  m.Title,
			URL:    m.WebURL,
			OpenedBy: changelog.User{
				Name:     m.Author.Name,
				Username: m.Author.Username,
				URL:      m.Author.WebURL,
			},
			MergedBy: changelog.User{
				Name:     m.Merger.Name,
				Username: m.Merger.Username,
				URL:      m.Merger.WebURL,
			},
		})
	}

	return mergeGroup
}

func resolveReleases(sortedTags remote.Tags, futureTag remote.Tag, im issueMap, cm mergeMap, s spec.Spec) []changelog.Release {
	releases := []changelog.Release{}

	var tags remote.Tags
	if futureTag.IsZero() {
		tags = sortedTags
	} else {
		tags = append(remote.Tags{futureTag}, sortedTags...)
	}

	issueGroups := s.Issues.Groups()
	mergeGroups := s.Merges.Groups()

	for _, tag := range tags {
		// Every tag represents a new release
		release := changelog.Release{
			TagName: tag.Name,
			TagURL:  tag.WebURL,
			TagTime: tag.Time,
			// TODO: CompareURL: tag.CompareURL,
		}

		// Group issues for the current tag
		if issues, ok := im[tag.Name]; ok {
			leftIssues := issues

			for _, group := range issueGroups {
				f := func(i remote.Issue) bool {
					return i.Labels.Any(group.Labels...)
				}

				selected := issues.Select(f)
				leftIssues.Remove(f)

				if len(selected) > 0 {
					issueGroup := toIssueGroup(group.Title, selected)
					release.IssueGroups = append(release.IssueGroups, issueGroup)
				}
			}

			if len(leftIssues) > 0 {
				issueGroup := toIssueGroup("Closed Issues", leftIssues)
				release.IssueGroups = append(release.IssueGroups, issueGroup)
			}
		}

		// Group merges for the current tag
		if merges, ok := cm[tag.Name]; ok {
			leftMerges := merges

			for _, group := range mergeGroups {
				f := func(m remote.Merge) bool {
					return m.Labels.Any(group.Labels...)
				}

				selected := merges.Select(f)
				leftMerges.Remove(f)

				if len(selected) > 0 {
					mergeGroup := toMergeGroup(group.Title, selected)
					release.MergeGroups = append(release.MergeGroups, mergeGroup)
				}
			}

			if len(leftMerges) > 0 {
				mergeGroup := toMergeGroup("Merged Changes", leftMerges)
				release.MergeGroups = append(release.MergeGroups, mergeGroup)
			}
		}

		releases = append(releases, release)
	}

	return releases
}
