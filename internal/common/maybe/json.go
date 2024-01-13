package maybe

import (
	"encoding/json"
)

// NOTE: omitempty notation for JSON tag does not work on Maybe
// Since encoding/json/encode.go does not check structs for emptiness
// So, None maybe always marshals into `null` value
// See json_test.go TestJSONMarshallNone for concrete example of this case

const (
	null = "null"
)

func (m *Maybe[T]) UnmarshalJSON(bytes []byte) error {
	if string(bytes) == null {
		// reset state when null passed
		m.just = false
		var t T
		m.v = t
		return nil
	}

	var v T
	err := json.Unmarshal(bytes, &v)
	if err != nil {
		return err
	}

	m.v = v
	m.just = true

	return nil
}

func (m Maybe[T]) MarshalJSON() ([]byte, error) {
	if !m.just {
		return []byte("null"), nil
	}
	return json.Marshal(m.v)
}
