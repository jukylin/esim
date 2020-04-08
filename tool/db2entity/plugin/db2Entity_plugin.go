package main

import (
	"unsafe"
	"sort"
	"encoding/json"
	"reflect"
	"strings"
	"github.com/hashicorp/go-plugin"
	"github.com/jukylin/esim/tool/factory"
)


type InitFieldsReturn struct{
	Fields []string
	SpecFields []Field
}

type Field struct{
	Name string
	Str string
	Size int
	Type string
	TypeName string
}

type Fields []Field

func (f Fields) Len() int { return len(f) }

func (f Fields) Less(i, j int) bool {
	return f[i].Size <  f[j].Size
}

func (f Fields) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

type Return struct{
	Fields Fields
	Size int
}


func (ModelImp) Sort() string {

	db2entity := db2Entity{}

	originSize := unsafe.Sizeof(db2entity)

	getType := reflect.TypeOf(db2entity)

	var fields Fields

	
	field0 := Field{}
	field0.Name = "logger"
	field0.Str = "logger logger.Logger"
	field0.Size = int(getType.Field(0).Type.Size())
	fields = append(fields, field0)

	
	field1 := Field{}
	field1.Name = "withDisabledRepo"
	field1.Str = "withDisabledRepo bool"
	field1.Size = int(getType.Field(1).Type.Size())
	fields = append(fields, field1)

	
	field2 := Field{}
	field2.Name = "withDisabledDao"
	field2.Str = "withDisabledDao bool"
	field2.Size = int(getType.Field(2).Type.Size())
	fields = append(fields, field2)

	
	field3 := Field{}
	field3.Name = "withBoubctx"
	field3.Str = "withBoubctx string"
	field3.Size = int(getType.Field(3).Type.Size())
	fields = append(fields, field3)

	
	field4 := Field{}
	field4.Name = "withDisabledEntity"
	field4.Str = "withDisabledEntity bool"
	field4.Size = int(getType.Field(4).Type.Size())
	fields = append(fields, field4)

	
	field5 := Field{}
	field5.Name = "withEntityDir"
	field5.Str = "withEntityDir string"
	field5.Size = int(getType.Field(5).Type.Size())
	fields = append(fields, field5)

	
	field6 := Field{}
	field6.Name = "withPackage"
	field6.Str = "withPackage string"
	field6.Size = int(getType.Field(6).Type.Size())
	fields = append(fields, field6)

	
	field7 := Field{}
	field7.Name = "withStruct"
	field7.Str = "withStruct string"
	field7.Size = int(getType.Field(7).Type.Size())
	fields = append(fields, field7)

	
	field8 := Field{}
	field8.Name = "withInject"
	field8.Str = "withInject bool"
	field8.Size = int(getType.Field(8).Type.Size())
	fields = append(fields, field8)

	
	field9 := Field{}
	field9.Name = "dbConf"
	field9.Str = "dbConf dbConfig"
	field9.Size = int(getType.Field(9).Type.Size())
	fields = append(fields, field9)

	
	field10 := Field{}
	field10.Name = "ColumnsInter"
	field10.Str = "ColumnsInter ColumnsInter"
	field10.Size = int(getType.Field(10).Type.Size())
	fields = append(fields, field10)

	
	field11 := Field{}
	field11.Name = "conf"
	field11.Str = "conf config.Config"
	field11.Size = int(getType.Field(11).Type.Size())
	fields = append(fields, field11)

	

	sort.Sort(fields)

	re := &Return{}
	re.Fields = fields
	re.Size = int(originSize)

	by, _ := json.Marshal(re)
	return string(by)

}

func (ModelImp) InitField() string {
		db2entity := db2Entity{}

		initReturn := &InitFieldsReturn{}
	 	fields := &Fields{}

		getType := reflect.TypeOf(db2entity)
		structFields := getInitStr(getType, string(strings.ToLower(getType.Name())[0]), fields)

		initReturn.SpecFields = *fields
		initReturn.Fields = structFields
		j, _ := json.Marshal(initReturn)
		return string(j)
	}

	func getInitStr(getType reflect.Type, name string, specFilds *Fields) []string {
		typeNum := getType.NumField()
		var structFields []string
		var initStr string
		field  := Field{}

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


func KindToInit(refType reflect.Type, name string, specFilds *Fields) string {
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
