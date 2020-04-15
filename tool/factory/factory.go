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
func (this *esimFactory)  getPluralForm(word string) string {
	newWord := inflect.Pluralize(word)
	if newWord == word || newWord == "" {
		newWord = word + "s"
	}

	return newWord
}

func (this *esimFactory) Run(v *viper.Viper) error {

	err := this.bindInput(v)
	if err != nil {
		this.logger.Panicf(err.Error())
	}

	if this.FindStruct() == false{
		this.logger.Panicf("not found this struct %s", this.StructName)
	}
	this.copyOldStructInfo()

	if this.ExtendField() {
		err = this.buildNewStructFileContent()
		if err != nil {
			this.logger.Panicf(err.Error())
		}
	}

	this.logger.Debugf("fields len %d", this.oldStructInfo.Fields.Len())

	if this.oldStructInfo.Fields.Len() > 0 {
		this.structFieldIface.SetStructInfo(this.NewStructInfo)
		this.structFieldIface.SetStructName(this.StructName)
		this.structFieldIface.SetStructDir(this.structDir)
		this.structFieldIface.SetStructFileName(this.structFileName)
		this.structFieldIface.SetFilesName(this.filesName)
		this.structFieldIface.SetPackName(this.packName)

		if this.withSort == true{
			sortedField := this.structFieldIface.SortField(this.NewStructInfo.Fields)
			this.logger.Debugf("sorted fields %+v", sortedField.Fields)
			this.NewStructInfo.Fields = sortedField.Fields
		}

		this.InitField = this.structFieldIface.InitField(this.NewStructInfo.Fields)
	}

	this.genStr()

	this.assignStructTpl()

	this.executeNewTmpl()

	this.organizePart()

	if this.withPrint {
		this.printResult()
	}else{
		err = file_dir.EsimBackUpFile(this.structDir +
			string(filepath.Separator) + this.structFileName)
		if err != nil{
			this.logger.Warnf("backup err %s:%s", this.structDir +
				string(filepath.Separator) + this.structFileName,
				err.Error())
		}

		originContent := this.replaceOriginContent()

		res, err := imports.Process("", []byte(originContent), nil)
		if err != nil {
			this.logger.Panicf("%s:%s", err.Error(), originContent)
		}

		err = file_dir.EsimWrite(this.structDir +
			string(filepath.Separator) + this.structFileName,
				string(res))
		if err != nil {
			this.logger.Panicf(err.Error())
		}
	}

	return nil
}

func (this *esimFactory) assignStructTpl()  {
	this.StructTpl.StructName = this.StructName
	this.StructTpl.Fields = this.NewStructInfo.Fields
}

func (this *esimFactory) executeNewTmpl() {
	tmpl, err := template.New("factory").Funcs(templates.EsimFuncMap()).
		Parse(newTemplate)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, this)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	this.NewStructInfo.structStr = buf.String()
}

//replaceOriginContent gen a new struct file content
func (this *esimFactory) replaceOriginContent() string {
	var newContent string
	originContent := this.oldStructInfo.structFileContent
	newContent = originContent
	if this.oldStructInfo.importStr != ""{
		newContent = strings.Replace(newContent, this.oldStructInfo.importStr, "", 1)
	}

	newContent = strings.Replace(newContent, this.packStr, this.firstPart, 1)

	if this.secondPart != ""{
		newContent = strings.Replace(newContent, this.oldStructInfo.varStr, this.secondPart, 1)
	}

	newContent = strings.Replace(newContent, this.oldStructInfo.structStr, this.thirdPart, 1)
	this.NewStructInfo.structFileContent = newContent

	return newContent
}

//printResult println file content to terminal
func (this *esimFactory) printResult()  {
	src := this.firstPart + "\n"
	src += this.secondPart + "\n"
	src += this.thirdPart + "\n"

	res, err := imports.Process("", []byte(src), nil)
	if err != nil{
		this.logger.Panicf(err.Error())
	}else{
		println(string(res))
	}
}

//organizePart  organize pack, import, var, struct
func (this *esimFactory) organizePart()  {
	this.firstPart = this.packStr + "\n"
	this.firstPart += this.NewStructInfo.importStr + "\n"

	if this.oldStructInfo.varStr != ""{
		this.secondPart = this.NewStructInfo.varStr
	}else{
		//merge firstPart and secondPart
		this.firstPart += this.NewStructInfo.varStr
	}

	this.thirdPart = this.NewStructInfo.structStr
}

//copy oldStructInfo to NewStructInfo
func (this *esimFactory) copyOldStructInfo()  {
	copyStructInfo := *this.oldStructInfo
	this.NewStructInfo = &copyStructInfo
}

