package generate

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/changelog/markdown"
	"github.com/moorara/changelog/internal/git"
	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/internal/remote/github"
	"github.com/moorara/changelog/internal/remote/gitlab"
	"github.com/moorara/changelog/internal/spec"
	"github.com/moorara/changelog/pkg/log"
)

// Generator is the changelog generator.
type Generator struct {
	spec       spec.Spec
	logger     log.Logger
	gitRepo    git.Repo
	remoteRepo remote.Repo
	processor  changelog.Processor
}

// New creates a new changelog generator.
func New(s spec.Spec, logger log.Logger, gitRepo git.Repo) *Generator {
	var remoteRepo remote.Repo
	switch s.Repo.Platform {
	case spec.PlatformGitHub:
		remoteRepo = github.NewRepo(logger, s.Repo.Path, s.Repo.AccessToken)
	case spec.PlatformGitLab:
		remoteRepo = gitlab.NewRepo(logger, s.Repo.Path, s.Repo.AccessToken)
	}

	return &Generator{
		spec:       s,
		logger:     logger,
		gitRepo:    gitRepo,
		remoteRepo: remoteRepo,
		processor:  markdown.NewProcessor(logger, s.General.File),
	}
}

// resolveTags determines the new tags that should be added to the changelog.
// sortedTags are expected to be sorted from the most recent to the least recent.
// Similarly, chlog.Existing are expected to be sorted from the most recent to the least recent.
// The return value is the list of new tags for generating changelog for them.
func (g *Generator) resolveTags(sortedTags remote.Tags, chlog *changelog.Changelog) (remote.Tags, error) {
	g.logger.Debug("Resolving new tags for changelog ...")

	mapFunc := func(t remote.Tag) string {
		return t.Name
	}

	// Select those tags that are not in changelog
	newTags := sortedTags.Select(func(t remote.Tag) bool {
		for _, release := range chlog.Existing {
			if t.Name == release.TagName {
				return false
			}
		}
		return true
	})

	// Resolve the from tag
	if from := g.spec.Tags.From; from != "" {
		i := newTags.Index(from)
		if i == -1 {
			return nil, fmt.Errorf("from-tag can be one of %s", newTags.Map(mapFunc))
		}
		// new tags are also sorted from the most recent to the least recent
		newTags = newTags[:i+1]
	}

	// Resolve the to tag
	if to := g.spec.Tags.To; to != "" {
		i := newTags.Index(to)
		if i == -1 {
			return nil, fmt.Errorf("to-tag can be one of %s", newTags.Map(mapFunc))
		}
		// new tags are also sorted from the most recent to the least recent
		newTags = newTags[i:]
	}

	// Resolve the future tag
	// The future tag should be the most recent tag (at index zero) if any
	if future := g.spec.Tags.Future; future != "" {
		if _, ok := sortedTags.Find(future); ok {
			return nil, fmt.Errorf("future tag cannot be same as an existing tag: %s", future)
		}

		futureTag := g.remoteRepo.FutureTag(future)
		newTags = append(remote.Tags{futureTag}, newTags...)
	}

	g.logger.Infof("Resolved new tags for changelog: %s", newTags.Map(mapFunc))

	return newTags, nil
}

func (g *Generator) resolveCommitMap(ctx context.Context, branch remote.Branch, sortedTags remote.Tags) (commitMap, error) {
	commitMap := commitMap{}

	// Resolve which commits are in the branch
	branchCommits, err := g.remoteRepo.FetchParentCommits(ctx, branch.Commit.Hash)
	if err != nil {
		return nil, err
	}

	for _, c := range branchCommits {
		if rev, ok := commitMap[c.Hash]; ok {
			rev.Branch = branch.Name
		} else {
			commitMap[c.Hash] = &revisions{
				Branch: branch.Name,
			}
		}
	}

	// Resolve which commits are in the each tag
	// sortedTags are sorted from the most recent to the least recent
	for _, tag := range sortedTags {
		// The first tag can be a future tag without a commit
		if !tag.Commit.IsZero() {
			tagCommits, err := g.remoteRepo.FetchParentCommits(ctx, tag.Commit.Hash)
			if err != nil {
				return nil, err
			}

			for _, c := range tagCommits {
				if rev, ok := commitMap[c.Hash]; ok {
					rev.Tags = append(rev.Tags, tag.Name)
				} else {
					commitMap[c.Hash] = &revisions{
						Tags: []string{tag.Name},
					}
				}
			}
		}
	}

	return commitMap, nil
}

