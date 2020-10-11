package markdown

import (
	"log"
	"testing"
	"time"

	"github.com/moorara/changelog/internal/changelog"
	"github.com/stretchr/testify/assert"
)

func TestNewProcessor(t *testing.T) {
	tests := []struct {
		name     string
		logger   *log.Logger
		filename string
	}{
		{
			name:     "OK",
			logger:   &log.Logger{},
			filename: "CHANGELOG.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			proc := NewProcessor(tc.logger, tc.filename)
			assert.NotNil(t, proc)

			p, ok := proc.(*processor)
			assert.True(t, ok)

			assert.Equal(t, tc.logger, p.logger)
			assert.Equal(t, tc.filename, p.filename)
			assert.Empty(t, p.doc)
		})
	}
}

func TestProcessorParse(t *testing.T) {
	tests := []struct {
		name              string
		p                 *processor
		opts              changelog.ParseOptions
		expectedChangelog *changelog.Changelog
		expectedError     string
	}{
		{
			name: "FileNotExist",
			p:    &processor{},
			opts: changelog.ParseOptions{},
			expectedChangelog: &changelog.Changelog{
				Title: "Changelog",
			},
			expectedError: "",
		},
		{
			name: "Success",
			p: &processor{
				filename: "test/CHANGELOG.md",
			},
			opts: changelog.ParseOptions{},
			expectedChangelog: &changelog.Changelog{
				Title: "Changelog",
				Releases: []changelog.Release{
					{
						GitTag:    "v0.1.1",
						URL:       "https://github.com/moorara/changelog/tree/v0.1.1",
						Timestamp: time.Date(2020, time.October, 11, 0, 0, 0, 0, time.UTC),
					},
					{
						GitTag:    "v0.1.0",
						URL:       "https://github.com/moorara/changelog/tree/v0.1.0",
						Timestamp: time.Date(2020, time.October, 10, 0, 0, 0, 0, time.UTC),
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

func TestProcessorRender(t *testing.T) {
	tests := []struct {
		name           string
		p              *processor
		chlog          *changelog.Changelog
		expectedString string
		expectedError  error
	}{
		{
			name:           "OK",
			p:              &processor{},
			chlog:          &changelog.Changelog{},
			expectedString: "&{Title: New:[] Releases:[]}",
			expectedError:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			str, err := tc.p.Render(tc.chlog)

			assert.Equal(t, tc.expectedString, str)
			assert.Equal(t, tc.expectedError, err)
		})
	}
}
