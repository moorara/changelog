package git

import (
	"regexp"
	"testing"
	"time"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

var (
	t1, _ = time.Parse(time.RFC3339, "2020-10-12T09:00:00-04:00")
	t2, _ = time.Parse(time.RFC3339, "2020-10-22T16:00:00-04:00")

	JohnDoe = Signature{
		Name:  "John Doe",
		Email: "john@doe.com",
		Time:  t1,
	}

	JaneDoe = Signature{
		Name:  "Jane Doe",
		Email: "jane@doe.com",
		Time:  t2,
	}

	commit1 = Commit{
		Hash:      "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		Author:    JohnDoe,
		Committer: JohnDoe,
		Message:   "foo",
	}

	commit2 = Commit{
		Hash:      "0251a422d2038967eeaaaa5c8aa76c7067fdef05",
		Author:    JaneDoe,
		Committer: JaneDoe,
		Message:   "bar",
	}

	tag1 = Tag{
		Type:   Lightweight,
		Hash:   "25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378",
		Name:   "v0.1.0",
		Commit: commit1,
	}

	tag2Message = "Release v0.2.0"
	tag2        = Tag{
		Type: Annotated,
		Hash: "4ff025213432eee78526e2a75f2e043d34962b5a",
		Name: "v0.2.0",
		Tagger: &Signature{
			Name:  "John Doe",
			Email: "john@doe.com",
			Time:  t1,
		},
		Message: &tag2Message,
		Commit:  commit2,
	}
)

func TestSignature(t *testing.T) {
	tests := []struct {
		name            string
		s1, s2          Signature
		expectedBefore  bool
		expectedAfter   bool
		expectedString1 string
		expectedString2 string
	}{
		{
			name:            "Before",
			s1:              JohnDoe,
			s2:              JaneDoe,
			expectedBefore:  true,
			expectedAfter:   false,
			expectedString1: "John Doe <john@doe.com> 2020-10-12T09:00:00-04:00",
			expectedString2: "Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00",
		},
		{
			name:            "After",
			s1:              JaneDoe,
			s2:              JohnDoe,
			expectedBefore:  false,
			expectedAfter:   true,
			expectedString1: "Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00",
			expectedString2: "John Doe <john@doe.com> 2020-10-12T09:00:00-04:00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedBefore, tc.s1.Before(tc.s2))
			assert.Equal(t, tc.expectedAfter, tc.s1.After(tc.s2))
			assert.Equal(t, tc.expectedString1, tc.s1.String())
			assert.Equal(t, tc.expectedString2, tc.s2.String())
		})
	}
}

func TestCommit(t *testing.T) {
	tests := []struct {
		name            string
		c1, c2          Commit
		expectedEqual   bool
		expectedBefore  bool
		expectedAfter   bool
		expectedString1 string
		expectedString2 string
	}{
		{
			name:            "Equal",
			c1:              commit1,
			c2:              commit1,
			expectedEqual:   true,
			expectedBefore:  false,
			expectedAfter:   false,
			expectedString1: "25aa2bd\nAuthor:    John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nCommitter: John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nfoo\n",
			expectedString2: "25aa2bd\nAuthor:    John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nCommitter: John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nfoo\n",
		},
		{
			name:            "Before",
			c1:              commit1,
			c2:              commit2,
			expectedEqual:   false,
			expectedBefore:  true,
			expectedAfter:   false,
			expectedString1: "25aa2bd\nAuthor:    John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nCommitter: John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nfoo\n",
			expectedString2: "0251a42\nAuthor:    Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00\nCommitter: Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00\nbar\n",
		},
		{
			name:            "After",
			c1:              commit2,
			c2:              commit1,
			expectedEqual:   false,
			expectedBefore:  false,
			expectedAfter:   true,
			expectedString1: "0251a42\nAuthor:    Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00\nCommitter: Jane Doe <jane@doe.com> 2020-10-22T16:00:00-04:00\nbar\n",
			expectedString2: "25aa2bd\nAuthor:    John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nCommitter: John Doe <john@doe.com> 2020-10-12T09:00:00-04:00\nfoo\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedEqual, tc.c1.Equal(tc.c2))
			assert.Equal(t, tc.expectedBefore, tc.c1.Before(tc.c2))
			assert.Equal(t, tc.expectedAfter, tc.c1.After(tc.c2))
			assert.Equal(t, tc.expectedString1, tc.c1.String())
			assert.Equal(t, tc.expectedString2, tc.c2.String())
		})
	}
}

