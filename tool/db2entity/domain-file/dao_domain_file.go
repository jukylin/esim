package domain_file

import (
	"path/filepath"
	"strings"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/tool/db2entity"
	"errors"
)

type daoDomainFile struct {
	writeTarget string

	withBoubctx string

	withDaoTarget string

	withDisableDao bool

	tplName string

	template string

	// data object of parsed template
	data *daoTpl

	logger log.Logger

	tpl templates.Tpl

	tableName string
}

type DaoDomainFileOption func(*daoDomainFile)

func NewDaoDomainFile(options ...DaoDomainFileOption) DomainFile {

	e := &daoDomainFile{}

	for _, option := range options {
		option(e)
	}

	e.tplName = "dao"

	e.template = daoTemplate

	return e
}

func WithDaoDomainFileLogger(logger log.Logger) DaoDomainFileOption {
	return func(e *daoDomainFile) {
		e.logger = logger
	}
}

func WithDaoDomainFileTpl(tpl templates.Tpl) DaoDomainFileOption {
	return func(e *daoDomainFile) {
		e.tpl = tpl
	}
}

//Disabled implements DomainFile.
func (ddf *daoDomainFile) Disabled() bool {
	return ddf.withDisableDao
}

//BindInput implements DomainFile.
func (ddf *daoDomainFile) BindInput(v *viper.Viper) error {

	ddf.tableName = v.GetString("table")
	if ddf.tableName == "" {
		return errors.New("table is empty")
	}

	ddf.withDisableDao = v.GetBool("disable_dao")
	if ddf.withDisableDao == false {

		ddf.withDaoTarget = v.GetString("dao_target")
		if ddf.withDaoTarget == "" {
			ddf.withDaoTarget = "internal" + string(filepath.Separator) + "infra " + string(filepath.Separator) + "dao"
		} else {
			ddf.withDaoTarget = strings.TrimLeft(ddf.withDaoTarget, ".") + string(filepath.Separator)
			ddf.withDaoTarget = strings.Trim(ddf.withDaoTarget, "/")
		}

		//check dao dir
		existsdao, err := file_dir.IsExistsDir(ddf.withDaoTarget)
		if err != nil {
			return err
		}

		if existsdao == false {
			return errors.New("dao dir not exists")
		}

		ddf.withDaoTarget = ddf.withDaoTarget + string(filepath.Separator)

		ddf.logger.Debugf("withDaoTarget %s", ddf.withDaoTarget)
	}

	return nil
}

//ParseCloumns implements DomainFile.
func (ddf *daoDomainFile) ParseCloumns(cs []Column, d2e *db2entity.Db2Entity) {

	daoTpl := NewDaoTpl(d2e.CamelStruct)

	if len(cs) < 1 {
		return
	}

	daoTpl.DataBaseName = d2e.DbConf.Database
	daoTpl.TableName = d2e.DbConf.Table

	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "context"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "github.com/jinzhu/gorm"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "errors"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "github.com/jukylin/esim/mysql"})
	daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: file_dir.GetGoProPath() + d2e.DirPathToImportPath(d2e.WithEntityTarget)})

	for _, column := range cs {
		nullable := false
		if column.IsNullAble == "YES" {
			nullable = true
		}

		if column.ColumnKey == "PRI" {
			daoTpl.PriKeyType = column.GetGoType(nullable)
			break
		}
	}

	ddf.data = daoTpl
}

//Execute implements DomainFile.
func (ddf *daoDomainFile) Execute() string {
	content, err := ddf.tpl.Execute(ddf.tplName, ddf.template, ddf.data)
	if err != nil {
		ddf.logger.Panicf(err.Error())
	}

	return content
}

//GetSavePath implements DomainFile.
func (ddf *daoDomainFile) GetSavePath() string  {
	return ddf.withDaoTarget + ddf.tableName + DOMAIN_File_EXT
}
