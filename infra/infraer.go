package infra

import (
	"go/ast"
	// "go/parser"
	"bytes"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/goast"
	"github.com/dave/dst/decorator/resolver/guess"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	domain_file "github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/spf13/viper"
	"go/token"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Infraer struct {
	logger log.Logger

	writer filedir.IfaceWriter

	execer pkg.Exec

	oldInfraInfo *Info

	newInfraInfo *Info

	injectInfos []*domain_file.InjectInfo

	hasInfraStruct bool

	withInfraDir string

	withInfraFile string

	newInfraContent string
}

type Info struct {
	imports pkg.Imports

	importStr string

	structInfo *templates.StructInfo

	structStr string

	specialStructName string

	specialVarName string

	infraSetArgs infraSetArgs

	infraSetStr string

	content string

	provides domain_file.Provides

	provideStr string
}

func NewInfo() *Info {
	ifaInfo := &Info{}

	ifaInfo.specialStructName = "Infra"

	ifaInfo.specialVarName = "infraSet"

	ifaInfo.infraSetArgs = infraSetArgs{}

	structInfo := templates.NewStructInfo()
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

	if infraer.injectInfos == nil {
		infraer.injectInfos = make([]*domain_file.InjectInfo, 0)
	}

	return infraer
}

func WithIfacerLogger(logger log.Logger) Option {
	return func(ir *Infraer) {
		ir.logger = logger
	}
}

func WithIfacerInfraInfo(infra *Info) Option {
	return func(ir *Infraer) {
		ir.oldInfraInfo = infra
		ir.newInfraInfo = infra
	}
}

func WithIfacerWriter(writer filedir.IfaceWriter) Option {
	return func(ir *Infraer) {
		ir.writer = writer
	}
}

func WithIfacerExecer(execer pkg.Exec) Option {
	return func(ir *Infraer) {
		ir.execer = execer
	}
}

// injectToInfra inject repo to infra.go and execute wire command.
func (ir *Infraer) Inject(v *viper.Viper, injectInfos []*domain_file.InjectInfo) bool {
	var err error

	if len(injectInfos) == 0 {
		ir.logger.Infof("Not need inject")
		return false
	}

	result := ir.bindInput(v)
	if !result {
		return false
	}

	ir.injectInfos = injectInfos

	// back up infra.go
	err = filedir.EsimBackUpFile(filedir.GetCurrentDir() + string(filepath.Separator) +
		ir.withInfraDir + ir.withInfraFile)
	if err != nil {
		ir.logger.Errorf("Back up err : %s", err.Error())
		return false
	}

	beautifulSource := ir.sourceInfraFile()
	if beautifulSource == "" {
		return false
	}

	parseResult := ir.constructNewInfra(beautifulSource)
	if !parseResult {
		return false
	}

	if ir.writeNewInfra() {
		ir.logger.Infof("inject success")
		return true
	}

	return false
}

func (ir *Infraer) bindInput(v *viper.Viper) bool {
	ir.withInfraDir = v.GetString("infra_dir")
	if ir.withInfraDir == "" {
		ir.withInfraDir = "internal" + string(filepath.Separator) + "infra" +
			string(filepath.Separator)
	} else {
		ir.withInfraDir = strings.TrimLeft(ir.withInfraDir, ".") + string(filepath.Separator)
		ir.withInfraDir = strings.Trim(ir.withInfraDir, "/") + string(filepath.Separator)
	}

	ir.withInfraFile = v.GetString("infra_file")
	if ir.withInfraFile == "" {
		ir.withInfraFile = "infra.go"
	}

	exists, err := filedir.IsExistsFile(ir.withInfraDir + ir.withInfraFile)
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

// constructNewInfra construct new infra use ast
func (ir *Infraer) constructNewInfra(srcStr string) bool {
	if len(ir.injectInfos) == 0 {
		return false
	}

	fset := token.NewFileSet()
	dec := decorator.NewDecoratorWithImports(fset, "infra", goast.New())
	f, err := dec.Parse(srcStr)
	if err != nil {
		if err != nil {
			ir.logger.Errorf(err.Error())
			return false
		}
	}

	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*dst.GenDecl); ok {
			if genDecl.Tok == token.TYPE {
				ir.constructNewType(genDecl)
			}
		}

		if genDecl, ok := decl.(*dst.GenDecl); ok {
			if genDecl.Tok == token.VAR {
				ir.constructNewVar(genDecl)
			}
		}
	}

	if !ir.hasInfraStruct {
		ir.logger.Errorf("not find %s", ir.oldInfraInfo.specialStructName)
		return false
	}

	funcDecls := ir.constructProvideFunc()
	if len(funcDecls) != 0 {
		for _, funcDecl := range funcDecls {
			f.Decls = append(f.Decls, funcDecl)
		}
	}

	var buf bytes.Buffer
	res := decorator.NewRestorerWithImports("infra", guess.New())
	if err := res.Fprint(&buf, f); err != nil {
		ir.logger.Errorf(err.Error())
		return false
	}
	ir.newInfraContent = buf.String()

	return true
}

