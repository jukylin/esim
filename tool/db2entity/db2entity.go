package db2entity

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"strings"
	"text/template"
	"path/filepath"
	"bytes"
	logger "github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	file_dir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/pkg/templates"
	"golang.org/x/tools/imports"
)

type db2Entity struct {
	withDisabledRepo bool

	withRepoTarget string

	withDisabledDao bool

	withDaoTarget string

	//true not create entity file
	//false create a new entity file in withEntityTarget
	withDisabledEntity bool

	withEntityTarget string

	//true inject repo to infra
	withInject bool

	logger logger.Logger

	withBoubctx string

	withPackage string

	withStruct string

	//Camel Form
	CamelStruct string

	ColumnsRepo ColumnsRepo

	dbConf dbConfig

	writer file_dir.IfaceWriter

	withInfraDir string

	withInfraFile string

	hasInfraStruct bool

	oldInfraInfo *infraInfo

	newInfraInfo *infraInfo

	execer pkg.Exec
}

type dbConfig struct {
	host string

	port int

	user string

	password string

	database string

	table string
}

type Db2EntityOption func(*db2Entity)

type Db2EntityOptions struct{}

func NewDb2Entity(options ...Db2EntityOption) *db2Entity {

	d := &db2Entity{}

	for _, option := range options {
		option(d)
	}

	if d.writer == nil{
		d.writer = file_dir.NewNullWrite()
	}

	if d.execer == nil{
		d.execer = &pkg.NullExec{}
	}

	return d
}

func (Db2EntityOptions) WithLogger(logger logger.Logger) Db2EntityOption {
	return func(d *db2Entity) {
		d.logger = logger
	}
}

func (Db2EntityOptions) WithColumnsInter(ColumnsRepo ColumnsRepo) Db2EntityOption {
	return func(d *db2Entity) {
		d.ColumnsRepo = ColumnsRepo
	}
}

func (Db2EntityOptions) WithIfaceWrite(writer file_dir.IfaceWriter) Db2EntityOption {
	return func(d *db2Entity) {
		d.writer = writer
	}
}

func (Db2EntityOptions) WithInfraInfo(infra *infraInfo) Db2EntityOption {
	return func(d *db2Entity) {
		d.oldInfraInfo = infra
		d.newInfraInfo = infra
	}
}

func (Db2EntityOptions) WithWriter(writer file_dir.IfaceWriter) Db2EntityOption {
	return func(d *db2Entity) {
		d.writer = writer
	}
}

func (Db2EntityOptions) WithExecer(execer pkg.Exec) Db2EntityOption {
	return func(d *db2Entity) {
		d.execer = execer
	}
}

