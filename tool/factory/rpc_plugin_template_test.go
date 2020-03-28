package factory

import (
	"testing"
	"text/template"
	"github.com/stretchr/testify/assert"
	"bytes"
	"github.com/jukylin/esim/tool/db2entity"
)


func TestExecuteRpcPluginTemplate(t *testing.T)  {

	Field1 := db2entity.Field{}
	Field1.Filed = "a int"

	Field2 := db2entity.Field{}
	Field2.Filed = "b string"

	fields := []db2entity.Field{}
	fields = append(fields, Field1, Field2)

	data := struct {
		StructName string
		Fields []db2entity.Field
	}{
		"Test",
		fields,
	}

	var buf bytes.Buffer
	tmpl, err := template.New("rpc_plugin").Funcs(EsimFuncMap()).
		Parse(rpcPluginTemplate)
	assert.Nil(t, err)
	err = tmpl.Execute(&buf, data)
	assert.Nil(t, err)
}


