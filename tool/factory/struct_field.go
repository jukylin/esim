package factory

import (
	"strings"
	"os"
	"reflect"
	"github.com/jukylin/esim/pkg/file-dir"
	"regexp"
	"io/ioutil"
	"path/filepath"
	"bytes"
	"os/exec"
	"fmt"
	"encoding/json"
	log2 "github.com/jukylin/esim/log"
	go_plugin "github.com/hashicorp/go-plugin"
	"github.com/hashicorp/go-hclog"
	"text/template"
	"github.com/jukylin/esim/pkg"
	"golang.org/x/tools/imports"
)

type StructFieldIface interface {
	SortField(fields []pkg.Field) *SortReturn

	InitField(fields []pkg.Field) *InitFieldsReturn

	Close()

	SetStructInfo(*structInfo)

	SetStructName(string)

	SetStructDir(string)

	SetStructFileName(string)

	SetFilesName(filesName []string)

	SetPackName(packName string)
}


type rpcPluginStructField struct{

	logger log2.Logger

	structDir string

	StructName string

	StructFileName string

	StrcutInfo *structInfo

	packName string

	filesName []string

	buildBeforeFunc func()

	Fields []pkg.Field

	oldImport []string

	pluginClient *go_plugin.Client

	model Model

	writer file_dir.IfaceWriter
}

var pluginMap = map[string]go_plugin.Plugin{
	"model": &ModelPlugin{},
}

var HandshakeConfig = go_plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func NewRpcPluginStructField(writer file_dir.IfaceWriter) *rpcPluginStructField {
	rpcPlugin := &rpcPluginStructField{}

	rpcPlugin.writer = writer

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

	// 复制文件
	// 改 package 名称
	reg, _ := regexp.Compile(`package[\s*]` + this.packName)

	for _, name := range this.filesName {
		if name == this.StructFileName{
			src := reg.ReplaceAll([]byte(this.StrcutInfo.structFileContent), []byte("package main"))
			this.writer.Write(targetDir + string(filepath.Separator) + this.StructFileName,
				string(src))
			continue
		}

		this.copyFile(targetDir + string(filepath.Separator) + name,
			this.structDir + string(filepath.Separator) + name, reg)
	}

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
func (this *rpcPluginStructField) copyFile(dstName, srcName string, reg *regexp.Regexp) {
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

	contents, err := ioutil.ReadAll(src)
	contents = reg.ReplaceAll(contents, []byte("package main"))

	dst.Write(contents)
}


// gen modelName_plugin.go
func (this *rpcPluginStructField) genStructPlugin(dir string)  {

	tmpl, err := template.New("rpc_plugin").Funcs(pkg.EsimFuncMap()).
		Parse(rpcPluginTemplate)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, this)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	src ,err := imports.Process("", buf.Bytes(), nil)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	file := dir + string(filepath.Separator) + this.StructName + "_plugin.go"
	err = ioutil.WriteFile(file, src, 0666)
	if err != nil {
		this.logger.Panicf(err.Error())
	}

	return
}


func (this *rpcPluginStructField) buildPlugin(dir string) error {
	cmd_line := fmt.Sprintf("go build -o %s/plugin %s", dir, dir)

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

	os.Chmod(dir + string(filepath.Separator) + "plugin", 0777)
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

	if this.StructName == ""{
		this.logger.Panicf("%s is empty", this.StructName)
	}

	this.pluginClient = go_plugin.NewClient(&go_plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
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


func (this *rpcPluginStructField) SortField(fields []pkg.Field) *SortReturn {

	this.Fields = fields

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


func (this *rpcPluginStructField) InitField(fields []pkg.Field) *InitFieldsReturn {
	this.Fields = fields

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
	this.StructName = structName
}


func (this *rpcPluginStructField) SetStructInfo(s *structInfo)  {
	this.StrcutInfo = s
}

func (this *rpcPluginStructField) SetStructFileName (structFileName string)  {
	this.StructFileName = structFileName
}


func (this *rpcPluginStructField) SetFilesName(filesName []string)  {
	this.filesName = filesName
}


func (this *rpcPluginStructField) SetPackName(packName string)  {
	this.packName = packName
}

func (this *rpcPluginStructField) Close()  {
	this.pluginClient.Kill()
	this.clear()
}

func (this *rpcPluginStructField) clear() {
	err := os.RemoveAll(this.structDir + string(filepath.Separator) + "plugin")
	if err != nil {
		this.logger.Panicf(err.Error())
	}
}

func (this *rpcPluginStructField) GenInitFieldStr(getType reflect.Type, fieldLink, initName string, specFilds *pkg.Fields) []string {
	typeNum := getType.NumField()
	var structFields []string
	var initStr string
	field := pkg.Field{}

	for i := 0; i < typeNum; i++ {
		switch getType.Field(i).Type.Kind() {
		case reflect.Array:
			structFields = append(structFields, "for k, _ := range "+fieldLink+"."+getType.Field(i).Name+" {")
			switch getType.Field(i).Type.Elem().Kind() {
			case reflect.Struct:
				structFields = append(structFields,
					this.GenInitFieldStr(getType.Field(i).Type.Elem(), fieldLink+"."+getType.Field(i).Name,
						initName+"."+getType.Field(i).Name, nil)...)
			default:
				initStr = this.KindToInit(getType.Field(i).Type.Elem())
				structFields = append(structFields, fieldLink+"."+getType.Field(i).Name+"[k] = "+initStr)
			}
			structFields = append(structFields, "}")
			continue
		case reflect.Map:
			structFields = append(structFields, "for k, _ := range "+fieldLink+"."+getType.Field(i).Name+" {")
			structFields = append(structFields, "delete("+fieldLink+"."+getType.Field(i).Name+", k)")
			structFields = append(structFields, "}")
			if specFilds != nil {
				field.Name = initName + "." + getType.Field(i).Name
				field.Type = "map"
				field.TypeName = getType.Field(i).Type.String()
				*specFilds = append(*specFilds, field)
			}
			continue
		case reflect.Struct:
			if getType.Field(i).Type.String() == "time.Time" {
				initStr = "time.Time{}"
			} else {
				structFields = append(structFields, this.GenInitFieldStr(getType.Field(i).Type,
					fieldLink+"."+getType.Field(i).Name, initName+"."+getType.Field(i).Name, nil)...)
				continue
			}
		case reflect.Slice:
			if specFilds != nil {
				field.Name = initName + "." + getType.Field(i).Name
				field.TypeName = getType.Field(i).Type.String()
				field.Type = "slice"
				*specFilds = append(*specFilds, field)
			}
			structFields = append(structFields, fieldLink+"."+getType.Field(i).Name+" = "+fieldLink+"."+getType.Field(i).Name+"[:0]")

			continue
		default:
			initStr = this.KindToInit(getType.Field(i).Type)
		}

		structFields = append(structFields, fieldLink+"."+getType.Field(i).Name+" = "+initStr)
	}

	return structFields
}


func (this *rpcPluginStructField) KindToInit(refType reflect.Type) string {
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