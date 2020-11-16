package github

import "sync"

type store struct {
	sync.Mutex
	m map[interface{}]interface{}
}

func newStore() *store {
	return &store{
		m: make(map[interface{}]interface{}),
	}
}

func (s *store) Save(key interface{}, val interface{}) {
	s.Lock()
	defer s.Unlock()

	s.m[key] = val
}

func (s *store) Load(key interface{}) (interface{}, bool) {
	s.Lock()
	defer s.Unlock()

	val, ok := s.m[key]
	return val, ok
}

func (s *store) Len() int {
	s.Lock()
	defer s.Unlock()

	return len(s.m)
}

func (s *store) ForEach(f func(interface{}, interface{}) error) error {
	s.Lock()
	defer s.Unlock()

	for k, v := range s.m {
		if err := f(k, v); err != nil {
			return err
		}
	}

	return nil
}
