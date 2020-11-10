[![Go Doc][godoc-image]][godoc-url]
[![Build Status][workflow-image]][workflow-url]
[![Go Report Card][goreport-image]][goreport-url]
[![Test Coverage][coverage-image]][coverage-url]
[![Maintainability][maintainability-image]][maintainability-url]

# Changelog

Changelog is a simple changelog generator for GitHub repositories.
It is mainly a reimplementation of the famous [github-changelog-generator](https://github.com/github-changelog-generator/github-changelog-generator) in Go.

## Why?

[github-changelog-generator](https://github.com/github-changelog-generator/github-changelog-generator) is great, battle-proven, and just works! So, why did I decide to reinvent the wheel?

It is not quite a reinvention!
For a long time, I was using the aforementioned Ruby Gem to generated changelogs for my repositories.
For creating releases on GitHub, I had to use another program to call the *github_changelog_generator* and then extract the newly added to changelog.
Here are some of the challenges I faced with this approach:

  - No control over the installed version of the Gem on developer machines
  - Installing and configuring a full Ruby environment in CI or containerized environments
  - Breaking changes surfacing from time to time due to dependency to an external program

I wish there was a version of this awesome piece of software in Go
(See https://github.com/github-changelog-generator/github-changelog-generator/issues/714).
So, I could copy a self-sufficient binary into my containers for generating changelogs.
Moreover, I could import it as a library with version management and make my tooling dependency-free.

The luxury of a fresh start allows you to:

  - Simplify and improve the user experience
  - Improve the performance and efficiency
  - Clean up workarounds that are no longer needed
  - Lay out a plan for further development in future

## Quick Start

### Install

### Examples

```bash
# Simply generated a changelog
changelog -access-token=$GITHUB_TOKEN

# Assign unreleased changes (changes without a tag) to a future tag that has not been yet created.
changelog -access-token=$GITHUB_TOKEN -future-tag v0.1.0
```

## Features

  - Single, dependency-free, and cross-platform binary
  - Generating changelog for issues and pull/merge requests
  - Creating changelog for unreleased changes (future or draft releases)
  - Filtering tags by name or regex
  - Filtering issues and pull/merge requests by labels
  - Grouping issues and pull/merge requests by labels
  - Grouping issues and pull/merge requests by milestone

## Expected Behavior

When you run the _changelog_ inside a Git directory, the following steps happen:

  1. Your remote repository is determined by the remote name `origin` (SSH and HTTPS URLs are supported).
  1. The existing changelog file (if any) will be compared against the list of Git tags and the list of tags without changelog will be resolved.
  1. The list of candidate tags will be further refined if the `exclude-tags` or/and `exclude-tags-regex` options are specified.
  1. A chain of API calls will be made to the remote platform (i.e. GitHub) and a list of **closed issues** and **merged pull/merge requests** will be retrieved.
  1. The list of issues will be filtered according to issues `selection`, `include-labels`, and `exclude-labels` options.
  1. The list of pull/merge requests will be filtered according to merges `selection`, `branch`, `include-labels`, and `exclude-labels` options.
  1. The list of issues will be grouped using the issues `grouping` option.
  1. The list of pull/merge requests will be grouped using the merges `grouping` option.
  1. Finally, the actual changelog will be generated and written to the changelog file.

## TODO

My goal with this changelog generator is to keep it simple and relevant.
It is only supposed to do one job and it should just work out-of-the-box.

  - Remote repository support:
    - [x] GitHub
    - [ ] GitLab
  - Changelog format:
    - [x] Markdown
    - [ ] HTML
  - Enhancements:
    - [ ] Updating GitHub release desctiptions
    - [ ] Switching to GitHub API v4 (GraphQL)


[godoc-url]: https://pkg.go.dev/github.com/moorara/changelog
[godoc-image]: https://godoc.org/github.com/moorara/changelog?status.svg
[workflow-url]: https://github.com/moorara/changelog/actions
[workflow-image]: https://github.com/moorara/changelog/workflows/Main/badge.svg
[goreport-url]: https://goreportcard.com/report/github.com/moorara/changelog
[goreport-image]: https://goreportcard.com/badge/github.com/moorara/changelog
[coverage-url]: https://codeclimate.com/github/moorara/changelog/test_coverage
[coverage-image]: https://api.codeclimate.com/v1/badges/698f0cb8dc342903b2ba/test_coverage
[maintainability-url]: https://codeclimate.com/github/moorara/changelog/maintainability
[maintainability-image]: https://api.codeclimate.com/v1/badges/698f0cb8dc342903b2ba/maintainability
