package changelog

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReleaseString(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2020-10-12T21:45:00-04:00")

	tests := []struct {
		name           string
		r              Release
		expectedString string
	}{
		{
			name: "OK",
			r: Release{
				GitTag:      "v0.1.0",
				URL:         "https://github.com/moorara/changelog/tree/v0.1.0",
				Timestamp:   ts,
				IssueGroups: []IssueGroup{},
				MergeGroups: []MergeGroup{},
			},
			expectedString: "v0.1.0 https://github.com/moorara/changelog/tree/v0.1.0 2020-10-12T21:45:00-04:00\n[]\n[]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectedString, tc.r.String())
		})
	}
}

func TestNewChangelog(t *testing.T) {
	changelog := NewChangelog()

	assert.NotNil(t, changelog)
	assert.NotNil(t, changelog.Title)
	assert.Len(t, changelog.New, 0)
	assert.Len(t, changelog.Releases, 0)
}
