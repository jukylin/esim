package db2entity

import (

	"strings"
	logger "github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/file-dir"

	"github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/spf13/viper"
	"path/filepath"
	"go/token"
	"go/parser"
	"go/ast"
	"io/ioutil"
	"go/format"
	"golang.org/x/tools/imports"
)

type Db2Entity struct {
	WithDisabledRepo bool

	WithRepoTarget string

	WithDisabledDao bool

	WithDaoTarget string

	//true not create entity file
	//false create a new entity file in withEntityTarget
	WithDisabledEntity bool

	WithEntityTarget string

	//true inject repo to infra
	withInject bool

	logger logger.Logger

	withBoubctx string

	withPackage string

	withStruct string

	//Camel Form
	CamelStruct string

	ColumnsRepo domain_file.ColumnsRepo

	DbConf *domain_file.DbConfig

	writer file_dir.IfaceWriter

	withInfraDir string

	withInfraFile string

	hasInfraStruct bool

	oldInfraInfo *infraInfo

	newInfraInfo *infraInfo

	execer pkg.Exec
}

type Db2EnOption func(*Db2Entity)

type Db2EnOptions struct{}

func NewDb2Entity(options ...Db2EnOption) *Db2Entity {

	d := &Db2Entity{}

	for _, option := range options {
		option(d)
	}

	if d.writer == nil {
		d.writer = file_dir.NewNullWrite()
	}

	if d.execer == nil {
		d.execer = &pkg.NullExec{}
	}

	if d.logger == nil {
		d.logger = logger.NewNullLogger()
	}

	return d
}


func (Db2EnOptions) WithLogger(logger logger.Logger) Db2EnOption {
	return func(d *Db2Entity) {
		d.logger = logger
	}
}


func (Db2EnOptions) WithColumnsInter(ColumnsRepo domain_file.ColumnsRepo) Db2EnOption {
	return func(d *Db2Entity) {
		d.ColumnsRepo = ColumnsRepo
	}
}

func (Db2EnOptions) WithIfaceWrite(writer file_dir.IfaceWriter) Db2EnOption {
	return func(d *Db2Entity) {
		d.writer = writer
	}
}


func (Db2EnOptions) WithInfraInfo(infra *infraInfo) Db2EnOption {
	return func(d *Db2Entity) {
		d.oldInfraInfo = infra
		d.newInfraInfo = infra
	}
}

func (Db2EnOptions) WithWriter(writer file_dir.IfaceWriter) Db2EnOption {
	return func(d *Db2Entity) {
		d.writer = writer
	}
}

func (Db2EnOptions) WithExecer(execer pkg.Exec) Db2EnOption {
	return func(d *Db2Entity) {
		d.execer = execer
	}
}

func (Db2EnOptions) WithDbConf(dbConf *domain_file.DbConfig) Db2EnOption {
	return func(d *Db2Entity) {
		d.DbConf = dbConf
	}
}

type infraInfo struct {

	imports pkg.Imports

	importStr string

	structInfo templates.StructInfo

	structStr string

	specialStructName string

	specialVarName string

	infraSetArgs infraSetArgs

	infraSetStr string

	content string
}

func NewInfraInfo() *infraInfo {
	ifaInfo := &infraInfo{}

	ifaInfo.specialStructName = "Infra"

	ifaInfo.specialVarName = "infraSet"

	ifaInfo.infraSetArgs = infraSetArgs{}

	structInfo := templates.StructInfo{}
	structInfo.StructName = ifaInfo.specialStructName
	ifaInfo.structInfo = structInfo

	ifaInfo.imports = pkg.Imports{}

	return ifaInfo
}

func (de *Db2Entity) Run(v *viper.Viper) error {
	de.bindInput(v)

	de.injectToInfra()

	return nil
}

func (de *Db2Entity) bindInput(v *viper.Viper) {

	de.DbConf.ParseConfig(v, de.logger)

	packageName := v.GetString("package")
	if packageName == "" {
		packageName = de.DbConf.Database
	}
	de.withPackage = packageName

	stuctName := v.GetString("struct")
	if stuctName == "" {
		stuctName = de.DbConf.Table
	}
	de.withStruct = stuctName
	de.CamelStruct = templates.SnakeToCamel(stuctName)

	de.bindInfra(v)
}



