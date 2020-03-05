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
	"regexp"
	"strconv"
	"strings"
	log2 "log"
	go_plugin "github.com/hashicorp/go-plugin"
	"golang.org/x/tools/imports"
	"github.com/hashicorp/go-hclog"
)

var (
	log logger.Logger
)

func init() {
	log = logger.NewLogger()
}

type SortReturn struct {
	Fields Fields `json:"fields"`
	Size   int    `json:"size"`
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

var HandshakeConfig = go_plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "BASIC_PLUGIN",
	MagicCookieValue: "hello",
}

type BuildPluginInfo struct {
	//模型目录 绝对路径
	modelDir string

	filesName []string

	//模型名称
	modelName string

	//模型文件名
	modelFileName string

	packName string

	packStr string

	oldFields []db2entity.Field

	oldStruct string

	newStruct string

	oldVar []Var

	oldVarBody []string

	VarStr string

	oldImport []string

	oldImportStr string

	willAppendImport []string

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

	//模型对应文件内容
	modelFileContent string
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

func HandleModel(v *viper.Viper) error {
	sname := v.GetString("sname")
	if sname == "" {
		return errors.New("请输入结构体名称")
	}

	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	plural := v.GetBool("plural")
	var structNamePlural string
	if plural == true {
		structNamePlural = getPluralWord(sname)
	}

	info, err := FindModel(wd, sname, structNamePlural)
	if err != nil {
		return err
	}

	if ExtendField(v, info) {
		err = ReWriteModelContent(info)
		if err != nil {
			return err
		}
	}

	if len(info.oldFields) > 0 {
		err = BuildPluginEnv(info, nil)
		if err != nil {
			return err
		}

		defer func() {
			if err := recover(); err != nil {
				log.Errorf("%v", err)
			}
			Clear(info)
		}()

		err = ExecPlugin(v, info)
		if err != nil {
			return err
		}
	}

	BuildFrame(v, info)

	err = file_dir.EsimBackUpFile(info.modelDir + "/" + info.modelFileName)
	if err != nil{
		log.Warnf("backup err %s:%s", info.modelDir + "/" + info.modelFileName, err.Error())
	}

	src, err := ReplaceContent(v, info)
	if err != nil {
		return err
	}

	res, err := imports.Process("", []byte(src), nil)
	if err != nil {
		return err
	}

	err = file_dir.EsimWrite(info.modelDir+"/"+info.modelFileName, string(res))
	if err != nil {
		return err
	}

	//err = db2entity.ExecGoFmt(info.modelFileName, info.modelDir)
	//if err != nil{
	//	return err
	//}

	err = Clear(info)
	return err
}

func BuildFrame(v *viper.Viper, info *BuildPluginInfo)  {
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
		info.VarStr = getVarStr(info.oldVar)
	}

	info.newObjStr = replaceFrame(frame, info)

	if v.GetBool("pool") == true {
		getHeader(info)
	}

	getTwoPart(info)
}

//@ 查找模型
//@ 解析模型
func FindModel(modelPath, modelName, modelNamePlural string) (*BuildPluginInfo, error) {

	modelPath = strings.TrimRight(modelPath, "/")

	exists, err := file_dir.IsExistsDir(modelPath)
	if err != nil {
		return nil, err
	}

	if exists == false {
		return nil, errors.New(modelPath + " dir not exists")
	}

	files, err := ioutil.ReadDir(modelPath)
	if err != nil {
		return nil, err
	}

	info := BuildPluginInfo{}
	info.modelName = modelName
	info.modelDir = modelPath
	info.pluralName = modelNamePlural

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

		found = false
		info.filesName = append(info.filesName, fileInfo.Name())

		if !fileInfo.IsDir() {
			src, err := ioutil.ReadFile(modelPath + "/" + fileInfo.Name())
			if err != nil {
				return nil, err
			}

			strSrc := string(src)
			fset := token.NewFileSet() // positions are relative to fset
			f, err := parser.ParseFile(fset, "", strSrc, parser.ParseComments)
			if err != nil {
				return nil, err
			}

			for _, decl := range f.Decls {
				if GenDecl, ok := decl.(*ast.GenDecl); ok {
					if GenDecl.Tok.String() == "type" {
						for _, specs := range GenDecl.Specs {
							if typeSpec, ok := specs.(*ast.TypeSpec); ok {

								if typeSpec.Name.String() == modelName {
									info.modelFileContent = strSrc
									info.modelFileName = fileInfo.Name()
									found = true
									info.packName = f.Name.String()
									info.packStr = strSrc[f.Name.Pos()-1 : f.Name.End()]
									info.oldFields = db2entity.GetOldFields(GenDecl, strSrc)
									info.oldStruct = string(src[GenDecl.TokPos-1 : typeSpec.Type.(*ast.StructType).Fields.Closing])
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
									info.oldVarBody = append(info.oldVarBody,
										strSrc[GenDecl.TokPos-1:GenDecl.Rparen])
								} else {
									info.oldVarBody = append(info.oldVarBody,
										strSrc[GenDecl.TokPos-1:typeSpec.End()])
								}

								fileVar.val = strSrc[typeSpec.Pos()-1 : typeSpec.End()]
								if typeSpec.Doc != nil {
									for _, doc := range typeSpec.Doc.List {
										fileVar.doc = append(fileVar.doc, doc.Text)
									}
								}
								fileVar.name = typeSpec.Names[0].Name
								info.oldVar = append(info.oldVar, fileVar)
							}
						}
					}

					if GenDecl.Tok.String() == "import" && found == true {
						for _, specs := range GenDecl.Specs {
							if typeSpec, ok := specs.(*ast.ImportSpec); ok {
								if typeSpec.Name.String() != "<nil>" {
									info.oldImport = append(info.oldImport,
										typeSpec.Name.String()+" "+typeSpec.Path.Value)
								} else {
									info.oldImport = append(info.oldImport, typeSpec.Path.Value)
								}
							}
						}
						info.oldImportStr = strSrc[GenDecl.Pos()-1 : GenDecl.End()]
					}
				}
			}
		}
	}

	if info.modelFileName != "" {
		return &info, nil
	} else {
		return nil, errors.New("not found model")
	}
}

