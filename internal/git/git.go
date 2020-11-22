package git

import (
	"fmt"
	"regexp"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

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
	GetRemoteInfo() (string, string, error)
	Commits() (Commits, error)
	Commit(string) (Commit, error)
	Tags() (Tags, error)
	Tag(string) (Tag, error)
	CommitsFromRevision(string) (Commits, error)
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

// GetRemoteInfo returns the domain part and path part of a Git remote repository URL.
// It assumes the remote repository is named origin.
func (r *repo) GetRemoteInfo() (string, string, error) {
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

func (r *repo) Commits() (Commits, error) {
	r.logger.Debug("Reading git commits ...")

	commitObjs, err := r.git.CommitObjects()
	if err != nil {
		return nil, err
	}

	commits := Commits{}

	_ = commitObjs.ForEach(func(commitObj *object.Commit) error {
		commit := toCommit(commitObj)
		commits[commit.Hash] = commit
		return nil
	})

	r.logger.Debugf("Git commits are read: %d", len(commits))

	return commits, nil
}

func (r *repo) Commit(hash string) (Commit, error) {
	r.logger.Debugf("Reading git commit %s ...", hash)

	commitObj, err := r.git.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return Commit{}, err
	}

	commit := toCommit(commitObj)

	r.logger.Debugf("Git commit %s is read", hash)

	return commit, nil
}

func (r *repo) Tags() (Tags, error) {
	r.logger.Debug("Reading git tags ...")

	refs, err := r.git.Tags()
	if err != nil {
		return nil, err
	}

	tags := Tags{}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		tagObj, err := r.git.TagObject(ref.Hash())
		switch err {
		// Annotated tag
		case nil:
			commitObj, err := r.git.CommitObject(tagObj.Target)
			if err != nil {
				return err
			}
			tag := toAnnotatedTag(tagObj, commitObj)
			tags = append(tags, tag)

		// Lightweight tag
		case plumbing.ErrObjectNotFound:
			commitObj, err := r.git.CommitObject(ref.Hash())
			if err != nil {
				return err
			}
			tag := toLightweightTag(ref, commitObj)
			tags = append(tags, tag)

		default:
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	r.logger.Infof("Git tags are read: %d", len(tags))

	return tags, nil
}

func (r *repo) Tag(name string) (Tag, error) {
	r.logger.Debugf("Reading git tag %s ...", name)

	ref, err := r.git.Tag(name)
	if err != nil {
		return Tag{}, err
	}

	var tag Tag

	tagObj, err := r.git.TagObject(ref.Hash())
	switch err {
	// Annotated tag
	case nil:
		commitObj, err := r.git.CommitObject(tagObj.Target)
		if err != nil {
			return Tag{}, err
		}
		tag = toAnnotatedTag(tagObj, commitObj)

	// Lightweight tag
	case plumbing.ErrObjectNotFound:
		commitObj, err := r.git.CommitObject(ref.Hash())
		if err != nil {
			return Tag{}, err
		}
		tag = toLightweightTag(ref, commitObj)

	default:
		return Tag{}, err
	}

	r.logger.Debugf("Git tag %s is read", name)

	return tag, nil
}

func (r *repo) CommitsFromRevision(rev string) (Commits, error) {
	r.logger.Debugf("Resolving git commits reachable from %s ...", rev)

	hash, err := r.git.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		return nil, err
	}

	commits := Commits{}
	err = r.commitsFromHash(commits, *hash)
	if err != nil {
		return nil, err
	}

	r.logger.Debugf("Resolved git commits reachable from %s: %d", rev, len(commits))

	return commits, nil
}

func (r *repo) commitsFromHash(commits Commits, hash plumbing.Hash) error {
	if _, ok := commits[hash.String()]; ok {
		return nil
	}

	commitObj, err := r.git.CommitObject(hash)
	if err != nil {
		return err
	}

	c := toCommit(commitObj)
	commits[c.Hash] = c

	for _, parentHash := range commitObj.ParentHashes {
		err := r.commitsFromHash(commits, parentHash)
		if err != nil {
			return err
		}
	}

	return nil
}
