package iface

import (
	"testing"
	"os"

	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
)

var Result = `package example1

import (
	"context"
)

type TestStub struct{}

func (this TestStub) Iface1(arg0 func(string) string) (result bool) {

	return result
}

func (this TestStub) Iface2(ctx context.Context, found *bool) (result bool, err error) {

	return result, err
}

func (this TestStub) Iface3() (f func(string) string) {

	return f
}

func (this TestStub) Iface4() map[string]string {
	var r0 map[string]string

	return r0
}
`

var ifacer *Iface

func TestMain(m *testing.M) {

	ifacer = NewIface()

	code := m.Run()

	os.Exit(code)
}

func TestIface(t *testing.T) {

	ifacer.OutFile = "./abc/test_stub.go"

	ifacer.StructName = "TestStub"

	//ifacePath := "./example"

	//iface.FindIface(ifacePath, "Test")

	err := ifacer.Process()
	assert.Nil(t, err)

	assert.Equal(t, Result, ifacer.Content)

	ifacer.Write()

	exists, err := file_dir.IsExistsDir("./abc")
	assert.Nil(t, err)
	assert.True(t, exists)

	assert.Nil(t, file_dir.RemoveDir("./abc"))
}


func TestIface_Run(t *testing.T) {
	v := viper.New()
	v.Set("out", "./abc/test_stub.go")

	v.Set("stname", "TestStub")

	v.Set("iname", "Test")

	v.Set("ipath", "./example/iface.go")

	ifacer := NewIface()
	err := ifacer.Run(v)
	assert.Nil(t, err)
	assert.Nil(t, file_dir.RemoveDir("./abc"))
}


func TestIface_ParsePackageImport(t *testing.T) {
	//ifacer := NewIface()

	//ifacer.Run()

	//iface := &Iface{}
	//iface.ParsePackageImport("", "/data/go/src/github.com/jukylin/esim/tool/iface")
}


func TestIface_GetUniqueImportName(t *testing.T)  {
	pkgName := "github.com/jukylin/esim/redis"

	importName, err := ifacer.getUniqueImportName(pkgName, 0)
	assert.Nil(t, err)
	assert.Equal(t, "redis", importName)

	importName, err = ifacer.getUniqueImportName(pkgName, 1)
	assert.Nil(t, err)
	assert.Equal(t, "esimredis", importName)

	importName, err = ifacer.getUniqueImportName(pkgName, 2)
	assert.Nil(t, err)
	assert.Equal(t, "jukylinesimredis", importName)

	importName, err = ifacer.getUniqueImportName(pkgName, 3)
	assert.Nil(t, err)
	assert.Equal(t, "githubcomjukylinesimredis", importName)

	importName, err = ifacer.getUniqueImportName(pkgName, 4)
	assert.Error(t, err)
}