func (this *esimFactory) bindInput(v *viper.Viper) error {
	sname := v.GetString("sname")
	if sname == "" {
		return errors.New("请输入结构体名称")
	}
	this.StructName = sname

	sdir := v.GetString("sdir")
	if sdir == ""{
		sdir = "."
	}

	dir, err := filepath.Abs(sdir)
	if err != nil{
		return err
	}
	this.structDir = strings.TrimRight(dir, "/")

	plural := v.GetBool("plural")
	if plural == true {
		this.withPlural = true
		this.pluralName = this.getPluralForm(sname)
	}

	this.withOption = v.GetBool("option")

	this.withGenConfOption = v.GetBool("gen_conf_option")

	this.withGenLoggerOption = v.GetBool("gen_logger_option")

	this.withSort = v.GetBool("sort")

	this.withImpIface = v.GetString("imp_iface")

	this.withPool = v.GetBool("pool")

	this.withStar = v.GetBool("star")

	this.withPrint = v.GetBool("print")

	this.WithNew = v.GetBool("new")

	return nil
}


func (this *esimFactory) genStr()  {

	this.genReturnVarStr()
	this.genStructInitStr()
	this.genSpecFieldInitStr()
	this.genReturnStr()

	if this.withOption == true{
		this.genOptionParam()
		this.genOptions()
	}

	if this.withPool == true && len(this.InitField.Fields) > 0{
		this.genPool()
	}

	if this.withPlural {
		this.genPlural()
	}

	this.NewStructInfo.varStr = this.NewStructInfo.vars.String()
}


//@ find struct
//@ parse struct
func (this *esimFactory) FindStruct() bool {

	exists, err := file_dir.IsExistsDir(this.structDir)
	if err != nil {
		this.logger.Panicf(err.Error())
	}

	if exists == false {
		this.logger.Panicf("%s dir not exists", this.structDir)
	}

	files, err := ioutil.ReadDir(this.structDir)
	if err != nil {
		this.logger.Panicf(err.Error())
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

		this.filesName = append(this.filesName, fileInfo.Name())

		if !fileInfo.IsDir() {
			src, err := ioutil.ReadFile(this.structDir + "/" + fileInfo.Name())
			if err != nil {
				this.logger.Panicf(err.Error())
			}

			strSrc := string(src)
			fset := token.NewFileSet() // positions are relative to fset
			f, err := parser.ParseFile(fset, "", strSrc, parser.ParseComments)
			if err != nil {
				this.logger.Panicf(err.Error())
			}

			for _, decl := range f.Decls {
				if GenDecl, ok := decl.(*ast.GenDecl); ok {
					if GenDecl.Tok.String() == "type" {
						for _, specs := range GenDecl.Specs {
							if typeSpec, ok := specs.(*ast.TypeSpec); ok {

								if typeSpec.Name.String() == this.StructName {
									this.oldStructInfo.structFileContent = strSrc
									this.structFileName = fileInfo.Name()
									this.found = true
									this.packName = f.Name.String()
									this.packStr = "package " + strSrc[f.Name.Pos()-1 : f.Name.End()]
									fields := pkg.Fields{}
									fields.ParseFromAst(GenDecl, strSrc)
									this.oldStructInfo.Fields = fields
									this.oldStructInfo.structStr = string(src[GenDecl.TokPos-1 : typeSpec.Type.(*ast.StructType).Fields.Closing])
								}
							}
						}
					}
				}
			}

			for _, decl := range f.Decls {
				if GenDecl, ok := decl.(*ast.GenDecl); ok {
					if GenDecl.Tok.String() == "var" && this.found == true {
						this.oldStructInfo.vars.ParseFromAst(GenDecl, strSrc)
					}

					if GenDecl.Tok.String() == "import" && this.found == true {
						imps := pkg.Imports{}
						imps.ParseFromAst(GenDecl)
						this.oldStructInfo.imports = imps
						this.oldStructInfo.importStr = strSrc[GenDecl.Pos()-1 : GenDecl.End()]
					}
				}
			}
		}
	}

	return this.found
}

