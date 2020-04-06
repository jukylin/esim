package db2entity

import (
	"bytes"
	"testing"
	"text/template"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/pkg"
)

func TestEntityTemplate(t *testing.T)  {
	tmpl, err := template.New("rpc_plugin").Funcs(pkg.EsimFuncMap()).
		Parse(entityTemplate)
	assert.Nil(t, err)

	var buf bytes.Buffer

	err = tmpl.Execute(&buf, nil)
	assert.Nil(t, err)
}





