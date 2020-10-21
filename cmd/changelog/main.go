package main

import (
	"context"
	"fmt"

	"github.com/moorara/flagit"

	"github.com/moorara/changelog/internal/generate"
	"github.com/moorara/changelog/internal/git"
	"github.com/moorara/changelog/internal/spec"
	"github.com/moorara/changelog/pkg/log"
	"github.com/moorara/changelog/version"
)

func main() {
	// CREATE DEPENDENCIES

	logger := log.New(log.Debug)

	gitRepo, err := git.NewRepo(logger, ".")
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

	logger.Debug(s)

	// RUNNING COMMANDS

	switch {
	case s.Help:
		if err := s.PrintHelp(); err != nil {
			logger.Fatal(err)
		}

	case s.Version:
		fmt.Println(version.String())

	default:
		g := generate.New(s, logger, gitRepo)
		ctx := context.Background()

		if err := g.Generate(ctx); err != nil {
			logger.Fatal(err)
		}
	}
}
