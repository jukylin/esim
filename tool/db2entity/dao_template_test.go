package db2entity

import (
	"bytes"
	"testing"
	"text/template"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/pkg"
)

func TestDaoTemplate(t *testing.T)  {
	tmpl, err := template.New("dao_template").Funcs(pkg.EsimFuncMap()).
		Parse(daoTemplate)
	assert.Nil(t, err)

	var imports pkg.Imports
	imports = append(imports, pkg.Import{Name : "time", Path: "time"})
	imports = append(imports, pkg.Import{Name : "sync", Path: "sync"})

	var buf bytes.Buffer
	daoTmp := daoTmp{}
	daoTmp.StructName = "User"
	daoTmp.Imports = imports
	daoTmp.DataBaseName = "test"
	daoTmp.TableName = "user"
	daoTmp.PriKeyType = "int"

	err = tmpl.Execute(&buf, daoTmp)
	assert.Nil(t, err)
	//println(buf.String())
}





