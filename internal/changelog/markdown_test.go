package changelog

import (
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMarkdownProcessor(t *testing.T) {
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
			proc := NewMarkdownProcessor(tc.logger, tc.filename)
			assert.NotNil(t, proc)

			p, ok := proc.(*markdownProcessor)
			assert.True(t, ok)

			assert.Equal(t, tc.logger, p.logger)
			assert.Equal(t, tc.filename, p.filename)
			assert.Empty(t, p.doc)
		})
	}
}

func TestMarkdownProcessorParse(t *testing.T) {
	tests := []struct {
		name              string
		p                 *markdownProcessor
		opts              ParseOptions
		expectedChangelog *Changelog
		expectedError     string
	}{
		{
			name: "FileNotExist",
			p:    &markdownProcessor{},
			opts: ParseOptions{},
			expectedChangelog: &Changelog{
				Title: "Changelog",
			},
			expectedError: "",
		},
		{
			name: "Success",
			p: &markdownProcessor{
				filename: "test/CHANGELOG.md",
			},
			opts: ParseOptions{},
			expectedChangelog: &Changelog{
				Title: "Changelog",
				Releases: []Release{
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

func TestMarkdownProcessorRender(t *testing.T) {
	tests := []struct {
		name           string
		p              *markdownProcessor
		chlog          *Changelog
		expectedString string
		expectedError  error
	}{
		{
			name:           "OK",
			p:              &markdownProcessor{},
			chlog:          &Changelog{},
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
