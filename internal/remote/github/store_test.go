package github

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/moorara/go-github"
)

func TestStore(t *testing.T) {
	tests := []struct {
		name string
		key  interface{}
		val  interface{}
	}{
		{
			name: "Int",
			key:  1000,
			val: github.Issue{
				Number: 1000,
			},
		},
		{
			name: "String",
			key:  "octocat",
			val: github.User{
				Login: "octocat",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newStore()
			s.Save(tc.key, tc.val)

			v, ok := s.Load(tc.key)
			assert.True(t, ok)
			assert.Equal(t, tc.val, v)

			l := s.Len()
			assert.Equal(t, 1, l)

			assert.NoError(t, s.ForEach(func(k, v interface{}) error {
				assert.Equal(t, tc.key, k)
				assert.Equal(t, tc.val, v)
				return nil
			}))

			assert.Error(t, s.ForEach(func(k, v interface{}) error {
				assert.Equal(t, tc.key, k)
				assert.Equal(t, tc.val, v)
				return errors.New("dummy")
			}))
		})
	}
}
