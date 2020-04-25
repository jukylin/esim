package domain_file

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)



func TestEntityDomainFile_ErrBindInput(t *testing.T) {
	v := viper.New()
	v.Set("boubctx", "a")
	v.Set("disabled_entity", false)
	v.Set("entity_target", "example")
	//v.Set("table", "test")

	err := testEntityDomainFile.BindInput(v)
	assert.Error(t, err)
}


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

	err := testEntityDomainFile.BindInput(v)
	assert.Nil(t, err)

	shareInfo := NewShareInfo()
	shareInfo.CamelStruct = "Test"

	testEntityDomainFile.ParseCloumns(Cols, shareInfo)
	content := testEntityDomainFile.Execute()
	assert.NotEmpty(t, content)

	savePath := testEntityDomainFile.GetSavePath()
	assert.Equal(t, "a/example/test.go", savePath)
	os.RemoveAll("a")
}