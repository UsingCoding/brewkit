package channel

// ProxyIn proxies in chan to out through mapping function
func ProxyIn[T, E any](in <-chan T, f func(T) E) <-chan E {
	out := make(chan E)

	go func() {
		for {
			v, ok := <-in
			if !ok {
				// in closed
				close(out)
				return
			}

			e := f(v)
			out <- e
		}
	}()

	return out
}
