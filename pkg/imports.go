package pkg

import (
	"bytes"
	"text/template"
	"go/format"
)

var importTmp = `import (
{{ range .Imports }}
{{ range $doc := .Doc}}{{$doc}}
{{end}}{{.Name}} "{{.Path}}"
{{end}}
)`

type Import struct{
	Name string

	Path string

	Doc []string
}


type Imports []Import


func (this Imports) Len() int {
	return len(this)
}


func (this Imports) String() (string, error) {

	if this.Len() < 0 {
		return "", nil
	}

	tmpl, err := template.New("import_template").Funcs(EsimFuncMap()).
		Parse(importTmp)
	if err != nil{
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {Imports}{this})
	if err != nil{
		return "", err
	}

	src, err := format.Source(buf.Bytes())
	if err != nil{
		return "", err
	}

	return string(src), nil
}
