package domain_file

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/stretchr/testify/assert"
)

func TestEntityTemplate(t *testing.T) {
	tmpl, err := template.New("entity_template").Funcs(templates.EsimFuncMap()).
		Parse(entityTemplate)
	assert.Nil(t, err)

	var imports pkg.Imports
	imports = append(imports, pkg.Import{Name: "time", Path: "time"},
		pkg.Import{Name: "sync", Path: "sync"})

	Field1 := pkg.Field{}
	Field1.Name = "id"
	Field1.Field = "id int"
	Field1.Tag = "`json:\"id\"`"

	Field2 := pkg.Field{}
	Field2.Name = "name"
	Field2.Field = "name string"
	Field2.Tag = "`json:\"name\"`"
	Field2.Doc = append(Field2.Doc, "//username \\r\\n is a test")

	var buf bytes.Buffer
	entityTpl := entityTpl{}
	entityTpl.StructName = "Entity"
	entityTpl.CurTimeStamp = append(entityTpl.CurTimeStamp, "CreateTime1", "CreateTime2")

	entityTpl.OnUpdateTimeStamp = append(entityTpl.OnUpdateTimeStamp, "LastUpdateTime")

	entityTpl.OnUpdateTimeStampStr = append(entityTpl.OnUpdateTimeStampStr,
		"last_update_time1", "last_update_time2")

	entityTpl.Imports = imports
	entityTpl.DelField = "is_del"

	structInfo := templates.StructInfo{}
	structInfo.StructName = entityTpl.StructName
	structInfo.Fields = append(structInfo.Fields, Field1, Field2)

	entityTpl.StructInfo = structInfo

	err = tmpl.Execute(&buf, entityTpl)
	assert.Nil(t, err)
	println(buf.String())
}
