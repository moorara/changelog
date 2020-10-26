package remote

import (
	"sort"
	"strings"
	"time"
)

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

// Change represents a single change.
// It can be an issue or a merge/pull/change request.
type Change struct {
	Number    int
	Title     string
	Labels    Labels
	Milestone string
	Timestamp time.Time
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

// Sort sorts the collection of changes by their timestamps from the most recent to the least recent.
func (c Changes) Sort() Changes {
	sorted := make(Changes, len(c))
	copy(sorted, c)

	sort.Slice(sorted, func(i, j int) bool {
		// The order of the tags should be from the most recent to the least recent
		return sorted[i].Timestamp.After(sorted[j].Timestamp)
	})

	return sorted
}
