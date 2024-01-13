package maybe

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewJust(t *testing.T) {
	m := NewJust(42)

	assert.Equal(t, 42, m.v)
	assert.True(t, m.just)
}

func TestNewNone(t *testing.T) {
	m := NewNone[int]()
	var zeroInt int

	assert.False(t, m.just)
	assert.Equal(t, zeroInt, m.v) // explicit none creates zero value for underlying type
}

func TestValid(t *testing.T) {
	valid := NewJust(42)
	none := NewNone[int]()

	assert.True(t, Valid(valid))
	assert.False(t, Valid(none))
}

func TestJust(t *testing.T) {
	v := 42
	m := NewJust(v)

	assert.Equal(t, v, Just(m)) // just should return same value

	// Just on none Maybe panics
	assert.Panics(t, func() {
		none := NewNone[int]()
		Just(none)
	})
}

func TestJustValid(t *testing.T) {
	v := 42
	m := NewJust(v)

	value, ok := JustValid(m)
	assert.Equal(t, v, value)
	assert.True(t, ok)

	none := NewNone[int]()
	_, ok = JustValid(none)
	assert.False(t, ok)
}

func TestMap(t *testing.T) {
	m1 := NewJust(42)

	m2 := Map(m1, func(value int) string {
		return fmt.Sprintf("%d", value)
	})

	assert.True(t, Valid(m2))
	assert.Equal(t, "42", Just(m2))

	m1 = NewNone[int]()

	m2 = Map(m1, func(value int) string {
		return fmt.Sprintf("%d", value)
	})

	assert.False(t, Valid(m2))
}

func TestMapErr(t *testing.T) {
	specialErr := stderrors.New("special error")

	m1 := NewJust(42)

	m2, err := MapErr(m1, func(value int) (string, error) {
		return fmt.Sprintf("%d", value), nil
	})

	assert.True(t, Valid(m2))
	assert.Equal(t, "42", Just(m2))
	assert.NoError(t, err)

	m2, err = MapErr(m1, func(value int) (string, error) {
		return fmt.Sprintf("%d", value), specialErr
	})

	assert.False(t, Valid(m2))
	assert.Equal(t, specialErr, err)

	m1 = NewNone[int]()

	m2, err = MapErr(m1, func(value int) (string, error) {
		return fmt.Sprintf("%d", value), nil
	})
	assert.False(t, Valid(m2))
	assert.NoError(t, err)

	m2, err = MapErr(m1, func(value int) (string, error) {
		return fmt.Sprintf("%d", value), specialErr
	})
	assert.False(t, Valid(m2))
	assert.NoError(t, err)
}

func TestMapNone(t *testing.T) {
	m1 := NewJust(42)
	m2 := NewNone[int]()

	mv1Value := MapNone(m1, func() int {
		return 61
	})
	assert.Equal(t, Just(m1), mv1Value)

	mv2Value := MapNone(m2, func() int {
		return 61
	})
	assert.Equal(t, 61, mv2Value)
}

func TestMapNoneErr(t *testing.T) {
	specialErr := stderrors.New("special error")

	m1 := NewJust(42)
	m2 := NewNone[int]()

	mv1Value, err := MapNoneErr(m1, func() (int, error) {
		return 61, nil
	})
	assert.Equal(t, Just(m1), mv1Value)
	assert.NoError(t, err)

	_, err = MapNoneErr(m1, func() (int, error) {
		return 61, specialErr
	})
	assert.NoError(t, err)

	mv2Value, err := MapNoneErr(m2, func() (int, error) {
		return 61, nil
	})
	assert.Equal(t, 61, mv2Value)
	assert.NoError(t, err)

	_, err = MapNoneErr(m2, func() (int, error) {
		return 61, specialErr
	})
	assert.Equal(t, specialErr, err)
}

func TestFromPtr(t *testing.T) {
	var nilPtr *int
	ptr := &struct{}{}

	m1 := FromPtr(nilPtr)
	assert.False(t, Valid(m1))

	m2 := FromPtr(ptr)
	assert.True(t, Valid(m2))
}

func TestToPtr(t *testing.T) {
	m1 := NewJust(42)
	m2 := NewNone[int]()

	assert.NotNil(t, ToPtr(m1))
	assert.Equal(t, 42, Just(m1))

	assert.Nil(t, ToPtr(m2))
}
