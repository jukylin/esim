package factory

import (
	"testing"
	"text/template"
	"github.com/stretchr/testify/assert"
	"bytes"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/templates"
)


func TestExecuteRpcPluginTemplate(t *testing.T)  {

	Field1 := pkg.Field{}
	Field1.Field = "a int"

	Field2 := pkg.Field{}
	Field2.Field = "b string"

	fields := make([]pkg.Field, 0)
	fields = append(fields, Field1, Field2)

	data := struct {
		StructName string
		Fields []pkg.Field
	}{
		"Test",
		fields,
	}

	var buf bytes.Buffer
	tmpl, err := template.New("rpc_plugin").Funcs(templates.EsimFuncMap()).
		Parse(rpcPluginTemplate)
	assert.Nil(t, err)
	err = tmpl.Execute(&buf, data)
	assert.Nil(t, err)
}


