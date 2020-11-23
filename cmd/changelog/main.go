package main

import (
	"context"
	"fmt"

	"github.com/moorara/flagit"

	"github.com/moorara/changelog/generate"
	"github.com/moorara/changelog/internal/git"
	"github.com/moorara/changelog/log"
	"github.com/moorara/changelog/spec"
	"github.com/moorara/changelog/version"
)

func main() {
	// CREATE DEPENDENCIES

	// We cannot enable the logger until the verbosity is known
	logger := log.New(log.None)

	gitRepo, err := git.NewRepo(logger, ".")
	if err != nil {
		logger.Fatal(err)
	}

	domain, path, err := gitRepo.GetRemoteInfo()
	if err != nil {
		logger.Fatal(err)
	}

	// READING SPEC

	s := spec.Default(domain, path)

	s, err = spec.FromFile(s)
	if err != nil {
		logger.Fatal(err)
	}

	if err := flagit.Populate(&s, false); err != nil {
		logger.Fatal(err)
	}

	// Update logger verbosity
	if s.General.Verbose {
		logger.ChangeVerbosity(log.Debug)
	} else if !s.General.Print {
		logger.ChangeVerbosity(log.Info)
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
		g, err := generate.New(s, logger)
		if err != nil {
			logger.Fatal(err)
		}

		ctx := context.Background()

		if _, err := g.Generate(ctx, s); err != nil {
			logger.Fatal(err)
		}
	}
}
