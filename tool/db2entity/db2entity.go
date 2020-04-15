package db2entity

import (
	"bytes"
	"fmt"
	logger "github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	file_dir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/spf13/viper"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"
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

func NewDb2EntityOptions() Db2EntityOptions {
	return Db2EntityOptions{}
}

func NewDb2Entity(options ...Db2EntityOption) *db2Entity {

	d := &db2Entity{}

	for _, option := range options {
		option(d)
	}

	if d.writer == nil {
		d.writer = file_dir.NewNullWrite()
	}

	if d.execer == nil {
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

type infraInfo struct {
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

func (de *db2Entity) Run(v *viper.Viper) error {

	de.bindInput(v)

	columns, err := de.ColumnsRepo.GetColumns(de.dbConf)
	if err != nil {
		de.logger.Fatalf(err.Error())
	}

	entityTmp := de.cloumnsToEntityTmp(columns)
	entityContent := de.executeTmpl("entity_tpl", entityTmp, entityTemplate)
	de.writer.Write(de.withEntityTarget+de.dbConf.table+".go", entityContent)

	daoTmp := de.cloumnsToDaoTmp(columns)
	daoContent := de.executeTmpl("entity_tpl", daoTmp, daoTemplate)
	de.writer.Write(de.withDaoTarget+de.dbConf.table+".go", daoContent)

	repoTmp := de.cloumnsToRepoTmp(columns)
	repoContent := de.executeTmpl("entity_tpl", repoTmp, repoTemplate)
	de.writer.Write(de.withRepoTarget+de.dbConf.table+".go", repoContent)

	de.injectToInfra()

	return nil
}

func (de *db2Entity) bindInput(v *viper.Viper) {

	de.bindDbConfig(v)

	packageName := v.GetString("package")
	if packageName == "" {
		packageName = de.dbConf.database
	}
	de.withPackage = packageName

	stuctName := v.GetString("struct")
	if stuctName == "" {
		stuctName = de.dbConf.table
	}
	de.withStruct = stuctName
	de.CamelStruct = templates.SnakeToCamel(stuctName)

	boubctx := v.GetString("boubctx")
	if boubctx != "" {
		de.withBoubctx = boubctx + string(filepath.Separator)
	}

	de.bindEntityDir(v)

	de.bindRepoDir(v)

	de.bindDaoDir(v)

	de.bindInfra(v)
}

func (de *db2Entity) bindEntityDir(v *viper.Viper) {
	if v.GetBool("disabled_entity") == false {

		de.withEntityTarget = v.GetString("entity_target")

		if de.withEntityTarget == "" {
			if de.withBoubctx != "" {
				de.withEntityTarget = "internal" + string(filepath.Separator) + "domain" + string(filepath.Separator) + de.withBoubctx + "entity"
			} else {
				de.withEntityTarget = "internal" + string(filepath.Separator) + "domain" + string(filepath.Separator) + "entity"
			}
		}

		entityTargetExists, err := file_dir.IsExistsDir(de.withEntityTarget)
		if err != nil {
			de.logger.Fatalf(err.Error())
		}

		if entityTargetExists == false {
			err = file_dir.CreateDir(de.withEntityTarget)
			if err != nil {
				de.logger.Fatalf(err.Error())
			}
		}

		de.withEntityTarget = de.withEntityTarget + string(filepath.Separator)
	}
}

func (de *db2Entity) bindRepoDir(v *viper.Viper) {
	de.withDisabledRepo = v.GetBool("disabled_repo")
	if de.withDisabledRepo == false {

		de.withRepoTarget = v.GetString("repo_target")
		if de.withRepoTarget == "" {
			de.withRepoTarget = "internal" + string(filepath.Separator) + "infra" + string(filepath.Separator) + "repo"
		} else {
			de.withRepoTarget = strings.TrimLeft(de.withRepoTarget, ".") + string(filepath.Separator)
			de.withRepoTarget = strings.Trim(de.withRepoTarget, "/") + string(filepath.Separator)
		}

		//repo 目录是否存在
		existsRepo, err := file_dir.IsExistsDir(de.withRepoTarget)
		if err != nil {
			de.logger.Fatalf(err.Error())
		}

		if existsRepo == false {
			de.logger.Fatalf("repo dir not exists")
		}
		de.withRepoTarget = de.withRepoTarget + string(filepath.Separator)
	}
}

func (de *db2Entity) bindDaoDir(v *viper.Viper) {
	de.withDisabledDao = v.GetBool("disabled_dao")
	if de.withDisabledDao == false {

		de.withDaoTarget = v.GetString("dao_target")
		if de.withDaoTarget == "" {
			de.withDaoTarget = "internal" + string(filepath.Separator) + "infra " + string(filepath.Separator) + "dao"
		} else {
			de.withDaoTarget = strings.TrimLeft(de.withDaoTarget, ".") + string(filepath.Separator)
			de.withDaoTarget = strings.Trim(de.withDaoTarget, "/") + string(filepath.Separator)
		}

		//dao 目录是否存在
		existsdao, err := file_dir.IsExistsDir(de.withDaoTarget)
		if err != nil {
			de.logger.Fatalf(err.Error())
		}

		if existsdao == false {
			de.logger.Fatalf("dao dir not exists")
		}
		de.withDaoTarget = de.withDaoTarget + string(filepath.Separator)
	}
}

func (de *db2Entity) bindDbConfig(v *viper.Viper) {
	dbConfig := dbConfig{}
	host := v.GetString("host")
	if host == "" {
		de.logger.Fatalf("host is empty")
	}
	dbConfig.host = host

	port := v.GetInt("port")
	if port == 0 {
		de.logger.Fatalf("port is 0")
	}
	dbConfig.port = port

	user := v.GetString("user")
	if user == "" {
		de.logger.Fatalf("user is empty")
	}
	dbConfig.user = user

	password := v.GetString("password")
	dbConfig.password = password

	database := v.GetString("database")
	if database == "" {
		de.logger.Fatalf("database is empty")
	}
	dbConfig.database = database

	table := v.GetString("table")
	if table == "" {
		de.logger.Fatalf("table is empty")
	}
	dbConfig.table = table

	de.dbConf = dbConfig
}

func (de *db2Entity) bindInfra(v *viper.Viper) {
	de.withInject = v.GetBool("inject")

	de.withInfraDir = v.GetString("infra_dir")
	if de.withInfraDir == "" {
		de.withInfraDir = "internal" + string(filepath.Separator) + "infra" + string(filepath.Separator)
	} else {
		de.withInfraDir = strings.TrimLeft(de.withInfraDir, ".") + string(filepath.Separator)
		de.withInfraDir = strings.Trim(de.withInfraDir, "/") + string(filepath.Separator)
	}

	if v.GetString("infra_file") == "" {
		de.withInfraFile = "infra.go"
	}

	exists, err := file_dir.IsExistsFile(de.withInfraDir + de.withInfraFile)
	if err != nil {
		de.logger.Fatalf(err.Error())
		return
	}

	if exists == false {
		de.logger.Fatalf("%s not exists", de.withInfraDir+de.withInfraFile)
	}
}

func (de *db2Entity) cloumnsToEntityTmp(columns []columns) entityTpl {

	entityTpl := entityTpl{}
	if len(columns) < 1 {
		return entityTpl
	}

	entityTpl.Imports = append(entityTpl.Imports, pkg.Import{Path: "github.com/jinzhu/gorm"})

	entityTpl.StructName = de.CamelStruct

	structInfo := templates.StructInfo{}

	for _, column := range columns {
		var colDefault string
		var valueType string

		field := pkg.Field{}

		fieldName := templates.SnakeToCamel(column.ColumnName)
		field.Name = fieldName

		nullable := false
		if column.IsNullAble == "YES" {
			nullable = true
		}

		valueType = de.mysqlTypeToGoType(column.DataType, nullable)
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
			field.Doc = append(field.Doc, "//"+column.ColumnComment)
		}

		primary := ""
		if column.ColumnKey == "PRI" {
			primary = ";primary_key"
		}

		if nullable == false {
			if column.ColumnDefault != "CURRENT_TIMESTAMP" && column.ColumnDefault != "" {
				colDefault = ";default:'" + column.ColumnDefault + "'"
			}
		}

		field.Tag = fmt.Sprintf("`gorm:\"column:%s%s%s\"`", column.ColumnName, primary, colDefault)

		entityTpl.DelField = de.checkDelField(column)

		field.Field = field.Name + " " + field.Type
		structInfo.Fields = append(structInfo.Fields, field)
	}

	structInfo.StructName = entityTpl.StructName

	entityTpl.StructInfo = structInfo

	return entityTpl
}

func (de *db2Entity) cloumnsToDaoTmp(columns []columns) daoTpl {
	daoTpl := daoTpl{}

	if len(columns) < 1 {
		return daoTpl
	}

	daoTpl.StructName = de.CamelStruct
	daoTpl.DataBaseName = de.dbConf.database
	daoTpl.TableName = de.dbConf.table

	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "context"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "github.com/jinzhu/gorm"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "errors"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "github.com/jukylin/esim/mysql"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: file_dir.GetGoProPath() + de.dirPathToImportPath(de.withEntityTarget)})

	for _, column := range columns {
		nullable := false
		if column.IsNullAble == "YES" {
			nullable = true
		}

		if column.ColumnKey == "PRI" {
			daoTpl.PriKeyType = de.mysqlTypeToGoType(column.DataType, nullable)
			break
		}
	}

	return daoTpl
}

func (de *db2Entity) cloumnsToRepoTmp(columns []columns) repoTpl {
	repoTpl := repoTpl{}

	if len(columns) < 1 {
		return repoTpl
	}

	repoTpl.StructName = de.CamelStruct
	repoTpl.TableName = de.dbConf.table

	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: "context"})
	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: "github.com/jukylin/esim/log"})
	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: file_dir.GetGoProPath() + de.dirPathToImportPath(de.withEntityTarget)})
	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: file_dir.GetGoProPath() + de.dirPathToImportPath(de.withDaoTarget)})

	for _, column := range columns {
		repoTpl.DelField = de.checkDelField(column)
	}

	return repoTpl
}

