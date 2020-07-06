package factory

import (
	"errors"
	"fmt"
	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/martinusso/inflect"
	"github.com/serenize/snaker"
	"github.com/spf13/viper"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	// "sort"
	"bytes"
	"github.com/dave/dst/decorator/resolver/gopackages"
	"github.com/davecgh/go-spew/spew"
	"sort"
	"strings"
)

var fset = token.NewFileSet()

type SortReturn struct {
	Fields pkg.Fields `json:"fields"`
}

type InitFieldsReturn struct {
	Fields     []string    `json:"fields"`
	SpecFields []pkg.Field `json:"SpecFields"`
}

type EsimFactory struct {
	// struct name which be search.
	StructName string

	// First letter uppercase of StructName.
	UpStructName string

	LowerStructName string

	ShortenStructName string

	// struct Absolute path
	structDir string

	filesName []string

	// package {{.packName}}
	packName string

	// package main => {{.packStr}}
	packStr string

	// File where the Struct is located
	structFileName string

	// Found Struct
	oldStructInfo *structInfo

	// Struct will be create
	NewStructInfo *structInfo

	// true if find the StructName
	// false if not
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

	// Struct plural form.
	pluralName string

	NewPluralStr string

	ReleasePluralStr string

	TypePluralStr string

	// option start.
	Option1 string

	Option2 string

	Option3 string

	Option4 string

	Option5 string

	// option end.

	OptionParam string

	logger log.Logger

	withSort bool

	withImpIface string

	withPool bool

	withStar bool

	WithNew bool

	writer filedir.IfaceWriter

	SpecFieldInitStr string

	ReturnStr string

	StructTpl *templates.StructInfo

	tpl templates.Tpl

	ot *optionTpl

	typeSpec *dst.TypeSpec

	// underlying type of a type.
	underType *types.Struct

	structPackage *decorator.Package

	dstFile *dst.File

	newStructField []*types.Var
}

type Option func(*EsimFactory)

func NewEsimFactory(options ...Option) *EsimFactory {
	factory := &EsimFactory{}

	for _, option := range options {
		option(factory)
	}

	if factory.logger == nil {
		factory.logger = log.NewLogger()
	}

	if factory.writer == nil {
		factory.writer = filedir.NewEsimWriter()
	}

	if factory.tpl == nil {
		factory.tpl = templates.NewTextTpl()
	}

	if factory.ot == nil {
		factory.ot = newOptionTpl(factory.tpl)
	}

	factory.oldStructInfo = &structInfo{
		vars: &pkg.Vars{},
	}

	factory.NewStructInfo = &structInfo{
		vars: &pkg.Vars{},
	}

	factory.structFieldIface = NewRPCPluginStructField(factory.writer, factory.logger)

	factory.StructTpl = templates.NewStructInfo(
		templates.WithTpl(factory.tpl),
		templates.WithLogger(factory.logger),
	)

	return factory
}

func WithEsimFactoryLogger(logger log.Logger) Option {
	return func(ef *EsimFactory) {
		ef.logger = logger
	}
}

func WithEsimFactoryWriter(writer filedir.IfaceWriter) Option {
	return func(ef *EsimFactory) {
		ef.writer = writer
	}
}

func WithEsimFactoryTpl(tpl templates.Tpl) Option {
	return func(ef *EsimFactory) {
		ef.tpl = tpl
	}
}

type structInfo struct {
	Fields pkg.Fields

	structStr string

	structFileContent string

	vars *pkg.Vars

	varStr string

	imports pkg.Imports

	importStr string

	ReturnVarStr string

	StructInitStr string
}

// getPluralWord Struct plural form.
// If plural is not obtained, add "s" at the end of the word.
func (ef *EsimFactory) getPluralForm(word string) string {
	newWord := inflect.Pluralize(word)
	if newWord == word || newWord == "" {
		newWord = word + "s"
	}

	return newWord
}

