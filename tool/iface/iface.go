package iface

import (
	"strings"

	"errors"

	"github.com/jukylin/esim/pkg/file-dir"

	"go/ast"

	"strconv"
	"text/template"
	"bytes"
	"golang.org/x/tools/imports"
	"github.com/spf13/viper"
	"github.com/vektra/mockery/mockery"
	"go/types"
	"path/filepath"
	"github.com/jukylin/esim/log"
)

type Iface struct {

	logger log.Logger

	Parser *mockery.Parser

	IfaceName string

	StructName string

	PackageName string

	Imports string

	ImportStr string

	found bool

	Methods []Method

	Content string

	Star string

	OutFile string

	NoConflictImport map[string]string
}

func NewIface() *Iface {

	iface := &Iface{}
	iface.Parser = mockery.NewParser([]string{})

	iface.NoConflictImport = make(map[string]string)

	iface.logger = log.NewLogger()

	return iface
}

type Method struct {
	FuncName string

	ArgStr string

	ReturnTypeStr string

	ReturnStr string

	InitReturnVarStr string

	ReturnVar []string
}

func (this *Iface) Run(v *viper.Viper) error {

	out_path := v.GetString("out")
	if out_path == ""{
		return errors.New("必须指定 out")
	}
	this.OutFile = out_path

	struct_name := v.GetString("stname")
	if struct_name == ""{
		return errors.New("必须指定 stname")
	}
	this.StructName = struct_name

	name := v.GetString("iname")
	if name == ""{
		return errors.New("必须指定 iname")
	}
	this.IfaceName = name

	star := v.GetBool("istar")
	if star == true {
		this.Star = "*"
	}

	iface_path := v.GetString("ipath")
	err := this.Parser.Parse(iface_path)
	if err != nil{
		return err
	}

	err = this.Parser.Load()
	if err != nil{
		return err
	}

	iface, err := this.Parser.Find(this.IfaceName)
	if err != nil{
		return err
	}
	//iface.File.Imports[0].
	//spew.Dump(iface.File.Decls.())
	//this.extra(iface)
	//return nil

	this.getNoConflictImport(iface.Pkg.Imports())

	for i := 0; i < iface.Type.NumMethods(); i++ {
		fn := iface.Type.Method(i)
		ftype := fn.Type().(*types.Signature)
		//fname := fn.Name()
		m := &Method{}
		m.ReturnStr = "return "

		m.FuncName = fn.Name()
		this.getArgStr(ftype.Params(), m)
		//spew.Dump(m)
		this.getReturnStr(ftype.Results(), m)

		//spew.Dump(ftype.Results(), fname)
	}
	return nil
	//err := this.FindIface(iface.File)
	if err != nil{
		return err
	}


	err = this.Process()
	if err != nil{
		return err
	}

	err = this.Write()
	if err != nil{
		return err
	}

	return nil
}

func (this *Iface) extra(p *mockery.Interface)  {
	this.PackageName = p.Pkg.Name()

	this.ImportStr = this.getNoConflictImport(p.Pkg.Imports())

}

func (this *Iface) getNoConflictImport(imports []*types.Package) string {
	for _, imp := range imports{
		if impPath, ok := this.NoConflictImport[imp.Name()]; ok {
			if impPath == imp.Path(){
				continue
			}

			//package name repeat
			level := 1
			flag := true
			for flag {
				importName := this.getUniqueImportName(imp.Path(), level)

				if _, ok := this.NoConflictImport[imp.Name()]; !ok {
					this.NoConflictImport[importName] = imp.Path()
					flag = false
				}
			}
		}else{
			this.NoConflictImport[imp.Name()] = imp.Path()
		}
	}

	return ""
}




//github.com/jukylin/esim/redis
//level
//		0 redis
//		1 esimredis
//		2 jukylinesimredis
//		3 githubcomjukylinesimredis
func (this *Iface) getUniqueImportName(pkgName string, level int) (string) {
	strs := strings.Split(pkgName, string(filepath.Separator))

	lenStr := len(strs)

	if lenStr - 1 < level{
		this.logger.DPanicf("%d out of range", level)
	}

	var importName string
	for _, str := range strs[lenStr - level - 1:] {
		if strings.Index(str, ".") > -1 {
			str = strings.Replace(str, ".", "", -1)
		}
		importName += str
	}

	return importName
}


