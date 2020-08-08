package domainfile

import (
	"path/filepath"
	"strings"

	"errors"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/serenize/snaker"
	"github.com/spf13/viper"
)

type daoDomainFile struct {
	withDaoTarget string

	withDisableDao bool

	name string

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

	e.name = "dao"

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

// Disabled implements DomainFile.
func (ddf *daoDomainFile) Disabled() bool {
	return ddf.withDisableDao
}

// BindInput implements DomainFile.
//nolint:dupl
func (ddf *daoDomainFile) BindInput(v *viper.Viper) error {
	ddf.withDisableDao = v.GetBool("disable_dao")
	if !ddf.withDisableDao {
		ddf.withDaoTarget = v.GetString("dao_target")
		if ddf.withDaoTarget == "" {
			ddf.withDaoTarget = "internal" + string(filepath.Separator) +
				"infra " + string(filepath.Separator) + "dao"
		} else {
			ddf.withDaoTarget = strings.TrimLeft(ddf.withDaoTarget, ".") +
				string(filepath.Separator)
			ddf.withDaoTarget = strings.Trim(ddf.withDaoTarget, "/")
		}

		// check dao dir
		existsdao, err := filedir.IsExistsDir(ddf.withDaoTarget)
		if err != nil {
			return err
		}

		if !existsdao {
			return errors.New("dao dir not exists")
		}

		ddf.withDaoTarget += string(filepath.Separator)

		ddf.logger.Debugf("withDaoTarget %s", ddf.withDaoTarget)
	}

	return nil
}

// ParseCloumns implements DomainFile.
func (ddf *daoDomainFile) ParseCloumns(cs Columns, shareInfo *ShareInfo) {
	daoTpl := newDaoTpl(shareInfo.CamelStruct)

	if cs.Len() == 0 {
		return
	}

	daoTpl.DataBaseName = shareInfo.DbConf.Database
	daoTpl.TableName = shareInfo.DbConf.Table
	ddf.tableName = shareInfo.DbConf.Table

	daoTpl.Imports = append(daoTpl.Imports,
		pkg.Import{Path: "context"},
		pkg.Import{Path: "github.com/jinzhu/gorm"},
		pkg.Import{Path: "errors"},
		pkg.Import{Path: "github.com/jukylin/esim/mysql"},
		pkg.Import{Path: filedir.GetGoProPath() +
			pkg.DirPathToImportPath(shareInfo.WithEntityTarget)})

	var hastTime bool
	for i := range cs {
		column := (&cs[i])

		fieldName := snaker.SnakeToCamel(column.ColumnName)

		nullable := false
		if column.IsNullAble == yesNull {
			nullable = true
		}

		if column.ColumnKey == pri {
			daoTpl.PriKeyType = column.GetGoType(nullable)
		}

		if column.IsOnUpdate() {
			hastTime = true
			daoTpl.OnUpdateTimeStamp = append(daoTpl.OnUpdateTimeStamp, fieldName)
			daoTpl.OnUpdateTimeStampStr = append(daoTpl.OnUpdateTimeStampStr,
				column.ColumnName)
		} else if column.IsCurrentTimeStamp() {
			hastTime = true
			daoTpl.CurTimeStamp = append(daoTpl.CurTimeStamp, fieldName)
		}
	}

	if hastTime {
		daoTpl.Imports = append(daoTpl.Imports, pkg.Import{Path: "time"})
	}

	ddf.data = daoTpl
}

// Execute implements DomainFile.
func (ddf *daoDomainFile) Execute() string {
	content, err := ddf.tpl.Execute(ddf.name, ddf.template, ddf.data)
	if err != nil {
		ddf.logger.Panicf(err.Error())
	}

	return content
}

// GetSavePath implements DomainFile.
func (ddf *daoDomainFile) GetSavePath() string {
	return ddf.withDaoTarget + ddf.tableName + DomainFileExt
}

func (ddf *daoDomainFile) GetName() string {
	return ddf.name
}

func (ddf *daoDomainFile) GetInjectInfo() *InjectInfo {
	return NewInjectInfo()
}
