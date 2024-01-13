package maps

// SubtractSet subtracts hashsets from map
// Result map contains values from source map
func SubtractSet[K comparable, V any](m map[K]V, sets ...Set[K]) map[K]V {
	if len(m) == 0 {
		return nil
	}
	if len(sets) == 0 {
		return m
	}

	res := make(map[K]V, len(m))
	for k, v := range m {
		res[k] = v
	}

	for i := 0; i < len(sets); i++ {
		m := sets[i]
		for k := range res {
			if _, exists := m[k]; exists {
				delete(res, k)
			}
		}
	}

	return res
}
