package factory

import (
	"strings"
	"os"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/tool/db2entity"
	"regexp"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"os/exec"
	"fmt"
	"encoding/json"
	log2 "github.com/jukylin/esim/log"
	go_plugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-hclog"
)

type StructFieldIface interface {
	SortField() *SortReturn

	InitField() *InitFieldsReturn

	Close()

	SetStructDir(string)

	SetStructName(string)
}


type rpcPluginStructField struct{

	logger log2.Logger

	structDir string

	structName string

	packName string

	filesName []string

	buildBeforeFunc func()

	oldFields []db2entity.Field

	oldImport []string

	pluginClient *go_plugin.Client

	model Model
}

var pluginMap = map[string]go_plugin.Plugin{
	"model": &ModelPlugin{},
}

var handshakeConfig = go_plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func NewRpcPluginStructField() *rpcPluginStructField {
	rpcPlugin := &rpcPluginStructField{}


	return rpcPlugin
}


func (this *rpcPluginStructField) buildPluginEnv() error {
	this.structDir = strings.TrimRight(this.structDir, string(filepath.Separator))

	targetDir := this.structDir + string(filepath.Separator) + "plugin"

	exists, err := file_dir.IsExistsDir(targetDir)
	if err != nil {
		return err
	}

	if !exists {
		err := file_dir.CreateDir(targetDir)
		if err != nil {
			return err
		}
	}

	//TODO 复制文件
	//TODO 改 package 名称
	for _, name := range this.filesName {
		this.copyFile(targetDir + string(filepath.Separator) + name,
			this.structDir + string(filepath.Separator) + name,
				this.packName)
	}

	//TODO 生成 modelName_plugin.go 文件
	this.genStructPlugin(targetDir)
	if err != nil {
		return err
	}

	if this.buildBeforeFunc != nil {
		this.buildBeforeFunc()
	}

	err = this.buildPlugin(targetDir)
	if err != nil {
		return err
	}

	return nil
}

//@ Copy File
//@ repackagename
func (this *rpcPluginStructField) copyFile(dstName, srcName string, packageName string) {
	src, err := os.Open(srcName)
	if err != nil {
		this.logger.Panicf(err.Error())
	}

	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		this.logger.Panicf(err.Error())
	}
	defer dst.Close()

	reg, _ := regexp.Compile(`package[\s*]` + packageName)

	contents, err := ioutil.ReadAll(src)
	contents = reg.ReplaceAll(contents, []byte("package main"))

	dst.Write(contents)
}


func (this *rpcPluginStructField) genStructPlugin(dir string)  {
	str := `package main

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
	Size int
	Type string
	TypeName string
}

type Fields []Field

func (f Fields) Len() int { return len(f) }

func (f Fields) Less(i, j int) bool {
	return f[i].Size < f[j].Size
}

func (f Fields) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
`
	str += "type Return struct{ \r\n"
	str += "Fields Fields `json:\"fields\"` \r\n"
	str += "Size int `json:\"size\"` \r\n"
	str += "} \r\n"

	str += this.getSortBody()

	str += this.genInitField()

	file := dir + string(filepath.Separator) + this.structName + "_plugin.go"

	err := ioutil.WriteFile(file, []byte(str), 0666)
	if err != nil {
		this.logger.Panicf(err.Error())
	}

	return
}


func (this *rpcPluginStructField) getSortBody() string {
	str := "func (ModelImp) Sort() string { \r\n"

	str += "	" + strings.ToLower(this.structName) + " := " + this.structName + "{} \r\n"

	str += "	originSize := unsafe.Sizeof(" + strings.ToLower(this.structName) + ")\r\n"

	str += "	getType := reflect.TypeOf(" + strings.ToLower(this.structName) + ") \n"

	str += "	var fields Fields\r\n"

	for k, field := range this.oldFields {
		str += "	field" + strconv.Itoa(k) + " := Field{}\r\n"

		str += "	field" + strconv.Itoa(k) + ".Name = \"" + field.Filed + "\"\r\n"
		str += "	field" + strconv.Itoa(k) + ".Size =  + int(getType.Field(" + strconv.Itoa(k) + ").Type.Size())\r\n"

		str += "	fields = append(fields, field" + strconv.Itoa(k) + ")\r\n"
	}

	str += "	sort.Sort(fields)\r\n"

	str += "	re := &Return{}\r\n"

	str += "	re.Fields = fields\r\n"
	str += "	re.Size = int(originSize)\r\n"

	str += "	by, _ := json.Marshal(re)\r\n"

	str += "	return string(by)\r\n"

	str += "}"

	return str
}


