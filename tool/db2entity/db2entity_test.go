package db2entity

import (
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"testing"
	"github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/jukylin/esim/pkg/templates"
	"os"
	"path/filepath"
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

	db2Entity := NewDb2Entity(
		db2EntityOptions.WithLogger(logger),
		db2EntityOptions.WithColumnsInter(StubsColumnsRepo),
		db2EntityOptions.WithWriter(file_dir.NewEsimWriter()),
		db2EntityOptions.WithInfraInfo(NewInfraInfo()),
		db2EntityOptions.WithExecer(pkg.NewNullExec()),
		db2EntityOptions.WithDbConf(domain_file.NewDbConfig()),
		db2EntityOptions.WithDomainFile(entityDomainFile, daoDomainFile, repoDomainFile),
		db2EntityOptions.WithShareInfo(shareInfo),
		db2EntityOptions.WithTpl(templates.NewTextTpl()),
		db2EntityOptions.WithDbConf(dbConf),
	)

	err := db2Entity.Run(v)
	assert.Nil(t, err)

	for path, _ := range db2Entity.domainContent {
		assert.FileExists(t, path)
	}

	for path, _ := range db2Entity.domainContent {
		os.Remove(path)
	}

	//err = file_dir.EsimRecoverFile(file_dir.GetCurrentDir() +
	//	string(filepath.Separator) + "example" + string(filepath.Separator) + "infra" + string(filepath.Separator) + "infra.go")
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
		db2EntityOptions.WithInfraInfo(NewInfraInfo()),
		db2EntityOptions.WithExecer(pkg.NewNullExec()),
		db2EntityOptions.WithDbConf(domain_file.NewDbConfig()),
		db2EntityOptions.WithDomainFile(entityDomainFile, daoDomainFile, repoDomainFile),
		db2EntityOptions.WithShareInfo(shareInfo),
		db2EntityOptions.WithTpl(templates.NewTextTpl()),
		db2EntityOptions.WithDbConf(dbConf),
	)

	err := db2Entity.Run(v)
	assert.Nil(t, err)

	err = file_dir.EsimRecoverFile(file_dir.GetCurrentDir() +
		string(filepath.Separator) + "example" + string(filepath.Separator) + "infra" + string(filepath.Separator) + "infra.go")
	assert.Nil(t, err)
}

func TestDb2Entity_BuildNewInfraContent(t *testing.T) {

	expected := `package infra

import (

	//sync
	//is a test
	"sync"

	"github.com/google/wire"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/redis"
)

var infraOnce sync.Once
var onceInfra *Infra

type Infra struct {

	//Esim
	*container.Esim

	//redis
	Redis redis.RedisClient

	check bool

	a int
}

var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideA,
)

func NewInfra() *Infra {
	infraOnce.Do(func() {
	})

	return onceInfra
}

// Close close the infra when app stop
func (this *Infra) Close() {
}

func (this *Infra) HealthCheck() []error {
	var errs []error
	return errs
}
func provideA() { println("test") }
`

	db2EntityOptions := Db2EnOptions{}
	StubsColumnsRepo := domain_file.StubsColumnsRepo{}

	db2Entity := NewDb2Entity(db2EntityOptions.WithLogger(log.NewLogger()),
		db2EntityOptions.WithColumnsInter(StubsColumnsRepo),
		db2EntityOptions.WithWriter(file_dir.NewEsimWriter()),
		db2EntityOptions.WithInfraInfo(NewInfraInfo()),
		db2EntityOptions.WithExecer(pkg.NewNullExec()),
		db2EntityOptions.WithDbConf(domain_file.NewDbConfig()),
	)

	assert.True(t, db2Entity.parseInfra(infraContent))

	injectInfo := domain_file.NewInjectInfo()
	injectInfo.Imports = append(injectInfo.Imports, pkg.Import{Path: "time"})
	injectInfo.Fields = append(injectInfo.Fields, pkg.Field{Field : "a int"})
	injectInfo.InfraSetArgs = append(injectInfo.InfraSetArgs, "provideA")
	injectInfo.Provides = append(injectInfo.Provides, domain_file.Provide{`func provideA() {println("test")}`})

	db2Entity.injectInfos = append(db2Entity.injectInfos, injectInfo)

	db2Entity.copyInfraInfo()

	db2Entity.processNewInfra()

	db2Entity.toStringNewInfra()

	db2Entity.buildNewInfraContent()

	assert.Equal(t, expected, db2Entity.makeCodeBeautiful(db2Entity.newInfraInfo.content))
}
