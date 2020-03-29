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
)


type SortReturn struct {
	Fields Fields `json:"fields"`
}


type InitFieldsReturn struct {
	Fields     []string `json:"fields"`
	SpecFields []Field  `json:"SpecFields"`
}


type Field struct {
	Name     string `json:"Name"`
	Size     int    `json:"Size"`
	Type     string `json:"Type"`
	TypeName string `json:"TypeName"`
}


type Fields []Field


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
	//false if not find
	found bool

	//Struct plural form
	StructNamePlural string

	withPlural bool

	withOption bool

	withGenLoggerOption bool

	withGenConfOption bool
	
	structFieldIface StructFieldIface

	newObjStr string

	ReleaseStr string

	firstPart string

	secondPart string

	thirdPart string

	InitField *InitFieldsReturn

	///模型的复数
	pluralName string

	NewPluralStr string

	ReleasePluralStr string
	///模型的复数

	///options start
	Option1 string

	Option2 string

	Option3 string

	Option4 string

	Option5 string

	Option6 string
	///options end

	NewStr string

	OptionParam string

	logger logger.Logger

	withSort bool

	withImpIface string

	withPool bool

	withStar bool

	writer file_dir.IfaceWrite

	SpecFieldInitStr string

	ReturnStr string
}

func NewEsimFactory() *esimFactory {
	factory := &esimFactory{}

	factory.oldStructInfo = &structInfo{}

	factory.NewStructInfo = &structInfo{}

	factory.logger = logger.NewLogger()

	factory.writer = file_dir.EsimWriter{}


	factory.structFieldIface = NewRpcPluginStructField(factory.writer)


	return factory
}

