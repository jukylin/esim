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
)

func TestDb2Entity_Run(t *testing.T) {

	db2EntityOptions := Db2EnOptions{}
	StubsColumnsRepo := StubsColumnsRepo{}

	db2Entity := NewDb2Entity(db2EntityOptions.WithLogger(log.NewLogger()),
		db2EntityOptions.WithColumnsInter(StubsColumnsRepo),
		db2EntityOptions.WithIfaceWrite(file_dir.NewEsimWriter()),
		db2EntityOptions.WithInfraInfo(NewInfraInfo()),
		db2EntityOptions.WithExecer(pkg.NewNullExec()),
	)

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

	err := db2Entity.Run(v)
	assert.Nil(t, err)

	os.Remove("./example" + string(filepath.Separator) + "entity" + string(filepath.Separator) + "test.go")
	os.Remove("./example" + string(filepath.Separator) + "dao" + string(filepath.Separator) + "test.go")
	os.Remove("./example" + string(filepath.Separator) + "repo" + string(filepath.Separator) + "test.go")
	err = file_dir.EsimRecoverFile(file_dir.GetCurrentDir() +
		string(filepath.Separator) + "example" + string(filepath.Separator) + "infra" + string(filepath.Separator) + "infra.go")
	assert.Nil(t, err)
}

func TestDb2Entity_CloumnsToEntityTmp(t *testing.T) {

	db2Entity := &db2Entity{}

	entityTmp := db2Entity.cloumnsToEntityTpl(cols)
	assert.Equal(t, 3, len(entityTmp.StructInfo.Fields))
}

func TestDb2Entity_DirPathToImportPath(t *testing.T) {
	db2Entity := &db2Entity{}
	importPaht := db2Entity.dirPathToImportPath("./a/b/c/")
	assert.Equal(t, "a/b/c", importPaht)
}

func TestDb2Entity_ParseInfra(t *testing.T) {
	db2Entity := &db2Entity{}
	assert.True(t, db2Entity.parseInfra(infraContent))
}

func TestDb2Entity_ProcessInfraInfo(t *testing.T) {
	db2EntityOptions := Db2EnOptions{}
	db2Entity := NewDb2Entity(
		db2EntityOptions.WithInfraInfo(NewInfraInfo()))

	db2Entity.withStruct = "Test"

	assert.True(t, db2Entity.processNewInfra())
}
