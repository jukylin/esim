package infra

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/spf13/viper"
	"golang.org/x/tools/imports"
)

type Infraer struct {
	logger log.Logger

	writer file_dir.IfaceWriter

	execer pkg.Exec

	oldInfraInfo *infraInfo

	newInfraInfo *infraInfo

	injectInfos []*domain_file.InjectInfo

	hasInfraStruct bool

	withInfraDir string

	withInfraFile string
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

	provides domain_file.Provides

	provideStr string
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

type Option func(*Infraer)

func NewInfraer(options ...Option) *Infraer {

	infraer := &Infraer{}

	for _, option := range options {
		option(infraer)
	}

	return infraer
}

func WithIfacerLogger(logger log.Logger) Option {
	return func(ir *Infraer) {
		ir.logger = logger
	}
}

func WithIfacerInfraInfo(infra *infraInfo) Option {
	return func(ir *Infraer) {
		ir.oldInfraInfo = infra
		ir.newInfraInfo = infra
	}
}

func WithIfacerWriter(writer file_dir.IfaceWriter) Option {
	return func(ir *Infraer) {
		ir.writer = writer
	}
}

func WithIfacerExecer(execer pkg.Exec) Option {
	return func(ir *Infraer) {
		ir.execer = execer
	}
}

//injectToInfra inject repo to infra.go and execute wire command
func (ir *Infraer) Inject(v *viper.Viper, injectInfos []*domain_file.InjectInfo) bool {
	var err error

	if len(injectInfos) < 0 {
		ir.logger.Errorf("not need inject")
		return false
	}

	result := ir.bindInput(v)
	if !result {
		return false
	}

	ir.injectInfos = injectInfos

	//back up infra.go
	err = file_dir.EsimBackUpFile(file_dir.GetCurrentDir() + string(filepath.Separator) + ir.withInfraDir + ir.withInfraFile)
	if err != nil {
		ir.logger.Errorf(err.Error())
		return false
	}

	beautifulSource := ir.sourceInfraFile()
	if beautifulSource == "" {
		return false
	}

	parseResult := ir.parseInfra(beautifulSource)
	if !parseResult {
		return false
	}

	if ir.hasInfraStruct {
		ir.copyInfraInfo()

		ir.processNewInfra()

		ir.toStringNewInfra()

		ir.buildNewInfraContent()

		ir.writeNewInfra()

	} else {
		ir.logger.Errorf("not found the %s", ir.oldInfraInfo.specialStructName)
		return false
	}

	ir.logger.Infof("inject success")

	return true
}

func (ir *Infraer) bindInput(v *viper.Viper) bool {

	ir.withInfraDir = v.GetString("infra_dir")
	if ir.withInfraDir == "" {
		ir.withInfraDir = "internal" + string(filepath.Separator) + "infra" + string(filepath.Separator)
	} else {
		ir.withInfraDir = strings.TrimLeft(ir.withInfraDir, ".") + string(filepath.Separator)
		ir.withInfraDir = strings.Trim(ir.withInfraDir, "/") + string(filepath.Separator)
	}

	if v.GetString("infra_file") == "" {
		ir.withInfraFile = "infra.go"
	}

	exists, err := file_dir.IsExistsFile(ir.withInfraDir + ir.withInfraFile)
	if err != nil {
		ir.logger.Errorf(err.Error())
		return false
	}

	if !exists {
		ir.logger.Errorf("%s not exists", ir.withInfraDir+ir.withInfraFile)
		return false
	}

	return true
}

//parseInfra parse infra.go 's content, find "import", "Infra" , "infraSet" and record origin syntax
func (ir *Infraer) parseInfra(srcStr string) bool {

	// positions are relative to fset
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", srcStr, parser.ParseComments)
	if err != nil {
		ir.logger.Errorf(err.Error())
		return false
	}

	for _, decl := range f.Decls {
		if GenDecl, ok := decl.(*ast.GenDecl); ok {
			if GenDecl.Tok.String() == "import" {
				imps := pkg.Imports{}
				imps.ParseFromAst(GenDecl)
				ir.oldInfraInfo.imports = imps
				ir.oldInfraInfo.importStr = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
			}

			if GenDecl.Tok.String() == "type" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.TypeSpec); ok {
						if typeSpec.Name.String() == ir.oldInfraInfo.specialStructName {
							ir.hasInfraStruct = true
							fields := pkg.Fields{}
							fields.ParseFromAst(GenDecl, srcStr)
							ir.oldInfraInfo.structInfo.Fields = fields
							ir.oldInfraInfo.structStr = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
						}
					}
				}
			}

			if GenDecl.Tok.String() == "var" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.ValueSpec); ok {
						for _, name := range typeSpec.Names {
							if name.String() == ir.oldInfraInfo.specialVarName {
								ir.oldInfraInfo.infraSetStr = srcStr[GenDecl.TokPos-1 : GenDecl.End()]
								ir.oldInfraInfo.infraSetArgs.Args = append(ir.oldInfraInfo.infraSetArgs.Args,
									ir.parseInfraSetArgs(GenDecl, srcStr)...)
							}
						}
					}
				}
			}
		}
	}

	if !ir.hasInfraStruct {
		ir.logger.Errorf("not find %s", ir.oldInfraInfo.specialStructName)
		return false
	}

	ir.oldInfraInfo.content = srcStr

	return true
}

