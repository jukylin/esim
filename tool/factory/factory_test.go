package factory

import (
	"os"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/log"
	"path/filepath"
	"github.com/jukylin/esim/pkg/templates"
)


func TestMain(m *testing.M) {

	setUp()

	code := m.Run()

	os.Exit(code)
}


var esimfactory *esimFactory


func setUp()  {
	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	esimfactory = NewEsimFactory(
		WithEsimFactoryLogger(logger),
		WithEsimFactoryWriter(file_dir.NewEsimWriter()),
		WithEsimFactoryTpl(templates.NewTextTpl()),
	)
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
	v.Set("new", true)
	//v.Set("print", true)

	esimfactory.Run(v)
	esimfactory.Close()
	file_dir.EsimRecoverFile(esimfactory.structDir +
		string(filepath.Separator) + esimfactory.structFileName)
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
	err := esimfactory.bindInput(v)
	assert.Nil(t, err)
}


func TestCopyOldStructInfo(t *testing.T)  {
	esimfactory.oldStructInfo.imports = esimfactory.oldStructInfo.imports[:0]
	esimfactory.oldStructInfo.imports = append(esimfactory.oldStructInfo.imports, pkg.Import{Path:"fmt"})
	esimfactory.oldStructInfo.structFileContent = "package main"
	esimfactory.copyOldStructInfo()
	assert.Equal(t, "fmt", esimfactory.NewStructInfo.imports[0].Path)

	esimfactory.NewStructInfo.varStr = "var ()"
	assert.NotEqual(t, esimfactory.oldStructInfo.varStr, esimfactory.NewStructInfo.varStr)
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
	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	esimfactory = NewEsimFactory(
		WithEsimFactoryLogger(logger),
		WithEsimFactoryWriter(file_dir.NewEsimWriter()),
		WithEsimFactoryTpl(templates.NewTextTpl()),
	)

	esimfactory.withOption = true
	esimfactory.withGenLoggerOption = true
	esimfactory.withGenConfOption = true

	result := esimfactory.extendField()
	assert.True(t, result)

	assert.Equal(t, 2, len(esimfactory.NewStructInfo.Fields))
	assert.Equal(t, 2, len(esimfactory.NewStructInfo.imports))

	esimfactory.writer = file_dir.NewNullWrite()

	afield := pkg.Field{}
	afield.Field = "a int"
	bfield := pkg.Field{}
	bfield.Field = "b string"
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