func (this *Iface) FindIface(f *ast.File) (error) {
	var strSrc string
	var ifaceName string
	this.PackageName = f.Name.String()

	for _, decl := range f.Decls {
		if GenDecl, ok := decl.(*ast.GenDecl); ok {
			if GenDecl.Tok.String() == "import" {
				this.ImportStr = strSrc[GenDecl.Pos()-1: GenDecl.End()-1]
				continue
			}

			for _, spec := range GenDecl.Specs {

				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if typeSpec.Name.String() == ifaceName &&
						typeSpec.Type.(*ast.InterfaceType).Interface.IsValid() {
						this.found = true

						for _, method := range typeSpec.Type.(*ast.InterfaceType).Methods.List {
							if funcType, ok := method.Type.(*ast.FuncType); ok {
								m := Method{}
								m.ReturnStr = "return "
								m.FuncName = method.Names[0].String()

								if len(funcType.Params.List) > 0 {
									var paramsLen int
									paramsLen = len(funcType.Params.List)
									for k, param := range funcType.Params.List {
										if len(param.Names) > 0 {
											m.ArgStr += param.Names[0].String() + " "
										} else {
											m.ArgStr += "arg" + strconv.Itoa(k) + " "
										}

										m.ArgStr += strSrc[param.Type.Pos()-1:param.Type.End()-1]

										if k < paramsLen-1 {
											m.ArgStr += ", "
										}
									}
								}


								if funcType.Results.NumFields() > 0 {
									m.ReturnTypeStr = strSrc[funcType.Results.Pos()-1: funcType.Results.End()-1 ]
									var returnVarName string
									for rk, funcResult := range funcType.Results.List{
										if len(funcResult.Names) > 0 {
											returnVarName = funcResult.Names[0].String()
										}else{
											returnVarName = "r" + strconv.Itoa(rk)
											m.InitReturnVarStr += "	var " + returnVarName + " "
											m.InitReturnVarStr += strSrc[funcResult.Type.Pos()- 1 : funcResult.Type.End() - 1 ] + " \r\n"
										}

										m.ReturnStr += returnVarName + ","
									}
								}
								m.ReturnStr = strings.Trim(m.ReturnStr, ",")
								this.Methods = append(this.Methods, m)
							}
						}
					}
				}
			}
		}
	}



	return nil
}


func (this *Iface) getImportArr(tuple *types.Tuple, m *Method)  {

}

func (this *Iface) getArgStr(tuple *types.Tuple, m *Method)  {
	if tuple.Len() > 0 {
		for i := 0; i < tuple.Len(); i++ {
			ArgInfo := tuple.At(i)
			//spew.Dump(ArgInfo)
			if ArgInfo.Name() != "" {
				m.ArgStr += ArgInfo.Name() + " "
			} else {
				m.ArgStr += "arg" + strconv.Itoa(i) + " "
			}

			m.ArgStr += ArgInfo.Type().String()

			if i < tuple.Len() - 1 {
				m.ArgStr += ", "
			}
		}
	}
}

func (this *Iface) getReturnStr(tuple *types.Tuple, m *Method)  {
	if tuple.Len() > 0 {
		for i := 0; i < tuple.Len(); i++ {
			//ArgInfo := tuple.At(i)
			//spew.Dump(ArgInfo)
			//spew.Dump(ArgInfo)
			//m.ReturnTypeStr = strSrc[funcType.Results.Pos()-1: funcType.Results.End()-1 ]
			//var returnVarName string
			//for rk, funcResult := range funcType.Results.List{
			//	if len(funcResult.Names) > 0 {
			//		returnVarName = funcResult.Names[0].String()
			//	}else{
			//		returnVarName = "r" + strconv.Itoa(rk)
			//		m.InitReturnVarStr += "	var " + returnVarName + " "
			//		m.InitReturnVarStr += strSrc[funcResult.Type.Pos()- 1 : funcResult.Type.End() - 1 ] + " \r\n"
			//	}
			//
			//	m.ReturnStr += returnVarName + ","
			//}

			m.ReturnStr = strings.Trim(m.ReturnStr, ",")
		}
	}
}

func (this *Iface) Process() error {
	tmpl, err := template.New("iface").Parse(ifaceTemplate)
	if err != nil{
		return err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, this)
	if err != nil{
		return err
	}

	src, err := imports.Process("", buf.Bytes(), nil)
	if err != nil{
		return err
	}

	this.Content = string(src)

	return nil
}

func (this *Iface) Write() error {
	return file_dir.EsimWrite(this.OutFile, this.Content)
}