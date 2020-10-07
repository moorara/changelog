package generate

import (
	"log"

	"github.com/moorara/changelog/internal/spec"
)

// Generate generates changelogs for a Git repository.
func Generate(logger *log.Logger, s *spec.Spec) error {
	logger.Printf("%+v\n", s)
	return nil
}
