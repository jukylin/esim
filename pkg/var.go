package pkg

import (
	"go/ast"
	"bytes"
	"text/template"
)

var varTpl = `var ({{ range .Vars }}
{{ range $doc := .Doc}}	{{$doc}}
{{end}}	{{.Body}} {{end}}
)`

type Var struct {
	Doc []string

	//name and val
	Body string

	Name []string

	//now is empty
	Val []string
}


type Vars []Var


func (this Vars) Len() int {
	return len(this)
}


func (this Vars) String() string {
	if this.Len() < 0 {
		return ""
	}

	tmpl, err := template.New("var_template").Parse(varTpl)
	if err != nil{
		panic(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {Vars}{this})
	if err != nil{
		panic(err.Error())
	}

	return buf.String()
}

func (this *Vars) ParseFromAst(GenDecl *ast.GenDecl, src string) {
	if GenDecl.Tok.String() == "var" {
		for _, specs := range GenDecl.Specs {
			if typeSpec, ok := specs.(*ast.ValueSpec); ok {
				varObj := Var{}

				varObj.Body = src[typeSpec.Pos()-1 : typeSpec.End()]
				if typeSpec.Doc != nil {
					for _, doc := range typeSpec.Doc.List {
						varObj.Doc = append(varObj.Doc, doc.Text)
					}
				}

				if len(typeSpec.Names) > 0 {
					for _, name := range typeSpec.Names {
						varObj.Name = append(varObj.Name, name.Name)
					}
				}

				*this = append(*this, varObj)
			}
		}
	}
}