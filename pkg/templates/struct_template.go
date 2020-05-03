package templates

import (
	"bytes"
	"text/template"

	"github.com/jukylin/esim/pkg"
)

type StructInfo struct {
	StructName string

	Fields pkg.Fields
}

func NewStructInfo() StructInfo {
	return StructInfo{}
}

var StructTemplate = `
type {{.StructName}} struct{
{{.Fields.String}}
}`

func (si StructInfo) String() string {
	if si.StructName == "" {
		return ""
	}

	tmpl, err := template.New("struct_template").Parse(StructTemplate)
	if err != nil {
		panic(err.Error())
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, si)
	if err != nil {
		panic(err.Error())
		return ""
	}

	return buf.String()
}
