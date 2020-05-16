package domain_file

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/stretchr/testify/assert"
)

func TestRepoTemplate(t *testing.T) {
	tmpl, err := template.New("repo_template").Funcs(templates.EsimFuncMap()).
		Parse(repoTemplate)
	assert.Nil(t, err)

	var imports pkg.Imports
	imports = append(imports, pkg.Import{Name: "time", Path: "time"},
		pkg.Import{Name: "sync", Path: "sync"})

	var buf bytes.Buffer
	repoTpl := newRepoTpl("User")
	repoTpl.TableName = "user"
	repoTpl.Imports = imports
	repoTpl.DelField = "is_del"

	err = tmpl.Execute(&buf, repoTpl)
	if err != nil {
		println(err.Error())
	}
	assert.Nil(t, err)
	//println(buf.String())
}
