package domainfile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProvides_String(t *testing.T) {
	provide1 := Provide{}
	provide1.Content = "func () {}"

	provide2 := Provide{}
	provide2.Content = `func () {
	println(123)
}`

	provides := Provides{}

	provides = append(provides, provide1, provide2)

	result := provides.String()
	assert.NotEmpty(t, result)
}
