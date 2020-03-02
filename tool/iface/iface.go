package iface

import (
	"strings"
	"io/ioutil"
	"errors"
	"path"
	"github.com/jukylin/esim/pkg/file-dir"
	"go/token"
	"go/ast"
	"go/parser"
	"strconv"
	"text/template"
	"bytes"
	"golang.org/x/tools/imports"
	"github.com/spf13/viper"
)

type Iface struct {

	StructName string

	PackageName string

	ImportStr string

	found bool

	Methods []Method

	Content string

	OutFile string
}


type Method struct {
	FuncName string

	ArgStr string

	ReturnStr string
}

func (this *Iface) Run(v *viper.Viper) error {
	out_path := v.GetString("out")
	if out_path == ""{
		return errors.New("必须指定 out")
	}
	this.OutFile = out_path

	struct_name := v.GetString("struct_name")
	if struct_name == ""{
		return errors.New("必须指定 struct_name")
	}
	this.StructName = struct_name

	name := v.GetString("name")
	if name == ""{
		return errors.New("必须指定 name")
	}
	this.StructName = struct_name

	iface_path := v.GetString("iface_path")

	err := this.FindIface(iface_path, name)
	if err != nil{
		return err
	}

	if this.found == false{
		return errors.New("没有找到 " + name)
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

func (this *Iface) FindIface(ifacePath string, ifaceName string) (error) {
	ifacePath = strings.TrimRight(ifacePath, "/")

	exists, err := file_dir.IsExistsDir(ifacePath)
	if err != nil {
		return err
	}

	if exists == false {
		return errors.New(ifacePath + " dir not exists")
	}

	files, err := ioutil.ReadDir(ifacePath)
	if err != nil {
		return err
	}

	for _, fileInfo := range files {

		if this.found {
			continue
		}

		ext := path.Ext(fileInfo.Name())
		if ext != ".go" {
			continue
		}

		if strings.Index(fileInfo.Name(), "_test") > -1 {
			continue
		}

		if !fileInfo.IsDir() {
			src, err := ioutil.ReadFile(ifacePath + "/" + fileInfo.Name())
			if err != nil {
				return err
			}

			strSrc := string(src)
			fset := token.NewFileSet() // positions are relative to fset
			f, err := parser.ParseFile(fset, "", strSrc, parser.ParseComments)
			if err != nil {
				return err
			}

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
										m.ReturnStr = strSrc[funcType.Results.Pos()-1: funcType.Results.End()-1 ]

										this.Methods = append(this.Methods, m)
									}
								}
							}
						}
					}
				}
			}
		}
	}


	return nil
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