func (ef *EsimFactory) Run(v *viper.Viper) error {
	defer func() {
		ef.Close()
	}()

	err := ef.bindInput(v)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	ps := ef.loadPackages()
	if !ef.findStruct(ps) {
		ef.logger.Panicf("Not found struct %s", ef.StructName)
	}

	//if !ef.parseStruct() {
	//	ef.logger.Panicf("not found struct %s", ef.StructName)
	//}

	if ef.withSort {
		ef.sortField()
	}

	if ef.withPool {
		decl := ef.constructVarPool()
		ef.dstFile.Decls = append(ef.dstFile.Decls, decl)
	}

	if ef.withOption {
		decl := ef.constructOptionTypeFunc()
		ef.dstFile.Decls = append(ef.dstFile.Decls, decl)
	}

	if ef.WithNew {
		decl := ef.constructNew()
		ef.dstFile.Decls = append(ef.dstFile.Decls, decl)
	}

	if ef.withPool {
		decl := ef.constructRelease()
		ef.dstFile.Decls = append(ef.dstFile.Decls, decl)
	}

	ef.extendFields(ps)

	//ef.InitField = &InitFieldsReturn{}
	//ef.structFieldIface.HandleField(ef.NewStructInfo.Fields, ef.InitField)

	//
	//ef.genStr()
	//
	//ef.assignStructTpl()
	//
	//ef.executeNewTmpl()
	//
	//ef.organizePart()

	if ef.withPrint {
		ef.printResult()
	} else {
		err = filedir.EsimBackUpFile(ef.structDir +
			string(filepath.Separator) + ef.structFileName)
		if err != nil {
			ef.logger.Warnf("backup err %s:%s", ef.structDir+
				string(filepath.Separator)+ef.structFileName,
				err.Error())
		}

		originContent := ef.replaceOriginContent()
		res, err := imports.Process("", []byte(originContent), nil)
		if err != nil {
			ef.logger.Panicf("%s : %s", err.Error(), originContent)
		}

		err = filedir.EsimWrite(ef.structDir+
			string(filepath.Separator)+ef.structFileName,
			string(res))
		if err != nil {
			ef.logger.Panicf(err.Error())
		}
	}

	return nil
}

func (ef *EsimFactory) loadPackages() []*decorator.Package {
	var conf loader.Config
	conf.TypeChecker.Sizes = types.SizesFor(build.Default.Compiler, build.Default.GOARCH)
	conf.Fset = fset

	pConfig := &packages.Config{}
	pConfig.Fset = fset
	pConfig.Mode = packages.LoadAllSyntax
	pConfig.Dir = ef.structDir
	ps, err := decorator.Load(pConfig)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	return ps
}

