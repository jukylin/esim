package container

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEsim(t *testing.T) {
	esim := NewEsim()

	assert.NotNil(t, esim.Logger)
	assert.NotNil(t, esim.Conf)
	assert.NotNil(t, esim.Tracer)
	assert.NotNil(t, esim.prometheus)
}
