package infra

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInfraSetTemplate(t *testing.T) {
	isa := infraSetArgs{}
	isa.Args = append(isa.Args, "provideRedis", "provideDb")

	result := isa.String()
	assert.NotEmpty(t, result)
}
