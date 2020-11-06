package remote

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	t1, _ = time.Parse(time.RFC3339, "2020-10-05T05:00:00-04:00")
	t2, _ = time.Parse(time.RFC3339, "2020-10-10T10:00:00-04:00")
	t3, _ = time.Parse(time.RFC3339, "2020-10-15T15:00:00-04:00")
	t4, _ = time.Parse(time.RFC3339, "2020-10-20T20:00:00-04:00")

	user1 = User{
		Name:     "The Octocat",
		Email:    "octocat@github.com",
		Username: "octocat",
		WebURL:   "https://github.com/octocat",
	}

	user2 = User{
		Name:     "The Octodog",
		Email:    "octodog@github.com",
		Username: "octodog",
		WebURL:   "https://github.com/octodog",
	}

	commit1 = Commit{
		Hash: "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		Time: t1,
	}

	commit2 = Commit{
		Hash: "0251a422d2038967eeaaaa5c8aa76c7067fdef05",
		Time: t2,
	}

	branch = Branch{
		Name:   "main",
		Commit: commit2,
	}

	tag1 = Tag{
		Name:   "v0.1.0",
		Time:   t1,
		Commit: commit1,
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.1.0",
	}

	tag2 = Tag{
		Name:   "v0.2.0",
		Time:   t2,
		Commit: commit2,
		WebURL: "https://github.com/octocat/Hello-World/tree/v0.2.0",
	}

	issue1 = Issue{
		Change: Change{
			Number:    1001,
			Title:     "Found a bug",
			Labels:    []string{"bug"},
			Milestone: "v1.0",
			Time:      t1,
			Author:    user1,
			WebURL:    "https://github.com/octocat/Hello-World/issues/1001",
		},
		Closer: user1,
	}

	issue2 = Issue{
		Change: Change{
			Number:    1002,
			Title:     "Add a feature",
			Labels:    []string{"enhancement"},
			Milestone: "v1.0",
			Time:      t2,
			Author:    user1,
			WebURL:    "https://github.com/octocat/Hello-World/issues/1002",
		},
		Closer: user2,
	}

	merge1 = Merge{
		Change: Change{
			Number:    1003,
			Title:     "Fixed a bug",
			Labels:    []string{"bug"},
			Milestone: "v1.0",
			Time:      t3,
			Author:    user1,
			WebURL:    "https://github.com/octocat/Hello-World/pull/1003",
		},
		Merger: user1,
		Commit: commit1,
	}

	merge2 = Merge{
		Change: Change{
			Number:    1004,
			Title:     "Added a feature",
			Labels:    []string{"enhancement"},
			Milestone: "v1.0",
			Time:      t4,
			Author:    user1,
			WebURL:    "https://github.com/octocat/Hello-World/pull/1004",
		},
		Merger: user2,
		Commit: commit2,
	}
)

func TestCommit(t *testing.T) {
	tests := []struct {
		name           string
		c              Commit
		expectedIsZero bool
		expectedString string
	}{
		{
			name:           "Zero",
			c:              Commit{},
			expectedIsZero: true,
			expectedString: "",
		},
		{
			name:           "Commit1",
			c:              commit1,
			expectedIsZero: false,
			expectedString: "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		},
		{
			name:           "Commit2",
			c:              commit2,
			expectedIsZero: false,
			expectedString: "0251a422d2038967eeaaaa5c8aa76c7067fdef05",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedIsZero, tc.c.IsZero())
			assert.Equal(t, tc.expectedString, tc.c.String())
		})
	}
}

func TestCommits_Any(t *testing.T) {
	tests := []struct {
		name           string
		c              Commits
		hash           string
		expectedAny    bool
		expectedAll    bool
		expectedString string
	}{
		{
			name:        "Found",
			c:           Commits{commit2, commit1},
			hash:        "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
			expectedAny: true,
		},
		{
			name:        "NotFound",
			c:           Commits{commit2, commit1},
			hash:        "c414d1004154c6c324bd78c69d10ee101e676059",
			expectedAny: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedAny, tc.c.Any(tc.hash))
		})
	}
}

