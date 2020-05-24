package domainfile

import (
	"os"
	"testing"

	"github.com/jukylin/esim/log"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestEntityDomainFile_BindInput(t *testing.T) {
	v := viper.New()
	v.Set("boubctx", boubctx)
	v.Set("disabled_entity", false)
	v.Set("entity_target", target)
	v.Set("table", testTable)

	err := testEntityDomainFile.BindInput(v)
	assert.Nil(t, err)
}

func TestEntityDomainFile(t *testing.T) {
	v := viper.New()
	v.Set("boubctx", boubctx)
	v.Set("disabled_entity", false)
	v.Set("entity_target", target)
	v.Set("table", testTable)
	v.Set("database", database)

	err := testEntityDomainFile.BindInput(v)
	assert.Nil(t, err)

	dbConf := NewDbConfig()
	dbConf.ParseConfig(v, log.NewNullLogger())

	shareInfo := NewShareInfo()
	shareInfo.CamelStruct = testStructName
	shareInfo.DbConf = dbConf

	testEntityDomainFile.ParseCloumns(Cols, shareInfo)
	content := testEntityDomainFile.Execute()
	assert.NotEmpty(t, content)

	savePath := testEntityDomainFile.GetSavePath()
	assert.Equal(t, "a/example/test.go", savePath)
	os.RemoveAll(boubctx)
}
