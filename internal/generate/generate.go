package generate

import (
	"log"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/internal/changelog/markdown"
	"github.com/moorara/changelog/internal/spec"
)

// Generate generates changelogs for a Git repository.
func Generate(logger *log.Logger, spec spec.Spec) error {
	var markdownProcessor changelog.Processor
	markdownProcessor = markdown.NewProcessor(logger, spec.General.File)

	changelog, err := markdownProcessor.Parse(changelog.ParseOptions{})
	if err != nil {
		return err
	}

	content, err := markdownProcessor.Render(changelog)
	if err != nil {
		return err
	}

	logger.Println(content)

	return nil
}
