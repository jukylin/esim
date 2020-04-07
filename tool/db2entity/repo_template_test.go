package db2entity

import (
	"bytes"
	"testing"
	"text/template"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/pkg"
)

func TestRepoTemplate(t *testing.T)  {
	tmpl, err := template.New("repo_template").Funcs(pkg.EsimFuncMap()).
		Parse(repoTemplate)
	assert.Nil(t, err)

	var imports pkg.Imports
	imports = append(imports, pkg.Import{Name : "time", Path: "time"})
	imports = append(imports, pkg.Import{Name : "sync", Path: "sync"})

	var buf bytes.Buffer
	repoTmp := repoTmp{}
	repoTmp.StructName = "User"
	repoTmp.TableName = "user"
	repoTmp.Imports = imports
	repoTmp.DelField = "is_del"

	err = tmpl.Execute(&buf, repoTmp)
	if err != nil{
		println(err.Error())
	}
	assert.Nil(t, err)
}





