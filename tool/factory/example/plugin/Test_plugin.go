package main

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

	test := Test{}

	originSize := unsafe.Sizeof(test)

	getType := reflect.TypeOf(test)

	var fields pkg.Fields

	
	field0 := pkg.Field{}
	field0.Name = "b"
	field0.Field = "b int64"
	field0.Size = int(getType.Field(0).Type.Size())
	fields = append(fields, field0)

	
	field1 := pkg.Field{}
	field1.Name = "c"
	field1.Field = "c int8"
	field1.Size = int(getType.Field(1).Type.Size())
	fields = append(fields, field1)

	
	field2 := pkg.Field{}
	field2.Name = "i"
	field2.Field = "i bool"
	field2.Size = int(getType.Field(2).Type.Size())
	fields = append(fields, field2)

	
	field3 := pkg.Field{}
	field3.Name = "f"
	field3.Field = "f float32"
	field3.Size = int(getType.Field(3).Type.Size())
	fields = append(fields, field3)

	
	field4 := pkg.Field{}
	field4.Name = "a"
	field4.Field = "a int32"
	field4.Size = int(getType.Field(4).Type.Size())
	fields = append(fields, field4)

	
	field5 := pkg.Field{}
	field5.Name = "h"
	field5.Field = "h []int"
	field5.Size = int(getType.Field(5).Type.Size())
	fields = append(fields, field5)

	
	field6 := pkg.Field{}
	field6.Name = "m"
	field6.Field = "m map[string]interface{}"
	field6.Size = int(getType.Field(6).Type.Size())
	fields = append(fields, field6)

	
	field7 := pkg.Field{}
	field7.Name = "e"
	field7.Field = "e string"
	field7.Size = int(getType.Field(7).Type.Size())
	fields = append(fields, field7)

	
	field8 := pkg.Field{}
	field8.Name = "g"
	field8.Field = "g byte"
	field8.Size = int(getType.Field(8).Type.Size())
	fields = append(fields, field8)

	
	field9 := pkg.Field{}
	field9.Name = "u"
	field9.Field = "u [3]string"
	field9.Size = int(getType.Field(9).Type.Size())
	fields = append(fields, field9)

	
	field10 := pkg.Field{}
	field10.Name = "d"
	field10.Field = "d int16"
	field10.Size = int(getType.Field(10).Type.Size())
	fields = append(fields, field10)

	
	field11 := pkg.Field{}
	field11.Name = "logger"
	field11.Field = "logger log.Logger"
	field11.Size = int(getType.Field(11).Type.Size())
	fields = append(fields, field11)

	
	field12 := pkg.Field{}
	field12.Name = "conf"
	field12.Field = "conf config.Config"
	field12.Size = int(getType.Field(12).Type.Size())
	fields = append(fields, field12)

	

	sort.Sort(fields)

	re := &Return{}
	re.Fields = fields
	re.Size = int(originSize)

	by, _ := json.Marshal(re)
	return string(by)

}

func (ModelImp) InitField() string {
		test := Test{}

		initReturn := &InitFieldsReturn{}
	 	fields := &pkg.Fields{}

		getType := reflect.TypeOf(test)
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
