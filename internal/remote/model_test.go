package remote

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	ts1, _ = time.Parse(time.RFC3339, "2020-10-24T09:00:00-04:00")
	ts2, _ = time.Parse(time.RFC3339, "2020-10-25T16:00:00-04:00")

	c1 = Change{
		Number:    1001,
		Title:     "Found a bug",
		Labels:    []string{"bug"},
		Milestone: "v1.0",
		Timestamp: ts1,
	}

	c2 = Change{
		Number:    1002,
		Title:     "Added a feature",
		Labels:    []string{"enhancement"},
		Milestone: "v1.0",
		Timestamp: ts2,
	}
)

func TestLabels_Any(t *testing.T) {
	tests := []struct {
		name           string
		labels         Labels
		names          []string
		expectedAny    bool
		expectedAll    bool
		expectedString string
	}{
		{
			name:           "Found",
			labels:         Labels{"bug", "documentation", "enhancement", "question"},
			names:          []string{"bug", "duplicate", "invalid"},
			expectedAny:    true,
			expectedString: "bug,documentation,enhancement,question",
		},
		{
			name:           "NotFound",
			labels:         Labels{"bug", "documentation", "enhancement", "question"},
			names:          []string{"duplicate", "invalid"},
			expectedAny:    false,
			expectedString: "bug,documentation,enhancement,question",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedAny, tc.labels.Any(tc.names...))
			assert.Equal(t, tc.expectedString, tc.labels.String())
		})
	}
}

func TestChanges_Select(t *testing.T) {
	tests := []struct {
		name            string
		changes         Changes
		f               func(Change) bool
		expectedChanges Changes
	}{
		{
			name:    "Labeled",
			changes: Changes{c1, c2, Change{}},
			f: func(c Change) bool {
				return len(c.Labels) > 0
			},
			expectedChanges: Changes{c1, c2},
		},
		{
			name:    "Unlabeled",
			changes: Changes{c1, c2, Change{}},
			f: func(c Change) bool {
				return len(c.Labels) == 0
			},
			expectedChanges: Changes{Change{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changes := tc.changes.Select(tc.f)

			assert.Equal(t, tc.expectedChanges, changes)
		})
	}
}

func TestChanges_Sort(t *testing.T) {
	tests := []struct {
		name            string
		changes         Changes
		expectedChanges Changes
	}{
		{
			name:            "OK",
			changes:         Changes{c1, c2},
			expectedChanges: Changes{c2, c1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changes := tc.changes.Sort()

			assert.Equal(t, tc.expectedChanges, changes)
		})
	}
}
