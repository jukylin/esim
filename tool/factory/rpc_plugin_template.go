package factory

import (
	"text/template"
	"strings"
)

var rpcPluginTemplate = `package main

import (
	"unsafe"
	"sort"
	"encoding/json"
	"reflect"
	"strings"
	"github.com/hashicorp/go-plugin"
	"github.com/jukylin/esim/tool/factory"
	"github.com/jukylin/esim/pkg"
)


type InitFieldsReturn struct{
	Fields []string
	SpecFields []pkg.Field
}

type Return struct{
	Fields pkg.Fields
	Size int
}

func (ModelImp) Sort() string {

	{{.StructName | tolower}} := {{.StructName}}{}

	originSize := unsafe.Sizeof({{.StructName | tolower}})

	getType := reflect.TypeOf({{.StructName | tolower}})

	var fields pkg.Fields

	{{range $i, $field := .Fields}}
	field{{$i}} := pkg.Field{}
	field{{$i}}.Name = "{{$field.Name}}"
	field{{$i}}.Field = "{{$field.Field}}"
	field{{$i}}.Size = int(getType.Field({{$i}}).Type.Size())
	fields = append(fields, field{{$i}})

	{{end}}

	sort.Sort(fields)

	re := &Return{}
	re.Fields = fields
	re.Size = int(originSize)

	by, _ := json.Marshal(re)
	return string(by)

}

func (ModelImp) InitField() string {
		{{.StructName | tolower}} := {{.StructName}}{}

		initReturn := &InitFieldsReturn{}
	 	fields := &pkg.Fields{}

		getType := reflect.TypeOf({{.StructName | tolower}})
		structFields := getInitStr(getType, string(strings.ToLower(getType.Name())[0]), fields)

		initReturn.SpecFields = *fields
		initReturn.Fields = structFields
		j, _ := json.Marshal(initReturn)
		return string(j)
	}

	func getInitStr(getType reflect.Type, name string, specFilds *pkg.Fields) []string {
		typeNum := getType.NumField()
		var structFields []string
		var initStr string
		field  := pkg.Field{}

		for i := 0; i < typeNum; i++ {
		switch getType.Field(i).Type.Kind() {
			case reflect.Array:
				structFields = append(structFields, "for k, _ := range this." + getType.Field(i).Name+" {")
					switch getType.Field(i).Type.Elem().Kind() {
					case reflect.Struct:
						structFields = append(structFields,
							getInitStr(getType.Field(i).Type.Elem(),
								"this." + getType.Field(i).Name + "[k]", nil)...)
					default:
						initStr = KindToInit(getType.Field(i).Type.Elem(),  name + "." + getType.Field(i).Name + "[k]", nil)
						structFields = append(structFields, "this." + getType.Field(i).Name+ "[k] = " + initStr)
					}
				structFields = append(structFields, "}")
				continue
			case reflect.Map:
				structFields = append(structFields, "for k, _ := range this." + getType.Field(i).Name+" {")
				structFields = append(structFields, "delete(this." + getType.Field(i).Name + ", k)")
				structFields = append(structFields, "}")
				if specFilds != nil {
					field.Name = name + "." + getType.Field(i).Name
					field.Type = "map"
					field.TypeName = getType.Field(i).Type.String()
					*specFilds = append(*specFilds, field)
				}
				continue
			case reflect.Struct:
				if getType.Field(i).Type.String() == "time.Time"{
					initStr = "time.Time{}"
				}else {
					structFields = append(structFields, getInitStr(getType.Field(i).Type,
						name+"."+getType.Field(i).Name, nil)...)
					continue
				}
			case reflect.Slice:
				if specFilds != nil {
					field.Name = name + "." + getType.Field(i).Name
					field.TypeName = getType.Field(i).Type.String()
					field.Type = "slice"
					*specFilds = append(*specFilds, field)
				}
				structFields = append(structFields, "this." + getType.Field(i).Name + " = " + "this." + getType.Field(i).Name + "[:0]")

				continue
			default:
				initStr = KindToInit(getType.Field(i).Type,
					name + "." + getType.Field(i).Name, specFilds)
			}

			structFields = append(structFields, "this." + getType.Field(i).Name + " = " + initStr)
		}

		return structFields
	}


func KindToInit(refType reflect.Type, name string, specFilds *pkg.Fields) string {
	var initStr string

	switch refType.Kind() {
	case reflect.String:
		initStr = "\"\""
	case reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32:
		initStr = "0"
	case reflect.Uint, reflect.Uint64, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		initStr = "0"
	case reflect.Bool:
		initStr = "false"
	case reflect.Float32, reflect.Float64:
		initStr = "0.00"
	case reflect.Complex64, reflect.Complex128:
		initStr = "0+0i"
	case reflect.Interface:
		initStr = "nil"
	case reflect.Uintptr:
		initStr = "0"
	case reflect.Invalid, reflect.Func, reflect.Chan, reflect.Ptr, reflect.UnsafePointer:
		initStr = "nil"
	case reflect.Slice:
		initStr = "nil"
	case reflect.Map:
		initStr = "nil"
	case reflect.Array:
		initStr = "nil"
	}

	return initStr
}

type ModelImp struct{}

func main() {

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: factory.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"model": &factory.ModelPlugin{Impl: &ModelImp{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
`


func EsimFuncMap() template.FuncMap {
	return template.FuncMap{
		"tolower" : strings.ToLower,
	}
}
