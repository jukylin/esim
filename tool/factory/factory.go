package factory

import (
	"bytes"
	"errors"
	"go/build"
	"go/token"
	"go/types"
	"path/filepath"
	"sort"
	"strings"

	"github.com/dave/dst"
	"github.com/dave/dst/decorator"
	"github.com/dave/dst/decorator/resolver/gopackages"
	"github.com/davecgh/go-spew/spew"
	"github.com/jukylin/esim/log"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/martinusso/inflect"
	"github.com/spf13/viper"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

var fset = token.NewFileSet()

type EsimFactory struct {
	// struct name which be search.
	StructName string

	// First letter uppercase of StructName.
	UpStructName string

	LowerStructName string

	ShortenStructName string

	// struct Absolute path
	structDir string

	// File where the Struct is located
	structFileName string

	// true if find the StructName
	// false if not
	found bool

	withPlural bool

	withOption bool

	withGenLoggerOption bool

	withGenConfOption bool

	withPrint bool

	logger log.Logger

	withSort bool

	withImpIface string

	withPool bool

	withStar bool

	WithNew bool

	writer filedir.IfaceWriter

	SpecFieldInitStr string

	typeSpec *dst.TypeSpec

	// underlying type of a type.
	underType *types.Struct

	structPackage *decorator.Package

	dstFile *dst.File
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

// getPluralWord Struct plural form.
// If plural is not obtained, add "s" at the end of the word.
//nolint:unused
func (ef *EsimFactory) getPluralForm(word string) string {
	newWord := inflect.Pluralize(word)
	if newWord == word || newWord == "" {
		newWord = word + "s"
	}

	return newWord
}

func (ef *EsimFactory) Run(v *viper.Viper) error {
	err := ef.bindInput(v)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	ps := ef.loadPackages()
	if !ef.findStruct(ps) {
		ef.logger.Panicf("Not found struct %s", ef.StructName)
	}

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

	ef.extendFields()

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

		originContent := ""
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
	pConfig.Mode = packages.NeedSyntax | packages.NeedName | packages.NeedFiles |
		packages.NeedCompiledGoFiles | packages.NeedTypesInfo | packages.NeedDeps |
		packages.NeedTypes | packages.NeedTypesSizes
	pConfig.Dir = ef.structDir
	ps, err := decorator.Load(pConfig)
	if err != nil {
		ef.logger.Panicf(err.Error())
	}

	return ps
}

// check exists and extra.
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

									if typeSpec.Name.String() == ef.StructName {
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
							{
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
				{
					Names: []*dst.Ident{dst.NewIdent("options")},
					Type: &dst.Ellipsis{
						Elt: dst.NewIdent(ef.UpStructName + "Option"),
					},
				},
			},
		}
	}

	stmts := make([]dst.Stmt, 0)
	stmts = append(stmts, ef.getStructInstan(), ef.getOptionBody())
	stmts = append(stmts, ef.getSpecialFieldStmt()...)

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
				{
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

// t := Test{}.
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
	for _, field := range ef.typeSpec.Type.(*dst.StructType).Fields.List {
		field.Decs.After = dst.EmptyLine
		switch _typ := field.Type.(type) {
		case *dst.ArrayType:
			if len(field.Names) != 0 {
				cloned := dst.Clone(_typ).(*dst.ArrayType)
				stmts = append(stmts, ef.constructSpecialFieldStmt(ef.ShortenStructName,
					field.Names[0].String(), cloned))
			}
		case *dst.MapType:
			if len(field.Names) != 0 {
				cloned := dst.Clone(_typ).(*dst.MapType)
				stmts = append(stmts, ef.constructSpecialFieldStmt(ef.ShortenStructName,
					field.Names[0].String(), cloned))
			}
		}
	}

	return stmts
}

//nolint:unparam,unused
func (ef *EsimFactory) getReleaseFieldStmt() []dst.Stmt {
	// numField := ef.underType.NumFields()
	for _, field := range ef.typeSpec.Type.(*dst.StructType).Fields.List {
		field.Decs.After = dst.EmptyLine
		// spew.Dump(ef.underType.Field(k).Type().String())

		switch _t := field.Type.(type) {
		case *dst.ArrayType:
		case *dst.MapType:
		default:
			spew.Dump(_t)
		}
	}

	return nil
}

func (ef *EsimFactory) TypeToInit(ident *dst.Ident) string {
	var initStr string

	switch ident.Name {
	case "string":
		initStr = "\"\""
	case "int", "int64", "int8", "int16", "int32":
		initStr = "0"
	case "uint", "uint64", "uint8", "uint16", "uint32":
		initStr = "0"
	case "bool":
		initStr = "false"
	case "float32", "float64":
		initStr = "0.00"
	case "complex64", "complex128":
		initStr = "0+0i"
		// case reflect.Interface:
		//	initStr = "nil"
		// case reflect.Uintptr:
		//	initStr = "0"
		// case reflect.Invalid, reflect.Func, reflect.Chan, reflect.Ptr, reflect.UnsafePointer:
		//	initStr = "nil"
		// case reflect.Slice:
		//	initStr = "nil"
		// case reflect.Map:
		//	initStr = "nil"
		// case reflect.Array:
		//	initStr = "nil"
	}

	return initStr
}

func (ef *EsimFactory) getNewFuncTypeReturn() *dst.FieldList {
	fieldList := &dst.FieldList{}
	if ef.withImpIface != "" {
		fieldList.List = []*dst.Field{
			{
				Type: dst.NewIdent(ef.withImpIface),
			},
		}
	} else if ef.withPool || ef.withStar {
		fieldList.List = []*dst.Field{
			{
				Type: &dst.StarExpr{
					X: dst.NewIdent(ef.StructName),
				},
			},
		}
	} else {
		fieldList.List = []*dst.Field{
			{
				Type: dst.NewIdent(ef.StructName),
			},
		}
	}

	return fieldList
}

func (ef *EsimFactory) constructVarPool() *dst.GenDecl {
	valueSpec := &dst.ValueSpec{
		Names: []*dst.Ident{
			{
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
										{
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

// printResult println file content to terminal.
func (ef *EsimFactory) printResult() {

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

// Extend logger and conf for new struct field.
// It must after constructs.
func (ef *EsimFactory) extendFields() {
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

	if !fieldExists {
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