// check exists and extra
func (ef *EsimFactory) findStruct(ps []*decorator.Package) bool {
	for _, p := range ps {
		for _, syntax := range p.Syntax {
			for _, decl := range syntax.Decls {
				if genDecl, ok := decl.(*dst.GenDecl); ok {
					if genDecl.Tok == token.TYPE {
						for _, spec := range genDecl.Specs {
							if typeSpec, ok := spec.(*dst.TypeSpec); ok {
								for _, def := range p.TypesInfo.Defs {
									if def == nil {
										continue
									}

									if _, ok := def.(*types.TypeName); !ok {
										continue
									}

									typ, ok := def.Type().(*types.Named)
									if !ok {
										continue
									}

									underType, ok := typ.Underlying().(*types.Struct)
									if !ok {
										continue
									}

									if typeSpec.Name.String() == def.Name() &&
										typeSpec.Name.String() == ef.StructName {
										ef.found = true
										ef.typeSpec = typeSpec
										ef.underType = underType
										ef.structPackage = p
										ef.dstFile = syntax
										// return true
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return false
}

func (ef *EsimFactory) constructOptionTypeFunc() *dst.GenDecl {
	if !ef.withOption {
		return &dst.GenDecl{}
	}

	genDecl := &dst.GenDecl{
		Tok: token.TYPE,
		Specs: []dst.Spec{
			&dst.TypeSpec{
				Name: dst.NewIdent(ef.StructName + "Option"),
				Type: &dst.FuncType{
					Func: true,
					Params: &dst.FieldList{
						List: []*dst.Field{
							&dst.Field{
								Type: &dst.StarExpr{
									X: dst.NewIdent(ef.StructName),
								},
							},
						},
					},
				},
			},
		},
		Decs: dst.GenDeclDecorations{
			NodeDecs: dst.NodeDecs{
				Before: dst.EmptyLine,
			},
		},
	}

	return genDecl
}

func (ef *EsimFactory) constructNew() *dst.FuncDecl {
	params := &dst.FieldList{}
	if ef.withOption {
		params = &dst.FieldList{
			Opening: true,
			List: []*dst.Field{
				&dst.Field{
					Names: []*dst.Ident{dst.NewIdent("options")},
					Type: &dst.Ellipsis{
						Elt: dst.NewIdent(ef.UpStructName + "Option"),
					},
				},
			},
		}
	}

	stmts := make([]dst.Stmt, 0)
	stmts = append(stmts, ef.getStructInstan())
	stmts = append(stmts, ef.getOptionBody())

	stmts = append(stmts, ef.getSpecialFieldStmt()...)
	ef.getReleaseFieldStmt()
	stmts = append(stmts,
		&dst.ReturnStmt{
			Results: []dst.Expr{
				dst.NewIdent(ef.ShortenStructName),
			},
			Decs: dst.ReturnStmtDecorations{
				NodeDecs: dst.NodeDecs{
					Before: dst.EmptyLine,
				},
			},
		})

	funcDecl := &dst.FuncDecl{
		Name: dst.NewIdent("New" + ef.UpStructName),
		Type: &dst.FuncType{
			Func:    true,
			Params:  params,
			Results: ef.getNewFuncTypeReturn(),
		},
		Body: &dst.BlockStmt{
			List: stmts,
		},
	}

	return funcDecl
}

func (ef *EsimFactory) constructRelease() *dst.FuncDecl {
	funcDecl := &dst.FuncDecl{
		Recv: &dst.FieldList{
			Opening: true,
			List: []*dst.Field{
				&dst.Field{
					Names: []*dst.Ident{
						dst.NewIdent("t"),
					},
					Type: &dst.StarExpr{
						X: dst.NewIdent("Test"),
					},
				},
			},
		},
		Name: dst.NewIdent("Reset"),
		Type: &dst.FuncType{
			Func: true,
			Params: &dst.FieldList{
				Opening: true,
			},
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{

				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun: &dst.SelectorExpr{
							X:   dst.NewIdent("testPool"),
							Sel: dst.NewIdent("Put"),
						},
						Args: []dst.Expr{
							dst.NewIdent("t"),
						},
					},
				},
			},
		},
	}

	_ = funcDecl

	return funcDecl
}

// t := Test{}
func (ef *EsimFactory) getStructInstan() dst.Stmt {
	if ef.withPool {
		return &dst.AssignStmt{
			Lhs: []dst.Expr{
				dst.NewIdent(ef.ShortenStructName),
			},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.TypeAssertExpr{
					X: &dst.CallExpr{
						Fun: &dst.SelectorExpr{
							X:   dst.NewIdent(ef.LowerStructName + "Pool"),
							Sel: dst.NewIdent("Get"),
						},
					},
					Type: &dst.StarExpr{
						X: dst.NewIdent(ef.UpStructName),
					},
				},
			},
		}
	} else if ef.withStar {
		return &dst.AssignStmt{
			Lhs: []dst.Expr{
				dst.NewIdent(ef.ShortenStructName),
			},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.UnaryExpr{
					Op: token.AND,
					X: &dst.CompositeLit{
						Type: dst.NewIdent(ef.UpStructName),
					},
				},
			},
		}
	} else {
		return &dst.AssignStmt{
			Lhs: []dst.Expr{
				dst.NewIdent(ef.ShortenStructName),
			},
			Tok: token.DEFINE,
			Rhs: []dst.Expr{
				&dst.CompositeLit{
					Type: dst.NewIdent(ef.UpStructName),
				},
			},
		}
	}
}

func (ef *EsimFactory) getOptionBody() dst.Stmt {
	return &dst.RangeStmt{
		Key:   dst.NewIdent("_"),
		Value: dst.NewIdent("option"),
		Tok:   token.DEFINE,
		X:     dst.NewIdent("options"),
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.ExprStmt{
					X: &dst.CallExpr{
						Fun: dst.NewIdent("option"),
						Args: []dst.Expr{
							dst.NewIdent(ef.ShortenStructName),
						},
					},
				},
			},
		},
		Decs: dst.RangeStmtDecorations{
			NodeDecs: dst.NodeDecs{
				Before: dst.EmptyLine,
			},
		},
	}
}

func (ef *EsimFactory) getSpecialFieldStmt() []dst.Stmt {
	stmts := make([]dst.Stmt, 0)
	numField := ef.underType.NumFields()
	for k, field := range ef.typeSpec.Type.(*dst.StructType).Fields.List {
		field.Decs.After = dst.EmptyLine
		if k < numField {
			switch field.Type.(type) {
			case *dst.ArrayType:
				if len(field.Names) != 0 {
					cloned := dst.Clone(field.Type.(*dst.ArrayType)).(*dst.ArrayType)
					stmts = append(stmts, ef.constructSpecialFieldStmt(ef.ShortenStructName,
						field.Names[0].String(), cloned))
				}
			case *dst.MapType:
				if len(field.Names) != 0 {
					cloned := dst.Clone(field.Type.(*dst.MapType)).(*dst.MapType)
					stmts = append(stmts, ef.constructSpecialFieldStmt(ef.ShortenStructName,
						field.Names[0].String(), cloned))
				}
			}
		}
	}

	return stmts
}

func (ef *EsimFactory) getReleaseFieldStmt() []dst.Stmt {
	numField := ef.underType.NumFields()
	for k, field := range ef.typeSpec.Type.(*dst.StructType).Fields.List {
		field.Decs.After = dst.EmptyLine
		if k < numField {
			spew.Dump(ef.underType.Field(k).Type().String())
		}

		//switch field.Type.(type) {
		//case *:
		//
		//case *dst.ArrayType:
		//
		//case *dst.MapType:
		//
		//}
	}

	return nil
}

func (ef *EsimFactory) getNewFuncTypeReturn() *dst.FieldList {
	fieldList := &dst.FieldList{}
	if ef.withImpIface != "" {
		fieldList.List = []*dst.Field{
			&dst.Field{
				Type: dst.NewIdent(ef.withImpIface),
			},
		}
	} else if ef.withPool || ef.withStar {
		fieldList.List = []*dst.Field{
			&dst.Field{
				Type: &dst.StarExpr{
					X: dst.NewIdent(ef.StructName),
				},
			},
		}
	} else {
		fieldList.List = []*dst.Field{
			&dst.Field{
				Type: dst.NewIdent(ef.StructName),
			},
		}
	}

	return fieldList
}

func (ef *EsimFactory) constructVarPool() *dst.GenDecl {
	valueSpec := &dst.ValueSpec{
		Names: []*dst.Ident{
			&dst.Ident{
				Name: ef.LowerStructName + "Pool",
			},
		},
		Values: []dst.Expr{
			&dst.CompositeLit{
				Type: &dst.SelectorExpr{
					X:   &dst.Ident{Name: "sync"},
					Sel: &dst.Ident{Name: "Pool"},
					Decs: dst.SelectorExprDecorations{
						NodeDecs: dst.NodeDecs{
							End: dst.Decorations{"\n"},
						},
					},
				},
				Elts: []dst.Expr{
					&dst.KeyValueExpr{
						Key: &dst.Ident{Name: "New"},
						Value: &dst.FuncLit{
							Type: &dst.FuncType{
								Func: true,
								Params: &dst.FieldList{
									Opening: true,
									Closing: true,
								},
								Results: &dst.FieldList{
									List: []*dst.Field{
										&dst.Field{
											Type: &dst.InterfaceType{
												Methods: &dst.FieldList{
													Opening: true,
													Closing: true,
												},
											},
										},
									},
								},
							},
							Body: &dst.BlockStmt{
								List: []dst.Stmt{
									&dst.ReturnStmt{
										Results: []dst.Expr{
											&dst.UnaryExpr{
												Op: token.AND,
												X: &dst.CompositeLit{
													Type: &dst.Ident{Name: ef.StructName},
												},
											},
										},
										Decs: dst.ReturnStmtDecorations{
											NodeDecs: dst.NodeDecs{
												Before: dst.NewLine,
											},
										},
									},
								},
							},
						},
						Decs: dst.KeyValueExprDecorations{
							NodeDecs: dst.NodeDecs{
								Before: dst.NewLine,
							},
						},
					},
				},
			},
		},
	}

	genDecl := &dst.GenDecl{
		Tok:    token.VAR,
		Lparen: true,
		Specs: []dst.Spec{
			valueSpec,
		},
	}

	return genDecl
}

func (ef *EsimFactory) constructSpecialFieldStmt(structName, field string, expr dst.Expr) dst.Stmt {
	stmt := &dst.IfStmt{
		Cond: &dst.BinaryExpr{
			X: &dst.SelectorExpr{
				X:   dst.NewIdent(structName),
				Sel: dst.NewIdent(field),
			},
			Op: token.EQL,
			Y:  dst.NewIdent("nil"),
		},
		Body: &dst.BlockStmt{
			List: []dst.Stmt{
				&dst.AssignStmt{
					Lhs: []dst.Expr{
						&dst.SelectorExpr{
							X:   dst.NewIdent(structName),
							Sel: dst.NewIdent(field),
						},
					},
					Tok: token.ASSIGN,
					Rhs: []dst.Expr{
						&dst.CallExpr{
							Fun: dst.NewIdent("make"),
							Args: []dst.Expr{
								expr,
								&dst.BasicLit{
									Kind:  token.INT,
									Value: "0",
								},
							},
						},
					},
				},
			},
		},
		Decs: dst.IfStmtDecorations{
			NodeDecs: dst.NodeDecs{
				Before: dst.EmptyLine,
			},
		},
	}

	return stmt
}

func (ef *EsimFactory) newContext() string {
	r := decorator.NewRestorerWithImports("root", gopackages.New(ef.structDir))

	var buf bytes.Buffer
	if err := r.Fprint(&buf, ef.dstFile); err != nil {
		panic(err)
	}

	return buf.String()
}

func (ef *EsimFactory) assignStructTpl() {
	ef.StructTpl.StructName = ef.StructName
	ef.StructTpl.Fields = ef.NewStructInfo.Fields
}

func (ef *EsimFactory) executeNewTmpl() {
	content, err := ef.tpl.Execute("factory", newTemplate, ef)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	ef.NewStructInfo.structStr = content
}

// replaceOriginContent gen a new struct file content.
func (ef *EsimFactory) replaceOriginContent() string {
	var newContent string
	originContent := ef.oldStructInfo.structFileContent
	newContent = originContent
	if ef.oldStructInfo.importStr != "" {
		newContent = strings.Replace(newContent, ef.oldStructInfo.importStr, "", 1)
	}

	newContent = strings.Replace(newContent, ef.packStr, ef.firstPart, 1)

	if ef.secondPart != "" {
		newContent = strings.Replace(newContent, ef.oldStructInfo.varStr, ef.secondPart, 1)
	}

	newContent = strings.Replace(newContent, ef.oldStructInfo.structStr, ef.thirdPart, 1)
	ef.NewStructInfo.structFileContent = newContent

	return newContent
}

// printResult println file content to terminal.
func (ef *EsimFactory) printResult() {
	src := ef.firstPart + "\n"
	src += ef.secondPart + "\n"
	src += ef.thirdPart + "\n"

	res, err := imports.Process("", []byte(src), nil)
	if err != nil {
		ef.logger.Panicf(err.Error())
	} else {
		println(string(res))
	}
}

// organizePart  organize pack, import, var, struct.
func (ef *EsimFactory) organizePart() {
	ef.firstPart = ef.packStr + "\n"
	ef.firstPart += ef.NewStructInfo.importStr + "\n"

	if ef.oldStructInfo.varStr != "" {
		ef.secondPart = ef.NewStructInfo.varStr
	} else {
		// merge firstPart and secondPart
		ef.firstPart += ef.NewStructInfo.varStr
	}

	ef.thirdPart = ef.NewStructInfo.structStr
}

// copy oldStructInfo to NewStructInfo.
func (ef *EsimFactory) copyOldStructInfo() {
	copyStructInfo := *ef.oldStructInfo
	ef.NewStructInfo = &copyStructInfo
}

func (ef *EsimFactory) bindInput(v *viper.Viper) error {
	sname := v.GetString("sname")
	if sname == "" {
		return errors.New("sname is empty")
	}
	ef.StructName = sname
	ef.UpStructName = templates.FirstToUpper(sname)
	ef.ShortenStructName = templates.Shorten(sname)
	ef.LowerStructName = strings.ToLower(sname)

	sdir := v.GetString("sdir")
	if sdir == "" {
		sdir = "."
	}

	dir, err := filepath.Abs(sdir)
	if err != nil {
		return err
	}
	ef.structDir = strings.TrimRight(dir, "/")

	plural := v.GetBool("plural")
	if plural {
		ef.withPlural = true
		ef.pluralName = ef.getPluralForm(sname)
	}

	ef.withOption = v.GetBool("option")

	ef.withGenConfOption = v.GetBool("oc")

	ef.withGenLoggerOption = v.GetBool("ol")

	ef.withSort = v.GetBool("sort")

	ef.withImpIface = v.GetString("imp_iface")

	ef.withPool = v.GetBool("pool")

	ef.withStar = v.GetBool("star")

	ef.withPrint = v.GetBool("print")

	ef.WithNew = v.GetBool("new")

	return nil
}

func (ef *EsimFactory) genStr() {
	ef.genReturnVarStr()
	ef.genStructInitStr()
	ef.genSpecFieldInitStr()
	ef.genReturnStr()

	if ef.withOption {
		ef.genOptionParam()
		ef.genOptions()
	}

	if ef.withPool && len(ef.InitField.Fields) > 0 {
		ef.genPool()
	}

	if ef.withPlural {
		ef.genPlural()
	}

	ef.NewStructInfo.varStr = ef.NewStructInfo.vars.String()
}

// Find struct.
// Parse struct.
func (ef *EsimFactory) parseStruct() bool {
	exists, err := filedir.IsExistsDir(ef.structDir)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	if !exists {
		ef.logger.Panicf("%s dir not exists", ef.structDir)
	}

	files, err := ioutil.ReadDir(ef.structDir)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	for _, fileInfo := range files {
		ext := path.Ext(fileInfo.Name())
		if ext != ".go" {
			continue
		}

		// 不复制测试文件
		if strings.Contains(fileInfo.Name(), "_test") {
			continue
		}

		ef.filesName = append(ef.filesName, fileInfo.Name())

		if !fileInfo.IsDir() && !ef.found {
			src, err := ioutil.ReadFile(ef.structDir + "/" + fileInfo.Name())
			if err != nil {
				ef.logger.Panicf(err.Error())
			}

			strSrc := string(src)
			fset := token.NewFileSet() // positions are relative to fset
			f, err := parser.ParseFile(fset, "", strSrc, parser.ParseComments)
			if err != nil {
				ef.logger.Panicf(err.Error())
			}

			ef.parseDecls(src, fileInfo, f)
		}
	}

	return ef.found
}

func (ef *EsimFactory) parseDecls(src []byte, fileInfo os.FileInfo, f *ast.File) {
	strSrc := string(src)

	// Must find the structName first.
	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Tok.String() == "type" {
				ef.parseType(genDecl, src, fileInfo, f)
			}
		}
	}

	for _, decl := range f.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok {
			if genDecl.Tok.String() == "var" && ef.found {
				ef.oldStructInfo.vars.ParseFromAst(genDecl, strSrc)
				ef.oldStructInfo.varStr = ef.oldStructInfo.vars.String()
			}

			if genDecl.Tok.String() == "import" && ef.found {
				ef.parseImport(genDecl, src)
			}
		}
	}
}

