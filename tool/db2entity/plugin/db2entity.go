package main

import (
	"github.com/jukylin/esim/config"
)

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"path/filepath"

	logger "github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	file_dir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
	"golang.org/x/tools/imports"
)

type db2Entity struct {
	logger logger.Logger

	withDisabledRepo bool

	withDisabledDao bool

	withBoubctx string

	//true not create entity file
	//false create a new entity file in withEntityTarget
	withDisabledEntity bool

	withEntityDir string

	withPackage string

	withStruct string

	withInject bool

	dbConf dbConfig

	ColumnsInter ColumnsInter

	conf config.Config
}

func NewDb2Entity(logger logger.Logger) *db2Entity {

	db2entity := &db2Entity{}
	db2entity.logger = logger

	return db2entity
}

type dbConfig struct {
	host string

	port int

	user string

	password string

	database string

	table string
}

func (this *db2Entity) Run(v *viper.Viper) error {

	this.inputBind(v)

	//var daoTarget string
	//var etar string
	//var entityDir string
	//var repoTarget string

	//if v.GetBool("disdaotar") == false {
	//
	//	daotar := v.GetString("daotar")
	//	if daotar == "" {
	//		daotar = "internal/infra/dao"
	//	}
	//
	//	daoDir := daotar + "/"
	//	daoTarget = daoDir + table + ".go"
	//	ex, err := file_dir.IsExistsFile(daoTarget)
	//	if err != nil {
	//		log.Fatalf(err.Error())
	//	}
	//
	//	if ex {
	//		log.Fatalf(daoDir + " exists")
	//	}
	//
	//	log.Infof("create dir ... %s", daoDir)
	//
	//	err = file_dir.CreateDir(daoDir)
	//	if err != nil {
	//		log.Fatalf(err.Error())
	//	}
	//}

	//if v.GetBool("disrepotar") == false {
	//
	//	repotar := v.GetString("repotar")
	//	if repotar == "" {
	//		repotar = "internal/infra/repo"
	//	}
	//
	//	repoTarget = repotar + "/" + table + ".go"
	//	ex, err := file_dir.IsExistsFile(repoTarget)
	//	if err != nil {
	//		log.Fatalf(err.Error())
	//	}
	//
	//	if ex {
	//		log.Fatalf(repotar + " exists")
	//	}
	//
	//	log.Infof("create dir ... %s", repotar)
	//
	//	err = file_dir.CreateDir(repotar)
	//	if err != nil {
	//		log.Fatalf(err.Error())
	//	}
	//}

	//columnDataTypes, err := GetColumnsFromMysqlTable(user, password, host, port, database, table)

	//if err != nil {
	//	this.logger.Fatalf(err.Error())
	//}

	//js := v.GetBool("json")

	//gorm := v.GetBool("gorm")

	//guregu := v.GetBool("guregu")

	//struc, genMysqlInfo, err := Generate(columnDataTypes, table,
	//	st, pk, js, gorm, guregu, v)
	//if err != nil {
	//	this.logger.Fatalf("Error in creating struct from json: " + err.Error())
	//}

	//if v.GetBool("disetar") == false {
	//
	//	struc, err := format.Source([]byte(struc))
	//	if err != nil {
	//		this.logger.Fatalf(err.Error())
	//	}
	//
	//	struc, err = imports.Process("", struc, nil)
	//	if err != nil{
	//		this.logger.Fatalf(err.Error())
	//	}
	//
	//	err = ioutil.WriteFile(etar, struc, 0666)
	//	if err != nil {
	//		this.logger.Fatalf(err.Error())
	//	}
	//	this.logger.Infof("create file  %s success", etar)
	//} else {
	//	this.logger.Infof("not create file  %s", etar)
	//}

	//if v.GetBool("disdaotar") == false {
	//
	//	daoStr, err := GenerateDao(table, st, pk, v, genMysqlInfo, boubctx)
	//	if err != nil {
	//		this.logger.Fatalf("Error in creating struct from json: " + err.Error())
	//	}
	//
	//	forDaoStr, err := format.Source([]byte(daoStr))
	//	if err != nil {
	//		this.logger.Fatalf(err.Error())
	//	}
	//
	//	forDaoStr, err = imports.Process("", forDaoStr, nil)
	//	if err != nil{
	//		this.logger.Fatalf(err.Error())
	//	}
	//
	//	err = ioutil.WriteFile(daoTarget, forDaoStr, 0666)
	//	if err != nil {
	//		this.logger.Fatalf(err.Error())
	//	}
	//
	//	this.logger.Infof("create file  %s success", daoTarget)
	//} else {
	//	this.logger.Infof("not create file  %s", daoTarget)
	//}

	//if v.GetBool("disrepotar") == false {
	//
	//	repoStr := GenerateRepo(table, st, v, boubctx)
	//
	//	forRepoStr, err := format.Source([]byte(repoStr))
	//	if err != nil {
	//		this.logger.Fatalf(err.Error())
	//	}
	//
	//	forRepoStr, err = imports.Process("", forRepoStr, nil)
	//	if err != nil{
	//		this.logger.Fatalf(err.Error())
	//	}
	//
	//	err = ioutil.WriteFile(repoTarget, forRepoStr, 0666)
	//	if err != nil {
	//		this.logger.Fatalf(err.Error())
	//	}
	//
	//	this.logger.Infof("create file  %s success", repoTarget)
	//} else {
	//	this.logger.Infof("not create file  %s", repoTarget)
	//}

	//inject := v.GetBool("inject")
	//if inject == true {
	//	pwd, _ := os.Getwd()
	//	goPath := os.Getenv("GOPATH") + "/src/"
	//	//项目路径
	//	proPath := strings.Replace(pwd, goPath, "", -1)
	//
	//	err = file_dir.EsimBackUpFile(goPath + "/" + proPath + "/internal/infra/infra.go")
	//	if err != nil{
	//		this.logger.Warnf("backup err %s:%s", proPath + "/internal/infra/infra.go", err.Error())
	//	}
	//
	//	Inject("infra", st, pk,
	//		st + "Repo", "DB" + st + "Repo", proPath+"/internal/infra/repo")
	//}

	return nil
}

