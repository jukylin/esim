package mysql

import (
	"testing"
	"context"
	"sync"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
)

var (
	test1Config = DbConfig{}
	test2Config = DbConfig{}
)

type TestStruct struct{
	Id int `json:"id"`
	Name string `json:"name"`
}

type UserStruct struct{
	Id int `json:"id"`
	Username string `json:"username"`
}

func init()  {
	test1Config.Db = "test_1"
	test1Config.Dsn = "root:123456@tcp(0.0.0.0:3306)/test_1?charset=utf8&parseTime=True&loc=Local"
	test1Config.MaxIdle = 10
	test1Config.MaxOpen = 100

	test2Config.Db = "test_2"
	test2Config.Dsn = "root:123456@tcp(0.0.0.0:3306)/test_2?charset=utf8&parseTime=True&loc=Local"
	test2Config.MaxIdle = 10
	test2Config.MaxOpen = 100
}

func TestInitAndSingleInstance(t *testing.T)  {

	mysqlClientOptions := MysqlClientOptions{}

	mysqlClient := NewMysqlClient(mysqlClientOptions.WithDbConfig([]DbConfig{test1Config}))
	ctx := context.Background()
	db := mysqlClient.GetCtxDb(ctx, "test_1")
	assert.NotNil(t, db)

	_, ok := mysqlClient.dbs["test_1"]
	assert.True(t, ok)

	assert.Equal(t, mysqlClient, NewMysqlClient())
}


func TestProxyPatternWithTwoInstance(t *testing.T)  {
	mysqlOnce = sync.Once{}

	mysqlClientOptions := MysqlClientOptions{}
	monitorProxyOptions := MonitorProxyOptions{}
	memConfig := config.NewMemConfig()
	//memConfig.Set("debug", true)

	mysqlClient := NewMysqlClient(
		mysqlClientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
		mysqlClientOptions.WithConf(memConfig),
		mysqlClientOptions.WithProxy(func() interface{} {
			return NewMonitorProxy(
				monitorProxyOptions.WithConf(memConfig),
				monitorProxyOptions.WithLogger(log.NewLogger()))
		}),
		)

	ctx := context.Background()
	db1 := mysqlClient.GetCtxDb(ctx, "test_1")
	assert.NotNil(t, db1)

	ts := &TestStruct{}
	db1.Table("test").First(ts)

	assert.Len(t, db1.GetErrors(), 0)

	db2 := mysqlClient.GetCtxDb(ctx, "test_2")
	assert.NotNil(t, db2)

	us := &UserStruct{}
	db2.Table("user").First(us)
	assert.Len(t, db1.GetErrors(), 0)
}

func TestMulProxyPatternWithOneInstance(t *testing.T)  {
	mysqlOnce = sync.Once{}

	mysqlClientOptions := MysqlClientOptions{}
	monitorProxyOptions := MonitorProxyOptions{}
	memConfig := config.NewMemConfig()
	//memConfig.Set("debug", true)

	spyProxy1 := newSpyProxy(log.NewLogger(), "spyProxy1")
	spyProxy2 := newSpyProxy(log.NewLogger(), "spyProxy2")
	monitorProxy := NewMonitorProxy(
		monitorProxyOptions.WithConf(memConfig),
		monitorProxyOptions.WithLogger(log.NewLogger()))

	mysqlClient := NewMysqlClient(
		mysqlClientOptions.WithDbConfig([]DbConfig{test1Config}),
		mysqlClientOptions.WithConf(memConfig),
		mysqlClientOptions.WithProxy(
			func() interface{} {
				return spyProxy1
			},
			func() interface{} {
				return spyProxy2
			},
			func() interface{} {
				return monitorProxy
			},
		))

	ctx := context.Background()
	db1 := mysqlClient.GetCtxDb(ctx, "test_1")
	assert.NotNil(t, db1)

	ts := &TestStruct{}
	db1.Table("test").First(ts)

	assert.Len(t, db1.GetErrors(), 0)

	assert.True(t, spyProxy1.QueryWasCalled)
	assert.False(t, spyProxy1.QueryRowWasCalled)
	assert.False(t, spyProxy1.ExecWasCalled)
	assert.False(t, spyProxy1.PrepareWasCalled)

	assert.True(t, spyProxy2.QueryWasCalled)
	assert.False(t, spyProxy2.QueryRowWasCalled)
	assert.False(t, spyProxy2.ExecWasCalled)
	assert.False(t, spyProxy2.PrepareWasCalled)
}



