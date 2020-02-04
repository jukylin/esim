package redis

import (
	"testing"
	"context"
	"sync"
	"time"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/log"
	"github.com/bilibili/kratos/pkg/cache/redis"
	"github.com/jukylin/esim/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
)

//使用 proxyConn 代理请求
func TestGetProxyConn(t *testing.T){
	poolRedisOnce = sync.Once{}

	redisClientOptions := RedisClientOptions{}
	redisClent := NewRedisClient(
		redisClientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := MonitorProxyOptions{}
				return NewMonitorProxy(
					monitorProxyOptions.WithLogger(log.NewLogger()),
				)
			},
			),
	)

	conn := redisClent.GetCtxRedisConn()
	assert.IsTypef(t, NewMonitorProxy(), conn, "MonitorProxy type")
	assert.NotNil(t, conn)
	conn.Close()
	redisClent.Close()
}


func TestGetNotProxyConn(t *testing.T){
	poolRedisOnce = sync.Once{}
	redisClent := NewRedisClient()

	conn := redisClent.GetCtxRedisConn()

	assert.IsTypef(t, NewFacadeProxy(), conn, "MediatorProxy type")
	assert.NotNil(t, conn)
	conn.Close()
	redisClent.Close()
}

func TestMonitorProxy_Do(t *testing.T) {
	poolRedisOnce = sync.Once{}

	redisClientOptions := RedisClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewRedisClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithProxy(
			func() interface {} {
				monitorProxyOptions := MonitorProxyOptions{}
				return NewMonitorProxy(
					monitorProxyOptions.WithLogger(log.NewLogger()),
				)
			},
		),
	)

	ctx := context.Background()

	conn := redisClent.GetCtxRedisConn()

	_, err := redis.String(conn.Do(ctx, "get", "name"))
	assert.Nil(t, err)
	conn.Close()
	redisClent.Close()
}


func TestMulLevelProxy_Do(t *testing.T) {
	poolRedisOnce = sync.Once{}

	redisClientOptions := RedisClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	spyProxy := NewSpyProxy(log.NewLogger(), "spyProxy")

	redisClent := NewRedisClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithProxy(
			func() interface {} {
				monitorProxyOptions := MonitorProxyOptions{}
				return NewMonitorProxy(
					monitorProxyOptions.WithLogger(log.NewLogger()),
				)
			},
			func() interface {} {
				return spyProxy
			},
		),
	)

	ctx := context.Background()

	conn := redisClent.GetCtxRedisConn()
	conn.Do(ctx, "get", "name")
	assert.True(t, spyProxy.DoWasCalled)
	conn.Close()

	redisClent.Close()
}


func TestMulGo_Do(t *testing.T) {
	poolRedisOnce = sync.Once{}

	redisClientOptions := RedisClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewRedisClient(
		redisClientOptions.WithConf(memConfig),
		//redisClientOptions.WithProxy(
		//	func() ContextConn {
		//		return NewStubsProxy(log.NewLogger(), "stubsProxy")
		//	},
		//),
	)

	wg := sync.WaitGroup{}
	//ctx := context.Background()
	wg.Add(2)

	go func() {
		conn := redisClent.GetRedisConn()
		name, err := redis.String(conn.Do( "get", "name"))
		assert.Nil(t, err)
		assert.Equal(t, "test", name)
		conn.Close()
		wg.Done()
	}()

	go func() {
		conn := redisClent.GetRedisConn()
		version, err := redis.String(conn.Do("get", "version"))
		assert.Nil(t, err)
		assert.Equal(t, "2.0", version)
		conn.Close()
		wg.Done()
	}()
	wg.Wait()

	redisClent.Close()
}

func Benchmark_MulGo_Do(b *testing.B) {

	redisClientOptions := RedisClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewRedisClient(
		redisClientOptions.WithConf(memConfig),
	)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			wg := sync.WaitGroup{}
			ctx := context.Background()
			wg.Add(b.N * 2)
			for j := 0; j < b.N; j++ {
				go func() {
					conn := redisClent.GetCtxRedisConn()
					name, err := redis.String(conn.Do(ctx, "get", "name"))
					assert.Nil(b, err)
					assert.Equal(b, "test", name)
					conn.Close()
					wg.Done()
				}()

				go func() {
					conn := redisClent.GetCtxRedisConn()
					version, err := redis.String(conn.Do(ctx, "get", "version"))
					assert.Nil(b, err)
					assert.Equal(b, "2.0", version)
					conn.Close()
					wg.Done()
				}()
			}
			wg.Wait()
		}
	})
	redisClent.Close()
}

func TestRedisClient_Stats(t *testing.T) {

	redisClientOptions := RedisClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewRedisClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithStateTicker(10 * time.Microsecond),
	)

	time.Sleep(100 * time.Millisecond)

	ctx := context.Background()
	conn := redisClent.GetCtxRedisConn()
	conn.Do(ctx, "get", "name")
	conn.Close()

	conn.Do(ctx, "get", "name")
	conn.Close()

	conn.Do(ctx, "get", "name")
	conn.Close()

	lab := prometheus.Labels{"stats": "active_count"}
	c, _ := redisStats.GetMetricWith(lab)
	metric := &io_prometheus_client.Metric{}
	c.Write(metric)
	assert.True(t, metric.Gauge.GetValue() >= 0)


	redisClent.Close()
}