func (this *db2Entity) inputBind(v *viper.Viper) {

	this.bindDbConfig(v)

	packageName := v.GetString("package")
	if packageName == "" {
		packageName = this.dbConf.database
	}
	this.withPackage = packageName

	stuctName := v.GetString("struct")
	if stuctName == "" {
		stuctName = this.dbConf.table
	}
	this.withStruct = stuctName

	this.withDisabledRepo = v.GetBool("disabled_repo")
	if this.withDisabledRepo == false {
		//repo 目录是存在
		existsRepo, err := file_dir.IsExistsDir("." + string(filepath.Separator) + "internal" + string(filepath.Separator) + "infra" + string(filepath.Separator) + "repo")
		if err != nil {
			this.logger.Fatalf(err.Error())
		}

		if existsRepo == false {
			this.logger.Fatalf("repo dir not exists")
		}
	}

	//写的时候才创建文件
	this.withDisabledDao = v.GetBool("disabled_dao")
	if this.withDisabledDao == false {
		//dao 目录是否存在
		existsdao, err := file_dir.IsExistsDir("." + string(filepath.Separator) + "internal" + string(filepath.Separator) + "infra " + string(filepath.Separator) + "dao")
		if err != nil {
			this.logger.Fatalf(err.Error())
		}

		if existsdao == false {
			this.logger.Fatalf("dao dir not exists")
		}

		//daoTarget := v.GetString("dao_target")
		//if daoTarget == "" {
		//	daoTarget = "internal/infra/dao"
		//}
		//
		//daoDir := daoTarget + "/"
		//daoTarget = daoDir + table + ".go"
		//ex, err := file_dir.IsExistsFile(daoTarget)
		//if err != nil {
		//	this.logger.Fatalf(err.Error())
		//}
		//
		//if ex {
		//	this.logger.Fatalf(daoDir + " exists")
		//}
		//
		//this.logger.Infof("create dir ... %s", daoDir)
		//
		//err = file_dir.CreateDir(daoDir)
		//if err != nil {
		//	this.logger.Fatalf(err.Error())
		//}

	}

	boubctx := v.GetString("boubctx")
	if boubctx != "" {
		this.withBoubctx = boubctx + string(filepath.Separator)
	}

	this.bindEntityDir(v)

}

