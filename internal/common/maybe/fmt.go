package maybe

import (
	"fmt"
)

func (m Maybe[T]) String() string {
	if !Valid(m) {
		return ""
	}

	// Convert to empty interface to make the go compiler satisfied
	var v interface{} = Just(m)

	return fmt.Sprintf("%s", v)
}

func (m Maybe[T]) GoString() string {
	if !Valid(m) {
		return ""
	}

	var v interface{} = Just(m)

	return fmt.Sprintf("%#v", v)
}
