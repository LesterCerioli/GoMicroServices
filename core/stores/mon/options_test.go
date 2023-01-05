package mon

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetSlowThreshold(t *testing.T) {
	assert.Equal(t, defaultSlowThreshold, slowThreshold.Load())
	SetSlowThreshold(time.Second)
	assert.Equal(t, time.Second, slowThreshold.Load())
}

func TestDefaultOptions(t *testing.T) {
	assert.Equal(t, defaultTimeout, defaultOptions().timeout)
}

func TestOptions_mgoOptions(t *testing.T) {
	o := defaultOptions()
	assert.Equal(t, 1, len(o.mgoOptions()))

	o = &options{}
	assert.Equal(t, 0, len(o.mgoOptions()))
}