func TestTagType(t *testing.T) {
	tests := []struct {
		name           string
		t              TagType
		expectedString string
	}{
		{
			name:           "Void",
			t:              Void,
			expectedString: "Void",
		},
		{
			name:           "Lightweight",
			t:              Lightweight,
			expectedString: "Lightweight",
		},
		{
			name:           "Annotated",
			t:              Annotated,
			expectedString: "Annotated",
		},
		{
			name:           "Invalid",
			t:              TagType(-1),
			expectedString: "Invalid",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.t.String())
		})
	}
}

func TestLightweightTag(t *testing.T) {
	tests := []struct {
		name        string
		ref         *plumbing.Reference
		commitObj   *object.Commit
		expectedTag Tag
	}{
		{
			name: "OK",
			ref: plumbing.NewHashReference(
				plumbing.ReferenceName("v0.1.0"),
				plumbing.NewHash("25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378"),
			),
			commitObj: &object.Commit{
				Hash: plumbing.NewHash("25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378"),
				Author: object.Signature{
					Name:  "John Doe",
					Email: "john@doe.com",
					When:  t1,
				},
				Committer: object.Signature{
					Name:  "John Doe",
					Email: "john@doe.com",
					When:  t1,
				},
				Message: "foo",
			},
			expectedTag: tag1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag := lightweightTag(tc.ref, tc.commitObj)

			assert.Equal(t, tc.expectedTag, tag)
		})
	}
}

func TestAnnotatedTag(t *testing.T) {
	t1, _ := time.Parse(time.RFC3339, "2020-10-12T09:00:00-04:00")
	t2, _ := time.Parse(time.RFC3339, "2020-10-22T16:00:00-04:00")

	tests := []struct {
		name        string
		tagObj      *object.Tag
		commitObj   *object.Commit
		expectedTag Tag
	}{
		{
			name: "OK",
			tagObj: &object.Tag{
				Hash: plumbing.NewHash("4ff025213432eee78526e2a75f2e043d34962b5a"),
				Name: "v0.2.0",
				Tagger: object.Signature{
					Name:  "John Doe",
					Email: "john@doe.com",
					When:  t1,
				},
				Message: "Release v0.2.0",
			},
			commitObj: &object.Commit{
				Hash: plumbing.NewHash("0251a422d2038967eeaaaa5c8aa76c7067fdef05"),
				Author: object.Signature{
					Name:  "Jane Doe",
					Email: "jane@doe.com",
					When:  t2,
				},
				Committer: object.Signature{
					Name:  "Jane Doe",
					Email: "jane@doe.com",
					When:  t2,
				},
				Message: "bar",
			},
			expectedTag: tag2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag := annotatedTag(tc.tagObj, tc.commitObj)

			assert.Equal(t, tc.expectedTag, tag)
		})
	}
}

func TestTag(t *testing.T) {
	tests := []struct {
		name            string
		t1, t2          Tag
		expectedEqual   bool
		expectedBefore  bool
		expectedAfter   bool
		expectedString1 string
		expectedString2 string
	}{
		{
			name:            "Equal",
			t1:              tag1,
			t2:              tag1,
			expectedEqual:   true,
			expectedBefore:  false,
			expectedAfter:   false,
			expectedString1: "Lightweight 25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 v0.1.0 Commit[25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 foo]",
			expectedString2: "Lightweight 25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 v0.1.0 Commit[25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 foo]",
		},
		{
			name:            "Before",
			t1:              tag1,
			t2:              tag2,
			expectedEqual:   false,
			expectedBefore:  true,
			expectedAfter:   false,
			expectedString1: "Lightweight 25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 v0.1.0 Commit[25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 foo]",
			expectedString2: "Annotated 4ff025213432eee78526e2a75f2e043d34962b5a v0.2.0 Commit[0251a422d2038967eeaaaa5c8aa76c7067fdef05 bar]",
		},
		{
			name:            "After",
			t1:              tag2,
			t2:              tag1,
			expectedEqual:   false,
			expectedBefore:  false,
			expectedAfter:   true,
			expectedString1: "Annotated 4ff025213432eee78526e2a75f2e043d34962b5a v0.2.0 Commit[0251a422d2038967eeaaaa5c8aa76c7067fdef05 bar]",
			expectedString2: "Lightweight 25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 v0.1.0 Commit[25aa2bdbaf10fa30b6db40c2c0a15d280ad9f378 foo]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
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
