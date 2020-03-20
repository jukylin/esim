package factory

var factoryTemplate = `
type {{.StructName}} struct{
{{ range .StructFields }}

{{.FiledCom}}
{{.FieldName}} {{.FieldType}} {{.FiledTag}}
{{ end -}}

}

{{.Options1}}

{{.Options2}}

func New{{.StructName}}({{.Options3}}) *StructName{

{{.NewStr}}

{{.Options4}}

{{.InitFileStr}}

{{.ReturnStr}}
}

{{.Options5}}

{{.Options6}}


`
