package factory

import (
	"errors"
	"github.com/martinusso/inflect"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/pkg/file-dir"
	logger "github.com/jukylin/esim/log"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path"
	"strings"
	"golang.org/x/tools/imports"
	"path/filepath"
	"text/template"
	"bytes"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/serenize/snaker"
)


type SortReturn struct {
	Fields pkg.Fields `json:"fields"`
}


type InitFieldsReturn struct {
	Fields     []string `json:"fields"`
	SpecFields []pkg.Field  `json:"SpecFields"`
}


type Var struct {
	doc  []string
	val  string
	name string
}

//+-----------+-----------+
//| firstPart |	package	  |
//|			  |	import	  |
//|	----------|	----------|
//| secondPart| var		  |
//|			  |	     	  |
//|	----------|	----------|
//| thirdPart | struct	  |
//|			  |	funcBody  |
//|	----------|	----------|
type esimFactory struct {

	//struct name which be search
	StructName string

	//struct Absolute path
	structDir string

	filesName []string

	//package {{.packName}}
	packName string

	//package main => {{.packStr}}
	packStr string

	//File where the Struct is located
	structFileName string

	//Found Struct
	oldStructInfo *structInfo

	//Struct will be create
	NewStructInfo *structInfo

	//true if find the StructName
	//false if not
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

	//Struct plural form
	pluralName string

	NewPluralStr string

	ReleasePluralStr string

	TypePluralStr string

	//option start
	Option1 string

	Option2 string

	Option3 string

	Option4 string

	Option5 string

	Option6 string
	//option end

	OptionParam string

	logger logger.Logger

	withSort bool

	withImpIface string

	withPool bool

	withStar bool

	WithNew bool

	writer file_dir.IfaceWriter

	SpecFieldInitStr string

	ReturnStr string

	StructTpl templates.StructInfo
}

func NewEsimFactory(logger logger.Logger) *esimFactory {
	factory := &esimFactory{}

	factory.oldStructInfo = &structInfo{}

	factory.NewStructInfo = &structInfo{}

	factory.logger = logger

	factory.writer = file_dir.NewEsimWriter()

	factory.structFieldIface = NewRpcPluginStructField(factory.writer, logger)

	factory.StructTpl = templates.StructInfo{}

	return factory
}

type structInfo struct{

	Fields pkg.Fields

	structStr string

	//
	structFileContent string

	vars pkg.Vars

	varBody []string

	varStr string

	imports pkg.Imports

	importStr string

	ReturnVarStr string

	StructInitStr string
}

//getPluralWord Struct plural form
//If plural is not obtained, add "s" at the end of the word
func (ef *esimFactory)  getPluralForm(word string) string {
	newWord := inflect.Pluralize(word)
	if newWord == word || newWord == "" {
		newWord = word + "s"
	}

	return newWord
}

