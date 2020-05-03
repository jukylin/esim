package factory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlural_String(t *testing.T) {
	var src string

	plural := NewPlural()
	plural.PluralName = "tests"
	plural.StructName = "Test"

	src = plural.NewString()
	assert.NotEmpty(t, src)

	src = plural.ReleaseString()
	assert.NotEmpty(t, src)

	src = plural.TypeString()
	assert.NotEmpty(t, src)
}
