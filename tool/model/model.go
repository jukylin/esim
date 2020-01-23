package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/martinusso/inflect"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/tool/db2entity"
	"github.com/jukylin/esim/pkg/file_dir"
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
	//模型目录
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

	newVar string

	oldImport []string

	oldImportStr string

	willAppendImport []string

	newObjStr string

	releaseStr string

	headerStr string

	bodyStr string

	oldNewFuncBody string

	oldReleaseFuncBody string

	InitField InitFieldsReturn

	///模型的复数
	hasPlural bool

	pluralName string

	oldPluralNewBody string

	newPluralNewBody string

	oldPluralReleaseBody string

	newPluralReleaseBody string

	oldPluralType string
	///模型的复数
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
	modelname := v.GetString("modelname")
	if modelname == "" {
		return errors.New("请输入模型名称")
	}

	modelpath, err := os.Getwd()
	if err != nil {
		return err
	}

	plural := v.GetBool("plural")
	var modelNamePlural string
	if plural == true {
		modelNamePlural = getPluralWord(modelname)
	}

	info, err := FindModel(modelpath, modelname, modelNamePlural)
	if err != nil {
		return err
	}

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

	err = WriteContent(v, info)
	if err != nil {
		return err
	}

	db2entity.ExecGoFmt(info.modelFileName, info.modelDir)

	err = Clear(info)
	return err
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

								if info.pluralName != "" && typeSpec.Name.String() == modelNamePlural {
									info.hasPlural = true
									info.oldPluralType = string(src[GenDecl.TokPos-1 : typeSpec.End()])
								}

								if typeSpec.Name.String() == modelName {
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

				if GenDecl, ok := decl.(*ast.FuncDecl); ok && found == true {
					if GenDecl.Name.String() == "New"+modelName {
						info.oldNewFuncBody = strSrc[GenDecl.Pos()-1 : GenDecl.End()]
					}
					if GenDecl.Name.String() == "Release" {
						for _, recvList := range GenDecl.Recv.List {
							if recvList.Type.(*ast.StarExpr).X.(*ast.Ident).String() == modelName {
								info.oldReleaseFuncBody = strSrc[GenDecl.Pos()-1 : GenDecl.End()]
							}
						}
					}

					if info.hasPlural == true {
						if GenDecl.Name.String() == "New"+info.pluralName {
							info.oldPluralNewBody = strSrc[GenDecl.Pos()-1 : GenDecl.End()]
						}

						if GenDecl.Name.String() == "Release" {
							for _, recvList := range GenDecl.Recv.List {
								if recvList.Type.(*ast.StarExpr).X.(*ast.Ident).String() == info.pluralName {
									info.oldPluralReleaseBody = strSrc[GenDecl.Pos()-1 : GenDecl.End()]
								}
							}
						}
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
	_, err = GenPlugin(info.modelName, info.oldFields, targetDir)
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

	if v.GetBool("pool") == true {

		initReturn := InitFieldsReturn{}
		json.Unmarshal([]byte(model.InitField()), &initReturn)
		//HandleNewStruct(info, newStrcut)
		info.InitField = initReturn
		HandleInitFieldsAndPool(v, info)
		HandelPlural(v, info)
		info.newVar = getVarStr(info.oldVar)
	}

	if v.GetBool("pool") == true {
		getHeader(info)
	}

	getTwoPart(info)

	return nil
}

func WriteContent(v *viper.Viper, info *BuildPluginInfo) error {

	src, err := ioutil.ReadFile(info.modelDir + "/" + info.modelFileName)
	if err != nil {
		return err
	}

	strSrc := string(src)

	if info.oldNewFuncBody != "" && v.GetBool("pool") == true {
		strSrc = strings.Replace(strSrc, info.oldNewFuncBody, "", -1)
	}

	if info.oldReleaseFuncBody != "" && v.GetBool("pool") == true {
		strSrc = strings.Replace(strSrc, info.oldReleaseFuncBody, "", -1)
	}

	if info.oldPluralNewBody != "" && v.GetBool("plural") == true {
		strSrc = strings.Replace(strSrc, info.oldPluralNewBody, "", -1)
	}

	if info.oldPluralReleaseBody != "" && v.GetBool("plural") == true {
		strSrc = strings.Replace(strSrc, info.oldPluralReleaseBody, "", -1)
	}

	if info.oldPluralType != "" && v.GetBool("plural") == true {
		strSrc = strings.Replace(strSrc, info.oldPluralType, "", -1)
	}

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

	dst, err := os.OpenFile(info.modelDir+"/"+info.modelFileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()

	dst.Write([]byte(strSrc))

	return nil
}

func GenPlugin(structName string, fields []db2entity.Field, dir string) (string, error) {
	str := `package main

import (
	"unsafe"
	"sort"
	"encoding/json"
	"reflect"
	"strings"
	"github.com/hashicorp/go-plugin"
	"github.com/jukylin/esim/tool/model"
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
		HandshakeConfig: model.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"model": &model.ModelPlugin{Impl: &ModelImp{}},
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
	if info.oldNewFuncBody == "" || v.GetBool("coverpool") == true {
		info.newObjStr = getNewObjStr(info)
	}

	if info.oldReleaseFuncBody == "" || v.GetBool("coverpool") == true {
		info.releaseStr = getReleaseObjStr(info, info.InitField.Fields)
	}
}

//复数池
func HandlePluralPool(v *viper.Viper, info *BuildPluginInfo) {
	if info.oldPluralNewBody == "" || v.GetBool("coverpool") == true {
		info.newPluralNewBody = getNewPluralObjStr(info)
	}

	if info.oldPluralReleaseBody == "" || v.GetBool("coverpool") == true {
		info.newPluralReleaseBody = getReleasePluralObjStr(info)
	}
}

func getNewObjStr(info *BuildPluginInfo) string {
	str := `func New` + info.modelName + `() *` + info.modelName + ` {
	` + strings.ToLower(info.modelName) + ` := ` + strings.ToLower(info.modelName) + `Pool.Get().(*` + info.modelName + `)
`

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

	str += `return ` + strings.ToLower(info.modelName) + `
}
`

	return str
}

func getReleaseObjStr(info *BuildPluginInfo, initFields []string) string {
	str := "func (" + strings.ToLower(info.modelName) + " *" + info.modelName + ") Release() {\n"

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
	importStr := "import( \n"
	for _, imp := range info.oldImport {
		importStr += "	" + imp + "\n"
	}
	importStr += ") \n"

	headerStr += importStr
	headerStr += "\n"
	headerStr += info.newVar

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
		if info.hasPlural == false {
			bodyStr += "type " + info.pluralName + " []" + info.modelName
		} else {
			bodyStr += info.oldPluralType
		}
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
