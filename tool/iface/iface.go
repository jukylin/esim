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
)

type Iface struct {

	packName string

	imports []string

	found bool

	Methods []Method
}


type Method struct {
	funcName string

	argStr string

	returnStr string
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

		ext := path.Ext(fileInfo.Name())
		if ext != ".go" {
			continue
		}

		//测试文件不copy
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

			for _, decl := range f.Decls {
				specs := decl.(*ast.GenDecl).Specs
				for _, spec := range specs {
					if typeSpec, ok := spec.(*ast.TypeSpec); ok {
						if typeSpec.Name.String() == ifaceName &&
							typeSpec.Type.(*ast.InterfaceType).Interface.IsValid(){
							this.found = true

							for _, method := range typeSpec.Type.(*ast.InterfaceType).Methods.List {
								if funcType, ok := method.Type.(*ast.FuncType); ok{
									m := Method{}
									m.funcName = method.Names[0].String()

									if len(funcType.Params.List) > 0{
										var paramsLen int
										paramsLen = len(funcType.Params.List)
										for k, param := range funcType.Params.List {
											if len(param.Names) > 0{
												m.argStr += param.Names[0].String() + " "
											}else{
												m.argStr += "arg" + strconv.Itoa(k) + " "
											}

											m.argStr += strSrc[param.Type.Pos() -1 :param.Type.End() - 1]

											if k < paramsLen - 1{
												m.argStr += ", "
											}
										}
									}
									m.returnStr = strSrc[funcType.Results.Pos() -1 : funcType.Results.End() -1 ]

									this.Methods = append(this.Methods, m)
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