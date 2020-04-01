package factory

import (
	"testing"
	"text/template"
	"github.com/stretchr/testify/assert"
	"bytes"
	"github.com/jukylin/esim/tool/db2entity"
)


func TestExecuteFactoryTemplate(t *testing.T)  {

	factory := esimFactory{}
	factory.StructName = "Test"
	s := &structInfo{}

	Field1 := db2entity.Field{}
	Field1.Filed = "a int"
	Field1.Doc = []string{"a", "int"}

	Field2 := db2entity.Field{}
	Field2.Filed = "b string"
	Field2.Doc = []string{"b", "string"}

	//fields := []db2entity.Field{}
	s.Fields = append(s.Fields, Field1, Field2)

	factory.NewStructInfo = s

	var buf bytes.Buffer
	tmpl, err := template.New("factory").Funcs(EsimFuncMap()).
		Parse(newTemplate)

	assert.Nil(t, err)
	err = tmpl.Execute(&buf, factory)
	if err != nil{
		println(err.Error())
	}
	println(buf.String())
	assert.Nil(t, err)
}


