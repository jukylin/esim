package domain_file

import (
	"os"
	"testing"

	"github.com/jukylin/esim/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestEntityDomainFile_BindInput(t *testing.T) {
	v := viper.New()
	v.Set("boubctx", "a")
	v.Set("disabled_entity", false)
	v.Set("entity_target", "example")
	v.Set("table", "test")

	err := testEntityDomainFile.BindInput(v)
	assert.Nil(t, err)
}

func TestEntityDomainFile(t *testing.T) {
	v := viper.New()
	v.Set("boubctx", "a")
	v.Set("disabled_entity", false)
	v.Set("entity_target", "example")
	v.Set("table", "test")
	v.Set("database", "test")

	err := testEntityDomainFile.BindInput(v)
	assert.Nil(t, err)

	dbConf := NewDbConfig()
	dbConf.ParseConfig(v, log.NewNullLogger())

	shareInfo := NewShareInfo()
	shareInfo.CamelStruct = "Test"
	shareInfo.DbConf = dbConf

	testEntityDomainFile.ParseCloumns(Cols, shareInfo)
	content := testEntityDomainFile.Execute()
	assert.NotEmpty(t, content)

	savePath := testEntityDomainFile.GetSavePath()
	assert.Equal(t, "a/example/test.go", savePath)
	os.RemoveAll("a")
}
