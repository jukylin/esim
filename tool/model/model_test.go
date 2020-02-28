package model

import (
	"github.com/spf13/viper"
	"github.com/jukylin/esim/pkg/file-dir"
	"os"
	"testing"
	"unsafe"
)

func getCurDir() string {
	modelpath, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}

	return modelpath
}

func delModelFile() {
	os.Remove(getCurDir() + "/plugin/model.go")
	os.Remove(getCurDir() + "/plugin/model_test.go")
}

func TestFindModel(t *testing.T) {
	modelName := "Test"
	modelPath := getCurDir()

	info, err := FindModel(modelPath, modelName, "")
	if err != nil {
		t.Error(err.Error())
		return
	}
	if info.packName == "" {
		t.Error("error")
		return
	}
}

func TestBuildVirEnv(t *testing.T) {

	modelName := "Test"
	modelPath := getCurDir()

	info, err := FindModel(modelPath, modelName, "")
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = BuildPluginEnv(info, delModelFile)
	if err != nil {
		t.Error(err.Error())
		return
	}

	e, err := file_dir.IsExistsDir("./plugin")
	if err != nil {
		t.Error(err.Error())
		return
	}

	if e == false {
		t.Error("plugin 创建失败")
		return
	}
	Clear(info)
}

func TestExecPlugin(t *testing.T) {

	modelName := "Test"
	modelPath := getCurDir()

	v := viper.New()
	v.Set("sort", false)
	v.Set("pool", false)

	info, err := FindModel(modelPath, modelName, "")
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = BuildPluginEnv(info, delModelFile)
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = ExecPlugin(v, info)
	if err != nil {
		t.Error(err.Error())
		return
	}

	Clear(info)
}

func TestWriteContent(t *testing.T) {
	modelName := "Test"
	modelPath := getCurDir()

	v := viper.New()
	v.Set("sort", true)
	v.Set("pool", false)
	v.Set("coverpool", false)
	v.Set("plural", false)

	info, err := FindModel(modelPath, modelName, getPluralWord(modelName))
	if err != nil {
		t.Error(err.Error())
		return
	}

	err = BuildPluginEnv(info, delModelFile)
	if err != nil{
		t.Error(err.Error())
		return
	}

	err = ExecPlugin(v, info)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = WriteContent(v, info)
	if err != nil {
		t.Error(err.Error())
		return
	}
	Clear(info)
}

func TestClear(t *testing.T) {
	modelName := "Test"
	modelPath := getCurDir()

	info, err := FindModel(modelPath, modelName, "")
	if err != nil {
		t.Error(err.Error())
	}
	Clear(info)
}

func TestSize(t *testing.T) {
	test := Test{}

	println(unsafe.Sizeof(test))
}

func TestNewFrame(t *testing.T)  {

	v := viper.New()
	v.Set("gen_logger_option", true)
	v.Set("gen_conf_option", true)
	v.Set("star", true)

	info := &BuildPluginInfo{}
	info.modelName = "TestFrame"

	NewVarStr(v, info)
	frame := NewFrame(v, info)
	getOptions(v, info)

	newFrame := replaceOptions(frame, info)
	println(newFrame)
}