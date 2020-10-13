package git

import (
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
			name:           "Lightweight",
			t:              Lightweight,
			expectedString: "Lightweight",
		},
		{
			name:           "Annotated",
			t:              Annotated,
			expectedString: "Annotated",
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
	ts, _ := time.Parse(time.RFC3339, "2020-10-12T21:45:00-04:00")

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
	ts, _ := time.Parse(time.RFC3339, "2020-10-12T21:45:00-04:00")

	tests := []struct {
		name        string
		tagObj      *object.Tag
		expectedTag Tag
	}{
		{
			name: "Annotated",
			tagObj: &object.Tag{
				Hash: plumbing.NewHash("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"),
				Name: "v0.1.0",
				Tagger: object.Signature{
					Name:  "Jane Doe",
					Email: "jane@doe.com",
					When:  ts,
				},
			},
			expectedTag: Tag{
				Type: Annotated,
				Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
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
			tag := annotatedTag(tc.tagObj)

			assert.Equal(t, tc.expectedTag, tag)
		})
	}
}

func TestTagString(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2020-10-12T21:45:00-04:00")

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
					Timestamp: ts,
				},
			},
			expectedString: "Lightweight aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa v0.1.0 [Jane Doe <jane@doe.com> 2020-10-12T21:45:00-04:00]",
		},
		{
			name: "Annotated",
			t: Tag{
				Type: Annotated,
				Hash: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
				Name: "v0.1.0",
				Tagger: Tagger{
					Name:      "Jane Doe",
					Email:     "jane@doe.com",
					Timestamp: ts,
				},
			},
			expectedString: "Annotated bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb v0.1.0 [Jane Doe <jane@doe.com> 2020-10-12T21:45:00-04:00]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.t.String())
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

func TestRepoGetRemoteInfo(t *testing.T) {
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

func TestRepoTags(t *testing.T) {
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