type infraInfo struct{
	imports pkg.Imports

	importStr string

	structInfo templates.StructInfo

	structStr string

	specialStructName string

	specialVarName string

	infraSetArgs infraSetArgs

	infraSetStr string

	content string
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

func (this *db2Entity) Run(v *viper.Viper) error {

	this.bindInput(v)

	columns, err := this.ColumnsRepo.GetColumns(this.dbConf)
	if err != nil{
		this.logger.Fatalf(err.Error())
	}

	entityTmp := this.cloumnsToEntityTmp(columns)
	entityContent := this.executeTmpl("entity_tpl", entityTmp, entityTemplate)
	this.writer.Write(this.withEntityTarget + this.dbConf.table + ".go", entityContent)

	daoTmp := this.cloumnsToDaoTmp(columns)
	daoContent := this.executeTmpl("entity_tpl", daoTmp, daoTemplate)
	this.writer.Write(this.withDaoTarget + this.dbConf.table + ".go", daoContent)

	repoTmp := this.cloumnsToRepoTmp(columns)
	repoContent := this.executeTmpl("entity_tpl", repoTmp, repoTemplate)
	this.writer.Write(this.withRepoTarget + this.dbConf.table + ".go", repoContent)

	this.injectToInfra()

	return nil
}

func (this *db2Entity) bindInput(v *viper.Viper) {

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
	this.CamelStruct = templates.SnakeToCamel(stuctName)

	boubctx := v.GetString("boubctx")
	if boubctx != "" {
		this.withBoubctx = boubctx + string(filepath.Separator)
	}

	this.bindEntityDir(v)

	this.bindRepoDir(v)

	this.bindDaoDir(v)

	this.bindInfra(v)
}

func (this *db2Entity) bindEntityDir(v *viper.Viper) {
	if v.GetBool("disabled_entity") == false {

		this.withEntityTarget = v.GetString("entity_target")

		if this.withEntityTarget == "" {
			if this.withBoubctx != "" {
				this.withEntityTarget = "internal" + string(filepath.Separator) + "domain" + string(filepath.Separator) + this.withBoubctx + "entity"
			} else {
				this.withEntityTarget = "internal" + string(filepath.Separator) + "domain" + string(filepath.Separator) + "entity"
			}
		}

		entityTargetExists, err := file_dir.IsExistsDir(this.withEntityTarget)
		if err != nil {
			this.logger.Fatalf(err.Error())
		}

		if entityTargetExists == false {
			err = file_dir.CreateDir(this.withEntityTarget)
			if err != nil {
				this.logger.Fatalf(err.Error())
			}
		}

		this.withEntityTarget = this.withEntityTarget + string(filepath.Separator)
	}
}

func (this *db2Entity) bindRepoDir(v *viper.Viper)  {
	this.withDisabledRepo = v.GetBool("disabled_repo")
	if this.withDisabledRepo == false {

		this.withRepoTarget = v.GetString("repo_target")
		if this.withRepoTarget == "" {
			this.withRepoTarget = "internal" + string(filepath.Separator) + "infra" + string(filepath.Separator) + "repo"
		} else {
			this.withRepoTarget = strings.TrimLeft(this.withRepoTarget, ".") + string(filepath.Separator)
			this.withRepoTarget = strings.Trim(this.withRepoTarget, "/") + string(filepath.Separator)
		}

		//repo 目录是否存在
		existsRepo, err := file_dir.IsExistsDir(this.withRepoTarget)
		if err != nil {
			this.logger.Fatalf(err.Error())
		}

		if existsRepo == false {
			this.logger.Fatalf("repo dir not exists")
		}
		this.withRepoTarget = this.withRepoTarget + string(filepath.Separator)
	}
}

func (this *db2Entity) bindDaoDir(v *viper.Viper)  {
	this.withDisabledDao = v.GetBool("disabled_dao")
	if this.withDisabledDao == false {

		this.withDaoTarget = v.GetString("dao_target")
		if this.withDaoTarget == "" {
			this.withDaoTarget = "internal" + string(filepath.Separator) + "infra " + string(filepath.Separator) + "dao"
		} else {
			this.withDaoTarget = strings.TrimLeft(this.withDaoTarget, ".") + string(filepath.Separator)
			this.withDaoTarget = strings.Trim(this.withDaoTarget, "/") + string(filepath.Separator)
		}

		//dao 目录是否存在
		existsdao, err := file_dir.IsExistsDir(this.withDaoTarget)
		if err != nil {
			this.logger.Fatalf(err.Error())
		}

		if existsdao == false {
			this.logger.Fatalf("dao dir not exists")
		}
		this.withDaoTarget = this.withDaoTarget + string(filepath.Separator)
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

func (this *db2Entity) bindInfra(v *viper.Viper)  {
	this.withInject = v.GetBool("inject")

	this.withInfraDir = v.GetString("infra_dir")
	if this.withInfraDir == ""{
		this.withInfraDir = "internal" + string(filepath.Separator) + "infra" + string(filepath.Separator)
	} else {
		this.withInfraDir = strings.TrimLeft(this.withInfraDir, ".") + string(filepath.Separator)
		this.withInfraDir = strings.Trim(this.withInfraDir, "/") + string(filepath.Separator)
	}

	if v.GetString("infra_file") == ""{
		this.withInfraFile = "infra.go"
	}

	exists, err := file_dir.IsExistsFile(this.withInfraDir + this.withInfraFile)
	if err != nil {
		this.logger.Fatalf(err.Error())
		return
	}

	if exists == false {
		this.logger.Fatalf("%s not exists", this.withInfraDir + this.withInfraFile)
	}
}

func (this *db2Entity) cloumnsToEntityTmp(columns []columns) entityTpl {

	entityTpl := entityTpl{}
	if len(columns) < 1 {
		return entityTpl
	}

	entityTpl.Imports = append(entityTpl.Imports, pkg.Import{Path: "github.com/jinzhu/gorm"})

	entityTpl.StructName = this.CamelStruct

	structInfo := templates.StructInfo{}

	for _, column := range columns {
		field := pkg.Field{}

		fieldName := templates.SnakeToCamel(column.ColumnName)
		field.Name = fieldName

		nullable := false
		if column.IsNullAble == "YES" {
			nullable = true
		}

		var valueType string
		valueType = this.mysqlTypeToGoType(column.DataType, nullable)
		if valueType == golangTime {
			entityTpl.Imports = append(entityTpl.Imports, pkg.Import{Path: "time"})
		} else if strings.Index(valueType, "sql.") != -1 {
			entityTpl.Imports = append(entityTpl.Imports, pkg.Import{Path: "database/sql"})
		}
		field.Type = valueType

		if column.ColumnDefault == "CURRENT_TIMESTAMP" {
			entityTpl.CurTimeStamp = append(entityTpl.CurTimeStamp, fieldName)
		}

		if column.Extra == "on update CURRENT_TIMESTAMP" {
			entityTpl.OnUpdateTimeStamp = append(entityTpl.OnUpdateTimeStamp, fieldName)
			entityTpl.OnUpdateTimeStampStr = append(entityTpl.OnUpdateTimeStampStr, column.ColumnName)
		}

		if column.ColumnComment != "" {
			column.ColumnComment = strings.Replace(column.ColumnComment, "\r", "\\r", -1)
			column.ColumnComment = strings.Replace(column.ColumnComment, "\n", "\\n", -1)
			field.Doc = append(field.Doc, "//" + column.ColumnComment)
		}

		primary := ""
		if column.ColumnKey == "PRI" {
			primary = ";primary_key"
		}

		col_default := ""
		if nullable == false {
			if column.ColumnDefault != "CURRENT_TIMESTAMP" && column.ColumnDefault != "" {
				col_default = ";default:'" + column.ColumnDefault + "'"
			}
		}

		field.Tag = fmt.Sprintf("`gorm:\"column:%s%s%s\"`", column.ColumnName, primary, col_default)

		entityTpl.DelField = this.checkDelField(column)

		field.Field = field.Name + " " + field.Type
		structInfo.Fields = append(structInfo.Fields, field)
	}

	structInfo.StructName = entityTpl.StructName

	entityTpl.StructInfo = structInfo

	return entityTpl
}

func (this *db2Entity) cloumnsToDaoTmp(columns []columns) daoTpl {
	daoTpl := daoTpl{}

	if len(columns) < 1 {
		return daoTpl
	}

	daoTpl.StructName = this.CamelStruct
	daoTpl.DataBaseName = this.dbConf.database
	daoTpl.TableName = this.dbConf.table

	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "context"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "github.com/jinzhu/gorm"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "errors"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "github.com/jukylin/esim/mysql"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: file_dir.GetGoProPath() + this.dirPathToImportPath(this.withEntityTarget)})

	for _, column := range columns {
		nullable := false
		if column.IsNullAble == "YES" {
			nullable = true
		}

		if column.ColumnKey == "PRI" {
			daoTpl.PriKeyType = this.mysqlTypeToGoType(column.DataType, nullable)
			break;
		}
	}

	return daoTpl
}

func (this *db2Entity) cloumnsToRepoTmp(columns []columns) repoTpl {
	repoTpl := repoTpl{}

	if len(columns) < 1 {
		return repoTpl
	}

	repoTpl.StructName = this.CamelStruct
	repoTpl.TableName = this.dbConf.table

	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: "context"})
	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: "github.com/jukylin/esim/log"})
	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: file_dir.GetGoProPath() + this.dirPathToImportPath(this.withEntityTarget)})
	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: file_dir.GetGoProPath() + this.dirPathToImportPath(this.withDaoTarget)})


	for _, column := range columns {
		repoTpl.DelField = this.checkDelField(column)
	}

	return repoTpl
}