func (ef *EsimFactory) parseType(genDecl *ast.GenDecl, src []byte,
	fileInfo os.FileInfo, f *ast.File) {
	strSrc := string(src)
	for _, specs := range genDecl.Specs {
		if typeSpec, ok := specs.(*ast.TypeSpec); ok {
			if typeSpec.Name.String() == ef.StructName {
				ef.oldStructInfo.structFileContent = strSrc
				ef.structFileName = fileInfo.Name()
				// found the struct
				ef.found = true

				ef.packName = f.Name.String()
				ef.packStr = "package " + strSrc[f.Name.Pos()-1:f.Name.End()]

				fields := pkg.Fields{}
				fields.ParseFromAst(genDecl, strSrc)
				ef.oldStructInfo.Fields = fields
				colsePos := typeSpec.Type.(*ast.StructType).Fields.Closing
				structStr := src[genDecl.TokPos-1 : colsePos]
				ef.oldStructInfo.structStr = string(structStr)
			}
		}
	}
}

func (ef *EsimFactory) parseImport(genDecl *ast.GenDecl, src []byte) {
	strSrc := string(src)
	imps := pkg.Imports{}
	imps.ParseFromAst(genDecl)
	ef.oldStructInfo.imports = imps
	ef.oldStructInfo.importStr = strSrc[genDecl.Pos()-1 : genDecl.End()]
}

