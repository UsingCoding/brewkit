package maybe

// Maybe is monad to provide clear and explicit semantic of null value
type Maybe[T any] struct {
	v    T
	just bool
}

// NewJust returns Just monad that has value
func NewJust[T any](v T) Maybe[T] {
	return Maybe[T]{
		v:    v,
		just: true,
	}
}

// NewNone used for explicit none value
func NewNone[T any]() Maybe[T] {
	return Maybe[T]{}
}

// Valid returns true when monad is Just and false when monad is None
func Valid[T any](maybe Maybe[T]) bool {
	return maybe.just
}

// Just returns underlying value of Monad on just value.
// Just panics when value is none
func Just[T any](maybe Maybe[T]) T {
	if !Valid(maybe) {
		panic("violated usage of maybe: Just on non Valid Maybe")
	}
	return maybe.v
}

// JustValid returns underlying value of Monad and
// true if Maybe is Valid
// false if Maybe is ! Valid
func JustValid[T any](maybe Maybe[T]) (v T, ok bool) {
	if !Valid(maybe) {
		ok = false
		return
	}

	return Just(maybe), true
}

// Map maps maybe underlying value to value of other type with consisting maybe state
// If maybe was Just new maybe will be Just, otherwise None
func Map[T any, E any](m Maybe[T], f func(T) E) Maybe[E] {
	if !Valid(m) {
		return Maybe[E]{}
	}

	return Maybe[E]{
		v:    f(Just(m)),
		just: true,
	}
}

// MapErr same as Map but supports errors
func MapErr[T any, E any](m Maybe[T], f func(T) (E, error)) (Maybe[E], error) {
	if !Valid(m) {
		return Maybe[E]{}, nil
	}

	v, err := f(Just(m))
	if err != nil {
		return Maybe[E]{}, err
	}

	return Maybe[E]{
		v:    v,
		just: true,
	}, nil
}

// MapNone returns underlying value on Valid Maybe or value from f
func MapNone[T any](m Maybe[T], f func() T) T {
	if !Valid(m) {
		return f()
	}
	return Just(m)
}

// MapNoneErr same as MapNone bu support errors
func MapNoneErr[T any](m Maybe[T], f func() (T, error)) (T, error) {
	if !Valid(m) {
		return f()
	}
	return Just(m), nil
}

// FromPtr creates maybe from ptr
func FromPtr[T any](t *T) Maybe[T] {
	if t == nil {
		return NewNone[T]()
	}
	return NewJust[T](*t)
}

// ToPtr map maybe to ptr
func ToPtr[T any](m Maybe[T]) *T {
	if m.just {
		return &m.v
	}
	return nil
}
