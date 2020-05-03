package domain_file

import (
	"os"
	"testing"

	file_dir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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

	dbConf := NewDbConfig()
	dbConf.Database = "test"
	dbConf.Table = "test"

	shareInfo := NewShareInfo()
	shareInfo.CamelStruct = "Test"
	shareInfo.DbConf = dbConf

	testDaoDomainFile.ParseCloumns(Cols, shareInfo)
	content := testDaoDomainFile.Execute()
	assert.NotEmpty(t, content)

	savePath := testDaoDomainFile.GetSavePath()
	assert.Equal(t, "example/test.go", savePath)
}
