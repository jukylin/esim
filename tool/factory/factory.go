package factory

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"fmt"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	file_dir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/martinusso/inflect"
	"github.com/serenize/snaker"
	"github.com/spf13/viper"
	"golang.org/x/tools/imports"
)

type SortReturn struct {
	Fields pkg.Fields `json:"fields"`
}

type InitFieldsReturn struct {
	Fields     []string    `json:"fields"`
	SpecFields []pkg.Field `json:"SpecFields"`
}

// +-----------+-----------+
// | firstPart |	package	  |
// |			  |	import	  |
// |	----------|	----------|
// | secondPart| var		  |
// |			  |	     	  |
// |	----------|	----------|
// | thirdPart | struct	  |
// |			  |	funcBody  |
// |	----------|	----------|
type EsimFactory struct {
	// struct name which be search
	StructName string

	// struct Absolute path
	structDir string

	filesName []string

	// package {{.packName}}
	packName string

	// package main => {{.packStr}}
	packStr string

	// File where the Struct is located
	structFileName string

	// Found Struct
	oldStructInfo *structInfo

	// Struct will be create
	NewStructInfo *structInfo

	// true if find the StructName
	// false if not
	found bool

	withPlural bool

	withOption bool

	withGenLoggerOption bool

	withGenConfOption bool

	withPrint bool

	structFieldIface StructFieldIface

	ReleaseStr string

	firstPart string

	secondPart string

	thirdPart string

	InitField *InitFieldsReturn

	// Struct plural form
	pluralName string

	NewPluralStr string

	ReleasePluralStr string

	TypePluralStr string

	// option start
	Option1 string

	Option2 string

	Option3 string

	Option4 string

	Option5 string

	// option end

	OptionParam string

	logger log.Logger

	withSort bool

	withImpIface string

	withPool bool

	withStar bool

	WithNew bool

	writer file_dir.IfaceWriter

	SpecFieldInitStr string

	ReturnStr string

	StructTpl templates.StructInfo

	tpl templates.Tpl
}

type Option func(*EsimFactory)

func NewEsimFactory(options ...Option) *EsimFactory {
	factory := &EsimFactory{}

	for _, option := range options {
		option(factory)
	}

	if factory.logger == nil {
		factory.logger = log.NewLogger()
	}

	if factory.writer == nil {
		factory.writer = file_dir.NewEsimWriter()
	}

	if factory.tpl == nil {
		factory.tpl = templates.NewTextTpl()
	}

	factory.oldStructInfo = &structInfo{}

	factory.NewStructInfo = &structInfo{}

	factory.structFieldIface = NewRPCPluginStructField(factory.writer, factory.logger)

	factory.StructTpl = templates.StructInfo{}

	return factory
}

func WithEsimFactoryLogger(logger log.Logger) Option {
	return func(ef *EsimFactory) {
		ef.logger = logger
	}
}

func WithEsimFactoryWriter(writer file_dir.IfaceWriter) Option {
	return func(ef *EsimFactory) {
		ef.writer = writer
	}
}

func WithEsimFactoryTpl(tpl templates.Tpl) Option {
	return func(ef *EsimFactory) {
		ef.tpl = tpl
	}
}

type structInfo struct {
	Fields pkg.Fields

	structStr string

	structFileContent string

	vars pkg.Vars

	varStr string

	imports pkg.Imports

	importStr string

	ReturnVarStr string

	StructInitStr string
}

// getPluralWord Struct plural form
// If plural is not obtained, add "s" at the end of the word
func (ef *EsimFactory) getPluralForm(word string) string {
	newWord := inflect.Pluralize(word)
	if newWord == word || newWord == "" {
		newWord = word + "s"
	}

	return newWord
}

