package git

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// Signature determines who and when created a commit or tag.
type Signature struct {
	Name  string
	Email string
	Time  time.Time
}

// Before determines if a given signature is chronologically before another signature.
func (s Signature) Before(t Signature) bool {
	return s.Time.Before(t.Time)
}

// After determines if a given signature is chronologically after another signature.
func (s Signature) After(t Signature) bool {
	return s.Time.After(t.Time)
}

func (s Signature) String() string {
	return fmt.Sprintf("%s <%s> %s", s.Name, s.Email, s.Time.Format(time.RFC3339))
}

// Commit represents a Git commit.
type Commit struct {
	Hash      string
	Author    Signature
	Committer Signature
	Message   string
	Parents   []string
}

func toCommit(commitObj *object.Commit) Commit {
	return Commit{
		Hash: commitObj.Hash.String(),
		Author: Signature{
			Name:  commitObj.Author.Name,
			Email: commitObj.Author.Email,
			Time:  commitObj.Author.When,
		},
		Committer: Signature{
			Name:  commitObj.Committer.Name,
			Email: commitObj.Committer.Email,
			Time:  commitObj.Committer.When,
		},
		Message: commitObj.Message,
	}
}

// Equal determines if two commits are the same.
// Two commits are the same if they both have the same hash.
func (c Commit) Equal(d Commit) bool {
	return c.Hash == d.Hash
}

// Before determines if a given commit is chronologically before another commit.
func (c Commit) Before(d Commit) bool {
	return c.Committer.Before(d.Committer)
}

// After determines if a given commit is chronologically after another commit.
func (c Commit) After(d Commit) bool {
	return c.Committer.After(d.Committer)
}

// ShortMessage returns a one-line truncated commit message.
func (c Commit) ShortMessage() string {
	message := strings.Split(c.Message, "\n")[0]
	if len(message) > 100 {
		message = message[:100] + " ..."
	}
	return message
}

func (c Commit) String() string {
	return fmt.Sprintf("%s %s", c.Hash[:7], c.ShortMessage())
}

// Text returns a multi-line commit string.
func (c Commit) Text() string {
	return fmt.Sprintf("%s\nAuthor:    %s\nCommitter: %s\n%s", c.Hash, c.Author, c.Committer, c.Message)
}

// Commits is a map of Git commits.
type Commits map[string]Commit

// Sort sorts the map of commits by their commit times from the most recent to the least recent.
func (c Commits) Sort() []Commit {
	sorted := make([]Commit, 0)
	for _, commit := range c {
		sorted = append(sorted, commit)
	}

	sort.Slice(sorted, func(i, j int) bool {
		// The order of the commits should be from the most recent to the least recent
		return sorted[i].Committer.After(sorted[j].Committer)
	})

	return sorted
}

// TagType determines type a Git tag.
type TagType int

const (
	// Void is not a real Git tag!
	Void TagType = iota
	// Lightweight is a lightweight Git tag.
	Lightweight
	// Annotated is an annotated Git tag.
	Annotated
)

func (t TagType) String() string {
	switch t {
	case Void:
		return "Void"
	case Lightweight:
		return "Lightweight"
	case Annotated:
		return "Annotated"
	default:
		return "Invalid"
	}
}

// Tag represents a Git tag.
type Tag struct {
	Type    TagType
	Hash    string
	Name    string
	Tagger  *Signature
	Message *string
	Commit  Commit
}

func toLightweightTag(ref *plumbing.Reference, commitObj *object.Commit) Tag {
	// It is assumed that the given reference is a tag reference
	name := strings.TrimPrefix(string(ref.Name()), "refs/tags/")

	return Tag{
		Type:   Lightweight,
		Hash:   ref.Hash().String(),
		Name:   name,
		Commit: toCommit(commitObj),
	}
}

func toAnnotatedTag(tagObj *object.Tag, commitObj *object.Commit) Tag {
	return Tag{
		Type: Annotated,
		Hash: tagObj.Hash.String(),
		Name: tagObj.Name,
		Tagger: &Signature{
			Name:  tagObj.Tagger.Name,
			Email: tagObj.Tagger.Email,
			Time:  tagObj.Tagger.When,
		},
		Message: &tagObj.Message,
		Commit:  toCommit(commitObj),
	}
}

// IsZero determines if a tag is a zero tag instance.
func (t Tag) IsZero() bool {
	return reflect.ValueOf(t).IsZero()
}

// Equal determines if two tags are the same.
// Two tags are the same if they both have the same name.
func (t Tag) Equal(u Tag) bool {
	return t.Name == u.Name
}

// Before determines if a given tag is chronologically before another tag.
// Two tags are compared using the commits they refer to.
func (t Tag) Before(u Tag) bool {
	return t.Commit.Before(u.Commit)
}

// After determines if a given tag is chronologically after another tag.
// Two tags are compared using the commits they refer to.
func (t Tag) After(u Tag) bool {
	return t.Commit.After(u.Commit)
}

func (t Tag) String() string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s %s %s Commit[%s %s]", t.Type, t.Hash, t.Name, t.Commit.Hash, t.Commit.ShortMessage())
}

// Tags is a list of Git tags.
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

// Sort sorts the list of tags by their times from the most recent to the least recent.
func (t Tags) Sort() Tags {
	sorted := make(Tags, len(t))
	copy(sorted, t)

	sort.Slice(sorted, func(i, j int) bool {
		// The order of the tags should be from the most recent to the least recent
		return sorted[i].After(sorted[j])
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

// Map converts a list of tags to a list of strings.
func (t Tags) Map(f func(t Tag) string) []string {
	mapped := []string{}
	for _, tag := range t {
		mapped = append(mapped, f(tag))
	}

	return mapped
}