// Extend logger and conf for new struct field.
// It must after constructs.
func (ef *EsimFactory) extendFields(ps []*decorator.Package) {
	if ef.withOption {
		if ef.withGenLoggerOption {
			ef.extendField(newLoggerFieldInfo())
		}

		if ef.withGenConfOption {
			ef.extendField(newConfigFieldInfo())
		}
	}
}

type FieldSizes []FieldSize

type FieldSize struct {
	Size int64

	Field *dst.Field

	Vars *types.Var
}

func (fs FieldSizes) Len() int { return len(fs) }

func (fs FieldSizes) Less(i, j int) bool {
	return fs[i].Size < fs[j].Size
}

func (fs FieldSizes) Swap(i, j int) { fs[i], fs[j] = fs[j], fs[i] }

func (fs FieldSizes) getFields() []*dst.Field {
	dstFields := make([]*dst.Field, 0)
	for _, f := range fs {
		dstFields = append(dstFields, f.Field)
	}

	return dstFields
}

// sortField ascending in byte size.
func (ef *EsimFactory) sortField() {
	fs := make(FieldSizes, 0)
	var size int64

	for _, field := range ef.typeSpec.Type.(*dst.StructType).Fields.List {
		field.Decs.After = dst.EmptyLine
		fs = append(fs, FieldSize{
			Size:  size,
			Field: field,
		},
		)
	}
	sort.Sort(fs)
	ef.typeSpec.Type.(*dst.StructType).Fields.List = fs.getFields()
}

