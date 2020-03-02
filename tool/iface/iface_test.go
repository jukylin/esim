package iface

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/pkg/file-dir"
)

var Result = `package example1

import (
	"context"
)

type TestStub struct{}

func (this *TestStub) Iface1(arg0 func(string) string) bool {

}

func (this *TestStub) Iface2(ctx context.Context, found *bool) (bool, error) {

}

func (this *TestStub) Iface3() func(string) string {

}
`

func TestIface_FindIface(t *testing.T) {
	iface := &Iface{}

	iface.OutFile = "./abc/test_stub.go"

	iface.StructName = "TestStub"

	ifacePath := "./example"

	iface.FindIface(ifacePath, "Test")

	err := iface.Gen()
	assert.Nil(t, err)

	assert.Equal(t, Result, iface.Content)

	iface.Write()

	exists, err := file_dir.IsExistsDir("./abc")
	assert.Nil(t, err)
	assert.True(t, exists)

	assert.Nil(t, file_dir.RemoveDir("./abc"))
}