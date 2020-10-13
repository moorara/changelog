package main

import (
	"log"
	"os"

	"github.com/moorara/flagit"

	"github.com/moorara/changelog/internal/generate"
	"github.com/moorara/changelog/internal/git"
	"github.com/moorara/changelog/internal/spec"
	"github.com/moorara/changelog/version"
)

func main() {
	// CREATE DEPENDENCIES

	logger := log.New(os.Stdout, "", 0)

	gitRepo, err := git.NewRepo(".")
	if err != nil {
		logger.Fatal(err)
	}

	// READING SPEC

	s, err := spec.Default(gitRepo)
	if err != nil {
		logger.Fatal(err)
	}

	s, err = spec.FromFile(s)
	if err != nil {
		logger.Fatal(err)
	}

	if err := flagit.Populate(&s, false); err != nil {
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
		g := generate.New(logger, gitRepo, s)
		if err := g.Generate(); err != nil {
			logger.Fatal(err)
		}
	}
}