func (g *Generator) resolveReleases(ctx context.Context, sortedTags remote.Tags, baseRev string, im issueMap, cm mergeMap) []changelog.Release {
	releases := []changelog.Release{}

	issueGroups := g.spec.Issues.Groups()
	mergeGroups := g.spec.Merges.Groups()

	for i, tag := range sortedTags {
		var compareURL string
		if j := i + 1; j < len(sortedTags) {
			compareURL = g.remoteRepo.CompareURL(sortedTags[j].Name, tag.Name)
		} else {
			compareURL = g.remoteRepo.CompareURL(baseRev, tag.Name)
		}

		// Every tag represents a new release
		release := changelog.Release{
			TagName:    tag.Name,
			TagURL:     tag.WebURL,
			TagTime:    tag.Time,
			CompareURL: compareURL,
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

// Generate generates changelogs for a Git repository.
func (g *Generator) Generate(ctx context.Context) error {
	// Parse the existing changelog if any
	chlog, err := g.processor.Parse(changelog.ParseOptions{})
	if err != nil {
		return err
	}

	// ==============================> FETCH RELEASE BRANCH <==============================

	var branch remote.Branch

	if g.spec.Merges.Branch == "" {
		branch, err = g.remoteRepo.FetchDefaultBranch(ctx)
	} else {
		branch, err = g.remoteRepo.FetchBranch(ctx, g.spec.Merges.Branch)
	}

	if err != nil {
		return err
	}

	// ==============================> FETCH AND FILTER TAGS <==============================

	tags, err := g.remoteRepo.FetchTags(ctx)
	if err != nil {
		return err
	}

	g.logger.Info("Sorting and filtering git tags ...")

	sortedTags := tags.Sort()
	sortedTags = sortedTags.Exclude(g.spec.Tags.Exclude...)

	if g.spec.Tags.ExcludeRegex != "" {
		re, err := regexp.CompilePOSIX(g.spec.Tags.ExcludeRegex)
		if err != nil {
			return err
		}
		sortedTags = sortedTags.ExcludeRegex(re)
	}

	newTags, err := g.resolveTags(sortedTags, chlog)
	if err != nil {
		return err
	}

	if len(newTags) == 0 {
		g.logger.Info("Changelog is up-to-date (no new tag or a future tag)")
		return nil
	}

	// ==============================> FETCH COMMITS FOR BRANCH AND TAGS <==============================

	// Fetch the first commit in case it is needed for comparing the first tag against it
	firstCommit, err := g.remoteRepo.FetchFirstCommit(ctx)
	if err != nil {
		return err
	}

	// Construct a map of commit hashes to branch and tags names
	commitMap, err := g.resolveCommitMap(ctx, branch, newTags)
	if err != nil {
		return err
	}

	// ==============================> FETCH & ORGANIZE ISSUES AND MERGES <==============================

	// Fetch issues and merges since the last tag on changelog
	var since time.Time
	if len(chlog.Existing) > 0 {
		since = chlog.Existing[0].TagTime
	}

	issues, merges, err := g.remoteRepo.FetchIssuesAndMerges(ctx, since)
	if err != nil {
		return err
	}

	sortedIssues, sortedMerges := filterByLabels(issues, merges, g.spec)
	g.logger.Infof("Filtered issues (%d) and pull/merge requests (%d) by labels", len(sortedIssues), len(sortedMerges))

	issueMap := resolveIssueMap(sortedIssues, newTags)
	mergeMap := resolveMergeMap(sortedMerges, newTags, commitMap)
	g.logger.Info("Partitioned issues and pull/merge requests by tag")

	chlog.New = g.resolveReleases(ctx, newTags, firstCommit.Hash, issueMap, mergeMap)
	g.logger.Info("Grouped issues and pull/merge requests by labels")

	// ==============================> UPDATE THE CHANGELOG <==============================

	content, err := g.processor.Render(chlog)
	if err != nil {
		return err
	}

	fmt.Print(content)

	return nil
}