//写的时候才创建文件
func (this *db2Entity) bindEntityDir(v *viper.Viper) {
	if v.GetBool("disabled_entity") == false {

		this.withEntityDir = v.GetString("entity_target")

		if this.withEntityDir == "" {
			if this.withBoubctx != "" {
				this.withEntityDir = "internal" + string(filepath.Separator) + "domain" + string(filepath.Separator) + this.withBoubctx + "entity"
			} else {
				this.withEntityDir = "internal" + string(filepath.Separator) + "domain" + string(filepath.Separator) + "entity"
			}
		}

		entityDirExists, err := file_dir.IsExistsDir(this.withEntityDir)
		if err != nil {
			this.logger.Fatalf(err.Error())
		}

		if entityDirExists == false {
			err = file_dir.CreateDir(this.withEntityDir)
			if err != nil {
				this.logger.Fatalf(err.Error())
			}
		}

		this.withEntityDir = this.withEntityDir + " + string(filepath.Separator) + "
		ex, err := file_dir.IsExistsFile(this.withEntityDir + this.dbConf.table + ".go")
		if err != nil {
			this.logger.Fatalf(err.Error())
		}

		if ex {
			this.logger.Fatalf(this.withEntityDir + this.dbConf.table + ".go" + " exists")
		}

		this.logger.Infof("creating dir ... %s", this.withEntityDir+this.dbConf.table+".go")

		err = file_dir.CreateDir(this.withEntityDir)
		if err != nil {
			this.logger.Fatalf(err.Error())
		}
	}
}

func (this *db2Entity) bindDbConfig(v *viper.Viper) {
	dbConfig := dbConfig{}
	host := v.GetString("host")
	if host == "" {
		this.logger.Fatalf("host is empty")
	}
	dbConfig.host = host

	port := v.GetInt("port")
	if port == 0 {
		this.logger.Fatalf("port is 0")
	}
	dbConfig.port = port

	user := v.GetString("user")
	if user == "" {
		this.logger.Fatalf("user is empty")
	}
	dbConfig.user = user

	password := v.GetString("password")
	if password == "" {
		this.logger.Fatalf("password is empty")
	}
	dbConfig.password = password

	database := v.GetString("database")
	if database == "" {
		this.logger.Fatalf("database is empty")
	}
	dbConfig.database = database

	table := v.GetString("table")
	if table == "" {
		this.logger.Fatalf("table is empty")
	}
	dbConfig.table = table

	this.dbConf = dbConfig
}

