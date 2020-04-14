package pkg

import (
	"bytes"
	"text/template"
	"go/ast"
	"strings"
)

var importTmp = `import (
{{ range .Imports }}
{{ range $doc := .Doc}}{{$doc}}
{{end}}{{.Name}} "{{.Path}}"{{end}}
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


func (this Imports) String() string {

	if this.Len() < 0 {
		return ""
	}

	tmpl, err := template.New("import_template").Parse(importTmp)
	if err != nil{
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {Imports}{this})
	if err != nil{
		panic(err.Error())
	}

	return buf.String()
}

func (this *Imports) ParseFromAst(GenDecl *ast.GenDecl) {
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
			*this = append(*this, imp)
		}
	}
}