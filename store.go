package todo

import "slices"

type Store struct {
	Lists []*List
}

func (s *Store) AddList(name string) {
	s.Lists = append(s.Lists, NewList(name))
}

func (s *Store) RemoveList(i int) {
	s.Lists = slices.Delete(s.Lists, i, i+1)
}
