package main

import (
	"log"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/moorara/changelog/internal/generate"
	"github.com/moorara/changelog/internal/spec"
	"github.com/moorara/changelog/version"
	"github.com/moorara/flagit"
)

func main() {
	// CREATE DEPENDENCIES

	logger := log.New(os.Stdout, "", 0)

	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err != nil {
		logger.Fatal(err)
	}

	// READING SPEC

	s, err := spec.Default(repo)
	if err != nil {
		logger.Fatal(err)
	}

	if err := flagit.Populate(s, false); err != nil {
		logger.Fatal(err)
	}

	// RUNNING COMMANDS

	switch {
	case s.Help:
		if err := s.PrintHelp(); err != nil {
			logger.Fatal(err)
		}

	case s.Version:
		logger.Println(version.String())

	default:
		if err := generate.Generate(logger, s); err != nil {
			logger.Fatal(err)
		}
	}
}