//extend logger and conf for new struct field
func (this *esimFactory) ExtendField() bool {

	var HasExtend bool
	if this.withOption == true {
		if this.withGenLoggerOption == true{
			HasExtend = true
			var foundLogField bool
			for _, field := range this.NewStructInfo.Fields {
				if strings.Contains(field.Field, "log.Logger") == true && foundLogField == false{
					foundLogField = true
				}
			}

			if foundLogField == false || len(this.NewStructInfo.Fields) == 0{
				fld := pkg.Field{}
				fld.Field = "logger log.Logger"
				fld.Name = "logger"
				this.NewStructInfo.Fields = append(this.NewStructInfo.Fields, fld)
			}

			var foundLogImport bool
			for _, oim := range this.NewStructInfo.imports{
				if oim.Path == "github.com/jukylin/esim/log"{
					foundLogImport = true
				}
			}

			if foundLogImport == false {
				this.appendNewImport("github.com/jukylin/esim/log")
			}
		}

		if this.withGenConfOption == true{
			HasExtend = true

			var foundConfField bool
			for _, field := range this.NewStructInfo.Fields {
				if strings.Contains(field.Field, "config.Config") ==
					true && foundConfField == false{
					foundConfField = true
				}
			}

			if foundConfField == false || len(this.NewStructInfo.Fields) == 0{
				fld := pkg.Field{}
				fld.Field = "conf config.Config"
				fld.Name = "conf"
				this.NewStructInfo.Fields = append(this.NewStructInfo.Fields, fld)
			}

			var foundConfImport bool
			for _, oim := range this.NewStructInfo.imports{
				if oim.Path == "github.com/jukylin/esim/config"{
					foundConfImport = true
				}
			}
			
			if foundConfImport == false {
				this.appendNewImport("github.com/jukylin/esim/config")
			}
		}
	}

	return HasExtend
}

//if struct field had extend logger or conf
// so build a new struct
func (this *esimFactory) buildNewStructFileContent() error {

	this.NewStructInfo.importStr = this.NewStructInfo.imports.String()

	if this.oldStructInfo.importStr != "" {
		this.NewStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
			this.oldStructInfo.importStr, this.NewStructInfo.importStr, -1)
	} else if this.packStr != "" {
		//not find import
		this.getFirstPart()
		this.NewStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
			this.packStr, this.firstPart, -1)
	}else{
		this.logger.Panicf("can't build the first part")
	}

	structInfo := templates.NewStructInfo()
	structInfo.StructName = this.StructName
	structInfo.Fields = this.NewStructInfo.Fields
	this.NewStructInfo.structStr = structInfo.String()

	this.NewStructInfo.structFileContent = strings.Replace(this.NewStructInfo.structFileContent,
		this.oldStructInfo.structStr, this.NewStructInfo.structStr, -1)

	src, err := imports.Process("", []byte(this.NewStructInfo.structFileContent), nil)
	if err != nil{
		this.logger.Panicf("%s : %s", err.Error(), this.NewStructInfo.structFileContent)
	}

	this.NewStructInfo.structFileContent = string(src)

	return nil
}

//
// func NewStruct() {{.ReturnvarStr}} {
// }
//
func (this *esimFactory) genReturnVarStr()  {
	if this.withImpIface != ""{
		this.NewStructInfo.ReturnVarStr = this.withImpIface
	}else if this.withPool == true || this.withStar == true{
		this.NewStructInfo.ReturnVarStr = "*" + this.StructName
	}else{
		this.NewStructInfo.ReturnVarStr = this.StructName
	}
}

//type Option func(c *{{.OptionParam}})
func (this *esimFactory) genOptionParam()  {
	if this.withPool == true || this.withStar == true{
		this.OptionParam = "*" + this.StructName
	}else{
		this.OptionParam = this.StructName
	}
}

// StructObj := Struct{} => {{.StructInitStr}}
func (this *esimFactory) genStructInitStr() {
	var structInitStr string
	if this.withStar == true{
		structInitStr = strings.ToLower(string(this.StructName[0]))  +
			" := &" + this.StructName + "{}"
	}else if this.withPool == true{
		structInitStr = strings.ToLower(string(this.StructName[0])) + ` := ` +
			strings.ToLower(this.StructName) + `Pool.Get().(*` +
				this.StructName + `)`
	}else{
		structInitStr = strings.ToLower(string(this.StructName[0]))  +
			" := " + this.StructName + "{}"
	}

	this.NewStructInfo.StructInitStr = structInitStr
}


//return {{.ReturnStr}}
func (this *esimFactory) genReturnStr() {
	this.ReturnStr = strings.ToLower(string(this.StructName[0]))
}


