package utils

// No sets?
type Set[T comparable] map[T]struct{}

func NewSet[T comparable]() Set[T] {
	return make(Set[T])
}

// Return whether or not the value was added (false if the set already had the value).
func (s Set[T]) Add(obj T) bool {
	_, exists := s[obj]
	if exists {
		return false
	}
	s[obj] = struct{}{}
	return true
}

// Return whether or not the value was removed (false if the set didn't have the value).
func (s Set[T]) Remove(obj T) bool {
	_, exists := s[obj]
	if !exists {
		return false
	}
	delete(s, obj)
	return true
}

func (s Set[T]) Contains(obj T) bool {
	_, exists := s[obj]
	return exists
}
