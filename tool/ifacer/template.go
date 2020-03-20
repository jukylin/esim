package ifacer

var ifaceTemplate = `package {{.PackageName}}

{{- $StructName := .StructName}}

{{- $Star := .Star}}

{{.UsingImportStr}}

type {{$StructName}} struct {}

{{ range .Methods }}
func (this {{$Star}}{{$StructName}}) {{.FuncName}}({{.ArgStr}}) {{.ReturnTypeStr}} {
{{.InitReturnVarStr}}

{{.ReturnStr}}
}
{{ end -}}
`
