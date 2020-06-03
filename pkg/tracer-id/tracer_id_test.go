package tracerid

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTracerID(t *testing.T) {
	assert.NotEmpty(t, TracerID()())
}