func (ef *EsimFactory) Run(v *viper.Viper) error {
	err := ef.bindInput(v)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	if !ef.parseStruct() {
		ef.logger.Panicf("not found struct %s", ef.StructName)
	}
	ef.copyOldStructInfo()

	if ef.extendField() {
		err = ef.buildNewStructFileContent()
		if err != nil {
			ef.logger.Panicf(err.Error())
		}
	}

	ef.logger.Debugf("fields len %d", ef.oldStructInfo.Fields.Len())

	if ef.oldStructInfo.Fields.Len() > 0 {
		ef.structFieldIface.SetStructInfo(ef.NewStructInfo)
		ef.structFieldIface.SetStructName(ef.StructName)
		ef.structFieldIface.SetStructDir(ef.structDir)
		ef.structFieldIface.SetStructFileName(ef.structFileName)
		ef.structFieldIface.SetFilesName(ef.filesName)
		ef.structFieldIface.SetPackName(ef.packName)

		if ef.withSort {
			sortedField := ef.structFieldIface.SortField(ef.NewStructInfo.Fields)
			ef.logger.Debugf("sorted fields %+v", sortedField.Fields)
			ef.NewStructInfo.Fields = sortedField.Fields
		}

		ef.InitField = ef.structFieldIface.InitField(ef.NewStructInfo.Fields)
	}

	ef.genStr()

	ef.assignStructTpl()

	ef.executeNewTmpl()

	ef.organizePart()

	if ef.withPrint {
		ef.printResult()
	} else {
		err = file_dir.EsimBackUpFile(ef.structDir +
			string(filepath.Separator) + ef.structFileName)
		if err != nil {
			ef.logger.Warnf("backup err %s:%s", ef.structDir+
				string(filepath.Separator)+ef.structFileName,
				err.Error())
		}

		originContent := ef.replaceOriginContent()

		res, err := imports.Process("", []byte(originContent), nil)
		if err != nil {
			ef.logger.Panicf("%s:%s", err.Error(), originContent)
		}

		err = file_dir.EsimWrite(ef.structDir+
			string(filepath.Separator)+ef.structFileName,
			string(res))
		if err != nil {
			ef.logger.Panicf(err.Error())
		}
	}

	return nil
}

func (ef *EsimFactory) assignStructTpl() {
	ef.StructTpl.StructName = ef.StructName
	ef.StructTpl.Fields = ef.NewStructInfo.Fields
}

func (ef *EsimFactory) executeNewTmpl() {
	content, err := ef.tpl.Execute("factory", newTemplate, ef)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	ef.NewStructInfo.structStr = content
}

// replaceOriginContent gen a new struct file content
func (ef *EsimFactory) replaceOriginContent() string {
	var newContent string
	originContent := ef.oldStructInfo.structFileContent
	newContent = originContent
	if ef.oldStructInfo.importStr != "" {
		newContent = strings.Replace(newContent, ef.oldStructInfo.importStr, "", 1)
	}

	newContent = strings.Replace(newContent, ef.packStr, ef.firstPart, 1)

	if ef.secondPart != "" {
		newContent = strings.Replace(newContent, ef.oldStructInfo.varStr, ef.secondPart, 1)
	}

	newContent = strings.Replace(newContent, ef.oldStructInfo.structStr, ef.thirdPart, 1)
	ef.NewStructInfo.structFileContent = newContent

	return newContent
}

// printResult println file content to terminal
func (ef *EsimFactory) printResult() {
	src := ef.firstPart + "\n"
	src += ef.secondPart + "\n"
	src += ef.thirdPart + "\n"

	res, err := imports.Process("", []byte(src), nil)
	if err != nil {
		ef.logger.Panicf(err.Error())
	} else {
		println(string(res))
	}
}

// organizePart  organize pack, import, var, struct
func (ef *EsimFactory) organizePart() {
	ef.firstPart = ef.packStr + "\n"
	ef.firstPart += ef.NewStructInfo.importStr + "\n"

	if ef.oldStructInfo.varStr != "" {
		ef.secondPart = ef.NewStructInfo.varStr
	} else {
		// merge firstPart and secondPart
		ef.firstPart += ef.NewStructInfo.varStr
	}

	ef.thirdPart = ef.NewStructInfo.structStr
}

// copy oldStructInfo to NewStructInfo
func (ef *EsimFactory) copyOldStructInfo() {
	copyStructInfo := *ef.oldStructInfo
	ef.NewStructInfo = &copyStructInfo
}

func (ef *EsimFactory) bindInput(v *viper.Viper) error {
	sname := v.GetString("sname")
	if sname == "" {
		return errors.New("sname is empty")
	}
	ef.StructName = sname

	sdir := v.GetString("sdir")
	if sdir == "" {
		sdir = "."
	}

	dir, err := filepath.Abs(sdir)
	if err != nil {
		return err
	}
	ef.structDir = strings.TrimRight(dir, "/")

	plural := v.GetBool("plural")
	if plural {
		ef.withPlural = true
		ef.pluralName = ef.getPluralForm(sname)
	}

	ef.withOption = v.GetBool("option")

	ef.withGenConfOption = v.GetBool("oc")

	ef.withGenLoggerOption = v.GetBool("ol")

	ef.withSort = v.GetBool("sort")

	ef.withImpIface = v.GetString("imp_iface")

	ef.withPool = v.GetBool("pool")

	ef.withStar = v.GetBool("star")

	ef.withPrint = v.GetBool("print")

	ef.WithNew = v.GetBool("new")

	return nil
}

