package domain_file

import (
	"path/filepath"
	"strings"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/pkg"
	"errors"
)

type repoDomainFile struct {

	withRepoTarget string

	withDisableRepo bool

	tplName string

	template string

	// data object of parsed template
	data *repoTpl

	logger log.Logger

	tpl templates.Tpl

	tableName string
}

type RepoDomainFileOption func(*repoDomainFile)

func NewRepoDomainFile(options ...RepoDomainFileOption) DomainFile {

	e := &repoDomainFile{}

	for _, option := range options {
		option(e)
	}

	e.tplName = "repo"

	e.template = repoTemplate

	return e
}

func WithRepoDomainFileLogger(logger log.Logger) RepoDomainFileOption {
	return func(e *repoDomainFile) {
		e.logger = logger
	}
}

func WithRepoDomainFileTpl(tpl templates.Tpl) RepoDomainFileOption {
	return func(e *repoDomainFile) {
		e.tpl = tpl
	}
}

//Disabled implements DomainFile.
func (rdf *repoDomainFile) Disabled() bool {
	return rdf.withDisableRepo
}

//BindInput implements DomainFile.
func (rdf *repoDomainFile) BindInput(v *viper.Viper) error {

	rdf.withDisableRepo = v.GetBool("disable_repo")
	if rdf.withDisableRepo == false {

		rdf.withRepoTarget = v.GetString("repo_target")
		if rdf.withRepoTarget == "" {
			rdf.withRepoTarget = "internal" + string(filepath.Separator) + "infra" + string(filepath.Separator) + "repo"
		} else {
			rdf.withRepoTarget = strings.TrimLeft(rdf.withRepoTarget, ".") + string(filepath.Separator)
			rdf.withRepoTarget = strings.Trim(rdf.withRepoTarget, "/")
		}

		//check repo dir
		existsRepo, err := file_dir.IsExistsDir(rdf.withRepoTarget)
		if err != nil {
			return err
		}

		if existsRepo == false {
			return errors.New("repo dir not exists")
		}

		rdf.withRepoTarget = rdf.withRepoTarget + string(filepath.Separator)

		rdf.logger.Debugf("withRepoTarget %s", rdf.withRepoTarget)
	}

	return nil
}

//ParseCloumns implements DomainFile.
func (rdf *repoDomainFile) ParseCloumns(cs []Column, info *ShareInfo) {

	repoTpl := NewRepoTpl(info.CamelStruct)

	if len(cs) < 1 {
		return
	}

	repoTpl.TableName = info.DbConf.Table
	rdf.tableName = info.DbConf.Table

	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: "context"})
	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: "github.com/jukylin/esim/log"})
	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: file_dir.GetGoProPath() + pkg.DirPathToImportPath(info.WithEntityTarget)})
	repoTpl.Imports = append(repoTpl.Imports, pkg.Import{Path: file_dir.GetGoProPath() + pkg.DirPathToImportPath(info.WithDaoTarget)})

	for _, column := range cs {
		repoTpl.DelField = column.CheckDelField()
	}

	rdf.data = repoTpl
}

//Execute implements DomainFile.
func (rdf *repoDomainFile) Execute() string {
	content, err := rdf.tpl.Execute(rdf.tplName, rdf.template, rdf.data)
	if err != nil {
		rdf.logger.Panicf(err.Error())
	}

	return content
}

//GetSavePath implements DomainFile.
func (rdf *repoDomainFile) GetSavePath() string  {
	return rdf.withRepoTarget + rdf.tableName + DOMAIN_File_EXT
}
