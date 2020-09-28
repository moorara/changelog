package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	tests := []struct {
		Version        string
		Commit         string
		Branch         string
		GoVersion      string
		BuildTool      string
		BuildTime      string
		expectedString string
	}{
		{
			Version:   "0.1.0",
			Commit:    "aaaaaaa",
			Branch:    "master",
			GoVersion: "1.15",
			BuildTool: "go",
			BuildTime: "2020-09-20T15:00:00",
			expectedString: `
  version:    0.1.0
  commit:     aaaaaaa
  branch:     master
  goVersion:  1.15
  buildTool:  go
  buildTime:  2020-09-20T15:00:00
`,
		},
	}

	for _, tc := range tests {
		Version = tc.Version
		Commit = tc.Commit
		Branch = tc.Branch
		GoVersion = tc.GoVersion
		BuildTool = tc.BuildTool
		BuildTime = tc.BuildTime

		assert.Contains(t, tc.expectedString, String())
	}
}
