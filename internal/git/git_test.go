package git

import (
	"regexp"
	"testing"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
)

func TestTagType(t *testing.T) {
	tests := []struct {
		name           string
		t              TagType
		expectedString string
	}{
		{
			name:           "Void",
			t:              Void,
			expectedString: "",
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

func TestTagger(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2020-10-12T21:45:00-04:00")

	tests := []struct {
		name           string
		t              Tagger
		expectedString string
	}{
		{
			name: "OK",
			t: Tagger{
				Name:      "Jane Doe",
				Email:     "jane@doe.com",
				Timestamp: ts,
			},
			expectedString: "Jane Doe <jane@doe.com> 2020-10-12T21:45:00-04:00",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.t.String())
		})
	}
}

func TestLightweightTag(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2020-10-10T10:00:00-04:00")

	tests := []struct {
		name        string
		ref         *plumbing.Reference
		commit      *object.Commit
		expectedTag Tag
	}{
		{
			name: "Lightweight",
			ref: plumbing.NewHashReference(
				plumbing.ReferenceName("v0.1.0"),
				plumbing.NewHash("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			),
			commit: &object.Commit{
				Committer: object.Signature{
					Name:  "Jane Doe",
					Email: "jane@doe.com",
					When:  ts,
				},
			},
			expectedTag: Tag{
				Type: Lightweight,
				Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Name: "v0.1.0",
				Tagger: Tagger{
					Name:      "Jane Doe",
					Email:     "jane@doe.com",
					Timestamp: ts,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag := lightweightTag(tc.ref, tc.commit)

			assert.Equal(t, tc.expectedTag, tag)
		})
	}
}

func TestAnnotatedTag(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2020-10-11T11:00:00-04:00")

	tests := []struct {
		name        string
		tagObj      *object.Tag
		expectedTag Tag
	}{
		{
			name: "Annotated",
			tagObj: &object.Tag{
				Hash: plumbing.NewHash("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
				Name: "v0.1.1",
				Tagger: object.Signature{
					Name:  "Jane Doe",
					Email: "jane@doe.com",
					When:  ts,
				},
			},
			expectedTag: Tag{
				Type: Annotated,
				Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				Name: "v0.1.1",
				Tagger: Tagger{
					Name:      "Jane Doe",
					Email:     "jane@doe.com",
					Timestamp: ts,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tag := annotatedTag(tc.tagObj)

			assert.Equal(t, tc.expectedTag, tag)
		})
	}
}

func TestTag_Comparisons(t *testing.T) {
	ts1, _ := time.Parse(time.RFC3339, "2020-10-10T10:00:00-04:00")
	ts2, _ := time.Parse(time.RFC3339, "2020-10-11T11:00:00-04:00")

	tag1 := Tag{
		Type: Lightweight,
		Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		Name: "v0.1.0",
		Tagger: Tagger{
			Name:      "Jane Doe",
			Email:     "jane@doe.com",
			Timestamp: ts1,
		},
	}

	tag2 := Tag{
		Type: Annotated,
		Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		Name: "v0.1.1",
		Tagger: Tagger{
			Name:      "Jane Doe",
			Email:     "jane@doe.com",
			Timestamp: ts2,
		},
	}

	tests := []struct {
		name           string
		t, u           Tag
		expectedEqual  bool
		expectedBefore bool
		expectedAfter  bool
	}{
		{
			name:           "Case#1",
			t:              tag1,
			u:              tag1,
			expectedEqual:  true,
			expectedBefore: false,
			expectedAfter:  false,
		},
		{
			name:           "Case#2",
			t:              tag1,
			u:              tag2,
			expectedEqual:  false,
			expectedBefore: true,
			expectedAfter:  false,
		},
		{
			name:           "Case#3",
			t:              tag2,
			u:              tag1,
			expectedEqual:  false,
			expectedBefore: false,
			expectedAfter:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("Equal", func(t *testing.T) {
				assert.Equal(t, tc.expectedEqual, tc.t.Equal(tc.u))
			})

			t.Run("Before", func(t *testing.T) {
				assert.Equal(t, tc.expectedBefore, tc.t.Before(tc.u))
			})

			t.Run("After", func(t *testing.T) {
				assert.Equal(t, tc.expectedAfter, tc.t.After(tc.u))
			})
		})
	}
}

func TestTag_String(t *testing.T) {
	ts1, _ := time.Parse(time.RFC3339, "2020-10-10T10:00:00-04:00")
	ts2, _ := time.Parse(time.RFC3339, "2020-10-11T11:00:00-04:00")
	ts3, _ := time.Parse(time.RFC3339, "2020-10-12T12:00:00-04:00")

	tests := []struct {
		name           string
		t              Tag
		expectedString string
	}{
		{
			name: "Lightweight",
			t: Tag{
				Type: Lightweight,
				Hash: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				Name: "v0.1.0",
				Tagger: Tagger{
					Name:      "Jane Doe",
					Email:     "jane@doe.com",
					Timestamp: ts1,
				},
			},
			expectedString: "Lightweight aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa v0.1.0 [Jane Doe <jane@doe.com> 2020-10-10T10:00:00-04:00]",
		},
		{
			name: "Annotated",
			t: Tag{
				Type: Annotated,
				Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				Name: "v0.1.1",
				Tagger: Tagger{
					Name:      "Jane Doe",
					Email:     "jane@doe.com",
					Timestamp: ts2,
				},
			},
			expectedString: "Annotated bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb v0.1.1 [Jane Doe <jane@doe.com> 2020-10-11T11:00:00-04:00]",
		},
		{
			name: "Excluded",
			t: Tag{
				Type: Lightweight,
				Hash: "cccccccccccccccccccccccccccccccccccccccc",
				Name: "v0.1.2",
				Tagger: Tagger{
					Name:      "Jane Doe",
					Email:     "jane@doe.com",
					Timestamp: ts3,
				},
			},
			expectedString: "Lightweight cccccccccccccccccccccccccccccccccccccccc v0.1.2 [Jane Doe <jane@doe.com> 2020-10-12T12:00:00-04:00]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.t.String())
		})
	}
}

func TestTags_Index(t *testing.T) {
	tag1 := Tag{Name: "v0.1.0"}
	tag2 := Tag{Name: "v0.1.1"}

	tests := []struct {
		name          string
		t             Tags
		tagName       string
		expectedIndex int
	}{
		{
			name:          "Found",
			t:             Tags{tag1, tag2},
			tagName:       "v0.1.1",
			expectedIndex: 1,
		},
		{
			name:          "NotFound",
			t:             Tags{tag1, tag2},
			tagName:       "v0.1.2",
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
	tag1 := Tag{Name: "v0.1.0"}
	tag2 := Tag{Name: "v0.1.1"}

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
			tagName:     "v0.1.1",
			expectedTag: tag2,
			expectedOK:  true,
		},
		{
			name:        "NotFound",
			t:           Tags{tag1, tag2},
			tagName:     "v0.1.2",
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
	ts1, _ := time.Parse(time.RFC3339, "2020-10-10T10:00:00-04:00")
	ts2, _ := time.Parse(time.RFC3339, "2020-10-11T11:00:00-04:00")

	tag1 := Tag{
		Name: "v0.1.0",
		Tagger: Tagger{
			Timestamp: ts1,
		},
	}

	tag2 := Tag{
		Name: "v0.1.1",
		Tagger: Tagger{
			Timestamp: ts2,
		},
	}

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
			name: "OK",
			t: Tags{
				{Name: "v0.1.0"},
				{Name: "v0.1.1-alpha"},
				{Name: "v0.1.1-beta"},
			},
			names: []string{"v0.1.1-alpha", "v0.1.1-beta"},
			expectedTags: Tags{
				{Name: "v0.1.0"},
			},
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
			name: "OK",
			t: Tags{
				{Name: "v0.1.0"},
				{Name: "v0.1.1-alpha"},
				{Name: "v0.1.1-beta"},
			},
			regex: regexp.MustCompile(`v\d\.\d\.\d-\w+`),
			expectedTags: Tags{
				{Name: "v0.1.0"},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tags := tc.t.ExcludeRegex(tc.regex)

			assert.Equal(t, tc.expectedTags, tags)
		})
	}
}

func TestNewRepo(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "OK",
			path: ".",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r, err := NewRepo(tc.path)

			assert.NoError(t, err)
			assert.NotNil(t, r)
			assert.NotNil(t, r.git)
		})
	}
}

func TestRepo_GetRemoteInfo(t *testing.T) {
	g, err := git.PlainOpen("../..")
	assert.NoError(t, err)

	tests := []struct {
		name           string
		git            *git.Repository
		expectedDomain string
		expectedPath   string
		expectedError  string
	}{
		{
			name:           "OK",
			git:            g,
			expectedDomain: "github.com",
			expectedPath:   "moorara/changelog",
			expectedError:  "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &Repo{git: tc.git}
			domain, path, err := r.GetRemoteInfo()

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedDomain, domain)
				assert.Equal(t, tc.expectedPath, path)
			} else {
				assert.Empty(t, domain)
				assert.Empty(t, path)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestRepo_Tags(t *testing.T) {
	g, err := git.PlainOpen("../..")
	assert.NoError(t, err)

	tests := []struct {
		name          string
		git           *git.Repository
		expectedError string
	}{
		{
			name:          "OK",
			git:           g,
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := &Repo{git: tc.git}
			tags, err := r.Tags()

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.NotNil(t, tags)
			} else {
				assert.Nil(t, tags)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}
