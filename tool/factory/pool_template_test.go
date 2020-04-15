package factory


import (
	"testing"
	"github.com/stretchr/testify/assert"
)



func TestPoolTpl_String(t *testing.T) {
	var result = `test = sync.Pool{
	New: func() interface{} {
		return &Test{}
	},
}
`

	poolTpl := NewPoolTpl()
	poolTpl.VarPoolName = "test"
	poolTpl.StructName = "Test"

	src := poolTpl.String()
	assert.Equal(t, result, src)
}