func (de *Db2Entity) bindInfra(v *viper.Viper) {
	de.withInject = v.GetBool("inject")

	de.withInfraDir = v.GetString("infra_dir")
	if de.withInfraDir == ""{
		de.withInfraDir = "internal" + string(filepath.Separator) + "infra" + string(filepath.Separator)
	} else {
		de.withInfraDir = strings.TrimLeft(de.withInfraDir, ".") + string(filepath.Separator)
		de.withInfraDir = strings.Trim(de.withInfraDir, "/") + string(filepath.Separator)
	}

	if v.GetString("infra_file") == "" {
		de.withInfraFile = "infra.go"
	}

	exists, err := file_dir.IsExistsFile(de.withInfraDir + de.withInfraFile)
	if err != nil {
		de.logger.Fatalf(err.Error())
		return
	}

	if exists == false {
		de.logger.Fatalf("%s not exists", de.withInfraDir + de.withInfraFile)
	}
}


// ./a/b/c/ => a/b/c
func (de *Db2Entity) DirPathToImportPath(dirPath string) string {
	path := strings.TrimLeft(dirPath, ".")
	path = strings.Trim(dirPath, "/")
	path = string(filepath.Separator) + path
	return path
}

//injectToInfra inject repo to infra.go and execute wire command
func (de *Db2Entity) injectToInfra() {

	if de.withInject == false {
		return
	}

	//back up infra.go
	err := file_dir.EsimBackUpFile(file_dir.GetCurrentDir() + string(filepath.Separator) + de.withInfraDir + de.withInfraFile)
	if err != nil {
		de.logger.Fatalf(err.Error())
		return
	}

	beautifulSource := de.sourceInfraFile()

	de.parseInfra(beautifulSource)

	if de.hasInfraStruct {
		de.copyInfraInfo()

		de.processNewInfra()

		de.toStringNewInfra()

		de.buildNewInfraString()

		de.writeNewInfra()

	} else {
		de.logger.Fatalf("not found the %s", de.oldInfraInfo.specialStructName)
	}

	de.logger.Infof("inject success")
}

//parseInfra parse infra.go 's content, find "import", "Infra" , "infraSet" and record origin syntax
func (de *Db2Entity) parseInfra(srcStr string) bool {

	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", srcStr, parser.ParseComments)
	if err != nil {
		de.logger.Fatalf(err.Error())
	}

	for _, decl := range f.Decls {
		if GenDecl, ok := decl.(*ast.GenDecl); ok {
			if GenDecl.Tok.String() == "import" {
				imps := pkg.Imports{}
				imps.ParseFromAst(GenDecl)
				de.oldInfraInfo.imports = imps
				de.oldInfraInfo.importStr = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
			}

			if GenDecl.Tok.String() == "type" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.TypeSpec); ok {
						if typeSpec.Name.String() == de.oldInfraInfo.specialStructName {
							de.hasInfraStruct = true
							fields := pkg.Fields{}
							fields.ParseFromAst(GenDecl, srcStr)
							de.oldInfraInfo.structInfo.Fields = fields
							de.oldInfraInfo.structStr = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
						}
					}
				}
			}

			if GenDecl.Tok.String() == "var" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.ValueSpec); ok {
						for _, name := range typeSpec.Names {
							if name.String() == de.oldInfraInfo.specialVarName {
								de.oldInfraInfo.infraSetStr = srcStr[GenDecl.TokPos-1 : GenDecl.End()]
								de.oldInfraInfo.infraSetArgs.Args = append(de.oldInfraInfo.infraSetArgs.Args,
									de.getInfraSetArgs(GenDecl, srcStr)...)
							}
						}
					}
				}
			}
		}
	}

	if de.hasInfraStruct == false {
		de.logger.Fatalf("not find %s", de.oldInfraInfo.specialStructName)
	}

	de.oldInfraInfo.content = srcStr

	//srcStr = strings.Replace(srcStr, oldImportStr, newImportStr, -1)
	//srcStr = strings.Replace(srcStr, oldStruct, newStruct, -1)
	//srcStr = strings.Replace(srcStr, oldSet, newSet, -1)
	//srcStr += provideFunc

	return true
}

