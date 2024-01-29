package maps

func FromSlice[K comparable, V, E any](s []V, f func(V) (K, E)) map[K]E {
	res := map[K]E{}
	for _, v := range s {
		k, newV := f(v)
		res[k] = newV
	}
	return res
}

func FromSliceErr[K comparable, V, E any](s []V, f func(V) (K, E, error)) (map[K]E, error) {
	res := map[K]E{}
	for _, v := range s {
		k, newV, err := f(v)
		if err != nil {
			return nil, err
		}
		res[k] = newV
	}
	return res, nil
}

func ToSlice[K comparable, V, E any](m map[K]E, f func(K, E) V) []V {
	res := make([]V, 0, len(m))
	for k, e := range m {
		res = append(res, f(k, e))
	}
	return res
}

func Map[K1 comparable, V1 any, K2 comparable, V2 any](m map[K1]V1, f func(K1, V1) (K2, V2)) map[K2]V2 {
	res := make(map[K2]V2, len(m))
	for k1, v1 := range m {
		k2, v2 := f(k1, v1)
		res[k2] = v2
	}
	return res
}
