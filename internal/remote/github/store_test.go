package github

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIssueStore(t *testing.T) {
	tests := []struct {
		name   string
		number int
		i      issue
	}{
		{
			name:   "OK",
			number: 1001,
			i:      gitHubIssue1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newIssueStore()
			s.Save(tc.number, tc.i)
			i, ok := s.Load(tc.number)

			assert.True(t, ok)
			assert.Equal(t, tc.i, i)

			assert.NoError(t, s.ForEach(func(number int, i issue) error {
				assert.Equal(t, tc.number, number)
				assert.Equal(t, tc.i, i)
				return nil
			}))

			assert.Error(t, s.ForEach(func(number int, i issue) error {
				assert.Equal(t, tc.number, number)
				assert.Equal(t, tc.i, i)
				return errors.New("dummy")
			}))
		})
	}
}

func TestPullRequestStore(t *testing.T) {
	tests := []struct {
		name   string
		number int
		p      pullRequest
	}{
		{
			name:   "OK",
			number: 1002,
			p:      gitHubPull1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newPullRequestStore()
			s.Save(tc.number, tc.p)
			p, ok := s.Load(tc.number)

			assert.True(t, ok)
			assert.Equal(t, tc.p, p)

			assert.NoError(t, s.ForEach(func(number int, p pullRequest) error {
				assert.Equal(t, tc.number, number)
				assert.Equal(t, tc.p, p)
				return nil
			}))

			assert.Error(t, s.ForEach(func(number int, p pullRequest) error {
				assert.Equal(t, tc.number, number)
				assert.Equal(t, tc.p, p)
				return errors.New("dummy")
			}))
		})
	}
}

func TestEventStore(t *testing.T) {
	tests := []struct {
		name   string
		number int
		e      event
	}{
		{
			name:   "ClosedEvent",
			number: 1001,
			e:      gitHubEvent1,
		},
		{
			name:   "MergedEvent",
			number: 1002,
			e:      gitHubEvent2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newEventStore()
			s.Save(tc.number, tc.e)
			e, ok := s.Load(tc.number)

			assert.True(t, ok)
			assert.Equal(t, tc.e, e)

			assert.NoError(t, s.ForEach(func(number int, e event) error {
				assert.Equal(t, tc.number, number)
				assert.Equal(t, tc.e, e)
				return nil
			}))

			assert.Error(t, s.ForEach(func(number int, e event) error {
				assert.Equal(t, tc.number, number)
				assert.Equal(t, tc.e, e)
				return errors.New("dummy")
			}))
		})
	}
}

func TestCommitStore(t *testing.T) {
	tests := []struct {
		name string
		sha  string
		c    commit
	}{
		{
			name: "OK",
			sha:  "6dcb09b5b57875f334f61aebed695e2e4193db5e",
			c:    gitHubCommit,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newCommitStore()
			s.Save(tc.sha, tc.c)
			c, ok := s.Load(tc.sha)

			assert.True(t, ok)
			assert.Equal(t, tc.c, c)

			assert.NoError(t, s.ForEach(func(sha string, c commit) error {
				assert.Equal(t, tc.sha, sha)
				assert.Equal(t, tc.c, c)
				return nil
			}))

			assert.Error(t, s.ForEach(func(sha string, c commit) error {
				assert.Equal(t, tc.sha, sha)
				assert.Equal(t, tc.c, c)
				return errors.New("dummy")
			}))
		})
	}
}

func TestUserStore(t *testing.T) {
	tests := []struct {
		name     string
		username string
		u        user
	}{
		{
			name:     "OK",
			username: "octocat",
			u:        gitHubUser1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			s := newUserStore()
			s.Save(tc.username, tc.u)
			u, ok := s.Load(tc.username)

			assert.True(t, ok)
			assert.Equal(t, tc.u, u)

			assert.NoError(t, s.ForEach(func(username string, u user) error {
				assert.Equal(t, tc.username, username)
				assert.Equal(t, tc.u, u)
				return nil
			}))

			assert.Error(t, s.ForEach(func(username string, u user) error {
				assert.Equal(t, tc.username, username)
				assert.Equal(t, tc.u, u)
				return errors.New("dummy")
			}))
		})
	}
}
