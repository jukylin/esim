package factory

var factoryTemplate = `
type {{.StructName}} struct{
{{ range .NewStructInfo.Fields }}
	{{ range .Doc}}
		T1
	{{end}}
	{{.Filed}} {{.Tag}}
{{ end }}

}

{{.Option1}}

{{.Option2}}

func New{{.StructName}}({{.Option3}}) {{.NewStructInfo.ReturnVarStr}}{

{{.NewStructInfo.StructInitStr}}

{{.Option4}}

{{.SpecFieldInitStr}}

	return {{.ReturnStr}}
}

{{.Option5}}

{{.Option6}}
`
