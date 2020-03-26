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
	"os"
	"path"
	"strings"
	"golang.org/x/tools/imports"
	"path/filepath"
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
	structName string

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
	newStructInfo *structInfo

	//true if find the structName
	//false if not find
	found bool

	//Struct plural form
	structNamePlural string

	withOption bool

	withGenLoggerOption bool

	withGenConfOption bool
	
	structFieldIface StructFieldIface

	newObjStr string

	releaseStr string

	firstPart string

	secondPart string

	thirdPart string

	InitField *InitFieldsReturn

	///模型的复数
	pluralName string

	newPluralNewBody string

	newPluralReleaseBody string
	///模型的复数

	///options start
	option1 string

	option2 string

	option3 string

	option4 string

	option5 string
	///options end

	NewStr string

	OptionParam string

	logger logger.Logger

	withSort bool

	withImpIface string

	withPool bool

	withStar bool

	writer file_dir.IfaceWrite
}

func NewEsimFactory() *esimFactory {
	factory := &esimFactory{}

	factory.oldStructInfo = &structInfo{}

	factory.newStructInfo = &structInfo{}

	factory.logger = logger.NewLogger()

	factory.structFieldIface = NewRpcPluginStructField()

	factory.writer = file_dir.EsimWriter{}

	return factory
}

type structInfo struct{

	fields []db2entity.Field

	structStr string

	//
	structFileContent string

	vars []Var

	varBody []string

	varStr string

	imports []string

	importStr string
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
		this.logger.Panicf("not found this struct %s", this.structName)
	}
	
	this.copyOldStructInfo()

	if this.ExtendField() {
		err = this.replaceStructContent()
		if err != nil {
			return err
		}
	}

	if len(this.oldStructInfo.fields) > 0 {
		if this.withSort == true{
			sortedField := this.structFieldIface.SortField()

			this.newStructInfo.structStr = this.genNewStruct(this.structName,
				sortedField.Fields, this.oldStructInfo.fields)
		}

		this.InitField = this.structFieldIface.InitField()
	}

	this.buildFrame()

	err = file_dir.EsimBackUpFile(this.structDir +
		string(filepath.Separator) + this.structFileName)
	if err != nil{
		this.logger.Warnf("backup err %s:%s", this.structDir +
			string(filepath.Separator) + this.structFileName,
				err.Error())
	}


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


//copy oldStructInfo to newStructInfo
func (this *esimFactory) copyOldStructInfo()  {
	copyStructInfo := *this.oldStructInfo
	copyStructInfo.structFileContent = ""
	this.newStructInfo = &copyStructInfo
}

