package factory

import (
	"testing"
	"text/template"
	"github.com/stretchr/testify/assert"
	"bytes"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/templates"
)


func TestExecuteFactoryTemplate(t *testing.T)  {

	factory := esimFactory{}
	factory.StructName = "Test"
	s := &structInfo{}

	Field1 := pkg.Field{}
	Field1.Field = "a int"
	Field1.Doc = []string{"//a", "//int"}

	Field2 := pkg.Field{}
	Field2.Field = "b string"
	Field2.Doc = []string{"//b", "//string"}

	//fields := []db2entity.Field{}
	s.Fields = append(s.Fields, Field1, Field2)

	factory.NewStructInfo = s

	structTpl := templates.StructInfo{}
	structTpl.StructName = factory.StructName
	structTpl.Fields = factory.NewStructInfo.Fields

	factory.StructTpl = structTpl

	var buf bytes.Buffer
	tmpl, err := template.New("factory").Funcs(templates.EsimFuncMap()).
		Parse(newTemplate)
	assert.Nil(t, err)

	err = tmpl.Execute(&buf, factory)
	if err != nil{
		println(err.Error())
	}
	//println(buf.String())
	assert.Nil(t, err)
}


