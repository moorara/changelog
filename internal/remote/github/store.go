package github

import "sync"

type userStore struct {
	sync.Mutex
	m map[string]user
}

func newUserStore() *userStore {
	return &userStore{
		m: make(map[string]user),
	}
}

func (s *userStore) Save(username string, u user) {
	s.Lock()
	defer s.Unlock()

	s.m[username] = u
}

func (s *userStore) Load(username string) (user, bool) {
	s.Lock()
	defer s.Unlock()

	u, ok := s.m[username]
	return u, ok
}

func (s *userStore) ForEach(f func(string, user) error) error {
	s.Lock()
	defer s.Unlock()

	for username, u := range s.m {
		if err := f(username, u); err != nil {
			return err
		}
	}

	return nil
}

type tagStore struct {
	sync.Mutex
	m map[string]tag
}

func newTagStore() *tagStore {
	return &tagStore{
		m: make(map[string]tag),
	}
}

func (s *tagStore) Save(name string, t tag) {
	s.Lock()
	defer s.Unlock()

	s.m[name] = t
}

func (s *tagStore) Load(name string) (tag, bool) {
	s.Lock()
	defer s.Unlock()

	c, ok := s.m[name]
	return c, ok
}

func (s *tagStore) ForEach(f func(string, tag) error) error {
	s.Lock()
	defer s.Unlock()

	for name, t := range s.m {
		if err := f(name, t); err != nil {
			return err
		}
	}

	return nil
}

type commitStore struct {
	sync.Mutex
	m map[string]commit
}

func newCommitStore() *commitStore {
	return &commitStore{
		m: make(map[string]commit),
	}
}

func (s *commitStore) Save(sha string, c commit) {
	s.Lock()
	defer s.Unlock()

	s.m[sha] = c
}

func (s *commitStore) Load(sha string) (commit, bool) {
	s.Lock()
	defer s.Unlock()

	c, ok := s.m[sha]
	return c, ok
}

func (s *commitStore) ForEach(f func(string, commit) error) error {
	s.Lock()
	defer s.Unlock()

	for sha, c := range s.m {
		if err := f(sha, c); err != nil {
			return err
		}
	}

	return nil
}

type issueStore struct {
	sync.Mutex
	m map[int]issue
}

func newIssueStore() *issueStore {
	return &issueStore{
		m: make(map[int]issue),
	}
}

func (s *issueStore) Save(number int, i issue) {
	s.Lock()
	defer s.Unlock()

	s.m[number] = i
}

func (s *issueStore) Load(number int) (issue, bool) {
	s.Lock()
	defer s.Unlock()

	i, ok := s.m[number]
	return i, ok
}

func (s *issueStore) ForEach(f func(int, issue) error) error {
	s.Lock()
	defer s.Unlock()

	for number, i := range s.m {
		if err := f(number, i); err != nil {
			return err
		}
	}

	return nil
}

type pullStore struct {
	sync.Mutex
	m map[int]pull
}

func newPullStore() *pullStore {
	return &pullStore{
		m: make(map[int]pull),
	}
}

func (s *pullStore) Save(number int, p pull) {
	s.Lock()
	defer s.Unlock()

	s.m[number] = p
}

func (s *pullStore) Load(number int) (pull, bool) {
	s.Lock()
	defer s.Unlock()

	p, ok := s.m[number]
	return p, ok
}

func (s *pullStore) ForEach(f func(int, pull) error) error {
	s.Lock()
	defer s.Unlock()

	for number, p := range s.m {
		if err := f(number, p); err != nil {
			return err
		}
	}

	return nil
}

type eventStore struct {
	sync.Mutex
	m map[int]event
}

func newEventStore() *eventStore {
	return &eventStore{
		m: make(map[int]event),
	}
}

func (s *eventStore) Save(number int, e event) {
	s.Lock()
	defer s.Unlock()

	s.m[number] = e
}

func (s *eventStore) Load(number int) (event, bool) {
	s.Lock()
	defer s.Unlock()

	e, ok := s.m[number]
	return e, ok
}

func (s *eventStore) ForEach(f func(int, event) error) error {
	s.Lock()
	defer s.Unlock()

	for number, e := range s.m {
		if err := f(number, e); err != nil {
			return err
		}
	}

	return nil
}
