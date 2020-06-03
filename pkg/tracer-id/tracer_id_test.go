package tracerid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTracerID(t *testing.T) {
	assert.NotEmpty(t, TracerID()())
}