func TestCommits_Map(t *testing.T) {
	tests := []struct {
		name         string
		c            Commits
		f            func(Commit) string
		expectedList []string
	}{
		{
			name: "OK",
			c:    Commits{commit2, commit1},
			f: func(c Commit) string {
				return c.Hash
			},
			expectedList: []string{"0251a422d2038967eeaaaa5c8aa76c7067fdef05", "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			list := tc.c.Map(tc.f)

			assert.Equal(t, tc.expectedList, list)
		})
	}
}

func TestBranch(t *testing.T) {
	tests := []struct {
		name           string
		b              Branch
		expectedString string
	}{
		{
			name:           "Zero",
			b:              Branch{},
			expectedString: "",
		},
		{
			name:           "Branch",
			b:              branch,
			expectedString: "main",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.b.String())
		})
	}
}

func TestTag(t *testing.T) {
	tests := []struct {
		name           string
		t              Tag
		expectedIsZero bool
		expectedString string
	}{
		{
			name:           "Zero",
			t:              Tag{},
			expectedIsZero: true,
			expectedString: "",
		},
		{
			name:           "Tag1",
			t:              tag1,
			expectedIsZero: false,
			expectedString: "v0.1.0 Commit[25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378]",
		},
		{
			name:           "Tag2",
			t:              tag2,
			expectedIsZero: false,
			expectedString: "v0.2.0 Commit[0251a422d2038967eeaaaa5c8aa76c7067fdef05]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedIsZero, tc.t.IsZero())
			assert.Equal(t, tc.expectedString, tc.t.String())
		})
	}
}

