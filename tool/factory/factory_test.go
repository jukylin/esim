package factory

import (
	//"github.com/spf13/viper"
	//"github.com/jukylin/esim/pkg/file-dir"
	"os"
	"testing"
	//"strings"
	"github.com/stretchr/testify/assert"
	//"golang.org/x/tools/imports"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/tool/db2entity"
	"github.com/spf13/viper"
)


func TestMain(m *testing.M) {


	setUp()


	code := m.Run()


	os.Exit(code)
}


var esimfactory *esimFactory


func setUp()  {
	esimfactory = NewEsimFactory()
}


func getCurDir() string {
	modelpath, err := os.Getwd()
	if err != nil {
		println(err.Error())
	}

	return modelpath
}


func TestEsimFactory_Run(t *testing.T) {
	v := viper.New()
	v.Set("sname", "Test")
	v.Set("option", true)
	v.Set("sort", true)
	v.Set("gen_logger_option", true)
	v.Set("gen_conf_option", true)
	v.Set("sdir", "./example")
	v.Set("pool", true)
	v.Set("plural", true)
	v.Set("print", true)

	esimfactory.Run(v)
	esimfactory.Close()
}

func TestEsimFactory_InputBind(t *testing.T) {
	v := viper.New()
	v.Set("sname", "Test")
	v.Set("option", true)
	v.Set("sort", true)
	v.Set("gen_logger_option", true)
	v.Set("gen_conf_option", true)
	v.Set("print", true)

	v.Set("sdir", "./example")
	err := esimfactory.inputBind(v)
	assert.Nil(t, err)
}

//func delModelFile() {
//	os.Remove(getCurDir() + "/plugin/model.go")
//	os.Remove(getCurDir() + "/plugin/model_test.go")
//}
//
//
//func TestFindModel(t *testing.T) {
//	modelName := "Test"
//	modelPath := getCurDir() + "/example"
//
//	info, err := FindModel(modelPath, modelName, "")
//	if err != nil {
//		t.Error(err.Error())
//		return
//	}
//	if info.packName == "" {
//		t.Error("error")
//		return
//	}
//}
//
//
//func TestFinalContent(t *testing.T) {
//
//	result := `package example
//
//type Test struct {
//	c int8
//
//	i bool
//
//	g byte
//
//	d int16
//
//	f float32
//
//	a int32
//
//	b int64
//
//	m map[string]interface{}
//
//	e string
//
//	h []int
//
//	u [3]string
//}
//
//type Tests []Test
//`
//
//	modelName := "Test"
//	modelPath := getCurDir() + "/example"
//
//	v := viper.New()
//	v.Set("sort", true)
//	v.Set("pool", false)
//	v.Set("coverpool", false)
//	v.Set("plural", false)
//
//	info, err := FindModel(modelPath, modelName, getPluralWord(modelName))
//	if err != nil {
//		t.Error(err.Error())
//		return
//	}
//
//	err = BuildPluginEnv(info, delModelFile)
//	if err != nil{
//		t.Error(err.Error())
//		return
//	}
//
//	err = ExecPlugin(v, info)
//	if err != nil {
//		t.Error(err.Error())
//		return
//	}
//
//	BuildFrame(v, info)
//
//	src, err := ReplaceContent(v, info)
//	if err != nil {
//		t.Error(err.Error())
//		return
//	}
//
//	res, err := imports.Process("", []byte(src), nil)
//	if err != nil {
//		t.Error(err.Error())
//		return
//	}
//
//	assert.Equal(t, result, string(res))
//
//	Clear(info)
//}
//
//
//func TestClear(t *testing.T) {
//	modelName := "Test"
//	modelPath := getCurDir() + "/example"
//
//	info, err := FindModel(modelPath, modelName, "")
//	if err != nil {
//		t.Error(err.Error())
//	}
//	Clear(info)
//}
//
//
//func TestNewFrame(t *testing.T)  {
//
//	v := viper.New()
//	v.Set("gen_logger_option", true)
//	v.Set("gen_conf_option", true)
//	v.Set("star", true)
//
//	info := &BuildPluginInfo{}
//	info.modelName = "TestFrame"
//
//	NewVarStr(v, info)
//	NewOptionParam(v, info)
//	frame := NewFrame(v, info)
//	getOptions(v, info)
//
//	newFrame := replaceFrame(frame, info)
//	assert.Empty(t, newFrame)
//}
//
//
//func TestGetNewImport(t *testing.T)  {
//
//	result := `import (
//        test
//        test2
//)
//`
//
//	imports := []string{"test", "test2"}
//
//	newImport := getNewImport(imports)
//
//	assert.Equal(t,  result, newImport)
//}



func TestCopyOldStructInfo(t *testing.T)  {
	esimfactory.oldStructInfo.imports = append(esimfactory.oldStructInfo.imports, "fmt")
	esimfactory.oldStructInfo.structFileContent = "package main"
	esimfactory.copyOldStructInfo()
	assert.Equal(t, "fmt", esimfactory.NewStructInfo.imports[0])
}


var replaceStructContent = `package main

import (
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
)

type test struct {
	logger log.Logger

	conf config.Config

	a int

	b string
}
`

func TestExtendFieldAndReplaceStructContent(t *testing.T)  {

	esimfactory.withOption = true
	esimfactory.withGenLoggerOption = true
	esimfactory.withGenConfOption = true

	result := esimfactory.ExtendField()
	assert.True(t, result)

	assert.Equal(t, 2, len(esimfactory.NewStructInfo.Fields))
	assert.Equal(t, 2, len(esimfactory.NewStructInfo.imports))

	esimfactory.writer = file_dir.NullWrite{}

	afield := db2entity.Field{}
	afield.Filed = "a int"
	bfield := db2entity.Field{}
	bfield.Filed = "b string"
	esimfactory.NewStructInfo.Fields = append(esimfactory.NewStructInfo.Fields, afield, bfield)

	esimfactory.oldStructInfo.structFileContent = `package main

import (
	"fmt"
	"sync"
)

type test struct {
	a int

	b string
}

`
	esimfactory.oldStructInfo.importStr = `
import (
	"fmt"
	"sync"
)
`

	esimfactory.oldStructInfo.structStr = `
type test struct {
	a int

	b string
}
`
	esimfactory.StructName = "test"
	err := esimfactory.buildNewStructFileContent()
	assert.Nil(t, err)
	assert.Equal(t, replaceStructContent, esimfactory.NewStructInfo.structFileContent)

	esimfactory.oldStructInfo.importStr = ""
	esimfactory.packStr = "package main"
	err = esimfactory.buildNewStructFileContent()
	assert.Nil(t, err)
	assert.Equal(t, replaceStructContent, esimfactory.NewStructInfo.structFileContent)
}


