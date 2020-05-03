package protoc

import (
	"testing"
	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
	"github.com/spf13/viper"
	"os"
)

func TestProtoc_Run(t *testing.T) {
	protocer := NewProtoc(
		WithProtocLogger(log.NewLogger()),
	)

	v := viper.New()
	v.Set("target", "./")
	v.Set("from_proto", "helloworld/helloworld.proto")

	result := protocer.Run(v)
	assert.True(t, result)
	os.Remove("./helloworld/helloworld.pb.go")
}

func TestProtoc_ParsePkgName(t *testing.T)  {
	protocer := NewProtoc(
		WithProtocLogger(log.NewLogger()),
	)

	packName, err := protocer.parsePkgName("./helloworld/helloworld.proto")
	assert.Nil(t, err)
	assert.Equal(t, "helloworld", packName)
}

func TestProtoc_NotPkgName(t *testing.T)  {
	protocer := NewProtoc(
		WithProtocLogger(log.NewLogger()),
	)
	_, err := protocer.parsePkgName("./helloworld/helloworld_not_pkg_name.proto")
	assert.Error(t, err)
}

func TestProtoc_ParseProtoPath(t *testing.T)  {
	protocer := NewProtoc(
		WithProtocLogger(log.NewLogger()),
	)
	protocer.fromProto = "./data/go/src/github.com/grpc/grpc/examples/helloworld.proto"
	protocer.parseProtoPath()
	assert.Equal(t, "./data/go/src/github.com/grpc/grpc/examples", protocer.protoPath)
}