//extend logger and conf
func ExtendField(v *viper.Viper, info *BuildPluginInfo) bool {

	var HasExtend bool
	if v.GetBool("option") == true {
		if v.GetBool("gen_logger_option") == true{
			HasExtend = true
			var foundLogField bool
			for _, field := range info.oldFields {
				if strings.Contains(field.Filed, "log.Logger") == true && foundLogField == false{
					foundLogField = true
				}
			}

			if foundLogField == false || len(info.oldFields) == 0{
				fld := db2entity.Field{}
				fld.Filed = "logger log.Logger"
				info.oldFields = append(info.oldFields, fld)
			}

			var foundLogImport bool
			for _, oim := range info.oldImport{
				if oim == "github.com/jukylin/esim/log"{
					foundLogImport = true
				}
			}

			if foundLogImport == false {
				appendImport(info, "github.com/jukylin/esim/log")
			}
		}

		if v.GetBool("gen_conf_option") == true{
			HasExtend = true

			var foundConfField bool
			for _, field := range info.oldFields {
				if strings.Contains(field.Filed, "config.Config") == true && foundConfField == false{
					foundConfField = true
				}
			}

			if foundConfField == false || len(info.oldFields) == 0{
				fld := db2entity.Field{}
				fld.Filed = "conf config.Config"
				info.oldFields = append(info.oldFields, fld)
			}

			var foundConfImport bool
			for _, oim := range info.oldImport{
				if oim == "github.com/jukylin/esim/config"{
					foundConfImport = true
				}
			}
			if foundConfImport == false {
				appendImport(info, "github.com/jukylin/esim/config")
			}
		}
	}
	return HasExtend
}

//有扩展属性才重写
func ReWriteModelContent(info *BuildPluginInfo) error {

	if info.oldImportStr != "" {
		info.modelFileContent = strings.Replace(info.modelFileContent, info.oldImportStr, getNewImport(info.oldImport), -1)
	} else if info.packStr != "" {
		getHeader(info)
		info.modelFileContent = strings.Replace(info.modelFileContent, info.packStr, info.headerStr, -1)
	}

	info.oldImportStr = getNewImport(info.oldImport)

	info.modelFileContent = strings.Replace(info.modelFileContent, info.oldStruct, db2entity.GetNewStruct(info.modelName, info.oldFields), -1)
	info.oldStruct = db2entity.GetNewStruct(info.modelName, info.oldFields)

	src, err := imports.Process("", []byte(info.modelFileContent), nil)
	if err != nil{
		return err
	}

	return file_dir.EsimWrite(info.modelDir+"/"+info.modelFileName, string(src))
}

