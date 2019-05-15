// DO NOT EDIT Code automatically generated.
//go:generate go run github.com/pierrec/packagen/cmd/packagen -v single -o set_int.go -prefix Int -mvtype Elem=int -rmtype Elem set.go
package set

// Set implements a basic set data structure.
type IntSet struct {
	data map[int]struct{}
}

func (s *IntSet) init() {
	if s.data == nil {
		s.data = make(map[int]struct{})
	}
}

// Len returns the number of elements in the set.
func (s *IntSet) Len() int {
	return len(s.data)
}

// Add adds the element to the set.
func (s *IntSet) Add(v int) bool {
	s.init()
	_, ok := s.data[v]
	if !ok {
		s.data[v] = struct{}{}
	}
	return ok
}

// Has returns whether or not the element in the set.
func (s *IntSet) Has(v int) bool {
	s.init()
	_, ok := s.data[v]
	return ok
}

// Del removes the element from the set.
func (s *IntSet) Del(v int) bool {
	s.init()
	_, ok := s.data[v]
	if ok {
		delete(s.data, v)
	}
	return ok
}

// UnionWith performs the union of the two sets into s.
func (s *IntSet) Unionwith(ss *IntSet) {
	if ss == nil || len(ss.data) == 0 {
		return
	}
	s.init()
	for v := range ss.data {
		s.data[v] = struct{}{}
	}
}

// IntersectionWith removes all the elements from s that are not in s and ss.
func (s *IntSet) Intersectionwith(ss *IntSet) {
	if ss == nil || len(ss.data) == 0 {
		s.data = nil
		return
	}
	if len(s.data) == 0 {
		return
	}
	s.init()
	for v := range s.data {
		if _, ok := ss.data[v]; !ok {
			delete(s.data, v)
		}
	}
}

// DifferenceWith sets s to all the elements not in the intersection of s and ss.
func (s *IntSet) DifferenceWith(ss *IntSet) {
	if ss == nil || len(ss.data) == 0 {
		return
	}
	s.init()
	if len(s.data) == 0 {
		for v := range ss.data {
			s.data[v] = struct{}{}
		}
		return
	}
	for v := range ss.data {
		if _, ok := s.data[v]; ok {
			delete(s.data, v)
		} else {
			s.data[v] = struct{}{}
		}
	}
}