func (ef *EsimFactory) genStr() {
	ef.genReturnVarStr()
	ef.genStructInitStr()
	ef.genSpecFieldInitStr()
	ef.genReturnStr()

	if ef.withOption {
		ef.genOptionParam()
		ef.genOptions()
	}

	if ef.withPool && len(ef.InitField.Fields) > 0 {
		ef.genPool()
	}

	if ef.withPlural {
		ef.genPlural()
	}

	ef.NewStructInfo.varStr = ef.NewStructInfo.vars.String()
}

// find struct
// parse struct
func (ef *EsimFactory) parseStruct() bool {
	exists, err := file_dir.IsExistsDir(ef.structDir)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	if !exists {
		ef.logger.Panicf("%s dir not exists", ef.structDir)
	}

	files, err := ioutil.ReadDir(ef.structDir)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	for _, fileInfo := range files {
		ext := path.Ext(fileInfo.Name())
		if ext != ".go" {
			continue
		}

		//不复制测试文件
		if strings.Contains(fileInfo.Name(), "_test") {
			continue
		}

		ef.filesName = append(ef.filesName, fileInfo.Name())

		if !fileInfo.IsDir() && !ef.found {

			src, err := ioutil.ReadFile(ef.structDir + "/" + fileInfo.Name())
			if err != nil {
				ef.logger.Panicf(err.Error())
			}

			strSrc := string(src)
			fset := token.NewFileSet() // positions are relative to fset
			f, err := parser.ParseFile(fset, "", strSrc, parser.ParseComments)
			if err != nil {
				ef.logger.Panicf(err.Error())
			}

			for _, decl := range f.Decls {
				if genDecl, ok := decl.(*ast.GenDecl); ok {
					if genDecl.Tok.String() == "type" {
						ef.parseType(genDecl, src, fileInfo, f)
					}

					if genDecl.Tok.String() == "var" && ef.found {
						ef.oldStructInfo.vars.ParseFromAst(genDecl, strSrc)
					}

					if genDecl.Tok.String() == "import" && ef.found {
						ef.parseImport(genDecl, src)
					}
				}
			}
		}
	}

	return ef.found
}

func (ef *EsimFactory) parseType(genDecl *ast.GenDecl, src []byte,
	fileInfo os.FileInfo, f *ast.File) {
	strSrc := string(src)
	for _, specs := range genDecl.Specs {
		if typeSpec, ok := specs.(*ast.TypeSpec); ok {
			if typeSpec.Name.String() == ef.StructName {
				ef.oldStructInfo.structFileContent = strSrc
				ef.structFileName = fileInfo.Name()
				//found the struct
				ef.found = true

				ef.packName = f.Name.String()
				ef.packStr = "package " + strSrc[f.Name.Pos()-1:f.Name.End()]

				fields := pkg.Fields{}
				fields.ParseFromAst(genDecl, strSrc)
				ef.oldStructInfo.Fields = fields
				colsePos := typeSpec.Type.(*ast.StructType).Fields.Closing
				structStr := src[genDecl.TokPos-1 : colsePos]
				ef.oldStructInfo.structStr = string(structStr)
			}
		}
	}
}

func (ef *EsimFactory) parseImport(genDecl *ast.GenDecl, src []byte) {
	strSrc := string(src)
	imps := pkg.Imports{}
	imps.ParseFromAst(genDecl)
	ef.oldStructInfo.imports = imps
	ef.oldStructInfo.importStr = strSrc[genDecl.Pos()-1 : genDecl.End()]
}

// extend logger and conf for new struct field
func (ef *EsimFactory) extendField() bool {
	var HasExtend bool
	if ef.withOption {
		if ef.withGenLoggerOption {
			HasExtend = true
			ef.extendLog()
		}

		if ef.withGenConfOption {
			HasExtend = true
			ef.extendConf()
		}
	}

	return HasExtend
}

