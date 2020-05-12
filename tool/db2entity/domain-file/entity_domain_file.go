package domain_file

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/serenize/snaker"
	"github.com/spf13/viper"
)

type entityDomainFile struct {
	writeTarget string

	withBoubctx string

	withEntityTarget string

	withDisbleEntity bool

	name string

	template string

	// data object of parsed template
	data entityTpl

	logger log.Logger

	tpl templates.Tpl

	tableName string
}

type EntityDomainFileOption func(*entityDomainFile)

func NewEntityDomainFile(options ...EntityDomainFileOption) DomainFile {

	e := &entityDomainFile{}

	for _, option := range options {
		option(e)
	}

	e.name = "entity"

	e.template = entityTemplate

	return e
}

func WithEntityDomainFileLogger(logger log.Logger) EntityDomainFileOption {
	return func(e *entityDomainFile) {
		e.logger = logger
	}
}

func WithEntityDomainFileTpl(tpl templates.Tpl) EntityDomainFileOption {
	return func(e *entityDomainFile) {
		e.tpl = tpl
	}
}

//Disabled implements DomainFile.
//EntityDomainFile never disable
func (edf *entityDomainFile) Disabled() bool {
	return edf.withDisbleEntity
}

//bindInput implements DomainFile.
func (edf *entityDomainFile) BindInput(v *viper.Viper) error {

	boubctx := v.GetString("boubctx")
	if boubctx != "" {
		edf.withBoubctx = boubctx + string(filepath.Separator)
	}

	edf.withDisbleEntity = v.GetBool("disable_entity")
	if !edf.withDisbleEntity {

		edf.withEntityTarget = v.GetString("entity_target")

		if edf.withEntityTarget == "" {
			if edf.withBoubctx != "" {
				edf.withEntityTarget = "internal" + string(filepath.Separator) + "domain" +
					string(filepath.Separator) + edf.withBoubctx + "entity"
			} else {
				edf.withEntityTarget = "internal" + string(filepath.Separator) + "domain" +
					string(filepath.Separator) + "entity"
			}
		} else {
			edf.withEntityTarget = strings.TrimLeft(edf.withEntityTarget, "/")
			edf.withEntityTarget = edf.withBoubctx + edf.withEntityTarget
		}

		entityTargetExists, err := file_dir.IsExistsDir(edf.withEntityTarget)
		if err != nil {
			return err
		}

		if !entityTargetExists {
			err = file_dir.CreateDir(edf.withEntityTarget)
			if err != nil {
				return err
			}
		}

		edf.withEntityTarget = edf.withEntityTarget + string(filepath.Separator)
	}

	edf.logger.Debugf("withEntityTarget %s", edf.withEntityTarget)

	return nil
}

//parseCloumns implements DomainFile.
func (edf *entityDomainFile) ParseCloumns(cs Columns, info *ShareInfo) {

	entityTpl := entityTpl{}

	if cs.Len() < 1 {
		return
	}

	edf.tableName = info.DbConf.Table

	entityTpl.Imports = append(entityTpl.Imports, pkg.Import{Path: "github.com/jinzhu/gorm"})

	entityTpl.StructName = info.CamelStruct

	structInfo := templates.StructInfo{}

	var colDefault string
	var valueType string
	var doc string
	var nullable bool
	var fieldName string

	for _, column := range cs {

		field := pkg.Field{}

		fieldName = snaker.SnakeToCamel(column.ColumnName)
		field.Name = fieldName

		if column.IsNullAble == "YES" {
			nullable = true
		}

		valueType = column.GetGoType(nullable)
		if column.IsTime(valueType) {
			entityTpl.Imports = append(entityTpl.Imports, pkg.Import{Path: "time"})
		} else if strings.Contains(valueType, "sql.") {
			entityTpl.Imports = append(entityTpl.Imports, pkg.Import{Path: "database/sql"})
		}
		field.Type = valueType

		if column.IsCurrentTimeStamp() {
			entityTpl.CurTimeStamp = append(entityTpl.CurTimeStamp, fieldName)
		}

		if column.IsOnUpdate() {
			entityTpl.OnUpdateTimeStamp = append(entityTpl.OnUpdateTimeStamp, fieldName)
			entityTpl.OnUpdateTimeStampStr = append(entityTpl.OnUpdateTimeStampStr, column.ColumnName)
		}

		doc = column.FilterComment()
		if doc != "" {
			field.Doc = append(field.Doc, "//"+doc)
		}

		primary := ""
		if column.IsPri() {
			primary = ";primary_key"
		}

		if !nullable {
			colDefault = column.GetDefCol()
		}

		field.Tag = fmt.Sprintf("`gorm:\"column:%s%s%s\"`", column.ColumnName, primary, colDefault)

		entityTpl.DelField = column.CheckDelField()

		field.Field = field.Name + " " + field.Type
		structInfo.Fields = append(structInfo.Fields, field)

		colDefault = ""
		nullable = false
	}

	structInfo.StructName = entityTpl.StructName

	entityTpl.StructInfo = structInfo

	edf.data = entityTpl
}

//execute implements DomainFile.
func (edf *entityDomainFile) Execute() string {
	content, err := edf.tpl.Execute(edf.name, edf.template, edf.data)
	if err != nil {
		edf.logger.Panicf(err.Error())
	}

	return content
}

//getSavePath implements DomainFile.
func (edf *entityDomainFile) GetSavePath() string {
	return edf.withEntityTarget + edf.tableName + DOMAIN_FILE_EXT
}

func (edf *entityDomainFile) GetName() string {
	return edf.name
}

func (edf *entityDomainFile) GetInjectInfo() *InjectInfo {
	return NewInjectInfo()
}