//checkDelField check column.ColumnName contains "is" and "del" char
func (this *db2Entity) checkDelField(column columns) string {
	if strings.Index(column.ColumnName, "del") != -1 &&
		strings.Index(column.ColumnName, "is") != -1 {
		return column.ColumnName
	}

	return ""
}

//executeTmpl parse template
func (this *db2Entity) executeTmpl(tmplName string, data interface{}, text string) string {
	tmpl, err := template.New(tmplName).Funcs(templates.EsimFuncMap()).
		Parse(text)
	if err != nil {
		this.logger.Fatalf(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		this.logger.Fatalf(err.Error())
	}

	return buf.String()
}

// ./a/b/c/ => a/b/c
func (this *db2Entity) dirPathToImportPath(dirPath string) string {
	path := strings.TrimLeft(dirPath, ".")
	path = strings.TrimLeft(dirPath, "/")
	path = strings.TrimRight(path, "/")
	path = string(filepath.Separator) + path
	return path
}

//injectToInfra inject repo to infra.go and execute wire command
func (this *db2Entity) injectToInfra()  {
	//back up infra.go
	err := file_dir.EsimBackUpFile(file_dir.GetCurrentDir() + string(filepath.Separator) + this.withInfraDir + this.withInfraFile)
	if err != nil {
		this.logger.Fatalf(err.Error())
		return
	}

	beautifulSource := this.sourceInfraFile()

	this.parseInfra(beautifulSource)

	if this.hasInfraStruct {
		this.copyInfraInfo()

		this.processNewInfra()

		this.toStringNewInfra()

		this.buildNewInfraString()

		this.writeNewInfra()

	} else {
		this.logger.Fatalf("not found the %s", this.oldInfraInfo.specialStructName)
	}

	this.logger.Infof("inject success")
}

//parseInfra parse infra.go 's content, find "import", "Infra" , "infraSet" and record origin syntax
func (this *db2Entity) parseInfra(srcStr string) bool {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", srcStr, parser.ParseComments)
	if err != nil {
		this.logger.Fatalf(err.Error())
	}

	//provideFunc := getProvideFunc(interName, instName)
	for _, decl := range f.Decls {

		if GenDecl, ok := decl.(*ast.GenDecl); ok {
			if GenDecl.Tok.String() == "import" {
				imps := pkg.Imports{}
				imps.ParseFromAst(GenDecl)
				this.oldInfraInfo.imports = imps
				this.oldInfraInfo.importStr = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
			}

			if GenDecl.Tok.String() == "type" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.TypeSpec); ok {
						if typeSpec.Name.String() == this.oldInfraInfo.specialStructName {
							this.hasInfraStruct = true
							fields := pkg.Fields{}
							fields.ParseFromAst(GenDecl, srcStr)
							this.oldInfraInfo.structInfo.Fields = fields
							this.oldInfraInfo.structStr = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
						}
					}
				}
			}

			if GenDecl.Tok.String() == "var" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.ValueSpec); ok {
						for _, name := range typeSpec.Names {
							if name.String() == this.oldInfraInfo.specialVarName {
								this.oldInfraInfo.infraSetStr = srcStr[GenDecl.TokPos-1 : GenDecl.End()]
								this.oldInfraInfo.infraSetArgs.Args = append(this.oldInfraInfo.infraSetArgs.Args,
									this.getInfraSetArgs(GenDecl, srcStr)...)
							}
						}
					}
				}
			}
		}
	}

	if this.hasInfraStruct  == false {
		this.logger.Fatalf("not find %s", this.oldInfraInfo.specialStructName)
	}

	this.oldInfraInfo.content = srcStr

	//srcStr = strings.Replace(srcStr, oldImportStr, newImportStr, -1)
	//srcStr = strings.Replace(srcStr, oldStruct, newStruct, -1)
	//srcStr = strings.Replace(srcStr, oldSet, newSet, -1)
	//srcStr += provideFunc

	return true
}