func TestTag_Comparison(t *testing.T) {
	tests := []struct {
		name           string
		t1, t2         Tag
		expectedEqual  bool
		expectedBefore bool
		expectedAfter  bool
	}{
		{
			name:           "Zero",
			t1:             Tag{},
			t2:             Tag{},
			expectedEqual:  true,
			expectedBefore: false,
			expectedAfter:  false,
		},
		{
			name:           "Equal",
			t1:             tag1,
			t2:             tag1,
			expectedEqual:  true,
			expectedBefore: false,
			expectedAfter:  false,
		},
		{
			name:           "Before",
			t1:             tag1,
			t2:             tag2,
			expectedEqual:  false,
			expectedBefore: true,
			expectedAfter:  false,
		},
		{
			name:           "After",
			t1:             tag2,
			t2:             tag1,
			expectedEqual:  false,
			expectedBefore: false,
			expectedAfter:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedEqual, tc.t1.Equal(tc.t2))
			assert.Equal(t, tc.expectedBefore, tc.t1.Before(tc.t2))
			assert.Equal(t, tc.expectedAfter, tc.t1.After(tc.t2))
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

func TestTags_First(t *testing.T) {
	t1 := tag1
	t2 := tag1

	tests := []struct {
		name        string
		t           Tags
		f           func(Tag) bool
		expectedTag Tag
		expectedOK  bool
	}{
		{
			name:        "NoTagNoPredicate",
			t:           nil,
			f:           nil,
			expectedTag: Tag{},
			expectedOK:  false,
		},
		{
			name:        "NoPredicate",
			t:           Tags{t1, t2},
			f:           nil,
			expectedTag: t1,
			expectedOK:  true,
		},
		{
			name: "Found",
			t:    Tags{t1, t2},
			f: func(t Tag) bool {
				return t.Name == "v0.1.0"
			},
			expectedTag: t1,
			expectedOK:  true,
		},
		{
			name: "NotFound",
			t:    Tags{t1, t2},
			f: func(t Tag) bool {
				return t.Name == "v0.3.0"
			},
			expectedTag: Tag{},
			expectedOK:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag, ok := tc.t.First(tc.f)

			assert.Equal(t, tc.expectedTag, tag)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestTags_Last(t *testing.T) {
	t1 := tag1
	t2 := tag1

	tests := []struct {
		name        string
		t           Tags
		f           func(Tag) bool
		expectedTag Tag
		expectedOK  bool
	}{
		{
			name:        "NoTagNoPredicate",
			t:           nil,
			f:           nil,
			expectedTag: Tag{},
			expectedOK:  false,
		},
		{
			name:        "NoPredicate",
			t:           Tags{t1, t2},
			f:           nil,
			expectedTag: t2,
			expectedOK:  true,
		},
		{
			name: "Found",
			t:    Tags{t1, t2},
			f: func(t Tag) bool {
				return t.Name == "v0.1.0"
			},
			expectedTag: t2,
			expectedOK:  true,
		},
		{
			name: "NotFound",
			t:    Tags{t1, t2},
			f: func(t Tag) bool {
				return t.Name == "v0.3.0"
			},
			expectedTag: Tag{},
			expectedOK:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag, ok := tc.t.Last(tc.f)

			assert.Equal(t, tc.expectedTag, tag)
			assert.Equal(t, tc.expectedOK, ok)
		})
	}
}

func TestTags_Sort(t *testing.T) {
	tests := []struct {
		name         string
		t            Tags
		expectedTags Tags
	}{
		{
			name:         "OK",
			t:            Tags{tag1, tag2},
			expectedTags: Tags{tag2, tag1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := tc.t.Sort()

			assert.Equal(t, tc.expectedTags, tags)
		})
	}
}

func TestTags_Reverse(t *testing.T) {
	tests := []struct {
		name         string
		t            Tags
		expectedTags Tags
	}{
		{
			name:         "OK",
			t:            Tags{tag1, tag2},
			expectedTags: Tags{tag2, tag1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := tc.t.Reverse()

			assert.Equal(t, tc.expectedTags, tags)
		})
	}
}

func TestTags_Select(t *testing.T) {
	tests := []struct {
		name             string
		t                Tags
		f                func(Tag) bool
		expectedSelected Tags
	}{
		{
			name: "Named",
			t:    Tags{tag1, tag2},
			f: func(t Tag) bool {
				return len(t.Name) > 0
			},
			expectedSelected: Tags{tag1, tag2},
		},
		{
			name: "Unnamed",
			t:    Tags{tag1, tag2},
			f: func(t Tag) bool {
				return len(t.Name) == 0
			},
			expectedSelected: Tags{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			selected := tc.t.Select(tc.f)

			assert.Equal(t, tc.expectedSelected, selected)
		})
	}
}

func TestTags_Remove(t *testing.T) {
	tests := []struct {
		name            string
		t               Tags
		f               func(Tag) bool
		expectedTags    Tags
		expextedRemoved Tags
	}{
		{
			name:            "Nil",
			t:               nil,
			f:               nil,
			expectedTags:    nil,
			expextedRemoved: nil,
		},
		{
			name: "OK",
			t:    Tags{tag1, tag2},
			f: func(t Tag) bool {
				return t.Name == "v0.1.0"
			},
			expectedTags:    Tags{tag2},
			expextedRemoved: Tags{tag1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			removed := tc.t.Remove(tc.f)

			assert.Equal(t, tc.expectedTags, tc.t)
			assert.Equal(t, tc.expextedRemoved, removed)
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

func TestTags_Map(t *testing.T) {
	tests := []struct {
		name         string
		t            Tags
		f            func(Tag) string
		expectedList []string
	}{
		{
			name: "OK",
			t:    Tags{tag2, tag1},
			f: func(t Tag) string {
				return t.Name
			},
			expectedList: []string{"v0.2.0", "v0.1.0"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			list := tc.t.Map(tc.f)

			assert.Equal(t, tc.expectedList, list)
		})
	}
}

func TestLabels(t *testing.T) {
	tests := []struct {
		name           string
		l              Labels
		expectedString string
	}{
		{
			name:           "OK",
			l:              Labels{"bug", "documentation", "enhancement", "question"},
			expectedString: "bug,documentation,enhancement,question",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.l.String())
		})
	}
}

func TestLabels_Any(t *testing.T) {
	tests := []struct {
		name        string
		l           Labels
		names       []string
		expectedAny bool
		expectedAll bool
	}{
		{
			name:        "Found",
			l:           Labels{"bug", "documentation", "enhancement", "question"},
			names:       []string{"bug", "duplicate", "invalid"},
			expectedAny: true,
		},
		{
			name:        "NotFound",
			l:           Labels{"bug", "documentation", "enhancement", "question"},
			names:       []string{"duplicate", "invalid"},
			expectedAny: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedAny, tc.l.Any(tc.names...))
		})
	}
}

func TestIssues_Sort(t *testing.T) {
	tests := []struct {
		name           string
		i              Issues
		expectedIssues Issues
	}{
		{
			name:           "OK",
			i:              Issues{issue1, issue2},
			expectedIssues: Issues{issue2, issue1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			issues := tc.i.Sort()

			assert.Equal(t, tc.expectedIssues, issues)
		})
	}
}

func TestIssues_Select(t *testing.T) {
	tests := []struct {
		name             string
		i                Issues
		f                func(Issue) bool
		expectedSelected Issues
	}{
		{
			name: "Labeled",
			i:    Issues{issue1, issue2, Issue{}},
			f: func(i Issue) bool {
				return len(i.Labels) > 0
			},
			expectedSelected: Issues{issue1, issue2},
		},
		{
			name: "Unlabeled",
			i:    Issues{issue1, issue2, Issue{}},
			f: func(i Issue) bool {
				return len(i.Labels) == 0
			},
			expectedSelected: Issues{Issue{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			selected := tc.i.Select(tc.f)

			assert.Equal(t, tc.expectedSelected, selected)
		})
	}
}

func TestIssues_Remove(t *testing.T) {
	tests := []struct {
		name            string
		i               Issues
		f               func(Issue) bool
		expectedIssues  Issues
		expextedRemoved Issues
	}{
		{
			name:            "Nil",
			i:               nil,
			f:               nil,
			expectedIssues:  nil,
			expextedRemoved: nil,
		},
		{
			name: "OK",
			i:    Issues{issue1, issue2},
			f: func(i Issue) bool {
				return i.Number == 1001
			},
			expectedIssues:  Issues{issue2},
			expextedRemoved: Issues{issue1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			removed := tc.i.Remove(tc.f)

			assert.Equal(t, tc.expectedIssues, tc.i)
			assert.Equal(t, tc.expextedRemoved, removed)
		})
	}
}

func TestMerges_Sort(t *testing.T) {
	tests := []struct {
		name           string
		m              Merges
		expectedMerges Merges
	}{
		{
			name:           "OK",
			m:              Merges{merge1, merge2},
			expectedMerges: Merges{merge2, merge1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			merges := tc.m.Sort()

			assert.Equal(t, tc.expectedMerges, merges)
		})
	}
}

func TestMerges_Select(t *testing.T) {
	tests := []struct {
		name             string
		m                Merges
		f                func(Merge) bool
		expectedSelected Merges
	}{
		{
			name: "Labeled",
			m:    Merges{merge1, merge2, Merge{}},
			f: func(m Merge) bool {
				return len(m.Labels) > 0
			},
			expectedSelected: Merges{merge1, merge2},
		},
		{
			name: "Unlabeled",
			m:    Merges{merge1, merge2, Merge{}},
			f: func(m Merge) bool {
				return len(m.Labels) == 0
			},
			expectedSelected: Merges{Merge{}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			selected := tc.m.Select(tc.f)

			assert.Equal(t, tc.expectedSelected, selected)
		})
	}
}

func TestMerges_Remove(t *testing.T) {
	tests := []struct {
		name            string
		m               Merges
		f               func(Merge) bool
		expectedMerges  Merges
		expextedRemoved Merges
	}{
		{
			name:            "Nil",
			m:               nil,
			f:               nil,
			expectedMerges:  nil,
			expextedRemoved: nil,
		},
		{
			name: "OK",
			m:    Merges{merge1, merge2},
			f: func(m Merge) bool {
				return m.Number == 1003
			},
			expectedMerges:  Merges{merge2},
			expextedRemoved: Merges{merge1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			removed := tc.m.Remove(tc.f)

			assert.Equal(t, tc.expectedMerges, tc.m)
			assert.Equal(t, tc.expextedRemoved, removed)
		})
	}
}
