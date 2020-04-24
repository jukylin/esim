package domain_file

import (
	"os"
	"testing"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/tool/db2entity"
	"github.com/jukylin/esim/pkg/file-dir"
)

func TestDaoDomainFile(t *testing.T) {

	dir := "example"
	file_dir.CreateDir(dir)
	defer func() {
		os.RemoveAll(dir)
	}()

	v := viper.New()
	v.Set("disable_dao", false)
	v.Set("dao_target", "example")
	v.Set("table", "test")

	err := testDaoDomainFile.BindInput(v)
	assert.Nil(t, err)

	d2e := db2entity.NewDb2Entity()
	d2e.CamelStruct = "Test"

	testDaoDomainFile.ParseCloumns(db2entity.Cols, d2e)
	content := testDaoDomainFile.Execute()
	assert.NotEmpty(t, content)

	savePath := testDaoDomainFile.GetSavePath()
	assert.Equal(t, "example/test.go", savePath)
}