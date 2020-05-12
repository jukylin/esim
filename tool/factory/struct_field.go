package factory

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"text/template"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	log2 "github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	file_dir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
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

type rpcPluginStructField struct {
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

	pluginClient *plugin.Client

	model Model

	writer file_dir.IfaceWriter
}

var pluginMap = map[string]plugin.Plugin{
	"model": &ModelPlugin{},
}

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func NewRPCPluginStructField(writer file_dir.IfaceWriter, logger log2.Logger) *rpcPluginStructField {
	rpcPlugin := &rpcPluginStructField{}

	rpcPlugin.writer = writer

	rpcPlugin.logger = logger

	return rpcPlugin
}

func (rps *rpcPluginStructField) buildPluginEnv() error {
	rps.structDir = strings.TrimRight(rps.structDir, string(filepath.Separator))

	targetDir := rps.structDir + string(filepath.Separator) + "plugin"

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
	reg, _ := regexp.Compile(`package[\s*]` + rps.packName)

	for _, name := range rps.filesName {
		if name == rps.StructFileName {
			src := reg.ReplaceAll([]byte(rps.StrcutInfo.structFileContent), []byte("package main"))
			rps.writer.Write(targetDir+string(filepath.Separator)+rps.StructFileName,
				string(src))
			continue
		}

		rps.copyFile(targetDir+string(filepath.Separator)+name,
			rps.structDir+string(filepath.Separator)+name, reg)
	}

	rps.genStructPlugin(targetDir)
	if err != nil {
		return err
	}

	if rps.buildBeforeFunc != nil {
		rps.buildBeforeFunc()
	}

	err = rps.buildPlugin(targetDir)
	if err != nil {
		return err
	}

	return nil
}

//@ Copy File
//@ repackagename
func (rps *rpcPluginStructField) copyFile(dstName, srcName string, reg *regexp.Regexp) {
	src, err := os.Open(srcName)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}
	defer dst.Close()

	contents, err := ioutil.ReadAll(src)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	contents = reg.ReplaceAll(contents, []byte("package main"))

	dst.Write(contents)
}

// gen modelName_plugin.go
func (rps *rpcPluginStructField) genStructPlugin(dir string) {

	tmpl, err := template.New("rpc_plugin").Funcs(templates.EsimFuncMap()).
		Parse(rpcPluginTemplate)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, rps)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	src, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		rps.logger.Panicf("%s : %s", err.Error(), buf.String())
	}

	file := dir + string(filepath.Separator) + rps.StructName + "_plugin.go"
	err = ioutil.WriteFile(file, src, 0666)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	return
}

func (rps *rpcPluginStructField) buildPlugin(dir string) error {
	cmdLine := fmt.Sprintf("go build -o %s/plugin %s", dir, dir)

	println(cmdLine)

	args := strings.Split(cmdLine, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	os.Chmod(dir+string(filepath.Separator)+"plugin", 0777)
	return nil
}

func (rps *rpcPluginStructField) dispense() {
	rpcClient, err := rps.pluginClient.Client()
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("model")
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	rps.model = raw.(Model)
}

func (rps *rpcPluginStructField) run() {

	if rps.structDir == "" {
		rps.logger.Panicf("%s is empty", rps.structDir)
	}

	if rps.StructName == "" {
		rps.logger.Panicf("%s is empty", rps.StructName)
	}

	rps.pluginClient = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(rps.structDir + string(filepath.Separator) + "plugin" + string(filepath.Separator) + "plugin"),
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: hclog.DefaultOutput,
			Level:  hclog.Error,
			Name:   "plugin",
		}),
	})

	rps.buildPluginEnv()

	rps.dispense()
}

func (rps *rpcPluginStructField) SortField(fields []pkg.Field) *SortReturn {

	rps.Fields = fields

	if rps.model == nil {
		rps.run()
	}

	sortReturn := &SortReturn{}
	err := json.Unmarshal([]byte(rps.model.Sort()), sortReturn)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	return sortReturn
}

func (rps *rpcPluginStructField) InitField(fields []pkg.Field) *InitFieldsReturn {
	rps.Fields = fields

	if rps.model == nil {
		rps.run()
	}

	initReturn := &InitFieldsReturn{}
	err := json.Unmarshal([]byte(rps.model.InitField()), initReturn)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	return initReturn
}

func (rps *rpcPluginStructField) SetStructDir(structDir string) {
	rps.structDir = structDir
}

func (rps *rpcPluginStructField) SetStructName(structName string) {
	rps.StructName = structName
}

func (rps *rpcPluginStructField) SetStructInfo(s *structInfo) {
	rps.StrcutInfo = s
}

func (rps *rpcPluginStructField) SetStructFileName(structFileName string) {
	rps.StructFileName = structFileName
}

func (rps *rpcPluginStructField) SetFilesName(filesName []string) {
	rps.filesName = filesName
}

func (rps *rpcPluginStructField) SetPackName(packName string) {
	rps.packName = packName
}

func (rps *rpcPluginStructField) Close() {
	rps.pluginClient.Kill()
	rps.clear()
}

func (rps *rpcPluginStructField) clear() {
	err := os.RemoveAll(rps.structDir + string(filepath.Separator) + "plugin")
	if err != nil {
		rps.logger.Panicf(err.Error())
	}
}

func (rps *rpcPluginStructField) GenInitFieldStr(getType reflect.Type, fieldLink, initName string, specFilds *pkg.Fields) []string {
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
					rps.GenInitFieldStr(getType.Field(i).Type.Elem(), fieldLink+"."+getType.Field(i).Name,
						initName+"."+getType.Field(i).Name, nil)...)
			default:
				initStr = rps.KindToInit(getType.Field(i).Type.Elem())
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
				structFields = append(structFields, rps.GenInitFieldStr(getType.Field(i).Type,
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
			initStr = rps.KindToInit(getType.Field(i).Type)
		}

		structFields = append(structFields, fieldLink+"."+getType.Field(i).Name+" = "+initStr)
	}

	return structFields
}

func (rps *rpcPluginStructField) KindToInit(refType reflect.Type) string {
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
