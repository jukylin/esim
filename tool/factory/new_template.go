package factory

var newTemplate = `
{{.StructTpl.String}}

{{.TypePluralStr}}

{{.Option1}}

{{if .WithNew}}
func New{{.StructName | snakeToCamel}}({{.Option2}}) {{.NewStructInfo.ReturnVarStr}}{

{{.NewStructInfo.StructInitStr}}

{{.Option3}}

{{.SpecFieldInitStr}}

	return {{.ReturnStr}}
}
{{end}}

{{.Option4}}

{{.Option5}}

{{.ReleaseStr}}

{{.NewPluralStr}}

{{.ReleasePluralStr}}
`
