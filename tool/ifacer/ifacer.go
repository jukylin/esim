package ifacer

import (
	"bytes"
	"errors"
	"github.com/jukylin/esim/log"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/mockery"
	"go/types"
	"golang.org/x/tools/imports"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"github.com/jukylin/esim/pkg/file-dir"
	//"github.com/davecgh/go-spew/spew"
	"fmt"
)

type Ifacer struct {
	logger log.Logger

	Parser *mockery.Parser

	IfaceName string

	StructName string

	IfacePath string

	PackageName string

	Methods []Method

	//implements interface context
	Content string

	Star string

	OutFile string

	//import from interface file
	PkgNoConflictImport map[string]string

	//import in the interface
	IfaceUsingIngImport map[string]string

	writer file_dir.IfaceWriter

	UsingImportStr string
}


func NewIface(writer file_dir.IfaceWriter) *Ifacer {

	ifacer := &Ifacer{}
	ifacer.Parser = mockery.NewParser([]string{})

	ifacer.PkgNoConflictImport = make(map[string]string)

	ifacer.IfaceUsingIngImport = make(map[string]string)

	ifacer.logger = log.NewLogger()

	ifacer.writer = writer

	return ifacer
}


type Method struct {
	FuncName string

	ArgStr string

	ReturnTypeStr string

	ReturnStr string

	InitReturnVarStr string

	ReturnVar []string
}


func (this *Ifacer) Run(v *viper.Viper) error {

	err := this.inputBind(v)
	if err != nil {
		return err
	}

	err = this.Parser.Parse(this.IfacePath)
	if err != nil {
		return err
	}

	err = this.Parser.Load()
	if err != nil {
		return err
	}

	iface, err := this.Parser.Find(this.IfaceName)
	if err != nil {
		return err
	}
	this.PackageName = iface.Pkg.Name()

	this.setNoConflictImport(iface.Pkg.Name(), iface.Pkg.Path())
	this.ManageNoConflictImport(iface.Pkg.Imports())

	this.GenMethods(iface.Type)

	this.getUsingImportStr()

	err = this.Process()
	if err != nil {
		return err
	}

	err = this.writer.Write(this.OutFile, this.Content)
	if err != nil {
		return err
	}

	return nil
}


func (this *Ifacer) inputBind(v *viper.Viper) error {

	name := v.GetString("iname")
	if name == "" {
		return errors.New("必须指定 iname")
	}
	this.IfaceName = name

	out_path := v.GetString("out")
	if out_path == "" {
		out_path = "./dummy_" + strings.ToLower(this.IfaceName) + ".go"
	}
	this.OutFile = out_path

	struct_name := v.GetString("stname")
	if struct_name == "" {
		struct_name = "Dummy" + this.IfaceName
	}
	this.StructName = struct_name

	star := v.GetBool("istar")
	if star == true {
		this.Star = "*"
	}

	this.IfacePath = v.GetString("ipath")

	return nil
}


func (this *Ifacer) GenMethods(interacer *types.Interface) {
	for i := 0; i < interacer.NumMethods(); i++ {
		fn := interacer.Method(i)
		ftype := fn.Type().(*types.Signature)
		m := &Method{}
		m.ReturnStr = "return "

		m.FuncName = fn.Name()
		this.getArgStr(ftype.Params(), m)
		this.getReturnStr(ftype.Results(), m)
		this.Methods = append(this.Methods, *m)
	}
}


func (this *Ifacer) getUsingImportStr() {
	this.UsingImportStr = "import ( \r\n"
	for impName, impPkg := range this.PkgNoConflictImport {
		this.UsingImportStr += "	" + impName + " \"" + impPkg + "\" \r\n"
	}

	this.UsingImportStr += ")"
}


func (this *Ifacer) ManageNoConflictImport(imports []*types.Package) bool {
	for _, imp := range imports {
		this.setNoConflictImport(imp.Name(), imp.Path())
	}

	return true
}


func (this *Ifacer) setNoConflictImport(importName string, importPath string) bool {
	if impPath, ok := this.PkgNoConflictImport[importName]; ok {
		if impPath == importPath {
			return true
		}

		//package name repeat
		level := 1
		flag := true
		for flag {
			importName := this.getUniqueImportName(importPath, level)

			if _, ok := this.PkgNoConflictImport[importName]; !ok {

				this.PkgNoConflictImport[importName] = importPath
				flag = false
			}
			level++
		}
	} else {
		this.PkgNoConflictImport[importName] = importPath
	}

	return true
}

