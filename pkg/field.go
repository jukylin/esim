package pkg

import (
	"bytes"
	"text/template"
)

var fieldTmp = `{{ range .Fields }}
{{ range $doc := .Doc}}{{$doc}}
{{end}}{{.Field}} {{.Tag}}
{{end}}`

//struct field
type Field struct {
	Name string

	Type string

	TypeName string

	//Name + type or type
	Field string

	Size int

	Doc []string

	Tag string
}

type Fields []Field


func (f Fields) Len() int { return len(f) }

func (f Fields) Less(i, j int) bool {
	return f[i].Size <  f[j].Size
}

func (f Fields) Swap(i, j int) { f[i], f[j] = f[j], f[i] }


func (this Fields) String() (string, error) {

	if this.Len() < 0 {
		return "", nil
	}

	tmpl, err := template.New("field_template").Funcs(EsimFuncMap()).
		Parse(fieldTmp)
	if err != nil{
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {Fields}{this})
	if err != nil{
		return "", err
	}

	return buf.String(), nil
}