type structInfo struct{

	Fields []db2entity.Field

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

//获取单词的复数形式
//识别不了或单复同形，后面直接加s
func (this *esimFactory)  getPluralWord(word string) string {
	newWord := inflect.Pluralize(word)
	if newWord == word || newWord == "" {
		newWord = word + "s"
	}

	return newWord
}

func (this *esimFactory) Run(v *viper.Viper) error {

	err := this.inputBind(v)
	if err != nil {
		return err
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

			this.NewStructInfo.structStr = this.genNewStruct(this.StructName,
				sortedField.Fields, this.oldStructInfo.Fields)
		}

		this.InitField = this.structFieldIface.InitField(this.NewStructInfo.Fields)
	}

	this.genStr()

	err = file_dir.EsimBackUpFile(this.structDir +
		string(filepath.Separator) + this.structFileName)
	if err != nil{
		this.logger.Warnf("backup err %s:%s", this.structDir +
			string(filepath.Separator) + this.structFileName,
				err.Error())
	}

	tmpl, err := template.New("factory").Funcs(EsimFuncMap()).
		Parse(factoryTemplate)
	if err != nil{
		this.logger.Panicf(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, this)
	if err != nil{
		this.logger.Panicf(err.Error())
	}
	println(buf.String())
	//res, err := imports.Process("", []byte(src), nil)
	//if err != nil {
	//	return err
	//}

	//err = file_dir.EsimWrite(this.structDir +
	//	string(filepath.Separator) + this.oldStructInfo.structFileName,
	//		string(res))
	//if err != nil {
	//	return err
	//}

	return err
}


//copy oldStructInfo to NewStructInfo
func (this *esimFactory) copyOldStructInfo()  {
	copyStructInfo := *this.oldStructInfo
	copyStructInfo.structFileContent = ""
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
		this.StructNamePlural = this.getPluralWord(sname)
	}

	this.withOption = v.GetBool("option")

	this.withGenConfOption = v.GetBool("gen_logger_option")

	this.withGenLoggerOption = v.GetBool("gen_conf_option")

	this.withSort = v.GetBool("sort")

	this.withImpIface = v.GetString("imp_iface")

	this.withPool = v.GetBool("pool")

	this.withStar = v.GetBool("star")


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

	//info.newObjStr = replaceFrame(frame, info)

	//if v.GetBool("pool") == true {
	//	this.getHeader(info)
	//}
	//
	//getTwoPart(info)
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

	var found bool
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
									this.packStr = strSrc[f.Name.Pos()-1 : f.Name.End()]
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
					if GenDecl.Tok.String() == "var" && found == true {

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

					if GenDecl.Tok.String() == "import" && found == true {
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
				if strings.Contains(field.Filed, "log.Logger") == true && foundLogField == false{
					foundLogField = true
				}
			}

			if foundLogField == false || len(this.NewStructInfo.Fields) == 0{
				fld := db2entity.Field{}
				fld.Filed = "logger log.Logger"
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
				if strings.Contains(field.Filed, "config.Config") == 
					true && foundConfField == false{
					foundConfField = true
				}
			}

			if foundConfField == false || len(this.NewStructInfo.Fields) == 0{
				fld := db2entity.Field{}
				fld.Filed = "conf config.Config"
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

	if this.oldStructInfo.importStr != "" {
		this.NewStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
			this.oldStructInfo.importStr, this.genImport(this.NewStructInfo.imports), -1)
	} else if this.packStr != "" {
		//not find import
		this.getFirstPart()
		this.NewStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
			this.packStr, this.firstPart, -1)
	}else{
		return errors.New("can't build the first part")
	}

	this.NewStructInfo.importStr = this.genImport(this.NewStructInfo.imports)

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


func (this *esimFactory) genReturnStr() {
	this.ReturnStr = strings.ToLower(string(this.StructName[0]))
}


//
//func (this *esimFactory) NewFrame(v *viper.Viper, info *esimFactory) string {
//	var newFrame string
//	newFrame = `
//
//{{options1}}
//
//{{options2}}
//
//func New` + strings.ToUpper(string(info.StructName[0])) +
//	string(info.StructName[1:]) + `({{options3}}) ` + info.NewVarStr + ` {
//
//	`+ GetNewStr(v, info) +`
//
//	{{options4}}
//
//	` + getInitStr(info) + `
//
//` + GetReturnStr(info) + `
//}
//
//{{options5}}
//
//`
//
//	return newFrame
//}


//func replaceFrame(newFrame string, info *esimFactory) string {
//	newFrame = strings.Replace(newFrame, "{{options1}}", info.option1, -1)
//
//	newFrame = strings.Replace(newFrame, "{{options2}}", info.option2, -1)
//
//	newFrame = strings.Replace(newFrame, "{{options3}}", info.option3, -1)
//
//	newFrame = strings.Replace(newFrame, "{{options4}}", info.option4, -1)
//
//	newFrame = strings.Replace(newFrame, "{{options5}}", info.option5, -1)
//
//	return newFrame
//}


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


//
//func ReplaceContent(v *viper.Viper, info *esimFactory) (string, error) {
//
//	src, err := ioutil.ReadFile(info.structDir + "/" + info.structFileName)
//	if err != nil {
//		return "", err
//	}
//
//	strSrc := string(src)
//
//	if info.headerStr != "" {
//
//		for _, varBody := range info.oldVarBody {
//			strSrc = strings.Replace(strSrc, varBody, " ", -1)
//		}
//
//		if info.oldImportStr != "" {
//			strSrc = strings.Replace(strSrc, info.oldImportStr, info.headerStr, -1)
//		} else if info.packStr != "" {
//			strSrc = strings.Replace(strSrc, info.packStr, info.headerStr, -1)
//		}
//	}
//
//	strSrc = strings.Replace(strSrc, info.oldStruct, info.bodyStr, -1)
//
//	return strSrc, nil
//}


func (this *esimFactory) genNewStruct(StructName string, fields Fields,
	oldFields []db2entity.Field) string {
	var newStruct string

	newStruct = "type " + StructName + " struct{ \r\n"

	for _, field := range fields {
		for _, ofield := range oldFields {
			if field.Name == ofield.Filed {
				if len(ofield.Doc) > 0 {
					for _, doc := range ofield.Doc {
						newStruct += "	" + doc + "\r\n"
					}
				}
				newStruct += "	" + field.Name + " " + ofield.Tag + "\r\n"
				newStruct += "\r\n"
			}
		}
	}
	newStruct += "} \r\n"

	return newStruct
}


//
//func HandleNewStruct(info *esimFactory, newStrcut string) bool {
//	src, err := ioutil.ReadFile(info.structDir + "/" + info.structFileName)
//	if err != nil {
//		log.Errorf(err.Error())
//		return false
//	}
//
//	strSrc := string(src)
//
//	strSrc = strings.Replace(strSrc, info.oldStruct, newStrcut, -1)
//
//	dst, err := os.OpenFile(info.structDir+"/"+info.structFileName, os.O_WRONLY|os.O_CREATE, 0644)
//	if err != nil {
//		log.Errorf(err.Error())
//		return false
//	}
//	defer dst.Close()
//
//	dst.Write([]byte(strSrc))
//
//	return true
//}


//初始化变量，生成临时对象池
func (this *esimFactory) genPool() bool {

	this.incrPoolVar(this.StructName)

	this.ReleaseStr = this.genReleaseStructStr(this.InitField.Fields)

	return true
}


//复数池
func (this *esimFactory) genPlural() bool {

	this.incrPoolVar(this.pluralName)

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
	str := "func (" + strings.ToLower(this.StructName) +
		this.NewStructInfo.varStr + ") Release() {\n"

	for _, field := range this.InitField.Fields {
		if strings.Contains(field, "time.Time") {
			this.appendNewImport("time")
		}
		str += "		" + field + "\n"
	}

	str += "		" + strings.ToLower(this.StructName) +
		"Pool.Put(" + strings.ToLower(this.StructName) + ")\n"
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

func (this *esimFactory) genReleasePluralStr() string {
	str := "func (" + strings.ToLower(this.pluralName) + " *" +
		this.pluralName + ") Release() {\n"

	str += "*" + strings.ToLower(this.pluralName) + " = (*" +
		strings.ToLower(this.pluralName) + ")[:0]\n"
	str += "		" + strings.ToLower(this.pluralName) +
		"Pool.Put(" + strings.ToLower(this.pluralName) + ")\n"
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

	this.firstPart += this.genImport(this.NewStructInfo.imports)
	this.firstPart += "\n"
}

//var
func (this *esimFactory) getSecondPart() {
	this.secondPart += this.oldStructInfo.varStr
}

//struct body
//func getTwoPart(info *esimFactory) {
//	bodyStr := ""
//
//	if info.newStruct != "" {
//		bodyStr += info.newStruct
//	} else {
//		bodyStr += info.oldStruct
//	}
//
//	if info.pluralName != "" {
//		bodyStr += "\n"
//		bodyStr += "type " + info.pluralName + " []" + info.StructName
//	}
//
//	if info.newObjStr != "" {
//		bodyStr += "\n\n"
//		bodyStr += info.newObjStr
//	}
//
//	if info.releaseStr != "" {
//		bodyStr += "\n"
//		bodyStr += info.releaseStr
//	}
//
//	if info.newPluralNewBody != "" {
//		bodyStr += "\n\n"
//		bodyStr += info.newPluralNewBody
//	}
//
//	if info.newPluralReleaseBody != "" {
//		bodyStr += "\n"
//		bodyStr += info.newPluralReleaseBody
//	}
//
//	info.bodyStr = bodyStr
//}


