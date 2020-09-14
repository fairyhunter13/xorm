package reflection

import (
	"reflect"
	"testing"

	"github.com/fairyhunter13/reflecthelper"
	"github.com/stretchr/testify/assert"
)

func TestRevertVal2Zero(t *testing.T) {
	t.Run("Simple zero val test", func(t *testing.T) {
		var x **int
		val := reflect.ValueOf(&x)
		RevertVal2Zero(val)

		assert.True(t, reflecthelper.IsValueZero(val))
		assert.Nil(t, x)
	})
	t.Run("Don't set zero val if it's not empty", func(t *testing.T) {
		var x **int
		h := 5
		k := &h
		x = &k
		val := reflect.ValueOf(&x)
		RevertVal2Zero(val)

		assert.False(t, reflecthelper.IsValueZero(val))
	})
	t.Run("Set nil pointer if it is int with value empty (0)", func(t *testing.T) {
		var x **int
		h := 0
		k := &h
		x = &k
		val := reflect.ValueOf(&x)
		RevertVal2Zero(val)

		assert.True(t, reflecthelper.IsValueZero(val))
		assert.Nil(t, x)
	})
	t.Run("invalid reflect value", func(t *testing.T) {
		var test interface{}
		val := reflect.ValueOf(test)
		RevertVal2Zero(val)

		assert.True(t, reflecthelper.IsValueZero(val))
		assert.Nil(t, test)
	})
}
