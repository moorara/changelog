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

### Help

<details>
  <summary>changelog -help</summary>

```
  changelog is a simple command-line tool for generating changelogs based on issues and pull/merge requests.
  It assumes the remote repository name is origin.

  Supported Remote Repositories:

    • GitHub (github.com)

  Usage: changelog [flags]

  Flags:

    -help                         Show the help text
    -version                      Print the version number

    -access-token                 The OAuth access token for making API calls
                                  The default value is read from the CHANGELOG_ACCESS_TOKEN environment variable

    -file                         The output file for the generated changelog (default: CHANGELOG.md)
    -base                         An optional file for appending the generated changelog to it 
                                  This option can only be used when generating the changelog for the first time
    -print                        Print the generated changelong to STDOUT (default: false)
                                  If this option is enabled, all logs will be disabled
    -verbose                      Show the vervbosity logs (default: false)

    -from-tag                     Changelog will be generated for all changes after this tag (default: last tag on changelog)
    -to-tag                       Changelog will be generated for all changes before this tag (default: last git tag)
    -future-tag                   A future tag for all unreleased changes (changes after the last git tag) 
    -exclude-tags                 These tags will be excluded from changelog 
    -exclude-tags-regex           A POSIX-compliant regex for excluding certain tags from changelog 

    -issues-selection             Include closed issues in changelog (values: none|all|labeled) (default: all)
    -issues-include-labels        Include issues with these labels 
    -issues-exclude-labels        Exclude issues with these labels (default: duplicate,invalid,question,wontfix)
    -issues-grouping              Grouping style for issues (values: simple|milestone|label) (default: label)
    -issues-summary-labels        Labels for summary group (default: summary,release-summary)
    -issues-removed-labels        Labels for removed group (default: removed)
    -issues-breaking-labels       Labels for breaking group (default: breaking,backward-incompatible)
    -issues-deprecated-labels     Labels for deprecated group (default: deprecated)
    -issues-feature-labels        Labels for feature group (default: feature)
    -issues-enhancement-labels    Labels for enhancement group (default: enhancement)
    -issues-bug-labels            Labels for bug group (default: bug)
    -issues-security-labels       Labels for security group (default: security)

    -merges-selection             Include merged pull/merge requests in changelog (values: none|all|labeled) (default: all)
    -merges-branch                Include pull/merge requests merged into this branch (default: default remote branch)
    -merges-include-labels        Include merges with these labels 
    -merges-exclude-labels        Exclude merges with these labels 
    -merges-grouping              Grouping style for pull/merge requests (values: simple|milestone|label) (default: simple)
    -merges-summary-labels        Labels for summary group 
    -merges-removed-labels        Labels for removed group 
    -merges-breaking-labels       Labels for breaking group 
    -merges-deprecated-labels     Labels for deprecated group 
    -merges-feature-labels        Labels for feature group 
    -merges-enhancement-labels    Labels for enhancement group 
    -merges-bug-labels            Labels for bug group 
    -merges-security-labels       Labels for security group 

    -release-url                  An external release URL with the '{tag}' placeholder for the release tag

  Examples:

    changelog
    changelog -access-token=<your-access-token>
```
</details>

### Spec File

You can check in a file in your repository for configuring how changelogs are generated.

<details>
  <summary>changelog.yaml</summary>

```yaml
general:
  file: CHANGELOG.md
  base: HISTORY.md
  print: true
  verbose: false

tags:
  exclude: [ prerelease, candidate ]
  exclude-regex: (.*)-(alpha|beta)

issues:
  selection: labeled
  include-labels: [ breaking, bug, defect, deprecated, enhancement, feature, highlight, improvement, incompatible, privacy, removed, security, summary ]
  exclude-labels: [ documentation, duplicate, invalid, question, wontfix ]
  grouping: milestone
  summary-labels: [ summary, highlight ]
  removed-labels: [ removed ]
  breaking-labels: [ breaking, incompatible ]
  deprecated-labels: [ deprecated ]
  feature-labels: [ feature ]
  enhancement-labels: [ enhancement, improvement ]
  bug-labels: [ bug, defect ]
  security-labels: [ security, privacy ]

merges:
  selection: labeled
  branch: production
  include-labels: [ breaking, bug, defect, deprecated, enhancement, feature, highlight, improvement, incompatible, privacy, removed, security, summary ]
  exclude-labels: [ documentation, duplicate, invalid, question, wontfix ]
  grouping: label
  summary-labels: [ summary, highlight ]
  removed-labels: [ removed ]
  breaking-labels: [ breaking, incompatible ]
  deprecated-labels: [ deprecated ]
  feature-labels: [ feature ]
  enhancement-labels: [ enhancement, improvement ]
  bug-labels: [ bug, defect ]
  security-labels: [ security, privacy ]

content:
  release-url: https://storage.artifactory.com/project/releases/{tag}
```
</details>

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
