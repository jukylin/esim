package factory

import (
	"testing"
	"github.com/jukylin/esim/tool/db2entity"
	"github.com/jukylin/esim/pkg/file-dir"
	"path/filepath"
	"github.com/stretchr/testify/assert"
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

	Field1 := db2entity.Field{}
	Field1.Filed = "b int64"

	Field2 := db2entity.Field{}
	Field2.Filed = "c int8"

	Field3 := db2entity.Field{}
	Field3.Filed = "i bool"

	Field4 := db2entity.Field{}
	Field4.Filed = "f float32"

	Field5 := db2entity.Field{}
	Field5.Filed = "a int32"

	Field6 := db2entity.Field{}
	Field6.Filed = "h []int"

	Field7 := db2entity.Field{}
	Field7.Filed = "m map[string]interface{}"

	Field8 := db2entity.Field{}
	Field8.Filed = "u [3]string"

	Field9 := db2entity.Field{}
	Field9.Filed = "g byte"

	fields := []db2entity.Field{}
	fields = append(fields, Field1, Field2, Field3, Field4,
		Field5, Field6, Field7, Field8, Field9)
	rpcPlugin.Fields = fields

	rpcPlugin.filesName = append(rpcPlugin.filesName, "example.go")
	rpcPlugin.packName = "example"
	sortResult := rpcPlugin.SortField()
	assert.Equal(t, 9, len(sortResult.Fields))
}