func getNewImport(imports []string) string {
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
func BuildPluginEnv(info *BuildPluginInfo, buildBeforeFunc func()) error {

	info.modelDir = strings.TrimRight(info.modelDir, "/")

	targetDir := info.modelDir + "/plugin"

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

	//TODO 复制文件
	//TODO 改 package 名称
	for _, name := range info.filesName {
		CopyFile(targetDir+"/"+name, info.modelDir+"/"+name, info.packName)
	}

	//TODO 生成 modelName _plugin.go 文件
	_, err = GenPlugin(info.modelName, info.oldFields, targetDir, info.oldImport)
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

//@ Copy File
//@ repackagename
func CopyFile(dstName, srcName string, packageName string) (bool, error) {
	src, err := os.Open(srcName)
	if err != nil {
		return false, err
	}

	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return false, err
	}
	defer dst.Close()

	reg, _ := regexp.Compile(`package[\s*]` + packageName)

	contents, err := ioutil.ReadAll(src)
	contents = reg.ReplaceAll(contents, []byte("package main"))

	dst.Write(contents)

	return true, nil
}

func ExecPlugin(v *viper.Viper, info *BuildPluginInfo) error {

	var pluginMap = map[string]go_plugin.Plugin{
		"model": &ModelPlugin{},
	}

	log2.SetOutput(ioutil.Discard)

	client := go_plugin.NewClient(&go_plugin.ClientConfig{
		HandshakeConfig: HandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(info.modelDir + "/plugin/plugin"),
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

		info.newStruct = BuildNewStruct(info.modelName, re.Fields, info.oldFields)
	}

	initReturn := InitFieldsReturn{}
	json.Unmarshal([]byte(model.InitField()), &initReturn)
	//HandleNewStruct(info, newStrcut)
	info.InitField = initReturn

	return nil
}


func NewVarStr(v *viper.Viper, info *BuildPluginInfo)  {
	if v.GetString("imp_iface") != ""{
		info.NewVarStr = v.GetString("imp_iface")
	}else if v.GetBool("pool") == true || v.GetBool("star") == true{
		info.NewVarStr = "*" + info.modelName
	}else{
		info.NewVarStr = info.modelName
	}
}

func NewOptionParam(v *viper.Viper, info *BuildPluginInfo)  {
	if v.GetBool("pool") == true || v.GetBool("star") == true{
		info.OptionParam = "*" + info.modelName
	}else{
		info.OptionParam = info.modelName
	}
}


func GetNewStr(v *viper.Viper, info *BuildPluginInfo) string {
	if v.GetBool("star") == true{
		return strings.ToLower(string(info.modelName[0]))  + " := &" + info.modelName + "{}"
	}else if v.GetBool("pool") == true{
		return strings.ToLower(string(info.modelName[0])) + ` := ` + strings.ToLower(info.modelName) + `Pool.Get().(*` + info.modelName + `)`
	}else{
		return strings.ToLower(string(info.modelName[0]))  + " := " + info.modelName + "{}"
	}

	return ""
}


func GetReturnStr(info *BuildPluginInfo) string {
	return "	return " + strings.ToLower(string(info.modelName[0]))
}

func NewFrame(v *viper.Viper, info *BuildPluginInfo) string {
	var newFrame string
	newFrame = `

{{options1}}

{{options2}}

func New` + strings.ToUpper(string(info.modelName[0])) + string(info.modelName[1:]) + `({{options3}}) ` + info.NewVarStr + ` {

	`+ GetNewStr(v, info) +`

	{{options4}}

	` + getInitStr(info) + `

` + GetReturnStr(info) + `
}

{{options5}}

`


	return newFrame
}

func replaceFrame(newFrame string, info *BuildPluginInfo) string {
	newFrame = strings.Replace(newFrame, "{{options1}}", info.option1, -1)

	newFrame = strings.Replace(newFrame, "{{options2}}", info.option2, -1)

	newFrame = strings.Replace(newFrame, "{{options3}}", info.option3, -1)

	newFrame = strings.Replace(newFrame, "{{options4}}", info.option4, -1)

	newFrame = strings.Replace(newFrame, "{{options5}}", info.option5, -1)

	return newFrame
}


func getOptions(v *viper.Viper, info *BuildPluginInfo)  {

	info.option1 = `type `+info.modelName+`Option func(`+ info.OptionParam +`)`

	info.option2 = `type `+info.modelName+`Options struct{}`

	info.option3 = `options ...`+info.modelName+`Option`

	info.option4 = `
	for _, option := range options {
		option(` + strings.ToLower(string(info.modelName[0])) + `)
	}`

	if v.GetBool("gen_conf_option") == true{

		info.option5 += `
func (`+info.modelName+`Options) WithConf(conf config.Config) `+info.modelName+`Option {
	return func(` + string(info.modelName[0]) + ` `+ info.NewVarStr +`) {
	` + string(info.modelName[0]) + `.conf = conf
	}
}
`

	}

	if v.GetBool("gen_logger_option") == true {
		info.option5 += `
func (`+info.modelName+`Options) WithLogger(logger log.Logger) `+info.modelName+`Option {
	return func(` + string(info.modelName[0]) + ` ` + info.NewVarStr + `) {
		` + string(info.modelName[0]) + `.logger = logger
	}
}
`
	}
}


func ReplaceContent(v *viper.Viper, info *BuildPluginInfo) (string, error) {

	src, err := ioutil.ReadFile(info.modelDir + "/" + info.modelFileName)
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

func HandleNewStruct(info *BuildPluginInfo, newStrcut string) bool {
	src, err := ioutil.ReadFile(info.modelDir + "/" + info.modelFileName)
	if err != nil {
		log.Errorf(err.Error())
		return false
	}

	strSrc := string(src)

	strSrc = strings.Replace(strSrc, info.oldStruct, newStrcut, -1)

	dst, err := os.OpenFile(info.modelDir+"/"+info.modelFileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		log.Errorf(err.Error())
		return false
	}
	defer dst.Close()

	dst.Write([]byte(strSrc))

	return true
}

//初始化变量，生成临时对象池
func HandleInitFieldsAndPool(v *viper.Viper, info *BuildPluginInfo) bool {

	HandleVar(info, info.modelName, nil)

	HandlePool(v, info)

	return true
}

//处理复数
func HandelPlural(v *viper.Viper, info *BuildPluginInfo) bool {

	if info.pluralName != "" {

		HandleVar(info, info.pluralName, nil)

		HandlePluralPool(v, info)
	}

	return true
}

func HandleVar(info *BuildPluginInfo, varName string, context *string) bool {
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
func HandlePool(v *viper.Viper, info *BuildPluginInfo) {
	//info.newObjStr = getNewObjStr(info)

	info.releaseStr = getReleaseObjStr(info, info.InitField.Fields)
}

//复数池
func HandlePluralPool(v *viper.Viper, info *BuildPluginInfo) {
	info.newPluralNewBody = getNewPluralObjStr(info)

	info.newPluralReleaseBody = getReleasePluralObjStr(info)
}

func getInitStr(info *BuildPluginInfo) string {
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

func getReleaseObjStr(info *BuildPluginInfo, initFields []string) string {
	str := "func (" + strings.ToLower(info.modelName) + info.NewVarStr + ") Release() {\n"

	for _, field := range info.InitField.Fields {
		if strings.Contains(field, "time.Time") {
			appendImport(info, "time")
		}
		str += "		" + field + "\n"
	}

	str += "		" + strings.ToLower(info.modelName) + "Pool.Put(" + strings.ToLower(info.modelName) + ")\n"
	str += "}"

	return str
}

func getNewPluralObjStr(info *BuildPluginInfo) string {
	str := `func New` + info.pluralName + `() *` + info.pluralName + ` {
	` + strings.ToLower(info.pluralName) + ` := ` +
		strings.ToLower(info.pluralName) + `Pool.Get().(*` + info.pluralName + `)
`

	str += `return ` + strings.ToLower(info.pluralName) + `
}
`

	return str
}

func getReleasePluralObjStr(info *BuildPluginInfo) string {
	str := "func (" + strings.ToLower(info.pluralName) + " *" + info.pluralName + ") Release() {\n"

	str += "*" + strings.ToLower(info.pluralName) + " = (*" + strings.ToLower(info.pluralName) + ")[:0]\n"
	str += "		" + strings.ToLower(info.pluralName) + "Pool.Put(" + strings.ToLower(info.pluralName) + ")\n"
	str += "}"

	return str
}

func appendImport(info *BuildPluginInfo, importName string) bool {
	var found bool
	importName = "\"" + importName + "\""
	for _, importStr := range info.oldImport {
		if importStr == importName {
			found = true
		}
	}

	if found == false {
		info.oldImport = append(info.oldImport, importName)
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

func getPoolVar(modelName string) Var {
	var poolVar Var
	poolVar.val = strings.ToLower(modelName) + `Pool = sync.Pool{
        New: func() interface{} {
                return &` + modelName + `{}
        },
	}
`
	poolVar.name = strings.ToLower(modelName) + `Pool`
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
func getHeader(info *BuildPluginInfo) {

	headerStr := ""
	if info.oldImportStr == "" {
		headerStr = info.packStr + "\n"
	}

	headerStr += getNewImport(info.oldImport)
	headerStr += "\n"
	headerStr += info.VarStr

	info.headerStr = headerStr
}

//struct body
func getTwoPart(info *BuildPluginInfo) {
	bodyStr := ""

	if info.newStruct != "" {
		bodyStr += info.newStruct
	} else {
		bodyStr += info.oldStruct
	}

	if info.pluralName != "" {
		bodyStr += "\n"
		bodyStr += "type " + info.pluralName + " []" + info.modelName
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

func Clear(info *BuildPluginInfo) error {
	err := os.RemoveAll(info.modelDir + "/plugin")
	if err != nil {
		return err
	} else {
		return nil
	}
}