func (this *esimFactory) genOptions()  {

	this.Option1 = `type ` + this.StructName + `Option func(` + this.OptionParam + `)`

	this.Option2 = `type ` + this.StructName + `Options struct{}`

	this.Option3 = `options ...` + this.StructName +`Option`

	this.Option4 = `
	for _, option := range options {
		option(` + strings.ToLower(string(this.StructName[0])) + `)
	}`

	if this.withGenConfOption == true{

		this.Option5 += `
func (` + this.StructName + `Options) WithConf(conf config.Config) ` + this.StructName + `Option {
	return func(` + string(this.StructName[0]) + ` `+ this.NewStructInfo.ReturnVarStr +`) {
	` + string(this.StructName[0]) + `.conf = conf
	}
}
`

	}

	if this.withGenLoggerOption == true {
		this.Option6 += `
func (` + this.StructName + `Options) WithLogger(logger log.Logger) ` + this.StructName + `Option {
	return func(` + string(this.StructName[0]) + ` ` + this.NewStructInfo.ReturnVarStr + `) {
		` + string(this.StructName[0]) + `.logger = logger
	}
}
`
	}
}


func (this *esimFactory) genPool() bool {

	this.incrPoolVar(this.StructName)

	this.ReleaseStr = this.genReleaseStructStr(this.InitField.Fields)

	return true
}


func (this *esimFactory) genPlural() bool {

	this.incrPoolVar(this.pluralName)

	plural := NewPlural()
	plural.StructName = this.StructName
	plural.PluralName = this.pluralName
	if this.withStar {
		plural.Star = "*"
	}

	this.TypePluralStr = plural.TypeString()

	this.NewPluralStr = plural.NewString()

	this.ReleasePluralStr = plural.ReleaseString()

	return true
}


func (this *esimFactory) incrPoolVar(StructName string) bool {
	poolName := strings.ToLower(StructName) + "Pool"
	if this.varNameExists(this.NewStructInfo.vars, poolName) == true {
		this.logger.Debugf("var is exists : %s", poolName)
	} else {

		this.NewStructInfo.vars = append(this.NewStructInfo.vars,
			this.appendPoolVar(poolName, StructName))
		this.appendNewImport("sync")
	}

	return true
}


func (this *esimFactory) genSpecFieldInitStr()  {
	var str string

	for _, f := range this.InitField.SpecFields {
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

	this.SpecFieldInitStr = str
	return
}


func (this *esimFactory) genReleaseStructStr(initFields []string) string {
	str := "func (this " + this.NewStructInfo.ReturnVarStr + ") Release() {\n"

	for _, field := range this.InitField.Fields {
		if strings.Contains(field, "time.Time") {
			this.appendNewImport("time")
		}
		str += "		" + field + "\n"
	}

	str += "		" + strings.ToLower(this.StructName) +
		"Pool.Put(this)\n"
	str += "}"

	return str
}


func (this *esimFactory) genNewPluralStr() string {
	str := `func New` + this.pluralName + `() *` + this.pluralName + ` {
	` + strings.ToLower(this.pluralName) + ` := ` +
		strings.ToLower(this.pluralName) + `Pool.Get().(*` + this.pluralName + `)
`

	str += `return ` + strings.ToLower(this.pluralName) + `
}
`

	return str
}

func (this *esimFactory) genTypePluralStr() string {
	return "type " + this.pluralName + " []" + this.StructName
}


func (this *esimFactory) genReleasePluralStr() string {
	str := "func (this *" +
		this.pluralName + ") Release() {\n"

	str += "*this = (*this)[:0]\n"
	str += "		" + strings.ToLower(this.pluralName) +
		"Pool.Put(this)\n"
	str += "}"

	return str
}


func (this *esimFactory) appendNewImport(importName string) bool {
	var found bool
	for _, imp := range this.NewStructInfo.imports {
		if imp.Path == importName {
			found = true
		}
	}

	if found == false {
		this.NewStructInfo.imports = append(this.NewStructInfo.imports, pkg.Import{Path: importName})
	}

	return true
}


func (this *esimFactory) appendPoolVar(pollVarName, StructName string) pkg.Var {
	var poolVar pkg.Var

	pooltpl := NewPoolTpl()
	pooltpl.VarPoolName = pollVarName
	pooltpl.StructName = StructName

	poolVar.Body = pooltpl.String()

	poolVar.Name = append(poolVar.Name, pollVarName)
	return poolVar
}


//变量是否存在
func (this *esimFactory) varNameExists(vars pkg.Vars, poolVarName string) bool {
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
func (this *esimFactory) getFirstPart() {

	if this.oldStructInfo.importStr == "" {
		this.firstPart += this.packStr + "\n\n"
	}

	this.firstPart += this.NewStructInfo.importStr

}


//var
func (this *esimFactory) getSecondPart() {
	this.secondPart = this.oldStructInfo.varStr
}


func (this *esimFactory) Close() {
	this.structFieldIface.Close()
}


