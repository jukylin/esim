package factory

import (
	"errors"
	"github.com/martinusso/inflect"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/tool/db2entity"
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

	writer file_dir.IfaceWrite

	SpecFieldInitStr string

	ReturnStr string
}

func NewEsimFactory(logger logger.Logger) *esimFactory {
	factory := &esimFactory{}

	factory.oldStructInfo = &structInfo{}

	factory.NewStructInfo = &structInfo{}

	factory.logger = logger

	factory.writer = file_dir.EsimWriter{}


	factory.structFieldIface = NewRpcPluginStructField(factory.writer)


	return factory
}

type structInfo struct{

	Fields pkg.Fields

	structStr string

	//
	structFileContent string

	vars []Var

	varBody []string

	varStr string

	imports []string

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

	err := this.inputBind(v)
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
			return err
		}
	}

	if len(this.oldStructInfo.Fields) > 0 {

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


func (this *esimFactory) executeNewTmpl() {
	tmpl, err := template.New("factory").Funcs(pkg.EsimFuncMap()).
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
	originContent := this.oldStructInfo.structFileContent
	if this.oldStructInfo.importStr != ""{
		originContent = strings.Replace(originContent, this.oldStructInfo.importStr, "", 1)
	}

	originContent = strings.Replace(originContent, this.packStr, this.firstPart, 1)

	if this.secondPart != ""{
		originContent = strings.Replace(originContent, this.oldStructInfo.varStr, this.secondPart, 1)
	}

	originContent = strings.Replace(originContent, this.oldStructInfo.structStr, this.thirdPart, 1)

	return originContent
}


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


func (this *esimFactory) inputBind(v *viper.Viper) error {
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

	this.withGenConfOption = v.GetBool("gen_logger_option")

	this.withGenLoggerOption = v.GetBool("gen_conf_option")

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

	this.genVarStr(this.NewStructInfo.vars)
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
									this.oldStructInfo.Fields = db2entity.GetOldFields(GenDecl, strSrc)
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

						for _, specs := range GenDecl.Specs {

							var fileVar Var
							if typeSpec, ok := specs.(*ast.ValueSpec); ok {

								//区别有括号和无括号
								if GenDecl.Rparen != 0 {
									this.oldStructInfo.varBody = append(this.oldStructInfo.varBody,
										strSrc[GenDecl.TokPos-1:GenDecl.Rparen])
								} else {
									this.oldStructInfo.varBody = append(this.oldStructInfo.varBody,
										strSrc[GenDecl.TokPos-1:typeSpec.End()])
								}

								fileVar.val = strSrc[typeSpec.Pos()-1 : typeSpec.End()]
								if typeSpec.Doc != nil {
									for _, doc := range typeSpec.Doc.List {
										fileVar.doc = append(fileVar.doc, doc.Text)
									}
								}
								fileVar.name = typeSpec.Names[0].Name
								this.oldStructInfo.vars = append(this.oldStructInfo.vars, fileVar)
							}
						}
					}

					if GenDecl.Tok.String() == "import" && this.found == true {
						for _, specs := range GenDecl.Specs {
							if typeSpec, ok := specs.(*ast.ImportSpec); ok {
								if typeSpec.Name.String() != "<nil>" {
									this.oldStructInfo.imports = append(this.oldStructInfo.imports,
										typeSpec.Name.String()+" "+typeSpec.Path.Value)
								} else {
									this.oldStructInfo.imports = append(this.oldStructInfo.imports, typeSpec.Path.Value)
								}
							}
						}
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
				if oim == "github.com/jukylin/esim/log"{
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
				if oim == "github.com/jukylin/esim/config"{
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

	this.NewStructInfo.importStr = this.genImport(this.NewStructInfo.imports)

	if this.oldStructInfo.importStr != "" {
		this.NewStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
			this.oldStructInfo.importStr, this.NewStructInfo.importStr, -1)
	} else if this.packStr != "" {
		//not find import
		this.getFirstPart()
		this.NewStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
			this.packStr, this.firstPart, -1)
	}else{
		return errors.New("can't build the first part")
	}

	this.NewStructInfo.structStr = db2entity.GetNewStruct(this.StructName, this.NewStructInfo.Fields)
	this.NewStructInfo.structFileContent = strings.Replace(this.NewStructInfo.structFileContent,
		this.oldStructInfo.structStr, this.NewStructInfo.structStr, -1)

	src, err := imports.Process("", []byte(this.NewStructInfo.structFileContent), nil)
	if err != nil{
		return err
	}

	this.NewStructInfo.structFileContent = string(src)

	return nil
}


func (this *esimFactory) genImport(imports []string) string {

	if len(imports) <= 0{
		return ""
	}

	var newImport string
	newImport +=
`import (
`
	for _, imp := range imports {
		newImport += `	` +imp+ `
`
	}

	newImport += `)
`

	return newImport
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

	this.TypePluralStr = this.genTypePluralStr()

	this.NewPluralStr = this.genNewPluralStr()

	this.ReleasePluralStr = this.genReleasePluralStr()

	return true
}


func (this *esimFactory) incrPoolVar(StructName string) bool {
	poolName := strings.ToLower(StructName) + "Pool"
	if this.varNameExists(this.NewStructInfo.vars, poolName) == true {
		this.logger.Debugf("变量已存在 : %s", poolName)
	} else {

		this.NewStructInfo.vars = append(this.NewStructInfo.vars,
			this.genPoolVar(poolName, StructName))
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
	importName = "\"" + importName + "\""
	for _, importStr := range this.NewStructInfo.imports {
		if importStr == importName {
			found = true
		}
	}

	if found == false {
		this.NewStructInfo.imports = append(this.NewStructInfo.imports, importName)
	}

	return true
}


func (this *esimFactory) genVarStr(vars []Var) {

	if len(vars) <= 0{
		return
	}

	varStr := "var ( \n"
	for _, varInfo := range vars {
		for _, doc := range varInfo.doc {
			varStr += "	" + doc + "\n"
		}
		varStr += "	" + varInfo.val + "\n"
	}
	varStr += ") \n"

	this.NewStructInfo.varStr = varStr
	return
}


func (this *esimFactory) genPoolVar(pollVarName, StructName string) Var {
	var poolVar Var
	poolVar.val = pollVarName + ` = sync.Pool{
        New: func() interface{} {
                return &` + StructName + `{}
        },
	}
`
	poolVar.name = pollVarName
	return poolVar
}


//变量是否存在
func (this *esimFactory) varNameExists(vars []Var, poolVarName string) bool {
	for _, varInfo := range vars {
		if varInfo.name == poolVarName {
			return true
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


