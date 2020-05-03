package templates

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/jukylin/esim/pkg"
	"github.com/stretchr/testify/assert"
)

func TestStructTemplate(t *testing.T) {
	tmpl, err := template.New("struct_template").Funcs(EsimFuncMap()).
		Parse(StructTemplate)
	assert.Nil(t, err)

	var buf bytes.Buffer
	structInfo := StructInfo{}
	structInfo.StructName = "Test"

	filed1 := pkg.Field{}
	filed1.Name = "a"
	filed1.Type = "int"
	filed1.Field = "a int"

	filed2 := pkg.Field{}
	filed2.Name = "b"
	filed2.Type = "string"
	filed2.Field = "b string"
	filed2.Doc = append(filed2.Doc, "//is a test")

	structInfo.Fields = append(structInfo.Fields, filed1, filed2)

	err = tmpl.Execute(&buf, structInfo)
	assert.Nil(t, err)
	//println(buf.String())
}
