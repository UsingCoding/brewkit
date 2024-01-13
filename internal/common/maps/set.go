package maps

// Set on map with empty struct as value
type Set[T comparable] map[T]struct{}

func (s *Set[T]) Add(v T) {
	(*s)[v] = struct{}{}
}

func (s *Set[T]) Remove(v T) {
	delete(*s, v)
}

func (s *Set[T]) Has(v T) bool {
	_, has := (*s)[v]
	return has
}

func SetFromSlice[T any, E comparable](s []T, f func(T) E) Set[E] {
	return FromSlice(s, func(v T) (E, struct{}) {
		return f(v), struct{}{}
	})
}

// FromMapKeys creates set from maps keys
func FromMapKeys[K comparable, V any](maps ...map[K]V) Set[K] {
	res := Set[K]{}
	for _, m := range maps {
		for k := range m {
			res.Add(k)
		}
	}
	return res
}
