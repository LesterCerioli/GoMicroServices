package selector

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	Register(defaultSelector{})
	selector, ok := selectorMap["defaultSelector"]
	assert.True(t, ok)
	assert.Equal(t, defaultSelector{}, selector)
}

func TestGet(t *testing.T) {
	Register(defaultSelector{})
	selector := Get("defaultSelector")
	assert.Equal(t, defaultSelector{}, selector)

	selector = Get("")
	assert.Equal(t, noneSelector{}, selector)
}

func TestNewSelectorContext(t *testing.T) {
	ctx := NewSelectorContext(context.Background(), "defaultSelector")
	assert.EqualValues(t, "defaultSelector", ctx.Value(selectKey{}))

	ctx = NewSelectorContext(context.Background(), "")
	assert.True(t, context.Background() == ctx)

}
func TestSelectorFromContext(t *testing.T) {
	ctx := context.WithValue(context.Background(), selectKey{}, "defaultSelector")
	assert.EqualValues(t, "defaultSelector", SelectorFromContext(ctx))
	assert.EqualValues(t, "", SelectorFromContext(context.Background()))
}
