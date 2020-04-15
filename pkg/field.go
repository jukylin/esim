package pkg

import (
	"bytes"
	"text/template"
	"go/ast"
)

var fieldTpl = `{{ range .Fields }}
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

	tmpl, err := template.New("field_template").Parse(fieldTpl)
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



func (this *Fields) ParseFromAst(GenDecl *ast.GenDecl, fileContent string) {
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.TypeSpec); ok {
			if structType, ok := spec.Type.(*ast.StructType); ok {
				for _, astField := range structType.Fields.List {
					var field Field
					if astField.Doc != nil {
						for _, doc := range astField.Doc.List {
							field.Doc = append(field.Doc, doc.Text)
						}
					}

					if astField.Tag != nil {
						field.Tag = astField.Tag.Value
					}

					var name string
					if len(astField.Names) > 0 {
						name = astField.Names[0].String()
						field.Name = name
					}

					field.Type = ParseExpr(astField.Type, fileContent)
					field.Field = field.Name + " " + field.Type
					*this = append(*this, field)
				}
			}
		}
	}
}