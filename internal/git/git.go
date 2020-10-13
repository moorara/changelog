package git

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
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

// TagType determines if a Git tag is lightweight or annotated.
type TagType bool

const (
	// Lightweight is a lightweight Git tag.
	Lightweight TagType = false
	// Annotated is an annotated Git tag.
	Annotated TagType = true
)

func (t TagType) String() string {
	switch t {
	case Lightweight:
		return "Lightweight"
	case Annotated:
		return "Annotated"
	default:
		return "Invalid"
	}
}

// Tagger is the one who created the tag.
type Tagger struct {
	Name      string
	Email     string
	Timestamp time.Time
}

func (t Tagger) String() string {
	return fmt.Sprintf("%s <%s> %s", t.Name, t.Email, t.Timestamp.Format(time.RFC3339))
}

// Tag represents a Git tag.
type Tag struct {
	Type   TagType
	Hash   string
	Name   string
	Tagger Tagger
}

func lightweightTag(ref *plumbing.Reference, commit *object.Commit) Tag {
	// It is assumed that the given reference is a tag reference
	name := strings.TrimPrefix(string(ref.Name()), "refs/tags/")

	return Tag{
		Type: Lightweight,
		Hash: ref.Hash().String(),
		Name: name,
		Tagger: Tagger{ // TODO: Is this the right tagger?
			Name:      commit.Committer.Name,
			Email:     commit.Committer.Email,
			Timestamp: commit.Committer.When,
		},
	}
}

func annotatedTag(tagObj *object.Tag) Tag {
	return Tag{
		Type: Annotated,
		Hash: tagObj.Hash.String(),
		Name: tagObj.Name,
		Tagger: Tagger{
			Name:      tagObj.Tagger.Name,
			Email:     tagObj.Tagger.Email,
			Timestamp: tagObj.Tagger.When,
		},
	}
}

func (t Tag) String() string {
	return fmt.Sprintf("%s %s %s [%s]", t.Type, t.Hash, t.Name, t.Tagger)
}

// Repo is a Git repository.
type Repo struct {
	git *git.Repository
}

// NewRepo creates a new instance of Repo.
func NewRepo(path string) (*Repo, error) {
	git, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err != nil {
		return nil, err
	}

	return &Repo{
		git: git,
	}, nil
}

// GetRemoteInfo returns the domain part and path part of a Git remote repository URL.
// It assumes the remote repository is named origin.
func (r *Repo) GetRemoteInfo() (string, string, error) {
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
		return matches[1], matches[2], nil
	} else if matches := sshRE.FindStringSubmatch(remoteURL); len(matches) == 6 {
		// Git remote url is using SSH protocol
		// Example: git@github.com:moorara/changelog.git --> matches = []string{"git@github.com:moorara/changelog.git", "github.com", "moorara/changelog, "moorara/", "changelog", ".git"}
		return matches[1], matches[2], nil
	}

	return "", "", fmt.Errorf("invalid git remote url: %s", remoteURL)
}

// Tags returns all tags for the Git repository.
func (r *Repo) Tags() ([]Tag, error) {
	tags := make([]Tag, 0)

	refs, err := r.git.Tags()
	if err != nil {
		return nil, err
	}

	err = refs.ForEach(func(ref *plumbing.Reference) error {
		tagObj, err := r.git.TagObject(ref.Hash())
		switch err {
		// Annotated tag
		case nil:
			tag := annotatedTag(tagObj)
			tags = append(tags, tag)

		// Lightweight tag
		case plumbing.ErrObjectNotFound:
			commit, err := r.git.CommitObject(ref.Hash())
			if err != nil {
				return err
			}

			tag := lightweightTag(ref, commit)
			tags = append(tags, tag)

		default:
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tags, nil
}