func (ef *esimFactory) Run(v *viper.Viper) error {

	err := ef.bindInput(v)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	if ef.parseStruct() == false{
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

		if ef.withSort == true{
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
	}else{
		err = file_dir.EsimBackUpFile(ef.structDir +
			string(filepath.Separator) + ef.structFileName)
		if err != nil{
			ef.logger.Warnf("backup err %s:%s", ef.structDir +
				string(filepath.Separator) + ef.structFileName,
				err.Error())
		}

		originContent := ef.replaceOriginContent()

		res, err := imports.Process("", []byte(originContent), nil)
		if err != nil {
			ef.logger.Panicf("%s:%s", err.Error(), originContent)
		}

		err = file_dir.EsimWrite(ef.structDir +
			string(filepath.Separator) + ef.structFileName,
				string(res))
		if err != nil {
			ef.logger.Panicf(err.Error())
		}
	}

	return nil
}

func (ef *esimFactory) assignStructTpl()  {
	ef.StructTpl.StructName = ef.StructName
	ef.StructTpl.Fields = ef.NewStructInfo.Fields
}

func (ef *esimFactory) executeNewTmpl() {
	tmpl, err := template.New("factory").Funcs(templates.EsimFuncMap()).
		Parse(newTemplate)
	if err != nil{
		ef.logger.Panicf(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, ef)
	if err != nil{
		ef.logger.Panicf(err.Error())
	}

	ef.NewStructInfo.structStr = buf.String()
}

//replaceOriginContent gen a new struct file content
func (ef *esimFactory) replaceOriginContent() string {
	var newContent string
	originContent := ef.oldStructInfo.structFileContent
	newContent = originContent
	if ef.oldStructInfo.importStr != ""{
		newContent = strings.Replace(newContent, ef.oldStructInfo.importStr, "", 1)
	}

	newContent = strings.Replace(newContent, ef.packStr, ef.firstPart, 1)

	if ef.secondPart != ""{
		newContent = strings.Replace(newContent, ef.oldStructInfo.varStr, ef.secondPart, 1)
	}

	newContent = strings.Replace(newContent, ef.oldStructInfo.structStr, ef.thirdPart, 1)
	ef.NewStructInfo.structFileContent = newContent

	return newContent
}

//printResult println file content to terminal
func (ef *esimFactory) printResult()  {
	src := ef.firstPart + "\n"
	src += ef.secondPart + "\n"
	src += ef.thirdPart + "\n"

	res, err := imports.Process("", []byte(src), nil)
	if err != nil{
		ef.logger.Panicf(err.Error())
	}else{
		println(string(res))
	}
}

//organizePart  organize pack, import, var, struct
func (ef *esimFactory) organizePart()  {
	ef.firstPart = ef.packStr + "\n"
	ef.firstPart += ef.NewStructInfo.importStr + "\n"

	if ef.oldStructInfo.varStr != ""{
		ef.secondPart = ef.NewStructInfo.varStr
	}else{
		//merge firstPart and secondPart
		ef.firstPart += ef.NewStructInfo.varStr
	}

	ef.thirdPart = ef.NewStructInfo.structStr
}

//copy oldStructInfo to NewStructInfo
func (ef *esimFactory) copyOldStructInfo()  {
	copyStructInfo := *ef.oldStructInfo
	ef.NewStructInfo = &copyStructInfo
}

func (ef *esimFactory) bindInput(v *viper.Viper) error {
	sname := v.GetString("sname")
	if sname == "" {
		return errors.New("sname is empty")
	}
	ef.StructName = sname

	sdir := v.GetString("sdir")
	if sdir == ""{
		sdir = "."
	}

	dir, err := filepath.Abs(sdir)
	if err != nil{
		return err
	}
	ef.structDir = strings.TrimRight(dir, "/")

	plural := v.GetBool("plural")
	if plural == true {
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


func (ef *esimFactory) genStr()  {

	ef.genReturnVarStr()
	ef.genStructInitStr()
	ef.genSpecFieldInitStr()
	ef.genReturnStr()

	if ef.withOption == true{
		ef.genOptionParam()
		ef.genOptions()
	}

	if ef.withPool == true && len(ef.InitField.Fields) > 0{
		ef.genPool()
	}

	if ef.withPlural {
		ef.genPlural()
	}

	ef.NewStructInfo.varStr = ef.NewStructInfo.vars.String()
}


//@ find struct
//@ parse struct
func (ef *esimFactory) parseStruct() bool {

	exists, err := file_dir.IsExistsDir(ef.structDir)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	if exists == false {
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
		if strings.Index (fileInfo.Name(), "_test") > -1 {
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
				if GenDecl, ok := decl.(*ast.GenDecl); ok {
					if GenDecl.Tok.String() == "type" {
						for _, specs := range GenDecl.Specs {
							if typeSpec, ok := specs.(*ast.TypeSpec); ok {
								if typeSpec.Name.String() == ef.StructName {
									ef.oldStructInfo.structFileContent = strSrc
									ef.structFileName = fileInfo.Name()
									//found the struct
									ef.found = true

									ef.packName = f.Name.String()
									ef.packStr = "package " + strSrc[f.Name.Pos()-1 : f.Name.End()]

									fields := pkg.Fields{}
									fields.ParseFromAst(GenDecl, strSrc)
									ef.oldStructInfo.Fields = fields
									ef.oldStructInfo.structStr = string(src[GenDecl.TokPos-1 : typeSpec.Type.(*ast.StructType).Fields.Closing])
								}
							}
						}
					}
				}
			}

			for _, decl := range f.Decls {
				if GenDecl, ok := decl.(*ast.GenDecl); ok {
					if GenDecl.Tok.String() == "var" && ef.found == true {
						ef.oldStructInfo.vars.ParseFromAst(GenDecl, strSrc)
					}

					if GenDecl.Tok.String() == "import" && ef.found == true {
						imps := pkg.Imports{}
						imps.ParseFromAst(GenDecl)
						ef.oldStructInfo.imports = imps
						ef.oldStructInfo.importStr = strSrc[GenDecl.Pos()-1 : GenDecl.End()]
					}
				}
			}
		}
	}

	return ef.found
}

//extend logger and conf for new struct field
func (ef *esimFactory) extendField() bool {

	var HasExtend bool
	if ef.withOption == true {
		if ef.withGenLoggerOption == true{
			HasExtend = true
			var foundLogField bool
			for _, field := range ef.NewStructInfo.Fields {
				if strings.Contains(field.Field, "log.Logger") == true && foundLogField == false{
					foundLogField = true
				}
			}

			if foundLogField == false || len(ef.NewStructInfo.Fields) == 0{
				fld := pkg.Field{}
				fld.Field = "logger log.Logger"
				fld.Name = "logger"
				ef.NewStructInfo.Fields = append(ef.NewStructInfo.Fields, fld)
			}

			var foundLogImport bool
			for _, oim := range ef.NewStructInfo.imports{
				if oim.Path == "github.com/jukylin/esim/log"{
					foundLogImport = true
				}
			}

			if foundLogImport == false {
				ef.appendNewImport("github.com/jukylin/esim/log")
			}
		}

		if ef.withGenConfOption == true{
			HasExtend = true

			var foundConfField bool
			for _, field := range ef.NewStructInfo.Fields {
				if strings.Contains(field.Field, "config.Config") ==
					true && foundConfField == false{
					foundConfField = true
				}
			}

			if foundConfField == false || len(ef.NewStructInfo.Fields) == 0{
				fld := pkg.Field{}
				fld.Field = "conf config.Config"
				fld.Name = "conf"
				ef.NewStructInfo.Fields = append(ef.NewStructInfo.Fields, fld)
			}

			var foundConfImport bool
			for _, oim := range ef.NewStructInfo.imports{
				if oim.Path == "github.com/jukylin/esim/config"{
					foundConfImport = true
				}
			}
			
			if foundConfImport == false {
				ef.appendNewImport("github.com/jukylin/esim/config")
			}
		}
	}

	return HasExtend
}

//if struct field had extend logger or conf
// so build a new struct
func (ef *esimFactory) buildNewStructFileContent() error {

	ef.NewStructInfo.importStr = ef.NewStructInfo.imports.String()

	if ef.oldStructInfo.importStr != "" {
		ef.NewStructInfo.structFileContent = strings.Replace(ef.oldStructInfo.structFileContent,
			ef.oldStructInfo.importStr, ef.NewStructInfo.importStr, -1)
	} else if ef.packStr != "" {
		//not find import
		ef.getFirstPart()
		ef.NewStructInfo.structFileContent = strings.Replace(ef.oldStructInfo.structFileContent,
			ef.packStr, ef.firstPart, -1)
	}else{
		ef.logger.Panicf("can't build the first part")
	}

	structInfo := templates.NewStructInfo()
	structInfo.StructName = ef.StructName
	structInfo.Fields = ef.NewStructInfo.Fields

	ef.NewStructInfo.structStr = structInfo.String()

	ef.NewStructInfo.structFileContent = strings.Replace(ef.NewStructInfo.structFileContent,
		ef.oldStructInfo.structStr, ef.NewStructInfo.structStr, -1)

	src, err := imports.Process("", []byte(ef.NewStructInfo.structFileContent), nil)
	if err != nil{
		ef.logger.Panicf("%s : %s", err.Error(), ef.NewStructInfo.structFileContent)
	}

	ef.NewStructInfo.structFileContent = string(src)

	return nil
}

//
// func NewStruct() {{.ReturnvarStr}} {
// }
//
func (ef *esimFactory) genReturnVarStr()  {
	if ef.withImpIface != ""{
		ef.NewStructInfo.ReturnVarStr = ef.withImpIface
	}else if ef.withPool == true || ef.withStar == true{
		ef.NewStructInfo.ReturnVarStr = "*" + ef.StructName
	}else{
		ef.NewStructInfo.ReturnVarStr = ef.StructName
	}
}

//type Option func(c *{{.OptionParam}})
func (ef *esimFactory) genOptionParam()  {
	if ef.withPool == true || ef.withStar == true{
		ef.OptionParam = "*" + ef.StructName
	}else{
		ef.OptionParam = ef.StructName
	}
}

// StructObj := Struct{} => {{.StructInitStr}}
func (ef *esimFactory) genStructInitStr() {
	var structInitStr string
	if ef.withStar == true{
		structInitStr = strings.ToLower(string(ef.StructName[0]))  +
			" := &" + ef.StructName + "{}"
	}else if ef.withPool == true{
		structInitStr = strings.ToLower(string(ef.StructName[0])) + ` := ` +
			templates.FirstToLower(snaker.SnakeToCamelLower(ef.StructName)) + `Pool.Get().(*` +
				ef.StructName + `)`
	}else{
		structInitStr = strings.ToLower(string(ef.StructName[0]))  +
			" := " + ef.StructName + "{}"
	}

	ef.NewStructInfo.StructInitStr = structInitStr
}


//return {{.ReturnStr}}
func (ef *esimFactory) genReturnStr() {
	ef.ReturnStr = strings.ToLower(string(ef.StructName[0]))
}


func (ef *esimFactory) genOptions()  {

	ef.Option1 = `type ` + ef.StructName + `Option func(` + ef.OptionParam + `)`

	ef.Option2 = `type ` + ef.StructName + `Options struct{}`

	ef.Option3 = `options ...` + ef.StructName +`Option`

	ef.Option4 = `
	for _, option := range options {
		option(` + strings.ToLower(string(ef.StructName[0])) + `)
	}`

	if ef.withGenConfOption == true{

		ef.Option5 += `
func (` + ef.StructName + `Options) WithConf(conf config.Config) ` + ef.StructName + `Option {
	return func(` + string(ef.StructName[0]) + ` `+ ef.NewStructInfo.ReturnVarStr +`) {
	` + string(ef.StructName[0]) + `.conf = conf
	}
}
`

	}

	if ef.withGenLoggerOption == true {
		ef.Option6 += `
func (` + ef.StructName + `Options) WithLogger(logger log.Logger) ` + ef.StructName + `Option {
	return func(` + string(ef.StructName[0]) + ` ` + ef.NewStructInfo.ReturnVarStr + `) {
		` + string(ef.StructName[0]) + `.logger = logger
	}
}
`
	}
}

func (ef *esimFactory) genPool() bool {

	ef.incrPoolVar(ef.StructName)

	ef.ReleaseStr = ef.genReleaseStructStr(ef.InitField.Fields)

	return true
}

func (ef *esimFactory) genPlural() bool {

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

	return true
}

func (ef *esimFactory) incrPoolVar(StructName string) bool {
	poolName := StructName + "Pool"
	if ef.varNameExists(ef.NewStructInfo.vars, poolName) == true {
		ef.logger.Debugf("var is exists : %s", poolName)
	} else {

		ef.NewStructInfo.vars = append(ef.NewStructInfo.vars,
			ef.appendPoolVar(poolName, StructName))
		ef.appendNewImport("sync")
	}

	return true
}

func (ef *esimFactory) genSpecFieldInitStr()  {
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
	return
}

func (ef *esimFactory) genReleaseStructStr(initFields []string) string {
	str := "func (" + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) + "  " + ef.NewStructInfo.ReturnVarStr + ") Release() {\n"

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

func (ef *esimFactory) appendNewImport(importName string) bool {
	var found bool
	for _, imp := range ef.NewStructInfo.imports {
		if imp.Path == importName {
			found = true
		}
	}

	if found == false {
		ef.NewStructInfo.imports = append(ef.NewStructInfo.imports, pkg.Import{Path: importName})
	}

	return true
}

func (ef *esimFactory) appendPoolVar(pollVarName, StructName string) pkg.Var {
	var poolVar pkg.Var

	pooltpl := NewPoolTpl()
	pooltpl.VarPoolName = pollVarName
	pooltpl.StructName = StructName

	poolVar.Body = pooltpl.String()

	poolVar.Name = append(poolVar.Name, pollVarName)
	return poolVar
}

//变量是否存在
func (ef *esimFactory) varNameExists(vars pkg.Vars, poolVarName string) bool {
	for _, varInfo := range vars {
		for _, varName := range varInfo.Name{
			if varName == poolVarName {
				return true
			}
		}
	}

	return false
}

//package + import
func (ef *esimFactory) getFirstPart() {

	if ef.oldStructInfo.importStr == "" {
		ef.firstPart += ef.packStr + "\n\n"
	}

	ef.firstPart += ef.NewStructInfo.importStr

}

//var
func (ef *esimFactory) getSecondPart() {
	ef.secondPart = ef.oldStructInfo.varStr
}

func (ef *esimFactory) Close() {
	ef.structFieldIface.Close()
}


