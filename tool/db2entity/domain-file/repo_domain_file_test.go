package domain_file

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestRepoDomainFile(t *testing.T) {
	v := viper.New()
	v.Set("repo_target", "example")
	v.Set("table", "test")

	err := testRepoDomainFile.BindInput(v)
	assert.Nil(t, err)

	shareInfo := NewShareInfo()
	shareInfo.CamelStruct = "Test"

	testRepoDomainFile.ParseCloumns(Cols, shareInfo)
	content := testRepoDomainFile.Execute()
	println(content)
	assert.NotEmpty(t, content)

	savePath := testRepoDomainFile.GetSavePath()
	assert.Equal(t, "example/test.go", savePath)
	os.RemoveAll("a")
}