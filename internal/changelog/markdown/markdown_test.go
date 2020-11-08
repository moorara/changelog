package markdown

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/moorara/changelog/pkg/log"
)

var (
	tagTime, _ = time.Parse(time.RFC3339, "2020-11-02T22:00:00-04:00")
	chlog      = &changelog.Changelog{
		New: []changelog.Release{
			{
				TagName:    "v0.2.0",
				TagURL:     "https://github.com/octocat/Hello-World/tree/v0.2.0",
				TagTime:    tagTime,
				CompareURL: "https://github.com/octocat/Hello-World/compare/v0.1.0...v0.2.0",
				IssueGroups: []changelog.IssueGroup{
					{
						Title: "Fixed Bugs",
						Issues: []changelog.Issue{
							{
								Number: 1001,
								Title:  "Fixed a bug",
								URL:    "https://github.com/octocat/Hello-World/issues/1001",
								OpenedBy: changelog.User{
									Name:     "The Octocat",
									Username: "octocat",
									URL:      "https://github.com/octocat",
								},
								ClosedBy: changelog.User{
									Name:     "The Octocat",
									Username: "octocat",
									URL:      "https://github.com/octocat",
								},
							},
						},
					},
				},
				MergeGroups: []changelog.MergeGroup{
					{
						Title: "Merged Changes",
						Merges: []changelog.Merge{
							{
								Number: 1002,
								Title:  "Add a feature",
								URL:    "https://github.com/octocat/Hello-World/pull/1002",
								OpenedBy: changelog.User{
									Name:     "The Octocat",
									Username: "octocat",
									URL:      "https://github.com/octocat",
								},
								MergedBy: changelog.User{
									Name:     "The Octodog",
									Username: "octodog",
									URL:      "https://github.com/octodog",
								},
							},
						},
					},
				},
			},
		},
	}
)

const expectedContent = `
## [v0.2.0](https://github.com/octocat/Hello-World/tree/v0.2.0) (2020-11-02)

[Compare Changes](https://github.com/octocat/Hello-World/compare/v0.1.0...v0.2.0)

**Fixed Bugs:**

  - Fixed a bug [#1001](https://github.com/octocat/Hello-World/issues/1001) ([octocat](https://github.com/octocat))

**Merged Changes:**

  - Add a feature [#1002](https://github.com/octocat/Hello-World/pull/1002) ([octocat](https://github.com/octocat), [octodog](https://github.com/octodog))


`

func TestNewProcessor(t *testing.T) {
	tests := []struct {
		name     string
		logger   log.Logger
		filename string
	}{
		{
			name:     "OK",
			logger:   log.New(log.None),
			filename: "CHANGELOG.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := NewProcessor(tc.logger, tc.filename)
			assert.NotNil(t, p)

			mp, ok := p.(*processor)
			assert.True(t, ok)

			assert.Equal(t, tc.logger, mp.logger)
			assert.Equal(t, tc.filename, mp.filename)
		})
	}
}

func TestProcessor_Parse(t *testing.T) {
	tests := []struct {
		name              string
		p                 *processor
		opts              changelog.ParseOptions
		expectedChangelog *changelog.Changelog
		expectedError     string
	}{
		{
			name: "FileNotExist",
			p: &processor{
				logger: log.New(log.None),
			},
			opts: changelog.ParseOptions{},
			expectedChangelog: &changelog.Changelog{
				Title: "Changelog",
			},
			expectedError: "",
		},
		{
			name: "Success",
			p: &processor{
				logger:   log.New(log.None),
				filename: "test/CHANGELOG.md",
			},
			opts: changelog.ParseOptions{},
			expectedChangelog: &changelog.Changelog{
				Title: "Changelog",
				Existing: []changelog.Release{
					{
						TagName: "v0.1.1",
						TagURL:  "https://github.com/octocat/Hello-World/tree/v0.1.1",
						TagTime: time.Date(2020, time.October, 11, 0, 0, 0, 0, time.UTC),
					},
					{
						TagName: "v0.1.0",
						TagURL:  "https://github.com/octocat/Hello-World/tree/v0.1.0",
						TagTime: time.Date(2020, time.October, 10, 0, 0, 0, 0, time.UTC),
					},
				},
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			chlog, err := tc.p.Parse(tc.opts)

			if tc.expectedError == "" {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedChangelog, chlog)
			} else {
				assert.Nil(t, chlog)
				assert.EqualError(t, err, tc.expectedError)
			}
		})
	}
}

func TestProcessor_Render(t *testing.T) {
	tests := []struct {
		name            string
		p               *processor
		chlog           *changelog.Changelog
		expectedContent string
		expectedError   error
	}{
		{
			name: "OK",
			p: &processor{
				logger: log.New(log.None),
			},
			chlog:           chlog,
			expectedContent: expectedContent,
			expectedError:   nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			str, err := tc.p.Render(tc.chlog)

			assert.Equal(t, tc.expectedContent, str)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