//sourceInfraFile Beautify infra.go
func (de *Db2Entity) sourceInfraFile() string {

	src, err := ioutil.ReadFile(de.withInfraDir + de.withInfraFile)
	if err != nil {
		de.logger.Fatalf(err.Error())
	}

	formatSrc, err := format.Source([]byte(src))
	if err != nil {
		de.logger.Fatalf(err.Error())
	}

	ioutil.WriteFile(de.withInfraDir + de.withInfraFile, formatSrc, 0666)

	return string(formatSrc)
}

func (de *Db2Entity) copyInfraInfo() {
	oldContent := *de.oldInfraInfo
	de.newInfraInfo = &oldContent
}

//processInfraInfo process newInfraInfo, append import, repo field and wire's provider
func (de *Db2Entity) processNewInfra() bool {

	field := pkg.Field{}
	field.Name = de.CamelStruct + "Repo"
	field.Type = " repo." + de.CamelStruct + "Repo"
	field.Field = field.Name + " " + field.Type
	de.newInfraInfo.structInfo.Fields = append(de.newInfraInfo.structInfo.Fields, field)

	de.newInfraInfo.infraSetArgs.Args = append(de.newInfraInfo.infraSetArgs.Args,
		"provide" + de.CamelStruct + "Repo" + ",")
	
	imp := pkg.Import{Path: file_dir.GetGoProPath() + de.DirPathToImportPath(de.WithRepoTarget)}

	de.newInfraInfo.imports = append(de.newInfraInfo.imports, imp)

	return true
}

func (de *Db2Entity) toStringNewInfra() {

	de.newInfraInfo.importStr = de.newInfraInfo.imports.String()

	de.newInfraInfo.structStr = de.newInfraInfo.structInfo.String()

	de.newInfraInfo.infraSetStr = de.newInfraInfo.infraSetArgs.String()

}

func (de *Db2Entity) buildNewInfraString() {

	oldContent := de.oldInfraInfo.content

	oldContent = strings.Replace(oldContent,
		de.oldInfraInfo.importStr, de.newInfraInfo.importStr, -1)

	oldContent = strings.Replace(oldContent,
		de.oldInfraInfo.structStr, de.newInfraInfo.structStr, -1)

	de.newInfraInfo.content = strings.Replace(oldContent,
		de.oldInfraInfo.infraSetStr, de.newInfraInfo.infraSetStr, -1)

	de.newInfraInfo.content += de.appendProvideFunc()
}

func (de *Db2Entity) appendProvideFunc() string {
	return de.executeTmpl("provide_tpl", domain_file.NewRepoTpl(de.CamelStruct), domain_file.ProvideTemplate)
}

func (de *Db2Entity) writeNewInfra() {

	sourceSrc, err := format.Source([]byte(de.newInfraInfo.content))
	if err != nil {
		de.logger.Fatalf(err.Error())
		return
	}

	processSrc, err := imports.Process("", sourceSrc, nil)
	if err != nil {
		de.logger.Fatalf(err.Error())
		return
	}

	de.writer.Write(de.withInfraDir + de.withInfraFile, string(processSrc))

	err = de.execer.ExecWire(de.withInfraDir)
	if err != nil {
		de.logger.Fatalf(err.Error())
	}
}

func (de *Db2Entity) getInfraSetArgs(GenDecl *ast.GenDecl, srcStr string) []string {
	var args []string
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.ValueSpec); ok {
			for _, value := range spec.Values {
				if callExpr, ok := value.(*ast.CallExpr); ok {
					for _, callArg := range callExpr.Args {
						args = append(args, pkg.ParseExpr(callArg, srcStr))
					}
				}
			}
		}
	}

	return args
}


