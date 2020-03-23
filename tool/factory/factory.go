package factory

import (
	"encoding/json"
	"errors"
	"fmt"
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
	"os/exec"
	"path"
	"strconv"
	"strings"
	log2 "log"
	go_plugin "github.com/hashicorp/go-plugin"
	"golang.org/x/tools/imports"
	"github.com/hashicorp/go-hclog"
	"path/filepath"
	"github.com/hashicorp/consul/command/info"
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



type esimFactory struct {

	oldStructInfo *structInfo

	newStructInfo *structInfo

	found bool

	withOption bool

	withGenLoggerOption bool

	withGenConfOption bool

	willAppendImport []string

	structFieldIface StructFieldIface

	newObjStr string

	releaseStr string

	headerStr string

	bodyStr string

	InitField InitFieldsReturn

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

	NewVarStr string

	OptionParam string

	logger logger.Logger
}

func NewEsimFactory() *esimFactory {
	factory := &esimFactory{}

	factory.oldStructInfo = &structInfo{}

	factory.newStructInfo = &structInfo{}

	factory.logger = logger.NewLogger()

	factory.structFieldIface = NewRpcPluginStructField()

	return factory
}

type structInfo struct{
	//struct Absolute path
	structDir string

	filesName []string

	//struct name which be search
	structName string

	structNamePlural string

	//结构体文件名
	structFileName string

	packName string

	packStr string

	fields []db2entity.Field

	structStr string

	//模型对应文件内容
	structFileContent string

	vars []Var

	varBody []string

	varStr string

	imports []string

	importStr string
}

//获取单词的复数形式
//识别不了或单复同形，后面直接加s
func getPluralWord(word string) string {
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
		this.logger.Panicf("%s not found this struct", this.oldStructInfo.structName)
	}

	if this.ExtendField() {
		err = this.ReWriteStructContent()
		if err != nil {
			return err
		}
	}

	if len(this.oldStructInfo.fields) > 0 {
		err = BuildPluginEnv(info, nil)
		if err != nil {
			return err
		}

		defer func() {
			if err := recover(); err != nil {
				this.logger.Errorf("%v", err)
			}
		}()

		err = ExecPlugin(v, info)
		if err != nil {
			return err
		}
	}

	BuildFrame(v, info)

	err = file_dir.EsimBackUpFile(info.oldStructInfo.structDir + string(filepath.Separator) + info.oldStructInfo.structFileName)
	if err != nil{
		this.logger.Warnf("backup err %s:%s", info.oldStructInfo.structDir + string(filepath.Separator) + info.oldStructInfo.structFileName, err.Error())
	}

	src, err := ReplaceContent(v, info)
	if err != nil {
		return err
	}

	res, err := imports.Process("", []byte(src), nil)
	if err != nil {
		return err
	}

	err = file_dir.EsimWrite(info.oldStructInfo.structDir + string(filepath.Separator) + info.oldStructInfo.structFileName, string(res))
	if err != nil {
		return err
	}

	return err
}


func (this *esimFactory) inputBind(v *viper.Viper) error {
	sname := v.GetString("sname")
	if sname == "" {
		return errors.New("请输入结构体名称")
	}
	this.oldStructInfo.structName = sname

	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	this.oldStructInfo.structDir = strings.TrimRight(wd, "/")

	plural := v.GetBool("plural")
	if plural == true {
		this.oldStructInfo.structNamePlural = getPluralWord(sname)
	}

	this.withOption = v.GetBool("option")

	this.withGenConfOption = v.GetBool("gen_logger_option")

	this.withGenLoggerOption = v.GetBool("gen_conf_option")

	return nil
}

func BuildFrame(v *viper.Viper, info *esimFactory)  {
	var frame string

	if v.GetBool("new") == true {
		NewVarStr(v, info)
		frame = NewFrame(v, info)
	}

	if v.GetBool("option") == true {
		NewOptionParam(v, info)
		getOptions(v, info)
	}

	if v.GetBool("pool") == true && len(info.InitField.Fields) > 0{

		HandleInitFieldsAndPool(v, info)
		HandelPlural(v, info)
		info.oldStructInfo.varStr = getVarStr(info.oldStructInfo.vars)
	}

	info.newObjStr = replaceFrame(frame, info)

	if v.GetBool("pool") == true {
		getHeader(info)
	}

	getTwoPart(info)
}