func (this *rpcPluginStructField) genInitField() string {
	var str string
	str = `
func (ModelImp) InitField() string {
		` + strings.ToLower(this.structName) + ` := ` + this.structName + `{}

		initReturn := &InitFieldsReturn{}
	 	fields := &Fields{}

		getType := reflect.TypeOf(` + strings.ToLower(this.structName) + `)
		structFields := getInitStr(getType, strings.ToLower(getType.Name()), fields)

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
				structFields = append(structFields, "for k, _ := range "+ name + "." + getType.Field(i).Name+" {")
					switch getType.Field(i).Type.Elem().Kind() {
					case reflect.Struct:
						structFields = append(structFields,
							getInitStr(getType.Field(i).Type.Elem(),
								name + "." + getType.Field(i).Name + "[k]", nil)...)
					default:
						initStr = KindToInit(getType.Field(i).Type.Elem(),  name + "." + getType.Field(i).Name + "[k]", nil)
						structFields = append(structFields, name + "." + getType.Field(i).Name+ "[k] = " + initStr)
					}
				structFields = append(structFields, "}")
				continue
			case reflect.Map:
				structFields = append(structFields, "for k, _ := range "+ name + "." + getType.Field(i).Name+" {")
				structFields = append(structFields, "delete(" + name + "." + getType.Field(i).Name + ", k)")
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
			default:
				initStr = KindToInit(getType.Field(i).Type,
					name + "." + getType.Field(i).Name, specFilds)
			}

			structFields = append(structFields, name + "." + getType.Field(i).Name + " = " + initStr)
		}

		return structFields
	}


func KindToInit(refType reflect.Type, name string, specFilds *Fields) string {
	var initStr string
	field  := Field{}

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
		if specFilds != nil {
			field.Name = name
			field.TypeName = refType.String()
			field.Type = "slice"
			*specFilds = append(*specFilds, field)
		}
		initStr = name + "[:0]"
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

	return str
}


func (this *rpcPluginStructField) buildPlugin(dir string) error {
	cmd_line := fmt.Sprintf("go build -o plugin %s", dir)

	println(cmd_line)

	args := strings.Split(cmd_line, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	os.Chmod(dir + "/plugin", 0777)
	return nil
}


func (this *rpcPluginStructField) dispense()  {
	rpcClient, err := this.pluginClient.Client()
	if err != nil {
		this.logger.Panicf(err.Error())
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("model")
	if err != nil {
		this.logger.Panicf(err.Error())
	}

	this.model = raw.(Model)
}


func (this *rpcPluginStructField) run()  {

	if this.structDir == ""{
		this.logger.Panicf("%s is empty", this.structDir)
	}

	if this.structName == ""{
		this.logger.Panicf("%s is empty", this.structName)
	}

	this.pluginClient = go_plugin.NewClient(&go_plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(this.structDir + string(filepath.Separator) + "plugin" + string(filepath.Separator) + "plugin"),
		Logger : hclog.New(&hclog.LoggerOptions{
			Output: hclog.DefaultOutput,
			Level:  hclog.Error,
			Name:   "plugin",
		}),
	})

	this.buildPluginEnv()

	this.dispense()
}


func (this *rpcPluginStructField) SortField() *SortReturn {

	if this.model == nil{
		this.run()
	}

	sortReturn := &SortReturn{}
	err := json.Unmarshal([]byte(this.model.Sort()), sortReturn)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	return sortReturn
}


func (this *rpcPluginStructField) InitField() *InitFieldsReturn {

	if this.model == nil{
		this.run()
	}

	initReturn := &InitFieldsReturn{}
	err := json.Unmarshal([]byte(this.model.InitField()), initReturn)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	return initReturn
}

func (this *rpcPluginStructField) SetStructDir(structDir string)  {
	this.structDir = structDir
}


func (this *rpcPluginStructField) SetStructName(structName string)  {
	this.structName = structName
}


func (this *rpcPluginStructField) Close()  {
	this.pluginClient.Kill()
	this.clear()
}

func (this *rpcPluginStructField) clear() {
	err := os.RemoveAll(this.structDir + string(filepath.Separator) + "plugin")
	this.logger.Panicf(err.Error())
}