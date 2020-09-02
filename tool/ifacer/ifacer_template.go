package ifacer

//nolint:lll
var ifacerTemplate = `package {{.PackageName}}

{{- $StructName := .StructName}}

{{- $Star := .Star}}

{{.UsingImportStr}}

type {{$StructName}} struct {}

{{ range .Methods }}
func ({{$StructName | shorten | firstToLower}} {{$Star}}{{$StructName}}) {{.FuncName}}({{.ArgStr}}) {{.ReturnTypeStr}} {
{{.InitReturnVarStr}}

{{.ReturnStr}}
}
{{ end -}}
`