func GenerateDao(tableName string, structName string, pkgName string,
	v *viper.Viper, genMysqlInfo generateMysqlInfo, boubctx string) ([]byte, error) {

	daoStr := `
package dao

import (
	"errors"
	"context"
	"` + file_dir.GetCurrentDir() + `/internal/domain/` + boubctx + `entity"
	"github.com/jinzhu/gorm"
	"github.com/jukylin/esim/mysql"
)


type ` + structName + `Dao struct{
	mysql *mysql.MysqlClient
}

func New` + structName + `Dao() *` + structName + `Dao {
	dao := &` + structName + `Dao{
		mysql : mysql.NewMysqlClient(),
	}

	return dao
}


//主库
func (this *` + structName + `Dao) GetDb(ctx context.Context) *gorm.DB  {
	return this.mysql.GetCtxDb(ctx, "` + pkgName + `").Table("` + tableName + `")
}

//从库
func (this *` + structName + `Dao) GetSlaveDb(ctx context.Context) *gorm.DB  {
	return this.mysql.GetCtxDb(ctx, "` + pkgName + `_slave").Table("` + tableName + `")
}


//返回 自增id，错误
func (this *` + structName + `Dao) Create(ctx context.Context, ` + GetFirstToLower(structName) +
		` *entity.` + structName + `) (` + genMysqlInfo.priKeyType + `, error){
	db := this.GetDb(ctx).Create(` + GetFirstToLower(structName) + `)
	if db.Error != nil{
		return ` + genMysqlInfo.priKeyType + `(0), db.Error
	}else{
		return ` + genMysqlInfo.priKeyType + `(` + GetFirstToLower(structName) + `.ID), nil
	}
}

//ctx, "name = ?", "test"
func (this *` + structName + `Dao) Count(ctx context.Context, query interface{}, args ...interface{}) (int64, error){
	var count int64
	db := this.GetSlaveDb(ctx).Where(query, args...).Count(&count)
	if db.Error != nil{
		return count, db.Error
	}else{
		return count, nil
	}
}

// ctx, "id,name", "name = ?", "test"
// 可通过 HasData() 判断是否有数据
func (this *` + structName + `Dao) Find(ctx context.Context, squery , wquery interface{}, args ...interface{}) (entity.` + structName + `, error){
	var ` + GetFirstToLower(structName) + ` entity.` + structName + `
	db := this.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).First(&` + GetFirstToLower(structName) + `)
	if db.Error != nil{
		return ` + strings.ToLower(string(structName[0])) + `, db.Error
	}else{
		return ` + strings.ToLower(string(structName[0])) + `, nil
	}
}


// ctx, "id,name", "name = ?", "test"
//最多取10条
func (this *` + structName + `Dao) List(ctx context.Context, squery , wquery interface{}, args ...interface{}) ([]entity.` + structName + `, error){
	` + GetFirstToLower(structName) + `s := []entity.` + structName + `{}
	db := this.GetSlaveDb(ctx).Select(squery).
		Where(wquery, args...).Limit(10).Find(&` + GetFirstToLower(structName) + `s)
	if db.Error != nil{
		return ` + GetFirstToLower(structName) + `s, db.Error
	}else{
		return ` + GetFirstToLower(structName) + `s, nil
	}
}

func (this *` + structName + `Dao) DelById(ctx context.Context, id ` + genMysqlInfo.priKeyType + `) (bool, error){
	var del` + structName + ` entity.` + structName + `

	if del` + structName + `.DelKey() == ""{
		return false, errors.New("找不到 is_del / is_deleted / is_delete 字段")
	}

	del` + structName + `.ID = id
	db := this.GetDb(ctx).Update(map[string]interface{}{del` + structName + `.DelKey(): 1})
	if db.Error != nil{
		return false, db.Error
	}else{
		return true, nil
	}
}

//ctx, map[string]interface{}{"name": "hello"}, "name = ?", "test"
//返回影响数
func (this *` + structName + `Dao) Update(ctx context.Context, update map[string]interface{}, query interface{}, args ...interface{}) (int64, error) {
	db := this.GetDb(ctx).Where(query, args).
		Updates(update)
	return db.RowsAffected, db.Error
}

`

	return []byte(daoStr), nil
}

func GenerateRepo(tableName string, structName string, v *viper.Viper, boubctx string) string {
	repoStr := `
package repo


import (
	"context"
	"` + file_dir.GetCurrentDir() + `/internal/domain/` + boubctx + `entity"
	"` + file_dir.GetCurrentDir() + `/internal/infra/dao"
	"github.com/jukylin/esim/log"
)


type ` + structName + `Repo interface {
	FindById(context.Context, int64) entity.` + structName + `
}

type DB` + structName + `Repo struct{

	logger log.Logger

	` + GetFirstStringToLower(structName) + `Dao *dao.` + structName + `Dao
}

func NewDB` + structName + `Repo(logger log.Logger) ` + structName + `Repo {
	repo := &DB` + structName + `Repo{
		logger : logger,
	}

	if repo.` + GetFirstStringToLower(structName) + `Dao == nil{
		repo.` + GetFirstStringToLower(structName) + `Dao = dao.New` + structName + `Dao()
	}


	return repo
}

func (this *DB` + structName + `Repo) FindById(ctx context.Context, id int64) entity.` + structName + ` {
	var ` + tableName + ` entity.` + structName + `
	var err error

	` + tableName + `, err = this.` + GetFirstStringToLower(structName) + `Dao.Find(ctx, "*", "id = ? ", id)

	if err != nil{
		this.logger.Errorf(err.Error())
		return ` + tableName + `
	}

	return ` + tableName + `
}

`

	return repoStr
}

