package domain_file

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/tool/db2entity"
)


func TestRepoDomainFile(t *testing.T) {
	v := viper.New()
	v.Set("repo_target", "example")
	v.Set("table", "test")

	err := testRepoDomainFile.BindInput(v)
	assert.Nil(t, err)

	d2e := db2entity.NewDb2Entity()
	d2e.CamelStruct = "Test"

	testRepoDomainFile.ParseCloumns(db2entity.Cols, d2e)
	content := testRepoDomainFile.Execute()
	println(content)
	assert.NotEmpty(t, content)

	savePath := testRepoDomainFile.GetSavePath()
	assert.Equal(t, "example/test.go", savePath)
	os.RemoveAll("a")
}