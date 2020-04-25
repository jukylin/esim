package pkg

import (
	"bytes"
	"go/ast"
	"strings"
	"text/template"
)

var importTpl = `import (
{{ range .Imports }}
{{ range $doc := .Doc}}{{$doc}}
{{end}}{{.Name}} "{{.Path}}"{{end}}
)`

type Import struct {
	Name string

	Path string

	Doc []string
}

type Imports []Import

func (is Imports) Len() int {
	return len(is)
}

func (is Imports) String() string {

	if is.Len() < 0 {
		return ""
	}

	tmpl, err := template.New("import_template").Parse(importTpl)
	if err != nil {
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct{ Imports }{is})
	if err != nil {
		panic(err.Error())
	}

	return buf.String()
}

func (is *Imports) ParseFromAst(GenDecl *ast.GenDecl) {
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.ImportSpec); ok {
			imp := Import{}
			if spec.Name.String() != "<nil>" {
				imp.Name = spec.Name.String()
			}

			if spec.Doc != nil {
				for _, test := range spec.Doc.List {
					imp.Doc = append(imp.Doc, test.Text)
				}
			}

			imp.Path = strings.Trim(spec.Path.Value, "\"")
			*is = append(*is, imp)
		}
	}
}
