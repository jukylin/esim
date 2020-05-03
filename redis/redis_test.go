package redis

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	logger := log.NewLogger()

	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}
	opt := &dockertest.RunOptions{
		Repository: "redis",
		Tag:        "latest",
	}

	resource, err := pool.RunWithOptions(opt, func(hostConfig *dc.HostConfig) {
		hostConfig.PortBindings = map[dc.Port][]dc.PortBinding{
			"6379/tcp": {{HostIP: "", HostPort: "6379"}},
		}
	})
	if err != nil {
		logger.Fatalf("Could not start resource: %s", err)
	}

	resource.Expire(60)

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		logger.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

//使用 proxyConn 代理请求
func TestGetProxyConn(t *testing.T) {
	poolRedisOnce = sync.Once{}

	redisClientOptions := ClientOptions{}
	redisClent := NewClient(
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

func TestGetNotProxyConn(t *testing.T) {
	poolRedisOnce = sync.Once{}

	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewClient(
		redisClientOptions.WithConf(memConfig),
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

func TestMonitorProxy_Do(t *testing.T) {
	poolRedisOnce = sync.Once{}

	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := MonitorProxyOptions{}
				return NewMonitorProxy(
					monitorProxyOptions.WithLogger(log.NewLogger()),
				)
			},
		),
	)

	ctx := context.Background()

	conn := redisClent.GetCtxRedisConn()

	_, err := String(conn.Do(ctx, "get", "name"))
	assert.Nil(t, err)
	//conn.Close()
	//redisClent.Close()
}

func TestMulLevelProxy_Do(t *testing.T) {
	poolRedisOnce = sync.Once{}

	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	spyProxy := NewSpyProxy(log.NewLogger(), "spyProxy")

	redisClent := NewClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithProxy(
			func() interface{} {
				monitorProxyOptions := MonitorProxyOptions{}
				return NewMonitorProxy(
					monitorProxyOptions.WithLogger(log.NewLogger()),
				)
			},
			func() interface{} {
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

	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithProxy(
			func() interface{} {
				return NewStubsProxy(log.NewLogger(), "stubsProxy")
			},
		),
	)

	conn := redisClent.GetRedisConn()
	conn.Do("set", "name", "test")
	conn.Do("set", "version", "2.0")
	conn.Close()
	wg := sync.WaitGroup{}
	//ctx := context.Background()
	wg.Add(2)

	go func() {
		conn := redisClent.GetRedisConn()
		name, err := String(conn.Do("get", "name"))
		assert.Nil(t, err)
		assert.Equal(t, "test", name)
		conn.Close()
		wg.Done()
	}()

	go func() {
		conn := redisClent.GetRedisConn()
		version, err := String(conn.Do("get", "version"))
		assert.Nil(t, err)
		assert.Equal(t, "2.0", version)
		conn.Close()
		wg.Done()
	}()
	wg.Wait()

	redisClent.Close()
}

func Benchmark_MulGo_Do(b *testing.B) {

	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewClient(
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
					name, err := String(conn.Do(ctx, "get", "name"))
					assert.Nil(b, err)
					assert.Equal(b, "test", name)
					conn.Close()
					wg.Done()
				}()

				go func() {
					conn := redisClent.GetCtxRedisConn()
					version, err := String(conn.Do(ctx, "get", "version"))
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

	redisClientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", true)

	redisClent := NewClient(
		redisClientOptions.WithConf(memConfig),
		redisClientOptions.WithStateTicker(10*time.Microsecond),
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
