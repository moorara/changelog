package git

import (
	"fmt"
	"regexp"

	"github.com/go-git/go-git/v5"

	"github.com/moorara/changelog/log"
)

var (
	idPattern       = `[A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]`
	domainPattern   = fmt.Sprintf(`%s\.[A-Za-z]{2,63}`, idPattern)
	repoPathPattern = fmt.Sprintf(`(%s/){1,20}(%s)`, idPattern, idPattern)
	httpsPattern    = fmt.Sprintf(`^https://(%s)/(%s)(.git)?$`, domainPattern, repoPathPattern)
	sshPattern      = fmt.Sprintf(`^git@(%s):(%s)(.git)?$`, domainPattern, repoPathPattern)
	httpsRE         = regexp.MustCompile(httpsPattern)
	sshRE           = regexp.MustCompile(sshPattern)
)

// Repo is a Git repository.
type Repo interface {
	GetRemote() (string, string, error)
}

type repo struct {
	logger log.Logger
	git    *git.Repository
}

// NewRepo creates a new instance of Repo.
func NewRepo(logger log.Logger, path string) (Repo, error) {
	git, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err != nil {
		return nil, err
	}

	logger.Debug("Git repository found")

	return &repo{
		logger: logger,
		git:    git,
	}, nil
}

// GetRemote returns the domain part and path part of a Git remote repository URL.
// It assumes the remote repository is named origin.
func (r *repo) GetRemote() (string, string, error) {
	r.logger.Debug("Reading git remote URL ...")

	// TODO: Should we handle all remote names and not just the origin?
	remote, err := r.git.Remote("origin")
	if err != nil {
		return "", "", err
	}

	// TODO: Should we handle all URLs and not just the first one?
	var remoteURL string
	if config := remote.Config(); len(config.URLs) > 0 {
		remoteURL = config.URLs[0]
	}

	// Parse the origin remote URL into a domain part a path part
	if matches := httpsRE.FindStringSubmatch(remoteURL); len(matches) == 6 {
		// Git remote url is using HTTPS protocol
		// Example: https://github.com/moorara/changelog.git --> matches = []string{"https://github.com/moorara/changelog.git", "github.com", "moorara/changelog", "moorara/", "changelog", ".git"}
		r.logger.Infof("Git remote URL: %s", remoteURL)
		return matches[1], matches[2], nil
	} else if matches := sshRE.FindStringSubmatch(remoteURL); len(matches) == 6 {
		// Git remote url is using SSH protocol
		// Example: git@github.com:moorara/changelog.git --> matches = []string{"git@github.com:moorara/changelog.git", "github.com", "moorara/changelog, "moorara/", "changelog", ".git"}
		r.logger.Infof("Git remote URL: %s", remoteURL)
		return matches[1], matches[2], nil
	}

	return "", "", fmt.Errorf("invalid git remote url: %s", remoteURL)
}