func TestMulProxyPatternWithTwoInstance(t *testing.T)  {
	mysqlOnce = sync.Once{}

	mysqlClientOptions := MysqlClientOptions{}
	memConfig := config.NewMemConfig()
	//memConfig.Set("debug", true)

	mysqlClient := NewMysqlClient(
		mysqlClientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
		mysqlClientOptions.WithConf(memConfig),
		mysqlClientOptions.WithProxy(
			func() interface{} {
				return newSpyProxy(log.NewLogger(), "spyProxy1")
			},
			func() interface{} {
				return newSpyProxy(log.NewLogger(), "spyProxy2")
			},
			func() interface{} {
				monitorProxyOptions := MonitorProxyOptions{}
				return NewMonitorProxy(
					monitorProxyOptions.WithConf(memConfig),
					monitorProxyOptions.WithLogger(log.NewLogger()))
			},
		),
	)

	ctx := context.Background()
	db1 := mysqlClient.GetCtxDb(ctx, "test_1")
	assert.NotNil(t, db1)

	ts := &TestStruct{}
	db1.Table("test").First(ts)

	assert.Len(t, db1.GetErrors(), 0)

	db2 := mysqlClient.GetCtxDb(ctx, "test_2")
	assert.NotNil(t, db2)

	us := &UserStruct{}
	db2.Table("user").First(us)

	assert.Len(t, db2.GetErrors(), 0)
}



func BenchmarkParallelGetDB(b *testing.B) {
	mysqlOnce = sync.Once{}

	b.ReportAllocs()
	b.ResetTimer()

	mysqlClientOptions := MysqlClientOptions{}
	monitorProxyOptions := MonitorProxyOptions{}
	memConfig := config.NewMemConfig()

	mysqlClient := NewMysqlClient(
		mysqlClientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
		mysqlClientOptions.WithConf(memConfig),
		mysqlClientOptions.WithProxy(func() interface{} {
			spyProxy := newSpyProxy(log.NewLogger(), "spyProxy")
			spyProxy.NextProxy(NewMonitorProxy(
				monitorProxyOptions.WithConf(memConfig),
				monitorProxyOptions.WithLogger(log.NewLogger())))

			return spyProxy
		}),
	)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			mysqlClient.GetCtxDb(ctx, "test_1")

			mysqlClient.GetCtxDb(ctx, "test_2")

		}
	})

	b.StopTimer()
}

func TestDummyProxy_Exec(t *testing.T) {
	mysqlClientOptions := MysqlClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	mysqlClient := NewMysqlClient(
		mysqlClientOptions.WithDbConfig([]DbConfig{test1Config}),
		mysqlClientOptions.WithConf(memConfig),
		mysqlClientOptions.WithProxy(
			func() interface{} {
				return newSpyProxy(log.NewLogger(), "spyProxy")
			},
			func() interface{} {
				return newDummyProxy(log.NewLogger(), "dummyProxy")
			}),
		)
	ctx := context.Background()
	db := mysqlClient.GetCtxDb(ctx, "test_2")
	assert.NotNil(t, db)

	db, ok := mysqlClient.dbs["test_2"]
	assert.True(t, ok)

	db.Table("user").Create(&UserStruct{})
	assert.Len(t, db.GetErrors(), 0)

	assert.Equal(t, db.RowsAffected, int64(0))
}