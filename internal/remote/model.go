package remote

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"
)

// User represents a user.
type User struct {
	Name     string
	Email    string
	Username string
	WebURL   string
}

// Commit represents a commit.
type Commit struct {
	Hash string
	Time time.Time
}

func (c Commit) String() string {
	return fmt.Sprintf("%s", c.Hash)
}

// Commits is a collection of commits.
type Commits []Commit

// Any returns true if at least one of the commits is equal to the given commit.
func (c Commits) Any(hash string) bool {
	for _, commit := range c {
		if commit.Hash == hash {
			return true
		}
	}

	return false
}

// Map converts a list of commits to a list of strings.
func (c Commits) Map(f func(c Commit) string) []string {
	mapped := []string{}
	for _, commit := range c {
		mapped = append(mapped, f(commit))
	}

	return mapped
}

// Branch represents a branch.
type Branch struct {
	Name   string
	Commit Commit
}

func (b Branch) String() string {
	return b.Name
}

// Tag represents a tag.
type Tag struct {
	Name       string
	Time       time.Time
	Commit     Commit
	WebURL     string
	CompareURL string
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
	return t.Time.Before(u.Time)
}

// After determines if a given tag is chronologically after another tag.
// Two tags are compared using the commits they refer to.
func (t Tag) After(u Tag) bool {
	return t.Time.After(u.Time)
}

func (t Tag) String() string {
	if t.IsZero() {
		return ""
	}
	return fmt.Sprintf("%s Commit[%s]", t.Name, t.Commit.Hash)
}

// Tags is a collection of tags.
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

// First returns the first tag that satisifies the given predicate.
func (t Tags) First(f func(Tag) bool) (Tag, bool) {
	for _, tag := range t {
		if f(tag) {
			return tag, true
		}
	}

	return Tag{}, false
}

// Last returns the last tag that satisifies the given predicate.
func (t Tags) Last(f func(Tag) bool) (Tag, bool) {
	for i := len(t) - 1; i >= 0; i-- {
		if f(t[i]) {
			return t[i], true
		}
	}

	return Tag{}, false
}

// Sort sorts the collection of tags by their times from the most recent to the least recent.
func (t Tags) Sort() Tags {
	sorted := make(Tags, len(t))
	copy(sorted, t)

	sort.Slice(sorted, func(i, j int) bool {
		// The order of the tags should be from the most recent to the least recent
		return sorted[i].Time.After(sorted[j].Time)
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

	return new
}

// Reverse returns a new list of tags with the reverse order.
func (t Tags) Reverse() Tags {
	l := len(t)
	reversed := make(Tags, l)

	for i := 0; i < l; i++ {
		reversed[l-1-i] = t[i]
	}

	return reversed
}

// Map converts a list of tags to a list of strings.
func (t Tags) Map(f func(t Tag) string) []string {
	mapped := []string{}
	for _, tag := range t {
		mapped = append(mapped, f(tag))
	}

	return mapped
}

// Labels is a collection of labels.
type Labels []string

// Any returns true if one of the labels is equal to any of the given names.
func (l Labels) Any(names ...string) bool {
	f := func(label string) bool {
		for _, name := range names {
			if label == name {
				return true
			}
		}
		return false
	}

	for _, label := range l {
		if f(label) {
			return true
		}
	}

	return false
}

func (l Labels) String() string {
	return strings.Join(l, ",")
}

// Change has the common fields of an issue or a merge/pull request.
type Change struct {
	Number    int
	Title     string
	Labels    Labels
	Milestone string
	Time      time.Time
	Author    User
	WebURL    string
}

// Issue represents an issue.
type Issue struct {
	Change
	Closer User
}

// Issues is a collection of issues.
type Issues []Issue

// Select returns a new collection of issues that satisfies the predicate f.
func (i Issues) Select(f func(Issue) bool) Issues {
	new := Issues{}
	for _, issue := range i {
		if f(issue) {
			new = append(new, issue)
		}
	}

	return new
}

// Sort sorts the collection of issues from the most recent to the least recent.
func (i Issues) Sort() Issues {
	sorted := make(Issues, len(i))
	copy(sorted, i)

	sort.Slice(sorted, func(i, j int) bool {
		// The order of the tags should be from the most recent to the least recent
		return sorted[i].Time.After(sorted[j].Time)
	})

	return sorted
}

// Remove removes any issue that satisfies the predicate f and returns the removed issues.
func (i *Issues) Remove(f func(Issue) bool) Issues {
	var removed Issues

	for n, issue := range *i {
		if f(issue) {
			removed = append(removed, (*i)[n])
			*i = append((*i)[:n], (*i)[n+1:]...)
		}
	}

	return removed
}

// Merge represents a merge/pull request.
type Merge struct {
	Change
	Merger User
	Commit Commit
}

// Merges is a collection of merges.
type Merges []Merge

// Select returns a new collection of merges that satisfies the predicate f.
func (m Merges) Select(f func(Merge) bool) Merges {
	new := Merges{}
	for _, merge := range m {
		if f(merge) {
			new = append(new, merge)
		}
	}

	return new
}

// Sort sorts the collection of merges from the most recent to the least recent.
func (m Merges) Sort() Merges {
	sorted := make(Merges, len(m))
	copy(sorted, m)

	sort.Slice(sorted, func(i, j int) bool {
		// The order of the tags should be from the most recent to the least recent
		return sorted[i].Time.After(sorted[j].Time)
	})

	return sorted
}

// Remove removes any merge that satisfies the predicate f and returns the removed merges.
func (m *Merges) Remove(f func(Merge) bool) Merges {
	var removed Merges

	for n, merge := range *m {
		if f(merge) {
			removed = append(removed, (*m)[n])
			*m = append((*m)[:n], (*m)[n+1:]...)
		}
	}

	return removed
}
