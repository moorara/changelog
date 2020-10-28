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

func (g *Generator) resolveTags(tags git.Tags, chlog *changelog.Changelog) (git.Tag, git.Tag, git.Tag, error) {
	g.logger.Debug("Resolving from and to tags ...")

	var ok bool
	var zero git.Tag
	var lastGitTag, lastChangelogTag git.Tag
	var fromTag, toTag, futureTag git.Tag

	// Resolve the last git tag
	if len(tags) == 0 {
		lastGitTag = git.Tag{} // Denotes the case where there is no git
	} else {
		lastGitTag = tags[0] // The most recent tag
	}

	// Resolve the last tag on changelog
	if len(chlog.Releases) == 0 {
		lastChangelogTag = git.Tag{} // Denotes the case where the changelog is empty
	} else {
		if lastChangelogTag, ok = tags.Find(chlog.Releases[0].GitTag); !ok {
			return zero, zero, zero, fmt.Errorf("changelog tag not found: %s", chlog.Releases[0].GitTag)
		}
	}

	// Resolve the from tag
	if g.spec.Tags.From == "" {
		fromTag = lastChangelogTag
	} else {
		if fromTag, ok = tags.Find(g.spec.Tags.From); !ok {
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
		if toTag, ok = tags.Find(g.spec.Tags.To); !ok {
			return zero, zero, zero, fmt.Errorf("to-tag not found: %s", g.spec.Tags.To)
		}

		if toTag.Before(fromTag) || toTag.Equal(fromTag) {
			return zero, zero, zero, errors.New("to-tag should be after the from-tag")
		}
	}

	// Resolve the future tag
	if g.spec.Tags.Future != "" {
		futureTag = git.Tag{
			Type: git.Void,
			Name: g.spec.Tags.Future,
		}
	}

	g.logger.Infof("From tag resolved: %s", fromTag.Name)
	g.logger.Infof("To tag resolved: %s", toTag.Name)
	g.logger.Infof("Future tag resolved: %s", futureTag.Name)

	return fromTag, toTag, futureTag, nil
}

func (g *Generator) filterByLabels(issues, merges remote.Changes) (remote.Changes, remote.Changes) {
	g.logger.Debug("Filtering issues pull/merge requests by labels ...")

	switch g.spec.Issues.Selection {
	case spec.SelectionNone:
		issues = remote.Changes{}

	case spec.SelectionAll:
		// All issues without labels or if any labels, they should include one of the given labels
		if len(g.spec.Issues.IncludeLabels) > 0 {
			issues = issues.Select(func(c remote.Change) bool {
				return len(c.Labels) == 0 || c.Labels.Any(g.spec.Issues.IncludeLabels...)
			})
		}
		if len(g.spec.Issues.ExcludeLabels) > 0 {
			issues = issues.Select(func(c remote.Change) bool {
				return len(c.Labels) == 0 || !c.Labels.Any(g.spec.Issues.ExcludeLabels...)
			})
		}

	case spec.SelectionLabeled:
		// Select only labeled issues
		issues = issues.Select(func(c remote.Change) bool {
			return len(c.Labels) > 0
		})
		if len(g.spec.Issues.IncludeLabels) > 0 {
			issues = issues.Select(func(c remote.Change) bool {
				return c.Labels.Any(g.spec.Issues.IncludeLabels...)
			})
		}
		if len(g.spec.Issues.ExcludeLabels) > 0 {
			issues = issues.Select(func(c remote.Change) bool {
				return !c.Labels.Any(g.spec.Issues.ExcludeLabels...)
			})
		}
	}

	g.logger.Infof("Filtered issues by labels: %d", len(issues))

	switch g.spec.Merges.Selection {
	case spec.SelectionNone:
		merges = remote.Changes{}

	case spec.SelectionAll:
		// All merges without labels or if any labels, they should include one of the given labels
		if len(g.spec.Merges.IncludeLabels) > 0 {
			merges = merges.Select(func(c remote.Change) bool {
				return len(c.Labels) == 0 || c.Labels.Any(g.spec.Merges.IncludeLabels...)
			})
		}
		if len(g.spec.Merges.ExcludeLabels) > 0 {
			merges = merges.Select(func(c remote.Change) bool {
				return len(c.Labels) == 0 || !c.Labels.Any(g.spec.Merges.ExcludeLabels...)
			})
		}

	case spec.SelectionLabeled:
		// Select only labeled merges
		merges = merges.Select(func(c remote.Change) bool {
			return len(c.Labels) > 0
		})
		if len(g.spec.Merges.IncludeLabels) > 0 {
			merges = merges.Select(func(c remote.Change) bool {
				return c.Labels.Any(g.spec.Merges.IncludeLabels...)
			})
		}
		if len(g.spec.Merges.ExcludeLabels) > 0 {
			merges = merges.Select(func(c remote.Change) bool {
				return !c.Labels.Any(g.spec.Merges.ExcludeLabels...)
			})
		}
	}

	g.logger.Infof("Filtered pull/merge requests by labels: %d", len(merges))

	return issues, merges
}

// Generate generates changelogs for a Git repository.
func (g *Generator) Generate(ctx context.Context) error {
	// PARSE THE EXISTING CHANGELOG IF ANY

	chlog, err := g.processor.Parse(changelog.ParseOptions{})
	if err != nil {
		return err
	}

	// ==============================> GET AND FILTER GIT TAGS <==============================

	tags, err := g.gitRepo.Tags()
	if err != nil {
		return err
	}

	g.logger.Info("Sorting and filtering git tags ...")

	tags = tags.Sort()
	tags = tags.Exclude(g.spec.Tags.Exclude...)

	if g.spec.Tags.ExcludeRegex != "" {
		re, err := regexp.CompilePOSIX(g.spec.Tags.ExcludeRegex)
		if err != nil {
			return err
		}
		tags = tags.ExcludeRegex(re)
	}

	fromTag, toTag, futureTag, err := g.resolveTags(tags, chlog)
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

	if (fromTag == toTag || toTag.IsZero()) && futureTag.IsZero() {
		g.logger.Debug("No new tag or a future tag is detected.")
		g.logger.Info("Changelog is up-to-date")
		return nil
	}

	// ==============================> FETCH ISSUES AND MERGES <==============================

	since := fromTag.Commit.Committer.Time
	issues, merges, err := g.remoteRepo.FetchIssuesAndMerges(ctx, since)
	if err != nil {
		return err
	}

	issues, merges = g.filterByLabels(issues, merges)

	// ==============================> UPDATE THE CHANGELOG <==============================

	_, err = g.processor.Render(chlog)
	if err != nil {
		return err
	}

	return nil
}