//checkDelField check column.ColumnName contains "is" and "del" char
func (de *db2Entity) checkDelField(column columns) string {
	if strings.Index(column.ColumnName, "del") != -1 &&
		strings.Index(column.ColumnName, "is") != -1 {
		return column.ColumnName
	}

	return ""
}

//executeTmpl parse template
func (de *db2Entity) executeTmpl(tmplName string, data interface{}, text string) string {
	tmpl, err := template.New(tmplName).Funcs(templates.EsimFuncMap()).
		Parse(text)
	if err != nil {
		de.logger.Fatalf(err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		de.logger.Fatalf(err.Error())
	}

	return buf.String()
}

// ./a/b/c/ => a/b/c
func (de *db2Entity) dirPathToImportPath(dirPath string) string {
	path := strings.TrimLeft(dirPath, ".")
	path = strings.TrimLeft(dirPath, "/")
	path = strings.TrimRight(path, "/")
	path = string(filepath.Separator) + path
	return path
}

//injectToInfra inject repo to infra.go and execute wire command
func (de *db2Entity) injectToInfra() {
	//back up infra.go
	err := file_dir.EsimBackUpFile(file_dir.GetCurrentDir() + string(filepath.Separator) + de.withInfraDir + de.withInfraFile)
	if err != nil {
		de.logger.Fatalf(err.Error())
		return
	}

	beautifulSource := de.sourceInfraFile()

	de.parseInfra(beautifulSource)

	if de.hasInfraStruct {
		de.copyInfraInfo()

		de.processNewInfra()

		de.toStringNewInfra()

		de.buildNewInfraString()

		de.writeNewInfra()

	} else {
		de.logger.Fatalf("not found the %s", de.oldInfraInfo.specialStructName)
	}

	de.logger.Infof("inject success")
}

//parseInfra parse infra.go 's content, find "import", "Infra" , "infraSet" and record origin syntax
func (de *db2Entity) parseInfra(srcStr string) bool {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", srcStr, parser.ParseComments)
	if err != nil {
		de.logger.Fatalf(err.Error())
	}

	//provideFunc := getProvideFunc(interName, instName)
	for _, decl := range f.Decls {

		if GenDecl, ok := decl.(*ast.GenDecl); ok {
			if GenDecl.Tok.String() == "import" {
				imps := pkg.Imports{}
				imps.ParseFromAst(GenDecl)
				de.oldInfraInfo.imports = imps
				de.oldInfraInfo.importStr = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
			}

			if GenDecl.Tok.String() == "type" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.TypeSpec); ok {
						if typeSpec.Name.String() == de.oldInfraInfo.specialStructName {
							de.hasInfraStruct = true
							fields := pkg.Fields{}
							fields.ParseFromAst(GenDecl, srcStr)
							de.oldInfraInfo.structInfo.Fields = fields
							de.oldInfraInfo.structStr = srcStr[GenDecl.Pos()-1 : GenDecl.End()]
						}
					}
				}
			}

			if GenDecl.Tok.String() == "var" {
				for _, specs := range GenDecl.Specs {
					if typeSpec, ok := specs.(*ast.ValueSpec); ok {
						for _, name := range typeSpec.Names {
							if name.String() == de.oldInfraInfo.specialVarName {
								de.oldInfraInfo.infraSetStr = srcStr[GenDecl.TokPos-1 : GenDecl.End()]
								de.oldInfraInfo.infraSetArgs.Args = append(de.oldInfraInfo.infraSetArgs.Args,
									de.getInfraSetArgs(GenDecl, srcStr)...)
							}
						}
					}
				}
			}
		}
	}

	if de.hasInfraStruct == false {
		de.logger.Fatalf("not find %s", de.oldInfraInfo.specialStructName)
	}

	de.oldInfraInfo.content = srcStr

	//srcStr = strings.Replace(srcStr, oldImportStr, newImportStr, -1)
	//srcStr = strings.Replace(srcStr, oldStruct, newStruct, -1)
	//srcStr = strings.Replace(srcStr, oldSet, newSet, -1)
	//srcStr += provideFunc

	return true
}