// Construction function
func (ir *Infraer) constructProvideFunc() []*dst.FuncDecl {
	decls := make([]*dst.FuncDecl, 0)
	for _, injectInfo := range ir.injectInfos {
		for _, provideRepoFun := range injectInfo.ProvideRepoFuns {
			funcDecl := &dst.FuncDecl{
				Name: provideRepoFun.FuncName,
				Type: &dst.FuncType{
					Func: true,
					Params: &dst.FieldList{
						Opening: true,
						List: []*dst.Field{
							&dst.Field{
								Names: []*dst.Ident{provideRepoFun.ParamName},
								Type: &dst.StarExpr{
									X: provideRepoFun.ParamType,
								},
							},
						},
						Closing: true,
					},
					Results: &dst.FieldList{
						List: []*dst.Field{
							&dst.Field{
								Type: provideRepoFun.Result,
							},
						},
					},
				},
				Body: &dst.BlockStmt{
					List: []dst.Stmt{
						&dst.ReturnStmt{
							Results: []dst.Expr{
								&dst.CallExpr{
									Fun: provideRepoFun.BodyFunc,
									Args: []dst.Expr{
										provideRepoFun.BodyFuncArg,
									},
								},
							},
							Decs: dst.ReturnStmtDecorations{NodeDecs: dst.NodeDecs{After: dst.NewLine}},
						},
					},
				},
				Decs: dst.FuncDeclDecorations{NodeDecs: dst.NodeDecs{Before: dst.EmptyLine}},
			}
			decls = append(decls, funcDecl)
		}
	}

	return decls
}

func (ir *Infraer) constructNewType(genDecl *dst.GenDecl) {
	for _, spec := range genDecl.Specs {
		if typeSpec, ok := spec.(*dst.TypeSpec); ok {
			if typeSpec.Name.String() == ir.oldInfraInfo.specialStructName {
				ir.hasInfraStruct = true
				fields := typeSpec.Type.(*dst.StructType).Fields.List
				for _, injectInfo := range ir.injectInfos {
					for _, injectField := range injectInfo.Fields {
						fields = append(fields, &dst.Field{
							Names: []*dst.Ident{
								dst.NewIdent(injectField.Name),
							},
							Type: dst.NewIdent(injectField.Type),
							Decs: dst.FieldDecorations{NodeDecs: dst.NodeDecs{Before: dst.EmptyLine}},
						})
					}
				}
				typeSpec.Type.(*dst.StructType).Fields.List = fields
			}
		}
	}
}

func (ir *Infraer) constructNewVar(genDecl *dst.GenDecl) {
	for _, spec := range genDecl.Specs {
		if valueSpec, ok := spec.(*dst.ValueSpec); ok {
			if valueSpec.Names[0].String() == ir.oldInfraInfo.specialVarName {
				for _, specVals := range valueSpec.Values {
					if callExpr, ok := specVals.(*dst.CallExpr); ok {
						for _, injectInfo := range ir.injectInfos {
							for _, infraSet := range injectInfo.InfraSetArgs {
								callExpr.Args = append(callExpr.Args,
									&dst.Ident{
										Name: infraSet,
										Decs: dst.IdentDecorations{
											NodeDecs: dst.NodeDecs{Before: dst.NewLine}},
									})
							}
						}
					}
				}
			}
		}
	}
}

// sourceInfraFile Beautify infra.go.
func (ir *Infraer) sourceInfraFile() string {
	src, err := ioutil.ReadFile(ir.withInfraDir + ir.withInfraFile)
	if err != nil {
		ir.logger.Errorf(err.Error())
		return ""
	}

	formatSrc := ir.makeCodeBeautiful(string(src))

	err = ioutil.WriteFile(ir.withInfraDir+ir.withInfraFile, []byte(formatSrc), 0600)
	if err != nil {
		ir.logger.Errorf(err.Error())
		return ""
	}

	return formatSrc
}

func (ir *Infraer) makeCodeBeautiful(src string) string {
	options := &imports.Options{}
	options.Comments = false
	options.TabIndent = true
	options.TabWidth = 8
	options.FormatOnly = true

	result, err := imports.Process("", []byte(src), options)
	if err != nil {
		ir.logger.Panicf("err %s : %s", err.Error(), src)
	}

	return string(result)
}

// writeNewInfra cover old infra.go's content.
func (ir *Infraer) writeNewInfra() bool {
	processSrc := ir.makeCodeBeautiful(ir.newInfraContent)

	err := ir.writer.Write(ir.withInfraDir+ir.withInfraFile, processSrc)
	if err != nil {
		ir.logger.Errorf(err.Error())
		return false
	}

	err = ir.execer.ExecWire(ir.withInfraDir)
	if err != nil {
		ir.logger.Errorf(err.Error())
		return false
	}

	return true
}

func (ir *Infraer) parseInfraSetArgs(genDecl *ast.GenDecl, srcStr string) []string {
	var args []string
	for _, specs := range genDecl.Specs {
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