func GenInitFile(pkgName string, v *viper.Viper) string {
	initStr := `
package ` + pkgName + `

	import (
		"gopkg.in/go-playground/mold.v2/modifiers"
	)

	var (
	conform  = modifiers.New()
)
`

	return initStr
}

func camelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:])
}

func GetFirstToLower(str string) string {
	return strings.ToLower(string(str[0]))
}

func GetFirstStringToLower(str string) string {
	return strings.ToLower(string(str[0])) + str[1:]
}

func Inject(structName string, fieldName, packageName, interfaceName string,
	instanceName string, importStr string) {

	infrDir := "./internal/infra/"

	infrFile := "infra.go"

	exists, err := file_dir.IsExistsFile(infrDir + infrFile)
	if err != nil {
		//log.Errorf(err.Error())
		return
	}

	if exists {
		src, err := ioutil.ReadFile(infrDir + infrFile)
		if err != nil {
			//log.Errorf(err.Error())
			return
		}

		//先整理下源文件
		formatSrc, err := format.Source([]byte(src))
		if err != nil {
			//log.Errorf(err.Error())
			return
		}

		ioutil.WriteFile(infrDir+infrFile, formatSrc, 0666)

		srcStr := string(formatSrc)

		source := handleInject(srcStr, "Infra",
			fieldName, packageName, interfaceName, instanceName, importStr)

		//整理，写入
		formatSrc, err = format.Source([]byte(source))
		if err != nil {
			//log.Errorf(err.Error())
			return
		}

		formatSrc, err = imports.Process("", formatSrc, nil)
		if err != nil {
			//log.Errorf(err.Error())
			return
		}

		ioutil.WriteFile(infrDir+infrFile, []byte(formatSrc), 0666)

		//err = ExecGoFmt(infrFile, infrDir)
		//if err != nil {
		//	log.Fatalf(err.Error())
		//}

		err = ExecWire(infrDir)
		if err != nil {
			//log.Fatalf(err.Error())
		}

		//log.Infof("注入成功")

	} else {
		//log.Errorf("不存在 %s", infrDir+infrFile)
	}
}

func handleInject(srcStr string, structName string, fieldName, packageName, interName string,
	instName string, importStr string) string {

	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", srcStr, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	var hasStruct bool
	var oldStruct string
	var newStruct string

	var oldImportStr string
	var newImportStr string

	provideFunc := getProvideFunc(interName, instName)

	var oldSet string
	var newSet string

	for _, decl := range f.Decls {

		if GenDecl, ok := decl.(*ast.GenDecl); ok {
			if GenDecl.Tok.String() == "import" {
				oldImports := getOldImports(GenDecl)
				newImports := append(oldImports, "\""+importStr+"\"")
				oldImportStr = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
				newImportStr = getNewImportStr(newImports)
			}

			if GenDecl.Tok.String() == "type" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.TypeSpec); ok {
						if typeSpec.Name.String() == structName {
							hasStruct = true
							oldStruct = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
							oldFields := GetOldFields(GenDecl, srcStr)
							newFields := append(oldFields, pkg.Field{Field: interName + " repo." + interName})
							newStruct = GetNewStruct(structName, newFields)
						}
					}
				}
			}

			if GenDecl.Tok.String() == "var" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.ValueSpec); ok {
						for _, name := range typeSpec.Names {
							if name.String() == "infraSet" {
								var oldArgs []string
								oldSet = srcStr[GenDecl.TokPos-1 : GenDecl.End()]
								oldArgs = append(oldArgs, "var infraSet = wire.NewSet(")
								oldArgs = append(oldArgs, getSet(GenDecl, srcStr)...)
								newArgs := append(oldArgs, "provide"+instName+",")
								newArgs = append(newArgs, ")")
								newSet = getNewSet(newArgs)
							}
						}
					}
				}
			}
		}
	}

	//println(srcStr)
	if hasStruct == false {
		//log.Errorf("不存在 %s", structName)
		return ""
	}

	srcStr = strings.Replace(srcStr, oldImportStr, newImportStr, -1)
	srcStr = strings.Replace(srcStr, oldStruct, newStruct, -1)
	srcStr = strings.Replace(srcStr, oldSet, newSet, -1)
	srcStr += provideFunc

	return srcStr
}

