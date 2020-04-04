package factory

var newTemplate = `
type {{.StructName}} struct{
{{ range .NewStructInfo.Fields }}
{{ range $a := .Doc}}{{$a}}
{{end}}{{.Field}} {{.Tag}}
{{end}}
}

{{.TypePluralStr}}

{{.Option1}}

{{.Option2}}

{{if .WithNew}}
func New{{.StructName}}({{.Option3}}) {{.NewStructInfo.ReturnVarStr}}{

{{.NewStructInfo.StructInitStr}}

{{.Option4}}

{{.SpecFieldInitStr}}

	return {{.ReturnStr}}
}
{{end}}

{{.Option5}}

{{.Option6}}

{{.ReleaseStr}}

{{.NewPluralStr}}

{{.ReleasePluralStr}}
`
