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
	URL      string
}

// Tag represents a tag.
type Tag struct {
	Name string
	SHA  string
	Time time.Time
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
	return fmt.Sprintf("%s %s", t.SHA, t.Name)
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

// MapToString customizes the string representation of the tags collection.
func (t Tags) MapToString(f func(t Tag) string) string {
	vals := []string{}
	for _, tag := range t {
		vals = append(vals, f(tag))
	}

	return strings.Join(vals, ", ")
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

// Change represents a change.
// It can be an issue or a merge/pull request.
type Change struct {
	Number         int
	Title          string
	Labels         Labels
	Milestone      string
	Time           time.Time
	Creator        User
	CloserOrMerger User
}

// Changes is a collection of changes.
type Changes []Change

// Select returns a new collection of changes that satisfies the predicate f.
func (c Changes) Select(f func(Change) bool) Changes {
	new := Changes{}
	for _, change := range c {
		if f(change) {
			new = append(new, change)
		}
	}

	return new
}

// Sort sorts the collection of changes by their times from the most recent to the least recent.
func (c Changes) Sort() Changes {
	sorted := make(Changes, len(c))
	copy(sorted, c)

	sort.Slice(sorted, func(i, j int) bool {
		// The order of the tags should be from the most recent to the least recent
		return sorted[i].Time.After(sorted[j].Time)
	})

	return sorted
}