//@ 查找模型
//@ 解析模型
func (this *esimFactory) FindStruct() bool {

	exists, err := file_dir.IsExistsDir(this.oldStructInfo.structDir)
	if err != nil {
		this.logger.Panicf(err.Error())
	}

	if exists == false {
		this.logger.Panicf("%s dir not exists", this.oldStructInfo.structDir)
	}

	files, err := ioutil.ReadDir(this.oldStructInfo.structDir)
	if err != nil {
		this.logger.Panicf(err.Error())
	}

	var found bool
	for _, fileInfo := range files {

		ext := path.Ext(fileInfo.Name())
		if ext != ".go" {
			continue
		}

		//测试文件不copy
		if strings.Index (fileInfo.Name(), "_test") > -1 {
			continue
		}

		this.oldStructInfo.filesName = append(this.oldStructInfo.filesName, fileInfo.Name())

		if !fileInfo.IsDir() {
			src, err := ioutil.ReadFile(this.oldStructInfo.structDir + "/" + fileInfo.Name())
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

								if typeSpec.Name.String() == this.oldStructInfo.structName {
									this.oldStructInfo.structFileContent = strSrc
									this.oldStructInfo.structFileName = fileInfo.Name()
									this.found = true
									this.oldStructInfo.packName = f.Name.String()
									this.oldStructInfo.packStr = strSrc[f.Name.Pos()-1 : f.Name.End()]
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

//extend logger and conf for oldstruct field
func (this *esimFactory) ExtendField() bool {

	var HasExtend bool
	if this.withOption == true {
		if this.withGenLoggerOption == true{
			HasExtend = true
			var foundLogField bool
			for _, field := range this.oldStructInfo.fields {
				if strings.Contains(field.Filed, "log.Logger") == true && foundLogField == false{
					foundLogField = true
				}
			}

			if foundLogField == false || len(this.oldStructInfo.fields) == 0{
				fld := db2entity.Field{}
				fld.Filed = "logger log.Logger"
				this.oldStructInfo.fields = append(this.oldStructInfo.fields, fld)
			}

			var foundLogImport bool
			for _, oim := range this.oldStructInfo.imports{
				if oim == "github.com/jukylin/esim/log"{
					foundLogImport = true
				}
			}

			if foundLogImport == false {
				this.appendOldImport("github.com/jukylin/esim/log")
			}
		}

		if this.withGenConfOption == true{
			HasExtend = true

			var foundConfField bool
			for _, field := range this.oldStructInfo.fields {
				if strings.Contains(field.Filed, "config.Config") == true && foundConfField == false{
					foundConfField = true
				}
			}

			if foundConfField == false || len(this.oldStructInfo.fields) == 0{
				fld := db2entity.Field{}
				fld.Filed = "conf config.Config"
				this.oldStructInfo.fields = append(this.oldStructInfo.fields, fld)
			}

			var foundConfImport bool
			for _, oim := range this.oldStructInfo.imports{
				if oim == "github.com/jukylin/esim/config"{
					foundConfImport = true
				}
			}
			if foundConfImport == false {
				this.appendOldImport("github.com/jukylin/esim/config")
			}
		}
	}
	return HasExtend
}

//有扩展属性才重写
func (this *esimFactory) ReWriteStructContent() error {

	if this.oldStructInfo.importStr != "" {
		this.oldStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
			this.oldStructInfo.importStr, this.getNewImport(this.oldStructInfo.imports), -1)
	} else if this.oldStructInfo.packStr != "" {
		this.getHeader()
		this.oldStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
			this.oldStructInfo.packStr, this.headerStr, -1)
	}

	this.oldStructInfo.importStr = this.getNewImport(this.oldStructInfo.imports)

	this.oldStructInfo.structFileContent = strings.Replace(this.oldStructInfo.structFileContent,
		this.oldStructInfo.structStr, db2entity.GetNewStruct(this.oldStructInfo.structName,
			this.oldStructInfo.fields), -1)
	this.oldStructInfo.structStr = db2entity.GetNewStruct(this.oldStructInfo.structName, this.oldStructInfo.fields)

	src, err := imports.Process("", []byte(this.oldStructInfo.structFileContent), nil)
	if err != nil{
		return err
	}

	return file_dir.EsimWrite(this.oldStructInfo.structDir +
		string(filepath.Separator) + this.oldStructInfo.structFileName,
			string(src))
}

func (this *esimFactory) getNewImport(imports []string) string {
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



//@ 建立虚拟环境
func BuildPluginEnv(info *esimFactory, buildBeforeFunc func()) error {

	info.structDir = strings.TrimRight(info.structDir, "/")

	targetDir := info.structDir + "/plugin"

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



	//TODO 生成 structName _plugin.go 文件
	_, err = GenPlugin(info.structName, info.oldFields, targetDir, info.oldImport)
	if err != nil {
		return err
	}

	if buildBeforeFunc != nil {
		buildBeforeFunc()
	}

	err = BuildPlugin(targetDir)
	if err != nil {
		return err
	}

	return nil
}


func ExecPlugin(v *viper.Viper, info *esimFactory) error {

	log2.SetOutput(ioutil.Discard)

	client := go_plugin.NewClient(&go_plugin.ClientConfig{
		HandshakeConfig: handshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(info.structDir + "/plugin/plugin"),
		Logger : hclog.New(&hclog.LoggerOptions{
			Output: hclog.DefaultOutput,
			Level:  hclog.Error,
			Name:   "plugin",
		}),
		})

	defer client.Kill()

	rpcClient, err := client.Client()
	if err != nil {
		return err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("model")
	if err != nil {
		return err
	}

	model := raw.(Model)

	if v.GetBool("sort") == true {

		re := &SortReturn{}

		json.Unmarshal([]byte(model.Sort()), re)

		info.newStruct = BuildNewStruct(info.structName, re.Fields, info.oldFields)
	}

	initReturn := InitFieldsReturn{}
	json.Unmarshal([]byte(model.InitField()), &initReturn)
	//HandleNewStruct(info, newStrcut)
	info.InitField = initReturn

	return nil
}


func NewVarStr(v *viper.Viper, info *esimFactory)  {
	if v.GetString("imp_iface") != ""{
		info.NewVarStr = v.GetString("imp_iface")
	}else if v.GetBool("pool") == true || v.GetBool("star") == true{
		info.NewVarStr = "*" + info.structName
	}else{
		info.NewVarStr = info.structName
	}
}

func NewOptionParam(v *viper.Viper, info *esimFactory)  {
	if v.GetBool("pool") == true || v.GetBool("star") == true{
		info.OptionParam = "*" + info.structName
	}else{
		info.OptionParam = info.structName
	}
}


func GetNewStr(v *viper.Viper, info *esimFactory) string {
	if v.GetBool("star") == true{
		return strings.ToLower(string(info.structName[0]))  + " := &" + info.structName + "{}"
	}else if v.GetBool("pool") == true{
		return strings.ToLower(string(info.structName[0])) + ` := ` + strings.ToLower(info.structName) + `Pool.Get().(*` + info.structName + `)`
	}else{
		return strings.ToLower(string(info.structName[0]))  + " := " + info.structName + "{}"
	}

	return ""
}


func GetReturnStr(info *esimFactory) string {
	return "	return " + strings.ToLower(string(info.structName[0]))
}

func NewFrame(v *viper.Viper, info *esimFactory) string {
	var newFrame string
	newFrame = `

{{options1}}

{{options2}}

func New` + strings.ToUpper(string(info.structName[0])) + string(info.structName[1:]) + `({{options3}}) ` + info.NewVarStr + ` {

	`+ GetNewStr(v, info) +`

	{{options4}}

	` + getInitStr(info) + `

` + GetReturnStr(info) + `
}

{{options5}}

`


	return newFrame
}

func replaceFrame(newFrame string, info *esimFactory) string {
	newFrame = strings.Replace(newFrame, "{{options1}}", info.option1, -1)

	newFrame = strings.Replace(newFrame, "{{options2}}", info.option2, -1)

	newFrame = strings.Replace(newFrame, "{{options3}}", info.option3, -1)

	newFrame = strings.Replace(newFrame, "{{options4}}", info.option4, -1)

	newFrame = strings.Replace(newFrame, "{{options5}}", info.option5, -1)

	return newFrame
}


func getOptions(v *viper.Viper, info *esimFactory)  {

	info.option1 = `type `+info.structName+`Option func(`+ info.OptionParam +`)`

	info.option2 = `type `+info.structName+`Options struct{}`

	info.option3 = `options ...`+info.structName+`Option`

	info.option4 = `
	for _, option := range options {
		option(` + strings.ToLower(string(info.structName[0])) + `)
	}`

	if v.GetBool("gen_conf_option") == true{

		info.option5 += `
func (`+info.structName+`Options) WithConf(conf config.Config) `+info.structName+`Option {
	return func(` + string(info.structName[0]) + ` `+ info.NewVarStr +`) {
	` + string(info.structName[0]) + `.conf = conf
	}
}
`

	}

	if v.GetBool("gen_logger_option") == true {
		info.option5 += `
func (`+info.structName+`Options) WithLogger(logger log.Logger) `+info.structName+`Option {
	return func(` + string(info.structName[0]) + ` ` + info.NewVarStr + `) {
		` + string(info.structName[0]) + `.logger = logger
	}
}
`
	}
}


func ReplaceContent(v *viper.Viper, info *esimFactory) (string, error) {

	src, err := ioutil.ReadFile(info.structDir + "/" + info.structFileName)
	if err != nil {
		return "", err
	}

	strSrc := string(src)

	if info.headerStr != "" {

		for _, varBody := range info.oldVarBody {
			strSrc = strings.Replace(strSrc, varBody, " ", -1)
		}

		if info.oldImportStr != "" {
			strSrc = strings.Replace(strSrc, info.oldImportStr, info.headerStr, -1)
		} else if info.packStr != "" {
			strSrc = strings.Replace(strSrc, info.packStr, info.headerStr, -1)
		}
	}

	strSrc = strings.Replace(strSrc, info.oldStruct, info.bodyStr, -1)

	return strSrc, nil
}

func GenPlugin(structName string, fields []db2entity.Field, dir string, oldImport []string) (string, error) {
	str := `package main

import (
	"unsafe"
	"sort"
	"encoding/json"
	"reflect"
	"strings"
	"github.com/hashicorp/go-plugin"
	"github.com/jukylin/esim/tool/factory"
)


type InitFieldsReturn struct{
	Fields []string
	SpecFields []Field
}

type Field struct{
	Name string
	Size int
	Type string
	TypeName string
}

type Fields []Field

func (f Fields) Len() int { return len(f) }

func (f Fields) Less(i, j int) bool {
	return f[i].Size < f[j].Size
}

func (f Fields) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
`
	str += "type Return struct{ \r\n"
	str += "Fields Fields `json:\"fields\"` \r\n"
	str += "Size int `json:\"size\"` \r\n"
	str += "} \r\n"

	str += GetSortBody(structName, fields)

	str += BuildInitField(structName)

	file := dir + "/" + structName + "_plugin.go"

	err := ioutil.WriteFile(file, []byte(str), 0666)
	if err != nil {
		println(err.Error())
	}
	return file, err
}

func GetSortBody(structName string, fields []db2entity.Field) string {
	str := "func (ModelImp) Sort() string { \r\n"

	str += "	" + strings.ToLower(structName) + " := " + structName + "{} \r\n"

	str += "	originSize := unsafe.Sizeof(" + strings.ToLower(structName) + ")\r\n"

	str += "	getType := reflect.TypeOf(" + strings.ToLower(structName) + ") \n"

	str += "	var fields Fields\r\n"

	for k, field := range fields {
		str += "	field" + strconv.Itoa(k) + " := Field{}\r\n"

		str += "	field" + strconv.Itoa(k) + ".Name = \"" + field.Filed + "\"\r\n"
		str += "	field" + strconv.Itoa(k) + ".Size =  + int(getType.Field(" + strconv.Itoa(k) + ").Type.Size())\r\n"

		str += "	fields = append(fields, field" + strconv.Itoa(k) + ")\r\n"
	}

	str += "	sort.Sort(fields)\r\n"

	str += "	re := &Return{}\r\n"

	str += "	re.Fields = fields\r\n"
	str += "	re.Size = int(originSize)\r\n"

	str += "	by, _ := json.Marshal(re)\r\n"

	str += "	return string(by)\r\n"

	str += "}"

	return str
}

func BuildNewStruct(structName string, fields Fields,
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

func BuildInitField(structName string) string {
	var str string
	str = `
func (ModelImp) InitField() string {
		` + strings.ToLower(structName) + ` := ` + structName + `{}

		initReturn := &InitFieldsReturn{}
	 	fields := &Fields{}

		getType := reflect.TypeOf(` + strings.ToLower(structName) + `)
		structFields := getInitStr(getType, strings.ToLower(getType.Name()), fields)

		initReturn.SpecFields = *fields
		initReturn.Fields = structFields
		j, _ := json.Marshal(initReturn)
		return string(j)
	}

	func getInitStr(getType reflect.Type, name string, specFilds *Fields) []string {
		typeNum := getType.NumField()
		var structFields []string
		var initStr string
		field  := Field{}

		for i := 0; i < typeNum; i++ {
		switch getType.Field(i).Type.Kind() {
			case reflect.Array:
				structFields = append(structFields, "for k, _ := range "+ name + "." + getType.Field(i).Name+" {")
					switch getType.Field(i).Type.Elem().Kind() {
					case reflect.Struct:
						structFields = append(structFields,
							getInitStr(getType.Field(i).Type.Elem(),
								name + "." + getType.Field(i).Name + "[k]", nil)...)
					default:
						initStr = KindToInit(getType.Field(i).Type.Elem(),  name + "." + getType.Field(i).Name + "[k]", nil)
						structFields = append(structFields, name + "." + getType.Field(i).Name+ "[k] = " + initStr)
					}
				structFields = append(structFields, "}")
				continue
			case reflect.Map:
				structFields = append(structFields, "for k, _ := range "+ name + "." + getType.Field(i).Name+" {")
				structFields = append(structFields, "delete(" + name + "." + getType.Field(i).Name + ", k)")
				structFields = append(structFields, "}")
				if specFilds != nil {
					field.Name = name + "." + getType.Field(i).Name
					field.Type = "map"
					field.TypeName = getType.Field(i).Type.String()
					*specFilds = append(*specFilds, field)
				}
				continue
			case reflect.Struct:
				if getType.Field(i).Type.String() == "time.Time"{
					initStr = "time.Time{}"
				}else {
					structFields = append(structFields, getInitStr(getType.Field(i).Type,
						name+"."+getType.Field(i).Name, nil)...)
					continue
				}
			default:
				initStr = KindToInit(getType.Field(i).Type,
					name + "." + getType.Field(i).Name, specFilds)
			}

			structFields = append(structFields, name + "." + getType.Field(i).Name + " = " + initStr)
		}

		return structFields
	}


func KindToInit(refType reflect.Type, name string, specFilds *Fields) string {
	var initStr string
	field  := Field{}

	switch refType.Kind() {
	case reflect.String:
		initStr = "\"\""
	case reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32:
		initStr = "0"
	case reflect.Uint, reflect.Uint64, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		initStr = "0"
	case reflect.Bool:
		initStr = "false"
	case reflect.Float32, reflect.Float64:
		initStr = "0.00"
	case reflect.Complex64, reflect.Complex128:
		initStr = "0+0i"
	case reflect.Interface:
		initStr = "nil"
	case reflect.Uintptr:
		initStr = "0"
	case reflect.Invalid, reflect.Func, reflect.Chan, reflect.Ptr, reflect.UnsafePointer:
		initStr = "nil"
	case reflect.Slice:
		if specFilds != nil {
			field.Name = name
			field.TypeName = refType.String()
			field.Type = "slice"
			*specFilds = append(*specFilds, field)
		}
		initStr = name + "[:0]"
	case reflect.Map:
		initStr = "nil"
	case reflect.Array:
		initStr = "nil"
	}

	return initStr
}

type ModelImp struct{}

func main() {

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: factory.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"model": &factory.ModelPlugin{Impl: &ModelImp{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}


`

	return str
}

func BuildPlugin(dir string) error {
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

func HandleNewStruct(info *esimFactory, newStrcut string) bool {
	src, err := ioutil.ReadFile(info.structDir + "/" + info.structFileName)
	if err != nil {
		log.Errorf(err.Error())
		return false
	}

	strSrc := string(src)

	strSrc = strings.Replace(strSrc, info.oldStruct, newStrcut, -1)

	dst, err := os.OpenFile(info.structDir+"/"+info.structFileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Errorf(err.Error())
		return false
	}
	defer dst.Close()

	dst.Write([]byte(strSrc))

	return true
}

//初始化变量，生成临时对象池
func HandleInitFieldsAndPool(v *viper.Viper, info *esimFactory) bool {

	HandleVar(info, info.structName, nil)

	HandlePool(v, info)

	return true
}

//处理复数
func HandelPlural(v *viper.Viper, info *esimFactory) bool {

	if info.pluralName != "" {

		HandleVar(info, info.pluralName, nil)

		HandlePluralPool(v, info)
	}

	return true
}

func HandleVar(info *esimFactory, varName string, context *string) bool {
	poolName := strings.ToLower(varName) + "Pool"
	if varNameExists(info.oldVar, poolName) == true {
		log.Debugf("变量已存在 : %s", poolName)
	} else {

		info.oldVar = append(info.oldVar, getPoolVar(varName))
		appendImport(info, "sync")
	}

	return true
}

//单数池
func HandlePool(v *viper.Viper, info *esimFactory) {
	//info.newObjStr = getNewObjStr(info)

	info.releaseStr = getReleaseObjStr(info, info.InitField.Fields)
}

//复数池
func HandlePluralPool(v *viper.Viper, info *esimFactory) {
	info.newPluralNewBody = getNewPluralObjStr(info)

	info.newPluralReleaseBody = getReleasePluralObjStr(info)
}

func getInitStr(info *esimFactory) string {
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

func getReleaseObjStr(info *esimFactory, initFields []string) string {
	str := "func (" + strings.ToLower(info.structName) + info.NewVarStr + ") Release() {\n"

	for _, field := range info.InitField.Fields {
		if strings.Contains(field, "time.Time") {
			appendImport(info, "time")
		}
		str += "		" + field + "\n"
	}

	str += "		" + strings.ToLower(info.structName) + "Pool.Put(" + strings.ToLower(info.structName) + ")\n"
	str += "}"

	return str
}

func getNewPluralObjStr(info *esimFactory) string {
	str := `func New` + info.pluralName + `() *` + info.pluralName + ` {
	` + strings.ToLower(info.pluralName) + ` := ` +
		strings.ToLower(info.pluralName) + `Pool.Get().(*` + info.pluralName + `)
`

	str += `return ` + strings.ToLower(info.pluralName) + `
}
`

	return str
}

func getReleasePluralObjStr(info *esimFactory) string {
	str := "func (" + strings.ToLower(info.pluralName) + " *" + info.pluralName + ") Release() {\n"

	str += "*" + strings.ToLower(info.pluralName) + " = (*" + strings.ToLower(info.pluralName) + ")[:0]\n"
	str += "		" + strings.ToLower(info.pluralName) + "Pool.Put(" + strings.ToLower(info.pluralName) + ")\n"
	str += "}"

	return str
}

func (this *esimFactory) appendOldImport(importName string) bool {
	var found bool
	importName = "\"" + importName + "\""
	for _, importStr := range this.oldStructInfo.imports {
		if importStr == importName {
			found = true
		}
	}

	if found == false {
		this.oldStructInfo.imports = append(this.oldStructInfo.imports, importName)
	}

	return true
}

func getVarStr(vars []Var) string {
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

func getPoolVar(structName string) Var {
	var poolVar Var
	poolVar.val = strings.ToLower(structName) + `Pool = sync.Pool{
        New: func() interface{} {
                return &` + structName + `{}
        },
	}
`
	poolVar.name = strings.ToLower(structName) + `Pool`
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
func (this *esimFactory) getHeader() {

	headerStr := ""
	if this.oldStructInfo.importStr == "" {
		headerStr = this.oldStructInfo.packStr + "\n"
	}

	headerStr += this.getNewImport(this.oldStructInfo.imports)
	headerStr += "\n"
	headerStr += this.oldStructInfo.varStr

	this.headerStr = headerStr
}

//struct body
func getTwoPart(info *esimFactory) {
	bodyStr := ""

	if info.newStruct != "" {
		bodyStr += info.newStruct
	} else {
		bodyStr += info.oldStruct
	}

	if info.pluralName != "" {
		bodyStr += "\n"
		bodyStr += "type " + info.pluralName + " []" + info.structName
	}

	if info.newObjStr != "" {
		bodyStr += "\n\n"
		bodyStr += info.newObjStr
	}

	if info.releaseStr != "" {
		bodyStr += "\n"
		bodyStr += info.releaseStr
	}

	if info.newPluralNewBody != "" {
		bodyStr += "\n\n"
		bodyStr += info.newPluralNewBody
	}

	if info.newPluralReleaseBody != "" {
		bodyStr += "\n"
		bodyStr += info.newPluralReleaseBody
	}

	info.bodyStr = bodyStr
}


