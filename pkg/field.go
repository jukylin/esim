package pkg

import (
	"bytes"
	"go/ast"
	"text/template"
)

var fieldTpl = `{{ range .Fields }}
{{ range $doc := .Doc}}{{$doc}}
{{end}}{{.Field}} {{.Tag}}
{{end}}`

// struct field
type Field struct {
	Name string

	Type string

	TypeName string

	// Name + type or type
	Field string

	Size int

	Doc []string

	Tag string
}

type Fields []Field

func (fs Fields) Len() int { return len(fs) }

func (fs Fields) Less(i, j int) bool {
	return fs[i].Size < fs[j].Size
}

func (fs Fields) Swap(i, j int) { fs[i], fs[j] = fs[j], fs[i] }

func (fs Fields) String() (string, error) {
	if fs.Len() < 0 {
		return "", nil
	}

	tmpl, err := template.New("field_template").Parse(fieldTpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct{ Fields }{fs})
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (fs *Fields) ParseFromAst(genDecl *ast.GenDecl, fileContent string) {
	for _, specs := range genDecl.Specs {
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
					*fs = append(*fs, field)
				}
			}
		}
	}
}
