package ifacer

import (
	"errors"
	"fmt"
	"go/types"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/mockery"
	"golang.org/x/tools/imports"
)

type Ifacer struct {
	logger log.Logger

	parser *mockery.Parser

	IfaceName string

	StructName string

	IfacePath string

	PackageName string

	Methods []Method

	Content string

	Star string

	OutFile string

	// import from interface file.
	pkgNoConflictImport map[string]pkg.Import

	// import in the interface.
	ifaceUsingIngImport map[string]string

	writer filedir.IfaceWriter

	tpl templates.Tpl

	UsingImportStr string
}

type Option func(*Ifacer)

func NewIfacer(options ...Option) *Ifacer {
	ifacer := &Ifacer{}

	for _, option := range options {
		option(ifacer)
	}

	ifacer.parser = mockery.NewParser([]string{})

	ifacer.pkgNoConflictImport = make(map[string]pkg.Import)

	ifacer.ifaceUsingIngImport = make(map[string]string)

	return ifacer
}

func WithIfacerLogger(logger log.Logger) Option {
	return func(f *Ifacer) {
		f.logger = logger
	}
}

func WithIfacerWriter(writer filedir.IfaceWriter) Option {
	return func(f *Ifacer) {
		f.writer = writer
	}
}

func WithIfacerTpl(tpl templates.Tpl) Option {
	return func(f *Ifacer) {
		f.tpl = tpl
	}
}

type Method struct {
	FuncName string

	ArgStr string

	ReturnTypeStr string

	ReturnStr string

	InitReturnVarStr string

	ReturnVar []string
}

func (f *Ifacer) Run(v *viper.Viper) error {
	err := f.bindInput(v)
	if err != nil {
		return err
	}

	err = f.parser.Parse(f.IfacePath)
	if err != nil {
		return err
	}

	err = f.parser.Load()
	if err != nil {
		return err
	}

	ifacer, err := f.parser.Find(f.IfaceName)
	if err != nil {
		return err
	}

	f.PackageName = ifacer.Pkg.Name()

	f.setNoConflictImport(ifacer.Pkg.Name(), ifacer.Pkg.Path())
	f.ManageNoConflictImport(ifacer.Pkg.Imports())

	f.GenMethods(ifacer.Type)

	f.getUsingImportStr()

	err = f.Process()
	if err != nil {
		return err
	}

	err = f.writer.Write(f.OutFile, f.Content)
	if err != nil {
		return err
	}

	return nil
}

func (f *Ifacer) bindInput(v *viper.Viper) error {
	name := v.GetString("iname")
	if name == "" {
		return errors.New("iname is empty")
	}
	f.IfaceName = name

	outPath := v.GetString("out")
	if outPath == "" {
		outPath = "./dummy_" + strings.ToLower(f.IfaceName) + ".go"
	}
	f.OutFile = outPath

	structName := v.GetString("stname")
	if structName == "" {
		structName = "Dummy" + f.IfaceName
	}
	f.StructName = structName

	star := v.GetBool("star")
	if star {
		f.Star = "*"
	}

	f.IfacePath = v.GetString("ipath")

	return nil
}

func (f *Ifacer) GenMethods(interacer *types.Interface) {
	for i := 0; i < interacer.NumMethods(); i++ {
		fn := interacer.Method(i)
		ftype := fn.Type().(*types.Signature)
		m := &Method{}
		m.ReturnStr = "return "

		m.FuncName = fn.Name()
		f.getArgStr(ftype.Params(), m, ftype.Variadic())
		f.getReturnStr(ftype.Results(), m)
		f.Methods = append(f.Methods, *m)
	}
}

func (f *Ifacer) getUsingImportStr() {
	imps := pkg.Imports{}
	for _, imp := range f.pkgNoConflictImport {
		imps = append(imps, imp)
	}

	f.UsingImportStr = imps.String()
}

func (f *Ifacer) ManageNoConflictImport(imps []*types.Package) bool {
	for _, imp := range imps {
		f.setNoConflictImport(imp.Name(), imp.Path())
	}

	return true
}

func (f *Ifacer) setNoConflictImport(importName, importPath string) {
	if impPath, ok := f.pkgNoConflictImport[importName]; ok {
		if impPath.Path == importPath {
			return
		}

		// package name repeat
		level := 1
		flag := true
		for flag {
			uniqueImportName := f.getUniqueImportName(importPath, level)
			if _, ok := f.pkgNoConflictImport[uniqueImportName]; !ok {
				imp := pkg.Import{}
				imp.Name = uniqueImportName
				imp.Path = importPath

				f.pkgNoConflictImport[uniqueImportName] = imp
				flag = false
			}
			level++
		}
	} else {
		imp := pkg.Import{}
		imp.Name = importName
		imp.Path = importPath
		f.pkgNoConflictImport[importName] = imp
	}
}

