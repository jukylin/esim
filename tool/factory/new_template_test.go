package factory

import (
	"testing"

	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/stretchr/testify/assert"
)

func TestExecuteFactoryTemplate(t *testing.T) {

	factory := EsimFactory{}
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

	tpl := templates.NewTextTpl()
	tmpl, err := tpl.Execute("factory", newTemplate, factory)
	//println(tmpl)
	assert.Nil(t, err)
	assert.NotEmpty(t, tmpl)
}
