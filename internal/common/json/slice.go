package json

import (
	"encoding/json"
)

// Slice - allows to unmarshal single record or many in one slice
type Slice[T any] []T

func (s *Slice[T]) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		// With no input, we preserve the existing value by returning nil and
		// leaving the target alone. This allows defining default values for
		// the type.
		return nil
	}

	// trying to unmarshal as slice of elements
	p := make([]T, 0, 1)
	err := json.Unmarshal(b, &p)
	if err == nil {
		*s = p
		return nil
	}

	var t T
	err = json.Unmarshal(b, &t)
	if err == nil {
		p = append(p, t)
		*s = p
		return nil
	}

	return err
}