func (ef *EsimFactory) extendField(fieldInfo extendFieldInfo) {
	fields := ef.typeSpec.Type.(*dst.StructType).Fields.List
	var fieldExists bool
	for _, field := range fields {
		if len(field.Names) != 0 {
			if field.Names[0].Name == fieldInfo.name {
				fieldExists = true
			}
		}
	}

	if fieldExists == false {
		fields = append(fields,
			&dst.Field{
				Names: []*dst.Ident{
					dst.NewIdent(fieldInfo.name),
				},
				Type: &dst.Ident{Name: fieldInfo.typeName, Path: fieldInfo.typePath},
				Decs: dst.FieldDecorations{
					NodeDecs: dst.NodeDecs{
						Before: dst.EmptyLine,
					},
				},
			},
		)
		ef.typeSpec.Type.(*dst.StructType).Fields.List = fields
	}

}

// If struct field had extend logger or conf
// so build a new struct.
func (ef *EsimFactory) buildNewStructFileContent() error {
	ef.NewStructInfo.importStr = ef.NewStructInfo.imports.String()

	if ef.oldStructInfo.importStr != "" {
		ef.NewStructInfo.structFileContent = strings.Replace(ef.oldStructInfo.structFileContent,
			ef.oldStructInfo.importStr, ef.NewStructInfo.importStr, -1)
	} else if ef.packStr != "" {
		// not found import
		ef.getFirstPart()
		ef.NewStructInfo.structFileContent = strings.Replace(ef.oldStructInfo.structFileContent,
			ef.packStr, ef.firstPart, -1)
	} else {
		return errors.New("can't build the first part")
	}

	structInfo := templates.NewStructInfo()
	structInfo.StructName = ef.StructName
	structInfo.Fields = ef.NewStructInfo.Fields

	ef.NewStructInfo.structStr = structInfo.String()
	ef.NewStructInfo.structFileContent = strings.Replace(ef.NewStructInfo.structFileContent,
		ef.oldStructInfo.structStr, ef.NewStructInfo.structStr, -1)

	src, err := imports.Process("", []byte(ef.NewStructInfo.structFileContent), nil)
	if err != nil {
		return fmt.Errorf("%s : %s", err.Error(), ef.NewStructInfo.structFileContent)
	}

	ef.NewStructInfo.structFileContent = string(src)

	return nil
}

