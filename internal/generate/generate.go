package generate

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/git"
	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/internal/spec"
)

// Generator is the changelog generator.
type Generator struct {
	spec       spec.Spec
	logger     *log.Logger
	gitRepo    *git.Repo
	remoteRepo remote.Repo
	processor  changelog.Processor
}

// New creates a new changelog generator.
func New(s spec.Spec, logger *log.Logger, gitRepo *git.Repo) *Generator {
	var remoteRepo remote.Repo
	switch s.Repo.Platform {
	case spec.PlatformGitHub:
		remoteRepo = remote.NewGitHubRepo(logger, s.Repo.Path, s.Repo.AccessToken)
	case spec.PlatformGitLab:
		remoteRepo = remote.NewGitLabRepo(logger, s.Repo.Path, s.Repo.AccessToken)
	}

	return &Generator{
		spec:       s,
		logger:     logger,
		gitRepo:    gitRepo,
		remoteRepo: remoteRepo,
		processor:  changelog.NewMarkdownProcessor(logger, s.General.File),
	}
}

// Generate generates changelogs for a Git repository.
func (g *Generator) Generate() error {
	// PARSE THE EXISTING CHANGELOG IF ANY

	chlog, err := g.processor.Parse(changelog.ParseOptions{})
	if err != nil {
		return err
	}

	// GET AND FILTER GIT TAGS

	tags, err := g.gitRepo.Tags()
	if err != nil {
		return err
	}

	tags = tags.Sort()
	tags = tags.Exclude(g.spec.Release.ExcludeTags...)

	if g.spec.Release.ExcludeTagsRegex != "" {
		re, err := regexp.CompilePOSIX(g.spec.Release.ExcludeTagsRegex)
		if err != nil {
			return err
		}
		tags = tags.ExcludeRegex(re)
	}

	fromTag, toTag, err := g.resolveTags(tags, chlog)
	if err != nil {
		return err
	}

	// TODO: Remove!
	fmt.Printf("FromTag: %s\nToTag: %s\n", fromTag, toTag)
	if g.spec.Release.FutureTag != "" {
		fmt.Printf("FutureTag: %s\n", g.spec.Release.FutureTag)
	}
	fmt.Println()

	// FETCH ISSUES AND MERGES

	issues, merges, err := g.remoteRepo.FetchClosedIssuesAndMerges()
	if err != nil {
		return err
	}

	// TODO: Remove!
	fmt.Printf("Issues: %s\nMerges: %s\n\n", issues, merges)

	// UPDATE THE CHANGELOG

	content, err := g.processor.Render(chlog)
	if err != nil {
		return err
	}

	// TODO: Remove!
	fmt.Println(content)

	return nil
}

func (g *Generator) resolveTags(tags git.Tags, chlog *changelog.Changelog) (git.Tag, git.Tag, error) {
	var ok bool
	var zero git.Tag
	var lastGitTag, lastChangelogTag git.Tag
	var fromTag, toTag git.Tag

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
			return zero, zero, fmt.Errorf("changelog tag not found: %s", chlog.Releases[0].GitTag)
		}
	}

	// Resolve the from tag
	if g.spec.Release.FromTag == "" {
		fromTag = lastChangelogTag
	} else {
		if fromTag, ok = tags.Find(g.spec.Release.FromTag); !ok {
			return zero, zero, fmt.Errorf("from-tag not found: %s", g.spec.Release.FromTag)
		}

		if fromTag.Before(lastChangelogTag) {
			fromTag = lastChangelogTag
		}
	}

	// Resolve the to tag
	if g.spec.Release.ToTag == "" {
		toTag = lastGitTag
	} else {
		if toTag, ok = tags.Find(g.spec.Release.ToTag); !ok {
			return zero, zero, fmt.Errorf("to-tag not found: %s", g.spec.Release.ToTag)
		}

		if toTag.Before(fromTag) || toTag.Equal(fromTag) {
			return zero, zero, errors.New("to-tag should be after the from-tag")
		}
	}

	return fromTag, toTag, nil
}
