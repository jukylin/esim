package iface


var ifaceTemplate = `

package {{.PackageName}}

{{- $StructName := .StructName}}

{{.ImportStr}}

type {{$StructName}} struct {}

{{ range .Methods }}
func (this *{{$StructName}}) {{.FuncName}}({{.ArgStr}}) {{.ReturnStr}} {

}
{{ end -}}`