//sourceInfraFile Beautify infra.go
func (this *db2Entity) sourceInfraFile() string {
	src, err := ioutil.ReadFile(this.withInfraDir + this.withInfraFile)
	if err != nil {
		this.logger.Fatalf(err.Error())
	}

	formatSrc, err := format.Source([]byte(src))
	if err != nil {
		this.logger.Fatalf(err.Error())
	}

	ioutil.WriteFile(this.withInfraDir + this.withInfraFile, formatSrc, 0666)

	return string(formatSrc)
}

func (this *db2Entity) copyInfraInfo()  {
	oldContent := *this.oldInfraInfo
	this.newInfraInfo = &oldContent
}

//processInfraInfo process newInfraInfo, append import, repo field and wire's provider
func (this *db2Entity) processNewInfra() bool {

	field := pkg.Field{}
	field.Name = this.CamelStruct + "Repo"
	field.Type = " repo." + this.CamelStruct + "Repo"
	field.Field = field.Name + " " + field.Type
	this.newInfraInfo.structInfo.Fields = append(this.newInfraInfo.structInfo.Fields, field)

	this.newInfraInfo.infraSetArgs.Args = append(this.newInfraInfo.infraSetArgs.Args,
		"provide" + this.CamelStruct + "Repo" + ",")

	imp := pkg.Import{Path: file_dir.GetGoProPath() + this.dirPathToImportPath(this.withRepoTarget)}
	this.newInfraInfo.imports = append(this.newInfraInfo.imports, imp)

	return true
}

