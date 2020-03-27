package factory

import (
	"strings"
	"os"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/tool/db2entity"
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
	"html/template"
)

type StructFieldIface interface {
	SortField() *SortReturn

	InitField() *InitFieldsReturn

	Close()

	SetStructDir(string)

	SetStructName(string)

	SetFields(fields []db2entity.Field)
}


type rpcPluginStructField struct{

	logger log2.Logger

	structDir string

	StructName string

	packName string

	filesName []string

	buildBeforeFunc func()

	Fields []db2entity.Field

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


// gen modelName_plugin.go
func (this *rpcPluginStructField) genStructPlugin(dir string)  {

	tmpl, err := template.New("rpc_plugin").Funcs(EsimFuncMap()).
		Parse(rpcPluginTemplate)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	err = tmpl.Execute(os.Stdout, this)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, this)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	file := dir + string(filepath.Separator) + this.StructName + "_plugin.go"
	err = ioutil.WriteFile(file, buf.Bytes(), 0666)
	if err != nil {
		this.logger.Panicf(err.Error())
	}

	return
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

	if this.StructName == ""{
		this.logger.Panicf("%s is empty", this.StructName)
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
	this.StructName = structName
}

func (this *rpcPluginStructField) SetFields(fields []db2entity.Field)  {
	this.Fields = fields
}

func (this *rpcPluginStructField) Close()  {
	this.pluginClient.Kill()
	this.clear()
}

func (this *rpcPluginStructField) clear() {
	err := os.RemoveAll(this.structDir + string(filepath.Separator) + "plugin")
	this.logger.Panicf(err.Error())
}