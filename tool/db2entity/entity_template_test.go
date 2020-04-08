package db2entity

import (
	"bytes"
	"testing"
	"text/template"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/pkg"
)

func TestEntityTemplate(t *testing.T)  {
	tmpl, err := template.New("entity_template").Funcs(pkg.EsimFuncMap()).
		Parse(entityTemplate)
	assert.Nil(t, err)

	var imports pkg.Imports
	imports = append(imports, pkg.Import{Name : "time", Path: "time"})
	imports = append(imports, pkg.Import{Name : "sync", Path: "sync"})

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
	entityTmp := entityTmp{}
	entityTmp.StructName = "Entity"
	entityTmp.CurTimeStamp = append(entityTmp.CurTimeStamp, "CreateTime1")
	entityTmp.CurTimeStamp = append(entityTmp.CurTimeStamp, "CreateTime2")

	entityTmp.OnUpdateTimeStamp = append(entityTmp.OnUpdateTimeStamp, "LastUpdateTime")

	entityTmp.OnUpdateTimeStampStr = append(entityTmp.OnUpdateTimeStampStr, "last_update_time1")
	entityTmp.OnUpdateTimeStampStr = append(entityTmp.OnUpdateTimeStampStr, "last_update_time2")

	entityTmp.Imports = imports
	entityTmp.Fields = append(entityTmp.Fields, Field1, Field2)
	entityTmp.DelField = "is_del"

	err = tmpl.Execute(&buf, entityTmp)
	assert.Nil(t, err)
	println(buf.String())
}