//
// func NewStruct() {{.ReturnvarStr}} {
// }
//
func (ef *EsimFactory) genReturnVarStr() {
	if ef.withImpIface != "" {
		ef.NewStructInfo.ReturnVarStr = ef.withImpIface
	} else if ef.withPool || ef.withStar {
		ef.NewStructInfo.ReturnVarStr = "*" + ef.StructName
	} else {
		ef.NewStructInfo.ReturnVarStr = ef.StructName
	}
}

// type Option func(c *{{.OptionParam}}).
func (ef *EsimFactory) genOptionParam() {
	if ef.withPool || ef.withStar {
		ef.OptionParam = "*" + ef.StructName
	} else {
		ef.OptionParam = ef.StructName
	}
}

// StructObj := Struct{} => {{.StructInitStr}}.
func (ef *EsimFactory) genStructInitStr() {
	var structInitStr string
	if ef.withStar {
		structInitStr = templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) +
			" := &" + ef.StructName + "{}"
	} else if ef.withPool {
		structInitStr = templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) + ` := ` +
			templates.FirstToLower(snaker.SnakeToCamelLower(ef.StructName)) + `Pool.Get().(*` +
			ef.StructName + `)`
	} else {
		structInitStr = templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) +
			" := " + ef.StructName + "{}"
	}

	ef.NewStructInfo.StructInitStr = structInitStr
}

