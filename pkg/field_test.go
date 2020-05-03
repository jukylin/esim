package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFields_EmptyString(t *testing.T) {
	fields := Fields{}
	str, err := fields.String()
	assert.Nil(t, err)
	assert.Empty(t, str)
}

func TestFields_String(t *testing.T) {

	fields := Fields{}

	field1 := Field{}
	field1.Field = "id int"
	field1.Tag = "`json:\"id\"`"
	field1.Doc = append(field1.Doc, "//id", "//is a test")

	field2 := Field{}
	field2.Field = "name string"
	field2.Tag = "`json:\"name\"`"
	field2.Doc = append(field2.Doc, "//name", "//is a test")

	fields = append(fields, field1, field2)

	_, err := fields.String()
	assert.Nil(t, err)
	//println(string(src))
}
