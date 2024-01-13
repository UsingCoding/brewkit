package maybe

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONMarshal(t *testing.T) {
	v := struct {
		M Maybe[int] `json:"m"`
	}{
		M: NewJust(42),
	}

	data, err := json.Marshal(v)
	assert.NoError(t, err)

	assert.Equal(t, `{"m":42}`, string(data))
}

func TestJSONMarshallNone(t *testing.T) {
	v := struct {
		M Maybe[int] `json:"m"`
	}{}

	data, err := json.Marshal(v)
	assert.NoError(t, err)

	assert.Equal(t, `{"m":null}`, string(data))
}

func TestJSONUnmarshal(t *testing.T) {
	var v struct {
		M Maybe[int] `json:"m"`
	}
	err := json.Unmarshal([]byte(`{"m":42}`), &v)
	assert.NoError(t, err)

	assert.True(t, Valid(v.M))
	assert.Equal(t, 42, Just(v.M))

	err = json.Unmarshal([]byte(`{"m":null}`), &v)
	assert.NoError(t, err)

	assert.False(t, Valid(v.M))

	err = json.Unmarshal([]byte(`{}`), &v)
	assert.NoError(t, err)

	assert.False(t, Valid(v.M))
}
