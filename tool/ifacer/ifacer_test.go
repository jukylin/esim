package ifacer

import (
	"testing"

	"github.com/jukylin/esim/log"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var Result = `package example1

import (
	context "context"

	redis "github.com/gomodule/redigo/redis"
	repo "github.com/jukylin/esim/tool/ifacer/example/repo"
)

type TestStub struct{}

func (ts TestStub) Close(arg0 string, arg1 int) error {
	var r0 error

	return r0
}

func (ts TestStub) Err() error {
	var r0 error

	return r0
}

func (ts TestStub) Iface1(arg0 func(string) string) (result bool, pool redis.Pool) {

	return
}

func (ts TestStub) Iface10(arg0 Close) {

	return
}

func (ts TestStub) Iface11(arg0 ...interface{}) {

	return
}

func (ts TestStub) Iface2(ctx context.Context, found *bool) (result bool, err error) {

	return
}

func (ts TestStub) Iface3() (f func(repo.Repo) string) {

	return
}

func (ts TestStub) Iface4(arg0 map[string]*redis.Client) map[string]string {
	var r0 map[string]string

	return r0
}

func (ts TestStub) Iface5(redisClient *redis.Client) *redis.Client {
	var r0 *redis.Client

	return r0
}

func (ts TestStub) Iface6(redisClient redis.Client) redis.Client {
	var r0 redis.Client

	return r0
}

func (ts TestStub) Iface7(arg0 chan<- bool, arg1 chan<- redis.Client) <-chan bool {
	var r0 <-chan bool

	return r0
}

func (ts TestStub) Iface8(rp repo.Repo) repo.Repo {
	var r0 repo.Repo

	return r0
}

func (ts TestStub) Iface9(arg0 TestStruct, arg1 []TestStruct, arg2 [3]TestStruct) {

	return
}
`

func TestIfacer_RunNullWrite(t *testing.T) {
	v := viper.New()
	v.Set("out", "./abc/test_stub.go")

	v.Set("stname", "TestStub")

	v.Set("iname", "Test")

	v.Set("ipath", "./example/iface.go")

	writer := &filedir.NullWrite{}
	ifacer := NewIfacer(
		WithIfacerLogger(log.NewLogger()),
		WithIfacerTpl(templates.NewTextTpl()),
		WithIfacerWriter(writer),
	)

	err := ifacer.Run(v)
	assert.Nil(t, err)
	assert.Equal(t, Result, ifacer.Content)
}

func TestIfacer_Write(t *testing.T) {
	v := viper.New()
	v.Set("out", "./abc/test_stub.go")

	v.Set("stname", "TestStub")

	v.Set("iname", "Test")

	v.Set("ipath", "./example/iface.go")

	writer := &filedir.EsimWriter{}
	ifacer := NewIfacer(
		WithIfacerLogger(log.NewLogger()),
		WithIfacerTpl(templates.NewTextTpl()),
		WithIfacerWriter(writer),
	)

	err := ifacer.Run(v)
	assert.Equal(t, Result, ifacer.Content)
	assert.Nil(t, err)
	err = filedir.RemoveDir("./abc")
	assert.Nil(t, err)
}

func TestIfacer_GetUniqueImportName(t *testing.T) {
	pkgName := "github.com/jukylin/esim/redis"

	writer := &filedir.NullWrite{}
	ifacer := NewIfacer(
		WithIfacerLogger(log.NewLogger()),
		WithIfacerTpl(templates.NewTextTpl()),
		WithIfacerWriter(writer),
	)

	importName := ifacer.getUniqueImportName(pkgName, 0)
	assert.Equal(t, "redis0", importName)

	importName = ifacer.getUniqueImportName(pkgName, 1)
	assert.Equal(t, "redis1", importName)
}

//nolint:scopelint
func TestIfacer_SetNoConflictImport(t *testing.T) {
	testCases := []struct {
		caseName   string
		importName string
		pkgName    string
		expected   string
	}{
		{"redis", "redis", "github.com/jukylin/esim/redis", "github.com/jukylin/esim/redis"},
		{"redis1", "redis", "github.com/jukylin/a/redis", "github.com/jukylin/a/redis"},
		{"redis2", "redis", "github.com/juky/a/redis", "github.com/juky/a/redis"},
	}

	writer := &filedir.NullWrite{}
	ifacer := NewIfacer(
		WithIfacerLogger(log.NewLogger(log.WithDebug(true))),
		WithIfacerTpl(templates.NewTextTpl()),
		WithIfacerWriter(writer),
	)

	for _, test := range testCases {
		t.Run(test.caseName, func(t *testing.T) {
			ifacer.setNoConflictImport(test.importName, test.pkgName)
			assert.Equal(t, test.expected, ifacer.pkgNoConflictImport[test.caseName].Path)
		})
	}
}
