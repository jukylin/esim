package domainfile

import (
	"os"
	"testing"

	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestDaoDomainFile(t *testing.T) {
	err := filedir.CreateDir(target)
	assert.Nil(t, err)

	defer func() {
		err := os.RemoveAll(target)
		assert.Nil(t, err)
	}()

	v := viper.New()
	v.Set("disable_dao", false)
	v.Set("dao_target", target)
	v.Set("table", testTable)

	err = testDaoDomainFile.BindInput(v)
	assert.Nil(t, err)

	dbConf := NewDbConfig()
	dbConf.Database = database
	dbConf.Table = testTable

	shareInfo := NewShareInfo()
	shareInfo.CamelStruct = testStructName
	shareInfo.DbConf = dbConf

	testDaoDomainFile.ParseCloumns(Cols, shareInfo)
	content := testDaoDomainFile.Execute()
	assert.NotEmpty(t, content)

	savePath := testDaoDomainFile.GetSavePath()
	assert.Equal(t, "example/test.go", savePath)
}
