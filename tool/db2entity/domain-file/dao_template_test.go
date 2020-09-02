package domainfile

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/stretchr/testify/assert"
)

func TestDaoTemplate(t *testing.T) {
	tmpl, err := template.New("dao_template").Funcs(templates.EsimFuncMap()).
		Parse(daoTemplate)
	assert.Nil(t, err)

	var imports pkg.Imports
	imports = append(imports, pkg.Import{Name: "time", Path: "time"},
		pkg.Import{Name: "sync", Path: "sync"})

	var buf bytes.Buffer
	daoTmp := newDaoTpl(userStructName)
	daoTmp.Imports = imports
	daoTmp.DataBaseName = database
	daoTmp.TableName = userTable
	daoTmp.PriKeyType = "int"

	daoTmp.CurTimeStamp = append(daoTmp.CurTimeStamp, "CreateTime1")
	daoTmp.CurTimeStamp = append(daoTmp.CurTimeStamp, "CreateTime2")
	daoTmp.OnUpdateTimeStamp = append(daoTmp.OnUpdateTimeStamp, "UpdateTime")
	daoTmp.OnUpdateTimeStampStr = append(daoTmp.OnUpdateTimeStampStr,
		"last_update_time1", "last_update_time2")

	err = tmpl.Execute(&buf, daoTmp)
	println(buf.String())

	assert.Nil(t, err)
}
