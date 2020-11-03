package changelog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChangelog(t *testing.T) {
	changelog := NewChangelog()

	assert.NotNil(t, changelog)
	assert.NotNil(t, changelog.Title)
	assert.Len(t, changelog.New, 0)
	assert.Len(t, changelog.Existing, 0)
}
