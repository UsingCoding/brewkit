package progresslabel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	label := MakeLabelf(HiddenLabel, "Create directory")

	l, payload := ParseLabel(label)

	assert.Equal(t, HiddenLabel, l)
	assert.Equal(t, "Create directory", payload)
}
