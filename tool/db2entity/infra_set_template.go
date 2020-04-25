package db2entity

import (
	"bytes"
	"github.com/jukylin/esim/pkg/templates"
	"text/template"
)

type infraSetArgs struct {
	Args []string
}

var infraSetTemplate = `var infraSet = wire.NewSet(
{{ range $arg := .Args}}	{{$arg}},
{{end}}
)
`

func (sa infraSetArgs) String() string {
	tmpl, err := template.New("infra_set_template").Funcs(templates.EsimFuncMap()).
		Parse(infraSetTemplate)
	if err != nil {
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, sa)
	if err != nil {
		panic(err.Error())
	}

	return buf.String()
}
