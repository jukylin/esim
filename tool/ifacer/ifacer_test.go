package ifacer

import (
	"os"
	"testing"

	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

var Result = `package example1

import (
	context "context"

	redis "github.com/jukylin/esim/redis"
	repo "github.com/jukylin/esim/tool/ifacer/example/repo"
)

type TestStub struct{}

func (this TestStub) Close(arg0 string, arg1 int) error {
	var r0 error

	return r0
}

func (this TestStub) Err() error {
	var r0 error

	return r0
}

func (this TestStub) Iface1(arg0 func(string) string) (result bool, pool redis.Pool) {

	return
}

func (this TestStub) Iface10(arg0 Close) {

	return
}

func (this TestStub) Iface2(ctx context.Context, found *bool) (result bool, err error) {

	return
}

func (this TestStub) Iface3() (f func(string) string) {

	return
}

func (this TestStub) Iface4() map[string]string {
	var r0 map[string]string

	return r0
}

func (this TestStub) Iface5(redisClient *redis.RedisClient) *redis.RedisClient {
	var r0 *redis.RedisClient

	return r0
}

func (this TestStub) Iface6(redisClient redis.RedisClient) redis.RedisClient {
	var r0 redis.RedisClient

	return r0
}

func (this TestStub) Iface7(arg0 chan<- bool) <-chan bool {
	var r0 <-chan bool

	return r0
}

func (this TestStub) Iface8(rp repo.Repo) repo.Repo {
	var r0 repo.Repo

	return r0
}

func (this TestStub) Iface9(arg0 TestStruct) {

	return
}
`

var ifacer *Ifacer

func TestMain(m *testing.M) {

	writer := &file_dir.NullWrite{}
	ifacer = NewIface(writer)

	code := m.Run()

	os.Exit(code)
}

func TestIfacer_RunNullWrite(t *testing.T) {
	v := viper.New()
	v.Set("out", "./abc/test_stub.go")

	v.Set("stname", "TestStub")

	v.Set("iname", "Test")

	v.Set("ipath", "./example/iface.go")

	err := ifacer.Run(v)
	assert.Equal(t, Result, ifacer.Content)
	assert.Nil(t, err)
}

func TestIfacer_Write(t *testing.T) {
	v := viper.New()
	v.Set("out", "./abc/test_stub.go")

	v.Set("stname", "TestStub")

	v.Set("iname", "Test")

	v.Set("ipath", "./example/iface.go")

	ifacer.writer = &file_dir.EsimWriter{}

	err := ifacer.Run(v)
	assert.Equal(t, Result, ifacer.Content)
	assert.Nil(t, err)
	file_dir.RemoveDir("./abc")
}

func TestIfacer_GetUniqueImportName(t *testing.T) {
	pkgName := "github.com/jukylin/esim/redis"

	importName := ifacer.getUniqueImportName(pkgName, 0)
	assert.Equal(t, "redis", importName)

	importName = ifacer.getUniqueImportName(pkgName, 1)
	assert.Equal(t, "esimredis", importName)

	importName = ifacer.getUniqueImportName(pkgName, 2)
	assert.Equal(t, "jukylinesimredis", importName)

	importName = ifacer.getUniqueImportName(pkgName, 3)
	assert.Equal(t, "githubcomjukylinesimredis", importName)

	shouldPanic := assert.Panics(t, func() {
		importName = ifacer.getUniqueImportName(pkgName, 4)
	})
	assert.True(t, shouldPanic)
}

func TestIfacer_SetNoConflictImport(t *testing.T) {

	testCases := []struct {
		caseName   string
		importName string
		pkgName    string
		expected   string
	}{
		{"redis", "redis", "github.com/jukylin/esim/redis", "github.com/jukylin/esim/redis"},
		{"aredis", "redis", "github.com/jukylin/a/redis", "github.com/jukylin/a/redis"},
		{"jukyaredis", "redis", "github.com/juky/a/redis", "github.com/juky/a/redis"},
		{"gitlabcomjukyaredis", "redis", "gitlab.com/juky/a/redis", "gitlab.com/juky/a/redis"},
	}

	for _, test := range testCases {
		t.Run(test.caseName, func(t *testing.T) {
			ifacer.setNoConflictImport(test.importName, test.pkgName)
			assert.Equal(t, test.expected, ifacer.PkgNoConflictImport[test.caseName])
		})
	}
}