//sourceInfraFile Beautify infra.go
func (de *db2Entity) sourceInfraFile() string {
	src, err := ioutil.ReadFile(de.withInfraDir + de.withInfraFile)
	if err != nil {
		de.logger.Fatalf(err.Error())
	}

	formatSrc, err := format.Source([]byte(src))
	if err != nil {
		de.logger.Fatalf(err.Error())
	}

	ioutil.WriteFile(de.withInfraDir+de.withInfraFile, formatSrc, 0666)

	return string(formatSrc)
}

func (de *db2Entity) copyInfraInfo() {
	oldContent := *de.oldInfraInfo
	de.newInfraInfo = &oldContent
}

//processInfraInfo process newInfraInfo, append import, repo field and wire's provider
func (de *db2Entity) processNewInfra() bool {

	field := pkg.Field{}
	field.Name = de.CamelStruct + "Repo"
	field.Type = " repo." + de.CamelStruct + "Repo"
	field.Field = field.Name + " " + field.Type
	de.newInfraInfo.structInfo.Fields = append(de.newInfraInfo.structInfo.Fields, field)

	de.newInfraInfo.infraSetArgs.Args = append(de.newInfraInfo.infraSetArgs.Args,
		"provide"+de.CamelStruct+"Repo"+",")

	imp := pkg.Import{Path: file_dir.GetGoProPath() + de.dirPathToImportPath(de.withRepoTarget)}
	de.newInfraInfo.imports = append(de.newInfraInfo.imports, imp)

	return true
}