//sourceInfraFile Beautify infra.go
func (ir *Infraer) sourceInfraFile() string {

	src, err := ioutil.ReadFile(ir.withInfraDir + ir.withInfraFile)
	if err != nil {
		ir.logger.Errorf(err.Error())
		return ""
	}

	formatSrc := ir.makeCodeBeautiful(string(src))

	ioutil.WriteFile(ir.withInfraDir+ir.withInfraFile, []byte(formatSrc), 0666)

	return formatSrc
}

func (ir *Infraer) copyInfraInfo() {
	oldContent := *ir.oldInfraInfo
	ir.newInfraInfo = &oldContent
}

//processInfraInfo process newInfraInfo, append import, repo field and wire's provider
func (ir *Infraer) processNewInfra() bool {

	for _, injectInfo := range ir.injectInfos {
		ir.newInfraInfo.structInfo.Fields = append(ir.newInfraInfo.structInfo.Fields, injectInfo.Fields...)

		ir.newInfraInfo.infraSetArgs.Args = append(ir.newInfraInfo.infraSetArgs.Args, injectInfo.InfraSetArgs...)

		ir.newInfraInfo.imports = append(ir.newInfraInfo.imports, injectInfo.Imports...)

		ir.newInfraInfo.provides = append(ir.newInfraInfo.provides, injectInfo.Provides...)
	}

	return true
}

func (ir *Infraer) toStringNewInfra() {

	ir.newInfraInfo.importStr = ir.newInfraInfo.imports.String()

	ir.newInfraInfo.structStr = ir.newInfraInfo.structInfo.String()

	ir.newInfraInfo.infraSetStr = ir.newInfraInfo.infraSetArgs.String()

	ir.newInfraInfo.provideStr = ir.newInfraInfo.provides.String()
}

func (ir *Infraer) buildNewInfraContent() {

	oldContent := ir.oldInfraInfo.content

	oldContent = strings.Replace(oldContent,
		ir.oldInfraInfo.importStr, ir.newInfraInfo.importStr, -1)

	oldContent = strings.Replace(oldContent,
		ir.oldInfraInfo.structStr, ir.newInfraInfo.structStr, -1)

	ir.newInfraInfo.content = strings.Replace(oldContent,
		ir.oldInfraInfo.infraSetStr, ir.newInfraInfo.infraSetStr, -1)

	ir.newInfraInfo.content += ir.newInfraInfo.provideStr
}

func (ir *Infraer) makeCodeBeautiful(src string) string {

	result, err := imports.Process("", []byte(src), nil)
	if err != nil {
		ir.logger.Panicf("err %s : %s", err.Error(), src)
		return ""
	}

	return string(result)
}

//writeNewInfra cover old infra.go's content
func (ir *Infraer) writeNewInfra() bool {

	processSrc := ir.makeCodeBeautiful(ir.newInfraInfo.content)

	ir.writer.Write(ir.withInfraDir+ir.withInfraFile, string(processSrc))

	err := ir.execer.ExecWire(ir.withInfraDir)
	if err != nil {
		ir.logger.Errorf(err.Error())
		return false
	}

	return true
}

func (ir *Infraer) parseInfraSetArgs(GenDecl *ast.GenDecl, srcStr string) []string {
	var args []string
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.ValueSpec); ok {
			for _, value := range spec.Values {
				if callExpr, ok := value.(*ast.CallExpr); ok {
					for _, callArg := range callExpr.Args {
						args = append(args, strings.Trim(pkg.ParseExpr(callArg, srcStr), ","))
					}
				}
			}
		}
	}

	return args
}
