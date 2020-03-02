package iface


var ifaceTemplate = `package {{.PackageName}}

{{- $StructName := .StructName}}

{{- $Star := .Star}}


{{.ImportStr}}

//@ Interface {{.IfaceName}}
type {{$StructName}} struct {}

{{ range .Methods }}
func (this {{$Star}}{{$StructName}}) {{.FuncName}}({{.ArgStr}}) {{.ReturnStr}} {

}
{{ end -}}
`