func (ef *EsimFactory) extendLog() {
	var foundLogField bool
	for _, field := range ef.NewStructInfo.Fields {
		if strings.Contains(field.Field, "log.Logger") && !foundLogField {
			foundLogField = true
		}
	}

	if !foundLogField || len(ef.NewStructInfo.Fields) == 0 {
		fld := pkg.Field{}
		fld.Field = "logger log.Logger"
		fld.Name = "logger"
		ef.NewStructInfo.Fields = append(ef.NewStructInfo.Fields, fld)
	}

	var foundLogImport bool
	for _, oim := range ef.NewStructInfo.imports {
		if oim.Path == "github.com/jukylin/esim/log" {
			foundLogImport = true
		}
	}

	if !foundLogImport {
		ef.appendNewImport("github.com/jukylin/esim/log")
	}
}

func (ef *EsimFactory) extendConf() {
	var foundConfField bool
	for _, field := range ef.NewStructInfo.Fields {
		if strings.Contains(field.Field, "config.Config") && !foundConfField {
			foundConfField = true
		}
	}

	if !foundConfField || len(ef.NewStructInfo.Fields) == 0 {
		fld := pkg.Field{}
		fld.Field = "conf config.Config"
		fld.Name = "conf"
		ef.NewStructInfo.Fields = append(ef.NewStructInfo.Fields, fld)
	}

	var foundConfImport bool
	for _, oim := range ef.NewStructInfo.imports {
		if oim.Path == "github.com/jukylin/esim/config" {
			foundConfImport = true
		}
	}

	if !foundConfImport {
		ef.appendNewImport("github.com/jukylin/esim/config")
	}
}

// if struct field had extend logger or conf
// so build a new struct
func (ef *EsimFactory) buildNewStructFileContent() error {
	ef.NewStructInfo.importStr = ef.NewStructInfo.imports.String()

	if ef.oldStructInfo.importStr != "" {
		ef.NewStructInfo.structFileContent = strings.Replace(ef.oldStructInfo.structFileContent,
			ef.oldStructInfo.importStr, ef.NewStructInfo.importStr, -1)
	} else if ef.packStr != "" {
		//not find import
		ef.getFirstPart()
		ef.NewStructInfo.structFileContent = strings.Replace(ef.oldStructInfo.structFileContent,
			ef.packStr, ef.firstPart, -1)
	} else {
		return errors.New("can't build the first part")
	}

	structInfo := templates.NewStructInfo()
	structInfo.StructName = ef.StructName
	structInfo.Fields = ef.NewStructInfo.Fields

	ef.NewStructInfo.structStr = structInfo.String()

	ef.NewStructInfo.structFileContent = strings.Replace(ef.NewStructInfo.structFileContent,
		ef.oldStructInfo.structStr, ef.NewStructInfo.structStr, -1)

	src, err := imports.Process("", []byte(ef.NewStructInfo.structFileContent), nil)
	if err != nil {
		return fmt.Errorf("%s : %s", err.Error(), ef.NewStructInfo.structFileContent)
	}

	ef.NewStructInfo.structFileContent = string(src)

	return nil
}

//
// func NewStruct() {{.ReturnvarStr}} {
// }
//
func (ef *EsimFactory) genReturnVarStr() {
	if ef.withImpIface != "" {
		ef.NewStructInfo.ReturnVarStr = ef.withImpIface
	} else if ef.withPool || ef.withStar {
		ef.NewStructInfo.ReturnVarStr = "*" + ef.StructName
	} else {
		ef.NewStructInfo.ReturnVarStr = ef.StructName
	}
}

// type Option func(c *{{.OptionParam}})
func (ef *EsimFactory) genOptionParam() {
	if ef.withPool || ef.withStar {
		ef.OptionParam = "*" + ef.StructName
	} else {
		ef.OptionParam = ef.StructName
	}
}

// StructObj := Struct{} => {{.StructInitStr}}
func (ef *EsimFactory) genStructInitStr() {
	var structInitStr string
	if ef.withStar {
		structInitStr = templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) +
			" := &" + ef.StructName + "{}"
	} else if ef.withPool {
		structInitStr = templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) + ` := ` +
			templates.FirstToLower(snaker.SnakeToCamelLower(ef.StructName)) + `Pool.Get().(*` +
			ef.StructName + `)`
	} else {
		structInitStr = templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) +
			" := " + ef.StructName + "{}"
	}

	ef.NewStructInfo.StructInitStr = structInitStr
}

// return {{.ReturnStr}}
func (ef *EsimFactory) genReturnStr() {
	ef.ReturnStr = templates.Shorten(snaker.SnakeToCamelLower(ef.StructName))
}

