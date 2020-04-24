package domain_file

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/tool/db2entity"
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

	d2e := db2entity.NewDb2Entity()
	d2e.CamelStruct = "Test"

	testEntityDomainFile.ParseCloumns(db2entity.Cols, d2e)
	content := testEntityDomainFile.Execute()
	assert.NotEmpty(t, content)

	savePath := testEntityDomainFile.GetSavePath()
	assert.Equal(t, "a/example/test.go", savePath)
	os.RemoveAll("a")
}