func (this *db2Entity) toStringNewInfra() {

	this.newInfraInfo.importStr = this.newInfraInfo.imports.String()

	this.newInfraInfo.structStr = this.newInfraInfo.structInfo.String()

	this.newInfraInfo.infraSetStr = this.newInfraInfo.infraSetArgs.String()

}

func (this *db2Entity) buildNewInfraString() {

	oldContent := this.oldInfraInfo.content

	oldContent = strings.Replace(oldContent,
		this.oldInfraInfo.importStr, this.newInfraInfo.importStr, -1)

	oldContent = strings.Replace(oldContent,
		this.oldInfraInfo.structStr, this.newInfraInfo.structStr, -1)

	this.newInfraInfo.content = strings.Replace(oldContent,
		this.oldInfraInfo.infraSetStr, this.newInfraInfo.infraSetStr, -1)

	this.newInfraInfo.content += this.appendProvideFunc()
}

func (this *db2Entity) appendProvideFunc() string {
	return this.executeTmpl("provide_tpl", struct{StructName string}{this.CamelStruct}, provideTemplate)
}

func (this *db2Entity) writeNewInfra()  {

	sourceSrc, err := format.Source([]byte(this.newInfraInfo.content))
	if err != nil {
		this.logger.Fatalf(err.Error())
		return
	}

	processSrc, err := imports.Process("", sourceSrc, nil)
	if err != nil {
		this.logger.Fatalf(err.Error())
		return
	}

	this.writer.Write(this.withInfraDir + this.withInfraFile, string(processSrc))

	err = this.execer.ExecWire(this.withInfraDir)
	if err != nil {
		this.logger.Fatalf(err.Error())
	}
}

func (this *db2Entity) getInfraSetArgs(GenDecl *ast.GenDecl, srcStr string) []string {
	var args []string
	for _, specs := range GenDecl.Specs {
		if spec, ok := specs.(*ast.ValueSpec); ok {
			for _, value := range spec.Values {
				if callExpr, ok := value.(*ast.CallExpr); ok {
					for _, callArg := range callExpr.Args {
						args = append(args, pkg.ParseExpr(callArg, srcStr))
					}
				}
			}
		}
	}

	return args
}