//github.com/jukylin/esim/redis
//level
//		0 redis
//		1 esimredis
//		2 jukylinesimredis
//		3 githubcomjukylinesimredis
func (this *Ifacer) getUniqueImportName(pkgName string, level int) string {
	strs := strings.Split(pkgName, string(filepath.Separator))

	lenStr := len(strs)

	if lenStr-1 < level {
		this.logger.Panicf("%d out of range", level)
	}

	var importName string
	for _, str := range strs[lenStr-level-1:] {
		if strings.Index(str, ".") > -1 {
			str = strings.Replace(str, ".", "", -1)
		}
		importName += str
	}

	return importName
}


func (this *Ifacer) getArgStr(tuple *types.Tuple, m *Method) {
	if tuple.Len() > 0 {
		for i := 0; i < tuple.Len(); i++ {
			ArgVar := tuple.At(i)
			if ArgVar.Name() != "" {
				m.ArgStr += ArgVar.Name() + " "
			} else {
				m.ArgStr += "arg" + strconv.Itoa(i) + " "
			}

			m.ArgStr += this.parseVar(ArgVar)

			if i < tuple.Len()-1 {
				m.ArgStr += ", "
			}
		}
	}
}


func (this *Ifacer) parseVar(varObj *types.Var) string {
	return this.parseVarType(varObj.Type())
}

func (this *Ifacer) parseVarType(typ types.Type) string {

	var varType string

	switch t := typ.(type) {
	case *types.Named:
		if t.Obj().Pkg() != nil{
			if t.Obj().Pkg().Name() != this.PackageName{
				varType += t.Obj().Pkg().Name() + "."
				this.IfaceUsingIngImport[t.Obj().Pkg().Name()] = t.Obj().Pkg().Path()
			}
		}
		varType += t.Obj().Name()
	case *types.Pointer:
		varType = "*"
		varType += this.parseVarType(t.Elem())
	case *types.Slice:
		varType = "[]"
		varType += this.parseVarType(t.Elem())
	case *types.Array:
		varType = fmt.Sprintf("[%d]", t.Len())
		varType += this.parseVarType(t.Elem())
	case *types.Chan:
		switch t.Dir() {
		case types.SendRecv:
			varType += "chan "
		case types.RecvOnly:
			varType += "<-chan "
		default:
			varType += "chan<- "
		}
		varType += this.parseVarType(t.Elem())
	case *types.Map:
		key := this.parseVarType(t.Key())
		val := this.parseVarType(t.Elem())
		varType = fmt.Sprintf("map[%s]%s", key, val)
	case *types.Signature:
		varType = fmt.Sprintf(
			"func (%s) (%s)",
			this.parseTypeTuple(t.Params()),
			this.parseTypeTuple(t.Results()),
		)
	default:
		varType = t.String()
	}

	return varType
}


func (this *Ifacer) parseTypeTuple(tup *types.Tuple) string {
	var parts []string

	for i := 0; i < tup.Len(); i++ {
		v := tup.At(i)

		parts = append(parts, this.parseVar(v))
	}

	return strings.Join(parts, " , ")
}


func (this *Ifacer) getReturnStr(tuple *types.Tuple, m *Method) {
	if tuple.Len() > 0 {
		m.ReturnTypeStr += "("
		for i := 0; i < tuple.Len(); i++ {
			ArgVar := tuple.At(i)
			m.ReturnTypeStr += ArgVar.Name() + " " + this.parseVar(ArgVar)
			var returnVarName string
			if ArgVar.Name() == "" {
				returnVarName = "r" + strconv.Itoa(i)
				m.InitReturnVarStr += "	var " + returnVarName + " "
				m.InitReturnVarStr += this.parseVar(ArgVar) + " \r\n"
			}

			m.ReturnStr += returnVarName + ","
			m.ReturnTypeStr += ","
		}

		m.ReturnTypeStr = strings.Trim(m.ReturnTypeStr, ",")
		m.ReturnTypeStr += ")"
		m.ReturnStr = strings.Trim(m.ReturnStr, ",")
	}
}

func (this *Ifacer) Process() error {
	tmpl, err := template.New("ifacer").Parse(ifacerTemplate)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, this)
	if err != nil {
		return err
	}

	src, err := imports.Process("", buf.Bytes(), nil)
	if err != nil {
		return err
	}

	this.Content = string(src)

	return nil
}