// return {{.ReturnStr}}.
func (ef *EsimFactory) genReturnStr() {
	ef.ReturnStr = templates.Shorten(snaker.SnakeToCamelLower(ef.StructName))
}

func (ef *EsimFactory) genOptions() {
	ef.Option1 = `type ` + ef.StructName + `Option func(` + ef.OptionParam + `)`

	ef.Option2 = `options ...` + ef.StructName + `Option`

	ef.Option3 = `
	for _, option := range options {
		option(` + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) + `)
	}`

	if ef.withGenConfOption {
		ef.Option4 = ef.ot.String(ef.StructName, ef.NewStructInfo.ReturnVarStr,
			"conf_template", confOptionTemplate)
	}

	if ef.withGenLoggerOption {
		ef.Option5 = ef.ot.String(ef.StructName, ef.NewStructInfo.ReturnVarStr,
			"logger_template", loggerOptionTemplate)
	}
}

func (ef *EsimFactory) genPool() {
	ef.incrPoolVar(ef.StructName)

	ef.ReleaseStr = ef.genReleaseStructStr()
}

func (ef *EsimFactory) genPlural() {
	ef.incrPoolVar(ef.pluralName)

	plural := NewPlural()
	plural.StructName = ef.StructName
	plural.PluralName = ef.pluralName
	if ef.withStar {
		plural.Star = "*"
	}

	ef.TypePluralStr = plural.TypeString()

	ef.NewPluralStr = plural.NewString()

	ef.ReleasePluralStr = plural.ReleaseString()
}

func (ef *EsimFactory) incrPoolVar(structName string) {
	poolName := structName + "Pool"
	if ef.varNameExists(*ef.NewStructInfo.vars, poolName) {
		ef.logger.Debugf("var is exists : %s", poolName)
	} else {
		*ef.NewStructInfo.vars = append(*ef.NewStructInfo.vars,
			ef.appendPoolVar(poolName, structName))
		ef.appendNewImport("sync")
	}
}

//nolint:goconst
func (ef *EsimFactory) genSpecFieldInitStr() {
	var str string

	for _, f := range ef.InitField.SpecFields {
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

	ef.SpecFieldInitStr = str
}

func (ef *EsimFactory) genReleaseStructStr() string {
	str := "func (" + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) +
		"  " + ef.NewStructInfo.ReturnVarStr + ") Release() {\n"

	for _, field := range ef.InitField.Fields {
		if strings.Contains(field, "time.Time") {
			ef.appendNewImport("time")
		}
		str += "		" + field + "\n"
	}

	str += "		" + templates.FirstToLower(snaker.SnakeToCamelLower(ef.StructName)) +
		"Pool.Put(" + templates.Shorten(snaker.SnakeToCamelLower(ef.StructName)) + ")\n"
	str += "}"

	return str
}

func (ef *EsimFactory) appendNewImport(importName string) {
	var found bool
	for _, imp := range ef.NewStructInfo.imports {
		if imp.Path == importName {
			found = true
		}
	}

	if !found {
		ef.NewStructInfo.imports = append(ef.NewStructInfo.imports, pkg.Import{Path: importName})
	}
}

func (ef *EsimFactory) appendPoolVar(pollVarName, structName string) pkg.Var {
	var poolVar pkg.Var

	pooltpl := NewPoolTpl()
	pooltpl.VarPoolName = pollVarName
	pooltpl.StructName = structName

	poolVar.Body = pooltpl.String()

	poolVar.Name = append(poolVar.Name, pollVarName)
	return poolVar
}

func (ef *EsimFactory) varNameExists(vars pkg.Vars, poolVarName string) bool {
	for _, varInfo := range vars {
		for _, varName := range varInfo.Name {
			if varName == poolVarName {
				return true
			}
		}
	}

	return false
}

// package + import.
func (ef *EsimFactory) getFirstPart() {
	if ef.oldStructInfo.importStr == "" {
		ef.firstPart += ef.packStr + "\n\n"
	}

	ef.firstPart += ef.NewStructInfo.importStr
}

func (ef *EsimFactory) Close() {
	ef.structFieldIface.Close()
}
