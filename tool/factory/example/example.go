package example

import (
	"github.com/jukylin/esim/pkg"
)

var (
	var1 = []string{"var1"} //nolint:unused,varcheck,deadcode
)

//nolint:unused,structcheck,maligned
type Test struct {
	b int64

	c int8

	i bool

	f float32

	a int32

	h []int

	m map[string]interface{}

	e string

	g byte

	u [3]string

	d int16

	pkg.Fields

	pkg.Field

	n func(interface{})
}
