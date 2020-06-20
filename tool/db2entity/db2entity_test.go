package db2entity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jukylin/esim/infra"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	domainfile "github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestDb2Entity_Run(t *testing.T) {
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

	logger := log.NewLogger(log.WithDebug(true))
	tpl := templates.NewTextTpl()

	dbConf := domainfile.NewDbConfig()
	dbConf.ParseConfig(v, logger)

	daoDomainFile := domainfile.NewDaoDomainFile(
		domainfile.WithDaoDomainFileLogger(logger),
		domainfile.WithDaoDomainFileTpl(tpl),
	)

	entityDomainFile := domainfile.NewEntityDomainFile(
		domainfile.WithEntityDomainFileLogger(logger),
		domainfile.WithEntityDomainFileTpl(tpl),
	)

	repoDomainFile := domainfile.NewRepoDomainFile(
		domainfile.WithRepoDomainFileLogger(logger),
		domainfile.WithRepoDomainFileTpl(tpl),
	)

	db2EntityOptions := Db2EnOptions{}
	StubsColumnsRepo := domainfile.StubsColumnsRepo{}

	shareInfo := domainfile.NewShareInfo()
	shareInfo.DbConf = dbConf

	writer := filedir.NewEsimWriter()

	db2Entity := NewDb2Entity(
		db2EntityOptions.WithLogger(logger),
		db2EntityOptions.WithColumnsInter(StubsColumnsRepo),
		db2EntityOptions.WithWriter(writer),
		db2EntityOptions.WithExecer(pkg.NewNullExec()),
		db2EntityOptions.WithDbConf(domainfile.NewDbConfig()),
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

	err = filedir.EsimRecoverFile(filedir.GetCurrentDir() +
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

	dbConf := domainfile.NewDbConfig()
	dbConf.ParseConfig(v, logger)

	daoDomainFile := domainfile.NewDaoDomainFile(
		domainfile.WithDaoDomainFileLogger(logger),
		domainfile.WithDaoDomainFileTpl(tpl),
	)

	entityDomainFile := domainfile.NewEntityDomainFile(
		domainfile.WithEntityDomainFileLogger(logger),
		domainfile.WithEntityDomainFileTpl(tpl),
	)

	repoDomainFile := domainfile.NewRepoDomainFile(
		domainfile.WithRepoDomainFileLogger(logger),
		domainfile.WithRepoDomainFileTpl(tpl),
	)

	db2EntityOptions := Db2EnOptions{}
	StubsColumnsRepo := domainfile.StubsColumnsRepo{}

	shareInfo := domainfile.NewShareInfo()
	shareInfo.DbConf = dbConf

	db2Entity := NewDb2Entity(
		db2EntityOptions.WithLogger(logger),
		db2EntityOptions.WithColumnsInter(StubsColumnsRepo),
		db2EntityOptions.WithWriter(filedir.NewErrWrite(2)),
		db2EntityOptions.WithExecer(pkg.NewNullExec()),
		db2EntityOptions.WithDbConf(domainfile.NewDbConfig()),
		db2EntityOptions.WithDomainFile(entityDomainFile, daoDomainFile, repoDomainFile),
		db2EntityOptions.WithShareInfo(shareInfo),
		db2EntityOptions.WithTpl(templates.NewTextTpl()),
		db2EntityOptions.WithDbConf(dbConf),
		db2EntityOptions.WithInfraer(infra.NewInfraer(
			infra.WithIfacerInfraInfo(infra.NewInfo()),
			infra.WithIfacerLogger(logger),
			infra.WithIfacerWriter(filedir.NewNullWrite()),
			infra.WithIfacerExecer(pkg.NewNullExec()),
		)),
	)

	err := db2Entity.Run(v)
	assert.Nil(t, err)

	err = filedir.EsimRecoverFile(filedir.GetCurrentDir() +
		string(filepath.Separator) + "example" + string(filepath.Separator) + "infra" +
		string(filepath.Separator) + "infra.go")
	assert.Nil(t, err)
}
