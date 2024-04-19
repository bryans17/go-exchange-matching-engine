package main

type int_set struct {
	m map[uint32]bool
}

func (s int_set) insert(key uint32) {
	s.m[key] = true
}

func (s int_set) has(key uint32) bool {
	_, ok := s.m[key]
	return ok
}

func (s int_set) remove(key uint32) {
	delete(s.m, key)
}