func (ef *EsimFactory) genOptions() {
	ef.Option1 = `type ` + ef.StructName + `Option func(` + ef.OptionParam + `)`

	ef.Option2 = `options ...` + ef.StructName + `Option`

	ef.Option3 = `
	for _, option := range options {
		option(` + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) + `)
	}`

	if ef.withGenConfOption {

		ef.Option4 += `
func With` + ef.StructName + `Conf(conf config.Config) ` + ef.StructName + `Option {
	return func(` + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) +
			` ` + ef.NewStructInfo.ReturnVarStr + `) {
	` + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) + `.conf = conf
	}
}
`
	}

	if ef.withGenLoggerOption {
		ef.Option5 += `
func With` + ef.StructName + `Logger(logger log.Logger) ` + ef.StructName + `Option {
	return func(` + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) +
			` ` + ef.NewStructInfo.ReturnVarStr + `) {
		` + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) + `.logger = logger
	}
}
`
	}
}

func (ef *EsimFactory) genPool() {
	ef.incrPoolVar(ef.StructName)

	ef.ReleaseStr = ef.genReleaseStructStr()
}

func (ef *EsimFactory) genPlural() {
	ef.incrPoolVar(ef.pluralName)

	plural := NewPlural()
	plural.StructName = ef.StructName
	plural.PluralName = ef.pluralName
	if ef.withStar {
		plural.Star = "*"
	}

	ef.TypePluralStr = plural.TypeString()

	ef.NewPluralStr = plural.NewString()

	ef.ReleasePluralStr = plural.ReleaseString()
}

func (ef *EsimFactory) incrPoolVar(structName string) {
	poolName := structName + "Pool"
	if ef.varNameExists(ef.NewStructInfo.vars, poolName) {
		ef.logger.Debugf("var is exists : %s", poolName)
	} else {

		ef.NewStructInfo.vars = append(ef.NewStructInfo.vars,
			ef.appendPoolVar(poolName, structName))
		ef.appendNewImport("sync")
	}
}

//nolint:goconst
func (ef *EsimFactory) genSpecFieldInitStr() {
	var str string

	for _, f := range ef.InitField.SpecFields {
		if f.Type == "slice" {
			str += `if ` + f.Name + ` == nil {
`
			newTypeName := strings.Replace(f.TypeName, "main.", "", -1)
			str += f.Name + ` = make(` + newTypeName + `, 0)
`
			str += `}
`
		}

		if f.Type == "map" {

			str += `if ` + f.Name + ` == nil {
`
			str += f.Name + ` = make(` + f.TypeName + `)
`
			str += `}
`
		}
	}

	ef.SpecFieldInitStr = str
}

func (ef *EsimFactory) genReleaseStructStr() string {
	str := "func (" + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) +
		"  " + ef.NewStructInfo.ReturnVarStr + ") Release() {\n"

	for _, field := range ef.InitField.Fields {
		if strings.Contains(field, "time.Time") {
			ef.appendNewImport("time")
		}
		str += "		" + field + "\n"
	}

	str += "		" + templates.FirstToLower(snaker.SnakeToCamelLower(ef.StructName)) +
		"Pool.Put(" + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) + ")\n"
	str += "}"

	return str
}

func (ef *EsimFactory) appendNewImport(importName string) {
	var found bool
	for _, imp := range ef.NewStructInfo.imports {
		if imp.Path == importName {
			found = true
		}
	}

	if !found {
		ef.NewStructInfo.imports = append(ef.NewStructInfo.imports, pkg.Import{Path: importName})
	}
}

func (ef *EsimFactory) appendPoolVar(pollVarName, structName string) pkg.Var {
	var poolVar pkg.Var

	pooltpl := NewPoolTpl()
	pooltpl.VarPoolName = pollVarName
	pooltpl.StructName = structName

	poolVar.Body = pooltpl.String()

	poolVar.Name = append(poolVar.Name, pollVarName)
	return poolVar
}

// 变量是否存在
func (ef *EsimFactory) varNameExists(vars pkg.Vars, poolVarName string) bool {
	for _, varInfo := range vars {
		for _, varName := range varInfo.Name {
			if varName == poolVarName {
				return true
			}
		}
	}

	return false
}

// package + import
func (ef *EsimFactory) getFirstPart() {
	if ef.oldStructInfo.importStr == "" {
		ef.firstPart += ef.packStr + "\n\n"
	}

	ef.firstPart += ef.NewStructInfo.importStr
}

func (ef *EsimFactory) Close() {
	ef.structFieldIface.Close()
}
