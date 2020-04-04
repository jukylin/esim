package factory

import (
	"testing"
	"github.com/jukylin/esim/pkg/file-dir"
	"path/filepath"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/pkg"
)

func TestRpcPluginStructField_SortField(t *testing.T) {

	writer := file_dir.EsimWriter{}

	rpcPlugin := NewRpcPluginStructField(writer)
	dir, err := filepath.Abs(".")
	assert.Nil(t, err)
	rpcPlugin.structDir =  dir + "/example"
	rpcPlugin.StructName = "Test"

	rpcPlugin.StrcutInfo = &structInfo{}
	rpcPlugin.StrcutInfo.structFileContent = `package example

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
}

type empty struct {}`

	Field1 := pkg.Field{}
	Field1.Field = "b int64"

	Field2 := pkg.Field{}
	Field2.Field = "c int8"

	Field3 := pkg.Field{}
	Field3.Field = "i bool"

	Field4 := pkg.Field{}
	Field4.Field = "f float32"

	Field5 := pkg.Field{}
	Field5.Field = "a int32"

	Field6 := pkg.Field{}
	Field6.Field = "h []int"

	Field7 := pkg.Field{}
	Field7.Field = "m map[string]interface{}"

	Field8 := pkg.Field{}
	Field8.Field = "u [3]string"

	Field9 := pkg.Field{}
	Field9.Field = "g byte"

	fields := []pkg.Field{}
	fields = append(fields, Field1, Field2, Field3, Field4,
		Field5, Field6, Field7, Field8, Field9)
	rpcPlugin.Fields = fields

	rpcPlugin.filesName = append(rpcPlugin.filesName, "example.go")
	rpcPlugin.packName = "example"
	sortResult := rpcPlugin.SortField(fields)
	assert.Equal(t, 9, len(sortResult.Fields))

	initResult := rpcPlugin.InitField(fields)
	assert.Equal(t, 15, len(initResult.Fields))
	rpcPlugin.clear()
}