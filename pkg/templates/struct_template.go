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


var StructTemplate = `
type {{.StructName}} struct{
{{.Fields.String}}
}`


func (this StructInfo) String() string {
	if this.StructName == "" {
		return ""
	}

	tmpl, err := template.New("struct_template").Parse(StructTemplate)
	if err != nil{
		panic(err.Error())
		return ""
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, this)
	if err != nil{
		panic(err.Error())
		return ""
	}

	return buf.String()
}