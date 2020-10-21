package git

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/moorara/changelog/pkg/log"
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

// TagType determines the type a Git tag.
type TagType int

const (
	// Void represents a tag and does not exist in Git.
	Void TagType = iota
	// Lightweight is a lightweight Git tag.
	Lightweight
	// Annotated is an annotated Git tag.
	Annotated
)

func (t TagType) String() string {
	switch t {
	case Void:
		return ""
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

// Equal determines if two tags are the same.
// Two tags are the same if they both have the same name.
func (t Tag) Equal(u Tag) bool {
	return t.Name == u.Name
}

// Before determines if a given tag is chronologically before the current tag.
func (t Tag) Before(u Tag) bool {
	return t.Tagger.Timestamp.Before(u.Tagger.Timestamp)
}

// After determines if a given tag is chronologically after the current tag.
func (t Tag) After(u Tag) bool {
	return t.Tagger.Timestamp.After(u.Tagger.Timestamp)
}

func (t Tag) String() string {
	return fmt.Sprintf("%s %s %s [%s]", t.Type, t.Hash, t.Name, t.Tagger)
}

// Tags is a list of tags.
type Tags []Tag

// Index looks up a tag by its name and returns its index if found.
// Index returns the index of a tag specified by its name, or -1 if not found.
func (t Tags) Index(name string) int {
	for i, tag := range t {
		if tag.Name == name {
			return i
		}
	}

	return -1
}

// Find looks up a tag by its name.
func (t Tags) Find(name string) (Tag, bool) {
	for _, tag := range t {
		if tag.Name == name {
			return tag, true
		}
	}

	return Tag{}, false
}

// Sort sorts the list of tags by their timestamps from the most recent to the least recent.
func (t Tags) Sort() Tags {
	sorted := make(Tags, len(t))
	copy(sorted, t)

	sort.Slice(sorted, func(i, j int) bool {
		// The order of the tags should be from the most recent to the least recent
		return sorted[i].Tagger.Timestamp.After(sorted[j].Tagger.Timestamp)
	})

	return sorted
}

// Exclude excludes the given tag names and returns a new list of tags.
func (t Tags) Exclude(names ...string) Tags {
	match := func(tag Tag) bool {
		for _, name := range names {
			if tag.Name == name {
				return true
			}
		}
		return false
	}

	new := Tags{}
	for _, tag := range t {
		if !match(tag) {
			new = append(new, tag)
		}
	}

	/* new := make(Tags, len(t))
	copy(new, t)
	for i := range new {
		for _, name := range names {
			if new[i].Name == name {
				new[i].Included = false
			}
		}
	} */

	return new
}

// ExcludeRegex excludes matched tags against the given regex and returns a new list of tags.
func (t Tags) ExcludeRegex(regex *regexp.Regexp) Tags {
	new := Tags{}
	for _, tag := range t {
		if !regex.MatchString(tag.Name) {
			new = append(new, tag)
		}
	}

	/* new := make(Tags, len(t))
	copy(new, t)
	for i := range new {
		if regex.MatchString(new[i].Name) {
			new[i].Included = false
		}
	} */

	return new
}

// Repo is a Git repository.
type Repo struct {
	logger log.Logger
	git    *git.Repository
}

// NewRepo creates a new instance of Repo.
func NewRepo(logger log.Logger, path string) (*Repo, error) {
	git, err := git.PlainOpenWithOptions(path, &git.PlainOpenOptions{
		DetectDotGit: true,
	})

	if err != nil {
		return nil, err
	}

	logger.Debug("Git repository found.")

	return &Repo{
		logger: logger,
		git:    git,
	}, nil
}

// GetRemoteInfo returns the domain part and path part of a Git remote repository URL.
// It assumes the remote repository is named origin.
func (r *Repo) GetRemoteInfo() (string, string, error) {
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

// Tags returns all tags for the Git repository.
func (r *Repo) Tags() (Tags, error) {
	r.logger.Debug("Reading git tag references ...")

	tags := Tags{}

	refs, err := r.git.Tags()
	if err != nil {
		return nil, err
	}

	r.logger.Debug("Parsing git tag references ...")

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

	r.logger.Debug("Git tags parsed.")

	return tags, nil
}