func (this *esimFactory) inputBind(v *viper.Viper) error {
	sname := v.GetString("sname")
	if sname == "" {
		return errors.New("请输入结构体名称")
	}
	this.structName = sname

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	this.structDir = strings.TrimRight(wd, "/")

	plural := v.GetBool("plural")
	if plural == true {
		this.structNamePlural = this.getPluralWord(sname)
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

func (this *esimFactory) buildFrame()  {

	this.NewVarStr()

	if this.withOption == true{
		this.NewOptionParam()
		this.getOptions()
	}

	if this.withPool == true && len(this.InitField.Fields) > 0{

		this.genInitFieldsAndPool()
		this.HandelPlural()
		this.oldStructInfo.varStr = this.genVarStr(this.oldStructInfo.vars)
	}

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

								if typeSpec.Name.String() == this.structName {
									this.oldStructInfo.structFileContent = strSrc
									this.structFileName = fileInfo.Name()
									this.found = true
									this.packName = f.Name.String()
									this.packStr = strSrc[f.Name.Pos()-1 : f.Name.End()]
									this.oldStructInfo.fields = db2entity.GetOldFields(GenDecl, strSrc)
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
			for _, field := range this.newStructInfo.fields {
				if strings.Contains(field.Filed, "log.Logger") == true && foundLogField == false{
					foundLogField = true
				}
			}

			if foundLogField == false || len(this.newStructInfo.fields) == 0{
				fld := db2entity.Field{}
				fld.Filed = "logger log.Logger"
				this.newStructInfo.fields = append(this.newStructInfo.fields, fld)
			}

			var foundLogImport bool
			for _, oim := range this.newStructInfo.imports{
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
			for _, field := range this.newStructInfo.fields {
				if strings.Contains(field.Filed, "config.Config") == 
					true && foundConfField == false{
					foundConfField = true
				}
			}

			if foundConfField == false || len(this.newStructInfo.fields) == 0{
				fld := db2entity.Field{}
				fld.Filed = "conf config.Config"
				this.newStructInfo.fields = append(this.newStructInfo.fields, fld)
			}

			var foundConfImport bool
			for _, oim := range this.newStructInfo.imports{
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
// so build a new struct and replace it
func (this *esimFactory) replaceStructContent() error {

	if this.oldStructInfo.importStr != "" {
		this.newStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
			this.oldStructInfo.importStr, this.genImport(this.newStructInfo.imports), -1)
	} else if this.packStr != "" {
		//not find import
		this.getFirstPart()
		this.newStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
			this.packStr, this.firstPart, -1)
	}

	this.newStructInfo.importStr = this.genImport(this.newStructInfo.imports)

	this.newStructInfo.structStr = db2entity.GetNewStruct(this.structName, this.newStructInfo.fields)
	this.newStructInfo.structFileContent = strings.Replace(this.newStructInfo.structFileContent,
		this.oldStructInfo.structStr, this.newStructInfo.structStr, -1)

	src, err := imports.Process("", []byte(this.newStructInfo.structFileContent), nil)
	if err != nil{
		return err
	}

	this.newStructInfo.structFileContent = string(src)

	return this.writer.Write(this.structDir +
		string(filepath.Separator) + this.structFileName,
		this.newStructInfo.structFileContent)
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
// func NewStruct() {{.varStr}} {
// }
//
func (this *esimFactory) NewVarStr()  {
	if this.withImpIface != ""{
		this.newStructInfo.varStr = this.withImpIface
	}else if this.withPool == true || this.withStar == true{
		this.newStructInfo.varStr = "*" + this.structName
	}else{
		this.newStructInfo.varStr = this.structName
	}
}


//type Option func(c *{{.OptionParam}})
func (this *esimFactory) NewOptionParam()  {
	if this.withPool == true || this.withStar == true{
		this.OptionParam = "*" + this.structName
	}else{
		this.OptionParam = this.structName
	}
}


func GetNewStr(v *viper.Viper, info *esimFactory) string {
	if v.GetBool("star") == true{
		return strings.ToLower(string(info.structName[0]))  + " := &" + info.structName + "{}"
	}else if v.GetBool("pool") == true{
		return strings.ToLower(string(info.structName[0])) + ` := ` +
			strings.ToLower(info.structName) + `Pool.Get().(*` +
				info.structName + `)`
	}else{
		return strings.ToLower(string(info.structName[0]))  + " := " + info.structName + "{}"
	}

	return ""
}


func GetReturnStr(info *esimFactory) string {
	return "	return " + strings.ToLower(string(info.structName[0]))
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
//func New` + strings.ToUpper(string(info.structName[0])) +
//	string(info.structName[1:]) + `({{options3}}) ` + info.NewVarStr + ` {
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

func replaceFrame(newFrame string, info *esimFactory) string {
	newFrame = strings.Replace(newFrame, "{{options1}}", info.option1, -1)

	newFrame = strings.Replace(newFrame, "{{options2}}", info.option2, -1)

	newFrame = strings.Replace(newFrame, "{{options3}}", info.option3, -1)

	newFrame = strings.Replace(newFrame, "{{options4}}", info.option4, -1)

	newFrame = strings.Replace(newFrame, "{{options5}}", info.option5, -1)

	return newFrame
}


func (this *esimFactory) getOptions()  {

	this.option1 = `type ` + this.structName + `Option func(` + this.OptionParam + `)`

	this.option2 = `type ` + this.structName + `Options struct{}`

	this.option3 = `options ...` + this.structName +`Option`

	this.option4 = `
	for _, option := range options {
		option(` + strings.ToLower(string(this.structName[0])) + `)
	}`

	if this.withGenConfOption == true{

		this.option5 += `
func (` + this.structName + `Options) WithConf(conf config.Config) ` + this.structName + `Option {
	return func(` + string(this.structName[0]) + ` `+ this.newStructInfo.varStr +`) {
	` + string(this.structName[0]) + `.conf = conf
	}
}
`

	}

	if this.withGenLoggerOption == true {
		this.option5 += `
func (` + this.structName + `Options) WithLogger(logger log.Logger) ` + this.structName + `Option {
	return func(` + string(this.structName[0]) + ` ` + this.newStructInfo.varStr + `) {
		` + string(this.structName[0]) + `.logger = logger
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



func (this *esimFactory) genNewStruct(structName string, fields Fields,
	oldFields []db2entity.Field) string {
	var newStruct string

	newStruct = "type " + structName + " struct{ \r\n"

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
func (this *esimFactory) genInitFieldsAndPool() bool {

	this.incrPoolVar(this.structName)

	this.genPool()

	return true
}


//处理复数
func (this *esimFactory) HandelPlural() bool {

	if this.pluralName != "" {

		this.incrPoolVar(this.pluralName)

		this.genPluralPool()
	}

	return true
}


func (this *esimFactory) incrPoolVar(structName string) bool {
	poolName := strings.ToLower(structName) + "Pool"
	if varNameExists(this.newStructInfo.vars, poolName) == true {
		this.logger.Debugf("变量已存在 : %s", poolName)
	} else {

		this.newStructInfo.vars = append(this.newStructInfo.vars,
			this.genPoolVar(poolName, structName))
		this.appendNewImport("sync")
	}

	return true
}


//单数池
func (this *esimFactory) genPool() {
	//info.newObjStr = getNewObjStr(info)

	this.releaseStr = this.genReleaseStructStr(this.InitField.Fields)
}


//复数池
func (this *esimFactory) genPluralPool() {
	this.newPluralNewBody = this.genNewPluralObjStr()

	this.newPluralReleaseBody = this.genReleasePluralObjStr()
}


func (this *esimFactory) genInitStr(info *esimFactory) string {
	var str string

	for _, f := range info.InitField.SpecFields {
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


	return str
}


func (this *esimFactory) genReleaseStructStr(initFields []string) string {
	str := "func (" + strings.ToLower(this.structName) +
		this.newStructInfo.varStr + ") Release() {\n"

	for _, field := range this.InitField.Fields {
		if strings.Contains(field, "time.Time") {
			this.appendNewImport("time")
		}
		str += "		" + field + "\n"
	}

	str += "		" + strings.ToLower(this.structName) +
		"Pool.Put(" + strings.ToLower(this.structName) + ")\n"
	str += "}"

	return str
}

func (this *esimFactory) genNewPluralObjStr() string {
	str := `func New` + this.pluralName + `() *` + this.pluralName + ` {
	` + strings.ToLower(this.pluralName) + ` := ` +
		strings.ToLower(this.pluralName) + `Pool.Get().(*` + this.pluralName + `)
`

	str += `return ` + strings.ToLower(this.pluralName) + `
}
`

	return str
}

func (this *esimFactory) genReleasePluralObjStr() string {
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
	for _, importStr := range this.newStructInfo.imports {
		if importStr == importName {
			found = true
		}
	}

	if found == false {
		this.newStructInfo.imports = append(this.newStructInfo.imports, importName)
	}

	return true
}

func (this *esimFactory) genVarStr(vars []Var) string {
	varStr := "var ( \n"
	for _, varInfo := range vars {
		for _, doc := range varInfo.doc {
			varStr += "	" + doc + "\n"
		}
		varStr += "	" + varInfo.val + "\n"
	}
	varStr += ") \n"

	return varStr
}


func (this *esimFactory) genPoolVar(pollVarName, structName string) Var {
	var poolVar Var
	poolVar.val = pollVarName + ` = sync.Pool{
        New: func() interface{} {
                return &` + structName + `{}
        },
	}
`
	poolVar.name = pollVarName
	return poolVar
}

//变量是否存在
func varNameExists(vars []Var, poolVarName string) bool {
	for _, varInfo := range vars {
		if varInfo.name == poolVarName {
			return true
		}
	}

	return false
}

//import + var
func (this *esimFactory) getFirstPart() {

	if this.newStructInfo.importStr == "" {
		this.firstPart += this.packStr + "\n"
	}

	this.firstPart += this.genImport(this.newStructInfo.imports)
	this.firstPart += "\n"
	this.firstPart += this.oldStructInfo.varStr
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
//		bodyStr += "type " + info.pluralName + " []" + info.structName
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