func getNewSet(args []string) string {
	var structStr string

	for _, a := range args {
		structStr += "	" + a + "\r\n"
	}
	return structStr
}

func getOldImports(GenDecl *ast.GenDecl) []string {
	var imports []string
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.ImportSpec); ok {
			var name string
			if spec.Name.String() != "<nil>" {
				name = spec.Name.String()
			}

			imports = append(imports, name+" "+spec.Path.Value)
		}
	}

	return imports
}

func GetOldFields(GenDecl *ast.GenDecl, strSrc string) []pkg.Field {
	var fields pkg.Fields
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.TypeSpec); ok {
			if structType, ok := spec.Type.(*ast.StructType); ok {
				for _, astField := range structType.Fields.List {
					var field pkg.Field
					if astField.Doc != nil {
						for _, doc := range astField.Doc.List {
							field.Doc = append(field.Doc, doc.Text)
						}
					}

					if astField.Tag != nil {
						field.Tag = astField.Tag.Value
					}
					var name string
					if len(astField.Names) > 0 {
						name = astField.Names[0].String()
						field.Name = name
						field.Field = name + " " + strSrc[astField.Type.Pos()-1:astField.Type.End()-1]
					} else {
						nameSplit := strings.Split(strSrc[astField.Type.Pos()-1:astField.Type.End()-1], ".")
						field.Name = nameSplit[len(nameSplit)-1]
						field.Field = strSrc[astField.Type.Pos()-1 : astField.Type.End()-1]
					}

					fields = append(fields, field)
				}
			}
		}
	}

	return fields
}

func getSet(GenDecl *ast.GenDecl, srcStr string) []string {
	var args []string
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.ValueSpec); ok {
			for _, value := range spec.Values {
				if callExpr, ok := value.(*ast.CallExpr); ok {
					for _, callArg := range callExpr.Args {
						if arg, ok := callArg.(*ast.CallExpr); ok {
							var x string
							var sel string
							if fun, ok := arg.Fun.(*ast.SelectorExpr); ok {
								x = fun.X.(*ast.Ident).String()
								sel = fun.Sel.String()
							}
							args = append(args, x+"."+sel+srcStr[arg.Lparen-1:arg.Rparen]+",")
						}

						if arg, ok := callArg.(*ast.Ident); ok {
							args = append(args, arg.String()+",")
						}
					}
				}
			}
		}
	}

	return args
}

func getNewImportStr(newImports []string) string {

	var importStr = "import (\r\n"

	for _, is := range newImports {
		importStr += "	" + is + "\r\n"
	}

	importStr += ")\r\n"

	return importStr
}

func getProvideFunc(interName, instName string) string {
	funcStr := `
func provide` + instName + `(esim *container.Esim) repo.` + interName + ` {
	return repo.New` + instName + `(esim.Logger)
}`

	return funcStr
}

func GetNewStruct(name string, fields []pkg.Field) string {
	var structStr string
	structStr = " type " + name + " struct {\r\n"

	for _, f := range fields {
		if len(f.Doc) > 0 {
			for _, d := range f.Doc {
				structStr += "	" + d + "\r\n"
			}
		}
		structStr += "	" + f.Field + "\r\n"
		structStr += "\r\n"
	}

	structStr += "}\r\n"

	return structStr
}

func GetFirstToUpper(str string) string {
	return strings.ToUpper(string(str[0])) + str[1:]
}

func ExecGoFmt(file string, dir string) error {
	cmd_line := fmt.Sprintf("go fmt %s", file)

	//log.Infof(cmd_line)

	args := strings.Split(cmd_line, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}

func ExecWire(dir string) error {
	cmd_line := fmt.Sprintf("wire")

	//log.Infof("dir %s, %s", dir, cmd_line)

	args := strings.Split(cmd_line, " ")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	cmd.Env = os.Environ()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return err
}
