package remote

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	t1, _ = time.Parse(time.RFC3339, "2020-10-24T09:00:00-04:00")
	t2, _ = time.Parse(time.RFC3339, "2020-10-25T16:00:00-04:00")

	tag1 = Tag{
		Name: "v0.1.0",
		SHA:  "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		Time: t1,
	}

	tag2 = Tag{
		Name: "v0.2.0",
		SHA:  "0251a422d2038967eeaaaa5c8aa76c7067fdef05",
		Time: t2,
	}

	change1 = Change{
		Number:    1001,
		Title:     "Found a bug",
		Labels:    []string{"bug"},
		Milestone: "v1.0",
		Time:      t1,
	}

	change2 = Change{
		Number:    1002,
		Title:     "Added a feature",
		Labels:    []string{"enhancement"},
		Milestone: "v1.0",
		Time:      t2,
	}
)

func TestTag(t *testing.T) {
	tests := []struct {
		name            string
		t1, t2          Tag
		expectedIsZero1 bool
		expectedIsZero2 bool
		expectedEqual   bool
		expectedBefore  bool
		expectedAfter   bool
		expectedString1 string
		expectedString2 string
	}{
		{
			name:            "Zero",
			t1:              Tag{},
			t2:              Tag{},
			expectedIsZero1: true,
			expectedIsZero2: true,
			expectedEqual:   true,
			expectedBefore:  false,
			expectedAfter:   false,
			expectedString1: "",
			expectedString2: "",
		},
		{
			name:            "Equal",
			t1:              tag1,
			t2:              tag1,
			expectedIsZero1: false,
			expectedIsZero2: false,
			expectedEqual:   true,
			expectedBefore:  false,
			expectedAfter:   false,
			expectedString1: "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 v0.1.0",
			expectedString2: "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 v0.1.0",
		},
		{
			name:            "Before",
			t1:              tag1,
			t2:              tag2,
			expectedIsZero1: false,
			expectedIsZero2: false,
			expectedEqual:   false,
			expectedBefore:  true,
			expectedAfter:   false,
			expectedString1: "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 v0.1.0",
			expectedString2: "0251a422d2038967eeaaaa5c8aa76c7067fdef05 v0.2.0",
		},
		{
			name:            "After",
			t1:              tag2,
			t2:              tag1,
			expectedIsZero1: false,
			expectedIsZero2: false,
			expectedEqual:   false,
			expectedBefore:  false,
			expectedAfter:   true,
			expectedString1: "0251a422d2038967eeaaaa5c8aa76c7067fdef05 v0.2.0",
			expectedString2: "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 v0.1.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedIsZero1, tc.t1.IsZero())
			assert.Equal(t, tc.expectedIsZero2, tc.t2.IsZero())
			assert.Equal(t, tc.expectedEqual, tc.t1.Equal(tc.t2))
			assert.Equal(t, tc.expectedBefore, tc.t1.Before(tc.t2))
			assert.Equal(t, tc.expectedAfter, tc.t1.After(tc.t2))
			assert.Equal(t, tc.expectedString1, tc.t1.String())
			assert.Equal(t, tc.expectedString2, tc.t2.String())
		})
	}
}

func TestTags_Index(t *testing.T) {
	tests := []struct {
		name          string
		t             Tags
		tagName       string
		expectedIndex int
	}{
		{
			name:          "Found",
			t:             Tags{tag1, tag2},
			tagName:       "v0.2.0",
			expectedIndex: 1,
		},
		{
			name:          "NotFound",
			t:             Tags{tag1, tag2},
			tagName:       "v0.3.0",
			expectedIndex: -1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			index := tc.t.Index(tc.tagName)

			assert.Equal(t, tc.expectedIndex, index)
		})
	}
}

func TestTags_Find(t *testing.T) {
	tests := []struct {
		name        string
		t           Tags
		tagName     string
		expectedTag Tag
		expectedOK  bool
	}{
		{
			name:        "Found",
			t:           Tags{tag1, tag2},
			tagName:     "v0.2.0",
			expectedTag: tag2,
			expectedOK:  true,
		},
		{
			name:        "NotFound",
			t:           Tags{tag1, tag2},
			tagName:     "v0.3.0",
			expectedTag: Tag{},
			expectedOK:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag, ok := tc.t.Find(tc.tagName)

			assert.Equal(t, tc.expectedTag, tag)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestTags_Sort(t *testing.T) {
	tests := []struct {
		name         string
		tags         Tags
		expectedTags Tags
	}{
		{
			name:         "OK",
			tags:         Tags{tag1, tag2},
			expectedTags: Tags{tag2, tag1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := tc.tags.Sort()

			assert.Equal(t, tc.expectedTags, tags)
		})
	}
}

func TestTags_Exclude(t *testing.T) {
	tests := []struct {
		name         string
		t            Tags
		names        []string
		expectedTags Tags
	}{
		{
			name:         "OK",
			t:            Tags{tag1, tag2},
			names:        []string{"v0.2.0"},
			expectedTags: Tags{tag1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := tc.t.Exclude(tc.names...)

			assert.Equal(t, tc.expectedTags, tags)
		})
	}
}

func TestTags_ExcludeRegex(t *testing.T) {
	tests := []struct {
		name         string
		t            Tags
		regex        *regexp.Regexp
		expectedTags Tags
	}{
		{
			name:         "OK",
			t:            Tags{tag1, tag2},
			regex:        regexp.MustCompile(`v\d+\.2\.\d+`),
			expectedTags: Tags{tag1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := tc.t.ExcludeRegex(tc.regex)

			assert.Equal(t, tc.expectedTags, tags)
		})
	}
}

func TestTags_MapToString(t *testing.T) {
	tests := []struct {
		name           string
		t              Tags
		f              func(Tag) string
		expectedString string
	}{
		{
			name: "OK",
			t:    Tags{tag1, tag2},
			f: func(t Tag) string {
				return t.Name
			},
			expectedString: "v0.1.0, v0.2.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			str := tc.t.MapToString(tc.f)

			assert.Equal(t, tc.expectedString, str)
		})
	}
}

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
			changes: Changes{change1, change2, Change{}},
			f: func(c Change) bool {
				return len(c.Labels) > 0
			},
			expectedChanges: Changes{change1, change2},
		},
		{
			name:    "Unlabeled",
			changes: Changes{change1, change2, Change{}},
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
			changes:         Changes{change1, change2},
			expectedChanges: Changes{change2, change1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			changes := tc.changes.Sort()

			assert.Equal(t, tc.expectedChanges, changes)
		})
	}
}
