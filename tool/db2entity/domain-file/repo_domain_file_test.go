package domain_file

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg/file-dir"
)

func TestRepoDomainFile(t *testing.T) {
	dir := "example"
	file_dir.CreateDir(dir)

	v := viper.New()
	v.Set("repo_target", "example")
	v.Set("table", "test")

	err := testRepoDomainFile.BindInput(v)
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
	os.RemoveAll(dir)
}