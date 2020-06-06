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
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"golang.org/x/tools/imports"
)

type StructFieldIface interface {
	HandleField(fields []pkg.Field, data interface{})

	Close()

	SetStructInfo(*structInfo)

	SetStructName(string)

	SetStructDir(string)

	SetStructFileName(string)

	SetFilesName(filesName []string)

	SetPackName(packName string)
}

type RPCPluginStructField struct {
	logger log.Logger

	structDir string

	StructName string

	StructFileName string

	StrcutInfo *structInfo

	packName string

	filesName []string

	buildBeforeFunc func()

	Fields []pkg.Field

	pluginClient *plugin.Client

	model Model

	writer filedir.IfaceWriter
}

var pluginMap = map[string]plugin.Plugin{
	"model": &ModelPlugin{},
}

var HandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

func NewRPCPluginStructField(writer filedir.IfaceWriter,
	logger log.Logger) *RPCPluginStructField {
	rpcPlugin := &RPCPluginStructField{}

	rpcPlugin.writer = writer

	rpcPlugin.logger = logger

	return rpcPlugin
}

func (rps *RPCPluginStructField) buildPluginEnv() error {
	rps.structDir = strings.TrimRight(rps.structDir, string(filepath.Separator))

	targetDir := rps.structDir + string(filepath.Separator) + "plugin"
	exists, err := filedir.IsExistsDir(targetDir)
	if err != nil {
		return err
	}

	if !exists {
		err = filedir.CreateDir(targetDir)
		if err != nil {
			return err
		}
	}

	// copy go file
	// rename package
	reg, _ := regexp.Compile(`package[\s*]` + rps.packName)

	for _, name := range rps.filesName {
		if name == rps.StructFileName {
			src := reg.ReplaceAll([]byte(rps.StrcutInfo.structFileContent),
				[]byte("package main"))
			err = rps.writer.Write(targetDir+string(filepath.Separator)+rps.StructFileName,
				string(src))
			if err != nil {
				rps.logger.Errorf(err.Error())
			}
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

// Copy File.
// repackagename.
func (rps *RPCPluginStructField) copyFile(dstName, srcName string, reg *regexp.Regexp) {
	src, err := os.Open(srcName)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}
	defer dst.Close()

	contents, err := ioutil.ReadAll(src)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	contents = reg.ReplaceAll(contents, []byte("package main"))

	_, err = dst.Write(contents)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}
}

// gen modelName_plugin.go.
func (rps *RPCPluginStructField) genStructPlugin(dir string) {
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
	err = ioutil.WriteFile(file, src, 0600)
	if err != nil {
		rps.logger.Panicf(err.Error())
	}
}

func (rps *RPCPluginStructField) buildPlugin(dir string) error {
	cmdLine := fmt.Sprintf("build -o %s/plugin %s", dir, dir)

	rps.logger.Infof("%s %s", "go", cmdLine)

	args := strings.Split(cmdLine, " ")

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}

	return os.Chmod(dir+string(filepath.Separator)+"plugin", 0777)
}

func (rps *RPCPluginStructField) dispense() {
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

func (rps *RPCPluginStructField) run() {
	if rps.structDir == "" {
		rps.logger.Panicf("%s is empty", rps.structDir)
	}

	if rps.StructName == "" {
		rps.logger.Panicf("%s is empty", rps.StructName)
	}

	cmdPath := rps.structDir + string(filepath.Separator) + "plugin" +
		string(filepath.Separator) + "plugin"
	cmd := exec.Command(cmdPath)

	rps.pluginClient = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             cmd,
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: hclog.DefaultOutput,
			Level:  hclog.Error,
			Name:   "plugin",
		}),
	})

	err := rps.buildPluginEnv()
	if err != nil {
		rps.logger.Panicf(err.Error())
	}

	rps.dispense()
}

func (rps *RPCPluginStructField) HandleField(fields []pkg.Field, data interface{}) {
	rps.Fields = fields

	if rps.model == nil {
		rps.run()
	}

	switch d := data.(type) {
	case *SortReturn:
		err := json.Unmarshal([]byte(rps.model.Sort()), data)
		if err != nil {
			rps.logger.Panicf(err.Error())
		}
	case *InitFieldsReturn:
		err := json.Unmarshal([]byte(rps.model.InitField()), data)
		if err != nil {
			rps.logger.Panicf(err.Error())
		}
	default:
		rps.logger.Panicf("unknow type %T", d)
	}
}

func (rps *RPCPluginStructField) SetStructDir(structDir string) {
	rps.structDir = structDir
}

func (rps *RPCPluginStructField) SetStructName(structName string) {
	rps.StructName = structName
}

func (rps *RPCPluginStructField) SetStructInfo(s *structInfo) {
	rps.StrcutInfo = s
}

func (rps *RPCPluginStructField) SetStructFileName(structFileName string) {
	rps.StructFileName = structFileName
}

func (rps *RPCPluginStructField) SetFilesName(filesName []string) {
	rps.filesName = filesName
}

func (rps *RPCPluginStructField) SetPackName(packName string) {
	rps.packName = packName
}

func (rps *RPCPluginStructField) Close() {
	rps.pluginClient.Kill()
	rps.clear()
}

func (rps *RPCPluginStructField) clear() {
	err := os.RemoveAll(rps.structDir + string(filepath.Separator) + "plugin")
	if err != nil {
		rps.logger.Panicf(err.Error())
	}
}

func (rps *RPCPluginStructField) GenInitFieldStr(getType reflect.Type, fieldLink,
	initName string, specFilds *pkg.Fields) []string {
	typeNum := getType.NumField()
	var structFields []string
	var initStr string
	field := pkg.Field{}

	for i := 0; i < typeNum; i++ {
		switch getType.Field(i).Type.Kind() {
		case reflect.Array:
			structFields = append(structFields, "for k, _ := range "+
				fieldLink+"."+getType.Field(i).Name+" {")
			switch getType.Field(i).Type.Elem().Kind() {
			case reflect.Struct:
				structFields = append(structFields,
					rps.GenInitFieldStr(getType.Field(i).Type.Elem(),
						fieldLink+"."+getType.Field(i).Name,
						initName+"."+getType.Field(i).Name, nil)...)
			default:
				initStr = rps.KindToInit(getType.Field(i).Type.Elem())
				structFields = append(structFields, fieldLink+"."+
					getType.Field(i).Name+"[k] = "+initStr)
			}
			structFields = append(structFields, "}")
			continue
		case reflect.Map:
			structFields = append(structFields, "for k, _ := range "+fieldLink+"."+
				getType.Field(i).Name+" {",
				"delete("+fieldLink+"."+
					getType.Field(i).Name+", k)",
				"}")
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
					fieldLink+"."+getType.Field(i).Name,
					initName+"."+getType.Field(i).Name, nil)...)
				continue
			}
		case reflect.Slice:
			if specFilds != nil {
				field.Name = initName + "." + getType.Field(i).Name
				field.TypeName = getType.Field(i).Type.String()
				field.Type = "slice"
				*specFilds = append(*specFilds, field)
			}
			structFields = append(structFields, fieldLink+"."+getType.Field(i).Name+
				" = "+fieldLink+"."+getType.Field(i).Name+"[:0]")

			continue
		default:
			initStr = rps.KindToInit(getType.Field(i).Type)
		}

		structFields = append(structFields, fieldLink+"."+getType.Field(i).Name+
			" = "+initStr)
	}

	return structFields
}

//nolint:goconst
func (rps *RPCPluginStructField) KindToInit(refType reflect.Type) string {
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
