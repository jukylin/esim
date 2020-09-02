package container

import (
	"testing"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/prometheus"
	"github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

func TestNewEsim(t *testing.T) {
	esim := NewEsim()
	assert.Implements(t, (*log.Logger)(nil), esim.Logger)
	assert.Implements(t, (*config.Config)(nil), esim.Conf)
	assert.Implements(t, (*opentracing.Tracer)(nil), esim.Tracer)
	assert.IsType(t, (*prometheus.Prometheus)(nil), esim.prometheus)

	assert.NotNil(t, esim.Logger)
	assert.NotNil(t, esim.Conf)
	assert.NotNil(t, esim.Tracer)
	assert.NotNil(t, esim.prometheus)

	NewMockEsim()
}

func TestNewEsiStringm(t *testing.T) {
	esim := NewEsim()
	assert.NotEmpty(t, esim.String())
}