// Example:
// 	github.com/jukylin/esim/redis
//  level
//		0 redis
//		1 esimredis
//		2 jukylinesimredis
//		3 githubcomjukylinesimredis
func (f *Ifacer) getUniqueImportName(pkgName string, level int) string {
	strs := strings.Split(pkgName, string(filepath.Separator))

	f.logger.Debugf("pkgName %s", pkgName)

	var importName string

	lenStr := len(strs)
	importName = strs[lenStr-1] + strconv.Itoa(level)

	if strings.Contains(importName, ".") {
		importName = strings.Replace(importName, ".", "", -1)
	}

	if strings.Contains(importName, "-") {
		importName = strings.Replace(importName, "-", "", -1)
	}

	return importName
}

func (f *Ifacer) getArgStr(tuple *types.Tuple, m *Method, variadic bool) {
	for i := 0; i < tuple.Len(); i++ {
		ArgVar := tuple.At(i)
		if ArgVar.Name() != "" {
			m.ArgStr += ArgVar.Name() + " "
		} else {
			m.ArgStr += "arg" + strconv.Itoa(i) + " "
		}

		if i == tuple.Len()-1 {
			m.ArgStr += f.parseVar(ArgVar, variadic)
		} else {
			m.ArgStr += f.parseVar(ArgVar, false)
		}

		if i < tuple.Len()-1 {
			m.ArgStr += ", "
		}
	}
}

func (f *Ifacer) parseVar(varObj *types.Var, variadic bool) string {
	return f.parseVarType(varObj.Type(), variadic)
}

func (f *Ifacer) parseVarType(typ types.Type, variadic bool) string {
	var varType string

	switch t := typ.(type) {
	case *types.Named:
		if t.Obj().Pkg() != nil {
			if t.Obj().Pkg().Name() != f.PackageName {
				varType += t.Obj().Pkg().Name() + "."
				f.ifaceUsingIngImport[t.Obj().Pkg().Name()] = t.Obj().Pkg().Path()
			}
		}
		varType += t.Obj().Name()
	case *types.Pointer:
		varType = "*"
		varType += f.parseVarType(t.Elem(), false)
	case *types.Slice:
		if variadic {
			varType = "..."
		} else {
			varType = "[]"
		}
		varType += f.parseVarType(t.Elem(), false)
	case *types.Array:
		varType = fmt.Sprintf("[%d]", t.Len())
		varType += f.parseVarType(t.Elem(), false)
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			varType += "chan "
		case types.RecvOnly:
			varType += "<-chan "
		default:
			varType += "chan<- "
		}
		varType += f.parseVarType(t.Elem(), false)
	case *types.Map:
		key := f.parseVarType(t.Key(), false)
		val := f.parseVarType(t.Elem(), false)
		varType = fmt.Sprintf("map[%s]%s", key, val)
	case *types.Signature:
		varType = fmt.Sprintf(
			"func (%s) (%s)",
			f.parseTypeTuple(t.Params()),
			f.parseTypeTuple(t.Results()),
		)
	default:
		varType = t.String()
	}

	return varType
}

func (f *Ifacer) parseTypeTuple(tup *types.Tuple) string {
	var parts []string

	for i := 0; i < tup.Len(); i++ {
		v := tup.At(i)

		parts = append(parts, f.parseVar(v, false))
	}

	return strings.Join(parts, " , ")
}

func (f *Ifacer) getReturnStr(tuple *types.Tuple, m *Method) {
	if tuple.Len() > 0 {
		m.ReturnTypeStr += "("
		for i := 0; i < tuple.Len(); i++ {
			ArgVar := tuple.At(i)
			m.ReturnTypeStr += ArgVar.Name() + " " + f.parseVar(ArgVar, false)
			var returnVarName string
			if ArgVar.Name() == "" {
				returnVarName = "r" + strconv.Itoa(i)
				m.InitReturnVarStr += "	var " + returnVarName + " "
				m.InitReturnVarStr += f.parseVar(ArgVar, false) + " \r\n"
			}

			m.ReturnStr += returnVarName + ","
			m.ReturnTypeStr += ","
		}

		m.ReturnTypeStr = strings.Trim(m.ReturnTypeStr, ",")
		m.ReturnTypeStr += ")"
		m.ReturnStr = strings.Trim(m.ReturnStr, ",")
	}
}

// Process parsed template and formats and adjusts imports for the parsed content.
func (f *Ifacer) Process() error {
	content, err := f.tpl.Execute("ifacer", ifacerTemplate, f)
	if err != nil {
		return err
	}
	f.logger.Debugf("content : %s", content)
	src, err := imports.Process("", []byte(content), nil)
	if err != nil {
		return err
	}

	f.Content = string(src)

	return nil
}
