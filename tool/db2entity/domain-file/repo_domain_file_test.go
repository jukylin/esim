package domain_file

import (
	"os"
	"testing"

	"github.com/jukylin/esim/log"
	file_dir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRepoDomainFile(t *testing.T) {
	dir := "example"
	err := file_dir.CreateDir(dir)
	assert.Nil(t, err)

	v := viper.New()
	v.Set("repo_target", "example")
	v.Set("table", "test")

	err = testRepoDomainFile.BindInput(v)
	assert.Nil(t, err)

	dbConf := NewDbConfig()
	dbConf.ParseConfig(v, log.NewNullLogger())

	shareInfo := NewShareInfo()
	shareInfo.CamelStruct = "Test"
	shareInfo.DbConf = dbConf

	testRepoDomainFile.ParseCloumns(Cols, shareInfo)
	content := testRepoDomainFile.Execute()
	assert.NotEmpty(t, content)

	savePath := testRepoDomainFile.GetSavePath()
	assert.Equal(t, "example/test.go", savePath)
	err = os.RemoveAll(dir)
	assert.Nil(t, err)
}
