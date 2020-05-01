package protoc

import (
	"testing"
	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
)

func TestProtoc_Run(t *testing.T) {

}

func TestProtoc_ParsePkgName(t *testing.T)  {
	protocer := NewProtoc(
		WithProtocLogger(log.NewLogger()),
	)

	packName, err := protocer.parsePkgName("./example/helloworld.proto")
	assert.Nil(t, err)
	assert.Equal(t, "helloworld", packName)
}

func TestProtoc_NotPkgName(t *testing.T)  {
	protocer := NewProtoc(
		WithProtocLogger(log.NewLogger()),
	)
	_, err := protocer.parsePkgName("./example/helloworld_not_pkg_name.proto")
	assert.Error(t, err)
}