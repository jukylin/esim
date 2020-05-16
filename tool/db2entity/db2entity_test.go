package db2entity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jukylin/esim/infra"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	file_dir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	domain_file "github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestDb2Entity_Run(t *testing.T) {

	v := viper.New()
	v.Set("entity_target", "./example/entity")
	v.Set("dao_target", "./example/dao")
	v.Set("disable_entity", true)
	v.Set("repo_target", "./example/repo")
	v.Set("infra_dir", "./example/infra")

	v.Set("host", "127.0.0.1")
	v.Set("inject", true)
	v.Set("port", "3306")
	v.Set("user", "root")
	v.Set("passport", "")
	v.Set("database", "user")
	v.Set("table", "test_history")

	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	tpl := templates.NewTextTpl()

	dbConf := domain_file.NewDbConfig()
	dbConf.ParseConfig(v, logger)

	daoDomainFile := domain_file.NewDaoDomainFile(
		domain_file.WithDaoDomainFileLogger(logger),
		domain_file.WithDaoDomainFileTpl(tpl),
	)

	entityDomainFile := domain_file.NewEntityDomainFile(
		domain_file.WithEntityDomainFileLogger(logger),
		domain_file.WithEntityDomainFileTpl(tpl),
	)

	repoDomainFile := domain_file.NewRepoDomainFile(
		domain_file.WithRepoDomainFileLogger(logger),
		domain_file.WithRepoDomainFileTpl(tpl),
	)

	db2EntityOptions := Db2EnOptions{}
	StubsColumnsRepo := domain_file.StubsColumnsRepo{}

	shareInfo := domain_file.NewShareInfo()
	shareInfo.DbConf = dbConf

	writer := file_dir.NewEsimWriter()

	db2Entity := NewDb2Entity(
		db2EntityOptions.WithLogger(logger),
		db2EntityOptions.WithColumnsInter(StubsColumnsRepo),
		db2EntityOptions.WithWriter(writer),
		db2EntityOptions.WithExecer(pkg.NewNullExec()),
		db2EntityOptions.WithDbConf(domain_file.NewDbConfig()),
		db2EntityOptions.WithDomainFile(entityDomainFile, daoDomainFile, repoDomainFile),
		db2EntityOptions.WithShareInfo(shareInfo),
		db2EntityOptions.WithTpl(templates.NewTextTpl()),
		db2EntityOptions.WithDbConf(dbConf),
		db2EntityOptions.WithInfraer(infra.NewInfraer(
			infra.WithIfacerInfraInfo(infra.NewInfo()),
			infra.WithIfacerLogger(logger),
			infra.WithIfacerWriter(writer),
			infra.WithIfacerExecer(pkg.NewNullExec()),
		)),
	)

	err := db2Entity.Run(v)
	assert.Nil(t, err)

	for path, content := range db2Entity.domainContent {
		_ = content
		assert.FileExists(t, path)
	}

	for path := range db2Entity.domainContent {
		os.Remove(path)
	}

	err = file_dir.EsimRecoverFile(file_dir.GetCurrentDir() +
		string(filepath.Separator) + "example" + string(filepath.Separator) + "infra" +
		string(filepath.Separator) + "infra.go")
	assert.Nil(t, err)
}

func TestDb2Entity_ErrWrite(t *testing.T) {

	v := viper.New()
	v.Set("entity_target", "./example/entity")
	v.Set("dao_target", "./example/dao")
	v.Set("repo_target", "./example/repo")
	v.Set("infra_dir", "./example/infra")

	v.Set("host", "127.0.0.1")
	v.Set("inject", true)
	v.Set("port", "3306")
	v.Set("user", "root")
	v.Set("passport", "")
	v.Set("database", "user")
	v.Set("table", "test_history")

	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
	tpl := templates.NewTextTpl()

	dbConf := domain_file.NewDbConfig()
	dbConf.ParseConfig(v, logger)

	daoDomainFile := domain_file.NewDaoDomainFile(
		domain_file.WithDaoDomainFileLogger(logger),
		domain_file.WithDaoDomainFileTpl(tpl),
	)

	entityDomainFile := domain_file.NewEntityDomainFile(
		domain_file.WithEntityDomainFileLogger(logger),
		domain_file.WithEntityDomainFileTpl(tpl),
	)

	repoDomainFile := domain_file.NewRepoDomainFile(
		domain_file.WithRepoDomainFileLogger(logger),
		domain_file.WithRepoDomainFileTpl(tpl),
	)

	db2EntityOptions := Db2EnOptions{}
	StubsColumnsRepo := domain_file.StubsColumnsRepo{}

	shareInfo := domain_file.NewShareInfo()
	shareInfo.DbConf = dbConf

	db2Entity := NewDb2Entity(
		db2EntityOptions.WithLogger(logger),
		db2EntityOptions.WithColumnsInter(StubsColumnsRepo),
		db2EntityOptions.WithWriter(file_dir.NewErrWrite(2)),
		db2EntityOptions.WithExecer(pkg.NewNullExec()),
		db2EntityOptions.WithDbConf(domain_file.NewDbConfig()),
		db2EntityOptions.WithDomainFile(entityDomainFile, daoDomainFile, repoDomainFile),
		db2EntityOptions.WithShareInfo(shareInfo),
		db2EntityOptions.WithTpl(templates.NewTextTpl()),
		db2EntityOptions.WithDbConf(dbConf),
		db2EntityOptions.WithInfraer(infra.NewInfraer(
			infra.WithIfacerInfraInfo(infra.NewInfo()),
			infra.WithIfacerLogger(logger),
			infra.WithIfacerWriter(file_dir.NewNullWrite()),
			infra.WithIfacerExecer(pkg.NewNullExec()),
		)),
	)

	err := db2Entity.Run(v)
	assert.Nil(t, err)

	err = file_dir.EsimRecoverFile(file_dir.GetCurrentDir() +
		string(filepath.Separator) + "example" + string(filepath.Separator) + "infra" +
		string(filepath.Separator) + "infra.go")
	assert.Nil(t, err)
}
