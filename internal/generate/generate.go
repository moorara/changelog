package generate

import (
	"fmt"
	"log"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/git"
	"github.com/moorara/changelog/internal/remote"
	"github.com/moorara/changelog/internal/spec"
)

// Generator is the changelog generator.
type Generator struct {
	logger     *log.Logger
	gitRepo    *git.Repo
	remoteRepo remote.Repo
	processor  changelog.Processor
}

// New creates a new changelog generator.
func New(logger *log.Logger, gitRepo *git.Repo, s spec.Spec) *Generator {
	var remoteRepo remote.Repo
	switch s.Repo.Platform {
	case spec.PlatformGitHub:
		remoteRepo = remote.NewGitHubRepo(logger, s.Repo.Path, s.Repo.AccessToken)
	case spec.PlatformGitLab:
		remoteRepo = remote.NewGitLabRepo(logger, s.Repo.Path, s.Repo.AccessToken)
	}

	return &Generator{
		logger:     logger,
		gitRepo:    gitRepo,
		remoteRepo: remoteRepo,
		processor:  changelog.NewMarkdownProcessor(logger, s.General.File),
	}
}

// Generate generates changelogs for a Git repository.
func (g *Generator) Generate() error {
	changelog, err := g.processor.Parse(changelog.ParseOptions{})
	if err != nil {
		return err
	}

	tags, err := g.gitRepo.Tags()
	if err != nil {
		return err
	}

	// TODO: Remove!
	for _, t := range tags {
		fmt.Printf("%+v\n", t)
	}
	fmt.Println()

	content, err := g.processor.Render(changelog)
	if err != nil {
		return err
	}

	// TODO: Remove!
	g.logger.Println(content)

	return nil
}
