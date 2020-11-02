package generate

import (
	"context"
	"errors"
	"fmt"
	"regexp"

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

func (g *Generator) resolveTags(chlog *changelog.Changelog, sortedTags remote.Tags) (remote.Tag, remote.Tag, remote.Tag, error) {
	g.logger.Debug("Resolving from and to tags ...")

	var ok bool
	var zero remote.Tag
	var lastGitTag, lastChangelogTag remote.Tag
	var fromTag, toTag, futureTag remote.Tag

	// Resolve the last git tag
	if len(sortedTags) == 0 {
		lastGitTag = remote.Tag{} // Denotes the case where there is no git
	} else {
		lastGitTag = sortedTags[0] // The most recent tag
	}

	// Resolve the last tag on changelog
	if len(chlog.Releases) == 0 {
		lastChangelogTag = remote.Tag{} // Denotes the case where the changelog is empty
	} else {
		if lastChangelogTag, ok = sortedTags.Find(chlog.Releases[0].TagName); !ok {
			return zero, zero, zero, fmt.Errorf("changelog tag not found: %s", chlog.Releases[0].TagName)
		}
	}

	// Resolve the from tag
	if g.spec.Tags.From == "" {
		fromTag = lastChangelogTag
	} else {
		if fromTag, ok = sortedTags.Find(g.spec.Tags.From); !ok {
			return zero, zero, zero, fmt.Errorf("from-tag not found: %s", g.spec.Tags.From)
		}

		if fromTag.Before(lastChangelogTag) {
			g.logger.Debugf("From tag updated to the last tag on changelog: %s", lastChangelogTag.Name)
			fromTag = lastChangelogTag
		}
	}

	// Resolve the to tag
	if g.spec.Tags.To == "" {
		toTag = lastGitTag
	} else {
		if toTag, ok = sortedTags.Find(g.spec.Tags.To); !ok {
			return zero, zero, zero, fmt.Errorf("to-tag not found: %s", g.spec.Tags.To)
		}

		if toTag.Before(fromTag) || toTag.Equal(fromTag) {
			return zero, zero, zero, errors.New("to-tag should be after the from-tag")
		}
	}

	// Resolve the future tag
	if g.spec.Tags.Future != "" {
		futureTag = remote.Tag{
			Name: g.spec.Tags.Future,
		}
	}

	g.logger.Infof("From tag resolved: %s", fromTag.Name)
	g.logger.Infof("To tag resolved: %s", toTag.Name)
	g.logger.Infof("Future tag resolved: %s", futureTag.Name)

	return fromTag, toTag, futureTag, nil
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

	return commitMap, nil
}

// Generate generates changelogs for a Git repository.
func (g *Generator) Generate(ctx context.Context) error {
	// Parse the existing changelog if any
	chlog, err := g.processor.Parse(changelog.ParseOptions{})
	if err != nil {
		return err
	}

	// ==============================> FETCH BRANCH <==============================

	var branch remote.Branch

	if g.spec.Merges.Branch == "" {
		if branch, err = g.remoteRepo.FetchDefaultBranch(ctx); err != nil {
			return err
		}
	} else {
		if branch, err = g.remoteRepo.FetchBranch(ctx, g.spec.Merges.Branch); err != nil {
			return err
		}
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

	fromTag, toTag, futureTag, err := g.resolveTags(chlog, sortedTags)
	if err != nil {
		return err
	}

	/*
		Here is the logic for when to fetch changes for released and unreleased sections:

			| Since |  To | Future || Changes | Released | Unreleased |
			|-------|-----|--------||---------|----------|------------|
			|   0   |  0  |   0    ||    0    |     0    |      0     |
			|   0   |  0  |   1    ||    1    |     0    |      1     |
			|   0   |  1  |   0    ||    1    |     1    |      0     |
			|   0   |  1  |   1    ||    1    |     1    |      1     |
			|   1   |  0  |   0    ||    0    |     0    |      0     |
			|   1   |  0  |   1    ||    1    |     0    |      1     |
			|   1   |  1  |   0    ||    1    |     1    |      0     |
			|   1   |  1  |   1    ||    1    |     1    |      1     |
	*/

	if (fromTag.Equal(toTag) || toTag.IsZero()) && futureTag.IsZero() {
		g.logger.Debug("No new tag or a future tag is detected.")
		g.logger.Info("Changelog is up-to-date")
		return nil
	}

	// Resolving all tags eligible for changelog
	// Tags are sorted from the most recent to the least recent

	var candidateTags remote.Tags

	if fromTag.IsZero() && toTag.IsZero() {
		candidateTags = remote.Tags{}
	} else if fromTag.IsZero() {
		i := sortedTags.Index(toTag.Name)
		candidateTags = sortedTags[i:]
	} else if toTag.IsZero() {
		j := sortedTags.Index(fromTag.Name)
		candidateTags = sortedTags[:j+1]
	} else {
		i := sortedTags.Index(toTag.Name)
		j := sortedTags.Index(fromTag.Name)
		candidateTags = sortedTags[i : j+1]
	}

	g.logger.Debugf("New tags for generating changelog: %v", candidateTags.Map(func(t remote.Tag) string {
		return t.Name
	}))

	if !futureTag.IsZero() {
		g.logger.Debugf("Future tag for unreleased changes: %s", futureTag.Name)
	}

	// ==============================> FETCH & ORGANIZE ISSUES AND MERGES <==============================

	since := fromTag.Time
	issues, merges, err := g.remoteRepo.FetchIssuesAndMerges(ctx, since)
	if err != nil {
		return err
	}

	sortedIssues, sortedMerges := filterByLabels(issues, merges, g.spec)
	g.logger.Infof("Filtered issues (%d) and pull/merge requests (%d) by labels", len(sortedIssues), len(sortedMerges))

	// Construct a map of commit hashes to branch and tags names
	commitMap, err := g.resolveCommitMap(ctx, branch, candidateTags)
	if err != nil {
		return err
	}

	issueMap := resolveIssueMap(sortedIssues, candidateTags, futureTag)
	mergeMap := resolveMergeMap(sortedMerges, commitMap, futureTag)
	g.logger.Info("Partitioned issues and pull/merge requests by tag")

	chlog.New = resolveReleases(candidateTags, futureTag, issueMap, mergeMap, g.spec)
	g.logger.Info("Grouped issues and pull/merge requests by labels")

	// ==============================> UPDATE THE CHANGELOG <==============================

	_, err = g.processor.Render(chlog)
	if err != nil {
		return err
	}

	return nil
}