func (de *db2Entity) toStringNewInfra() {

	de.newInfraInfo.importStr = de.newInfraInfo.imports.String()

	de.newInfraInfo.structStr = de.newInfraInfo.structInfo.String()

	de.newInfraInfo.infraSetStr = de.newInfraInfo.infraSetArgs.String()

}

func (de *db2Entity) buildNewInfraString() {

	oldContent := de.oldInfraInfo.content

	oldContent = strings.Replace(oldContent,
		de.oldInfraInfo.importStr, de.newInfraInfo.importStr, -1)

	oldContent = strings.Replace(oldContent,
		de.oldInfraInfo.structStr, de.newInfraInfo.structStr, -1)

	de.newInfraInfo.content = strings.Replace(oldContent,
		de.oldInfraInfo.infraSetStr, de.newInfraInfo.infraSetStr, -1)

	de.newInfraInfo.content += de.appendProvideFunc()
}

func (de *db2Entity) appendProvideFunc() string {
	return de.executeTmpl("provide_tpl", struct{ StructName string }{de.CamelStruct}, provideTemplate)
}

func (de *db2Entity) writeNewInfra() {

	sourceSrc, err := format.Source([]byte(de.newInfraInfo.content))
	if err != nil {
		de.logger.Fatalf(err.Error())
		return
	}

	processSrc, err := imports.Process("", sourceSrc, nil)
	if err != nil {
		de.logger.Fatalf(err.Error())
		return
	}

	de.writer.Write(de.withInfraDir+de.withInfraFile, string(processSrc))

	err = de.execer.ExecWire(de.withInfraDir)
	if err != nil {
		de.logger.Fatalf(err.Error())
	}
}

func (de *db2Entity) getInfraSetArgs(GenDecl *ast.GenDecl, srcStr string) []string {
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

func (de *db2Entity) mysqlTypeToGoType(mysqlType string, nullable bool) string {
	switch mysqlType {
	case "tinyint", "int", "smallint", "mediumint":
		if nullable {
			return sqlNullInt
		}
		return golangInt
	case "bigint":
		if nullable {
			return sqlNullInt
		}
		return golangInt64
	case "char", "enum", "varchar", "longtext", "mediumtext", "text", "tinytext":
		if nullable {
			return sqlNullString
		}
		return "string"
	case "date", "datetime", "time", "timestamp":
		return golangTime
	case "decimal", "double":
		if nullable {
			return sqlNullFloat
		}
		return golangFloat64
	case "float":
		if nullable {
			return sqlNullFloat
		}
		return golangFloat32
	case "binary", "blob", "longblob", "mediumblob", "varbinary":
		return golangByteArray
	}
	return ""
}
