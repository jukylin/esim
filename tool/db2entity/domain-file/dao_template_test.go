package domain_file

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
	imports = append(imports, pkg.Import{Name: "time", Path: "time"}, pkg.Import{Name: "sync", Path: "sync"})

	var buf bytes.Buffer
	daoTmp := newDaoTpl("User")
	daoTmp.Imports = imports
	daoTmp.DataBaseName = "test"
	daoTmp.TableName = "user"
	daoTmp.PriKeyType = "int"

	err = tmpl.Execute(&buf, daoTmp)
	assert.Nil(t, err)
	//println(buf.String())
}
