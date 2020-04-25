package db2entity

import (
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"os"
	"path/filepath"
	"github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/jukylin/esim/pkg/templates"
)

func TestDb2Entity_Run(t *testing.T) {

	v := viper.New()
	v.Set("entity_target", "./example/entity")
	v.Set("dao_target", "./example/dao")
	v.Set("repo_target", "./example/repo")
	v.Set("infra_dir", "./example/infra")

	v.Set("host", "127.0.0.1")
	v.Set("port", "3306")
	v.Set("user", "root")
	v.Set("passport", "")
	v.Set("database", "user")
	v.Set("table", "test_history")

	logger := log.NewLogger()
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
		db2EntityOptions.WithIfaceWrite(file_dir.NewEsimWriter()),
		db2EntityOptions.WithInfraInfo(NewInfraInfo()),
		db2EntityOptions.WithExecer(pkg.NewNullExec()),
		db2EntityOptions.WithDbConf(domain_file.NewDbConfig()),
		db2EntityOptions.WithDomainFile(entityDomainFile, daoDomainFile, repoDomainFile),
		db2EntityOptions.WithShareInfo(shareInfo),
	)

	err := db2Entity.Run(v)
	assert.Nil(t, err)

	os.Remove("./example" + string(filepath.Separator) + "entity" + string(filepath.Separator) + "test.go")
	os.Remove("./example" + string(filepath.Separator) + "dao" + string(filepath.Separator) + "test.go")
	os.Remove("./example" + string(filepath.Separator) + "repo" + string(filepath.Separator) + "test.go")
	err = file_dir.EsimRecoverFile(file_dir.GetCurrentDir() +
		string(filepath.Separator) + "example" + string(filepath.Separator) + "infra" + string(filepath.Separator) + "infra.go")
	assert.Nil(t, err)
}

func TestDb2Entity_ParseInfra(t *testing.T) {
	db2EntityOptions := Db2EnOptions{}
	StubsColumnsRepo := domain_file.StubsColumnsRepo{}

	db2Entity := NewDb2Entity(db2EntityOptions.WithLogger(log.NewLogger()),
		db2EntityOptions.WithColumnsInter(StubsColumnsRepo),
		db2EntityOptions.WithIfaceWrite(file_dir.NewEsimWriter()),
		db2EntityOptions.WithInfraInfo(NewInfraInfo()),
		db2EntityOptions.WithExecer(pkg.NewNullExec()),
		db2EntityOptions.WithDbConf(domain_file.NewDbConfig()),
	)

	assert.True(t, db2Entity.parseInfra(infraContent))
}

func TestDb2Entity_ProcessInfraInfo(t *testing.T)  {
	db2EntityOptions := Db2EnOptions{}

	db2Entity := NewDb2Entity(
		db2EntityOptions.WithInfraInfo(NewInfraInfo()))

	db2Entity.withStruct = "Test"

	assert.True(t, db2Entity.processNewInfra())
}
