package redis

import (
	"log"
	"os"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/jukylin/esim/config"
	elog "github.com/jukylin/esim/log"
	"github.com/jukylin/esim/proxy"
	"github.com/prometheus/client_golang/prometheus"
)

var poolRedisOnce sync.Once
var onceRedisClient *RedisClient

type RedisClient struct {
	client *redis.Pool

	proxyConn []func() interface{}

	conf config.Config

	logger elog.Logger

	proxyNum int

	proxyInses []interface{}

	stateTicker time.Duration

	closeChan chan bool
}

type Option func(c *RedisClient)

type RedisClientOptions struct{}

func NewRedisClient(options ...Option) *RedisClient {
	return newPoolRedis(options...)
}

func newPoolRedis(options ...Option) *RedisClient {
	poolRedisOnce.Do(func() {

		onceRedisClient = &RedisClient{
			proxyConn:   make([]func() interface{}, 0),
			stateTicker: 10 * time.Second,
			closeChan:   make(chan bool, 1),
		}

		for _, option := range options {
			option(onceRedisClient)
		}

		if onceRedisClient.conf == nil {
			onceRedisClient.conf = config.NewNullConfig()
		}

		if onceRedisClient.logger == nil {
			onceRedisClient.logger = elog.NewLogger()
		}

		onceRedisClient.proxyNum = len(onceRedisClient.proxyConn)
		if onceRedisClient.proxyNum > 0 {
			onceRedisClient.proxyInses = proxy.NewProxyFactory().
				GetInstances("redis", onceRedisClient.proxyConn...)
		}

		redisMaxActive := onceRedisClient.conf.GetInt("redis_max_active")
		if redisMaxActive == 0 {
			redisMaxActive = 500
		}

		redisMaxIdle := onceRedisClient.conf.GetInt("redis_max_idle")
		if redisMaxActive == 0 {
			redisMaxIdle = 100
		}

		redisIdleTimeout := onceRedisClient.conf.GetInt("redis_idle_time_out")
		if redisIdleTimeout == 0 {
			redisIdleTimeout = 600
		}

		redis_etc1_host := onceRedisClient.conf.GetString("redis_host")
		if redis_etc1_host == "" {
			redis_etc1_host = "0.0.0.0"
		}
		redis_etc1_port := onceRedisClient.conf.GetString("redis_post")
		if redis_etc1_port == "" {
			redis_etc1_port = "6379"
		}

		redis_etc1_password := onceRedisClient.conf.GetString("redis_password")

		redis_read_time_out := onceRedisClient.conf.GetInt64("redis_read_time_out")
		if redis_read_time_out == 0 {
			redis_read_time_out = 300
		}

		redis_write_time_out := onceRedisClient.conf.GetInt64("redis_write_time_out")
		if redis_write_time_out == 0 {
			redis_write_time_out = 300
		}

		redis_conn_time_out := onceRedisClient.conf.GetInt64("redis_conn_time_out")
		if redis_conn_time_out == 0 {
			redis_conn_time_out = 300
		}

		onceRedisClient.client = &redis.Pool{
			MaxIdle:     redisMaxIdle,
			MaxActive:   redisMaxActive,
			IdleTimeout: time.Duration(redisIdleTimeout) * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", redis_etc1_host+":"+redis_etc1_port,
					redis.DialReadTimeout(time.Duration(redis_read_time_out)*time.Millisecond),
					redis.DialWriteTimeout(time.Duration(redis_write_time_out)*time.Millisecond),
					redis.DialConnectTimeout(time.Duration(redis_conn_time_out)*time.Millisecond))

				if err != nil {
					onceRedisClient.logger.Panicf("redis.Dial err: %s", err.Error())
					return nil, err
				}
				if redis_etc1_password != "" {
					if _, err := c.Do("AUTH", redis_etc1_password); err != nil {
						c.Close()
						onceRedisClient.logger.Panicf("redis.AUTH err: %s", err.Error())
						return nil, err
					}
				}
				// 选择db
				c.Do("SELECT", 0)

				if onceRedisClient.conf.GetBool("debug") == true {
					c = redis.NewLoggingConn(
						c, log.New(os.Stdout, "",
							log.Ldate|log.Ltime|log.Lshortfile), "")
				}
				return c, nil
			},
		}

		if onceRedisClient.conf.GetString("runmode") == "pro" {
			//conn success ？
			rc := onceRedisClient.client.Get()
			if rc.Err() != nil {
				onceRedisClient.logger.Panicf(rc.Err().Error())
			}
			rc.Close()
		}

		go onceRedisClient.Stats()

		onceRedisClient.logger.Infof("[redis] init success %s : %s", redis_etc1_host, redis_etc1_port)
	})

	return onceRedisClient
}

func (RedisClientOptions) WithConf(conf config.Config) Option {
	return func(r *RedisClient) {
		r.conf = conf
	}
}

func (RedisClientOptions) WithLogger(logger elog.Logger) Option {
	return func(r *RedisClient) {
		r.logger = logger
	}
}

func (RedisClientOptions) WithProxy(proxyConn ...func() interface{}) Option {
	return func(r *RedisClient) {
		r.proxyConn = append(r.proxyConn, proxyConn...)
	}
}

func (RedisClientOptions) WithStateTicker(stateTicker time.Duration) Option {
	return func(r *RedisClient) {
		r.stateTicker = stateTicker
	}
}

//使用原生redisgo
func (this *RedisClient) GetRedisConn() redis.Conn {

	rc := this.client.Get()

	return rc
}

//Recommended
func (this *RedisClient) GetCtxRedisConn() ContextConn {

	rc := this.client.Get()

	facadeProxy := NewFacadeProxy()
	facadeProxy.NextProxy(rc)

	var firstProxy ContextConn
	if this.proxyNum > 0 && rc.Err() == nil {
		firstProxy = this.proxyInses[len(this.proxyInses)-1].(ContextConn)
		firstProxy.(proxy.Proxy).NextProxy(facadeProxy)
	} else {
		firstProxy = facadeProxy
	}

	return firstProxy
}

func (this *RedisClient) Close() {
	this.client.Close()
	this.closeChan <- true
}

func (this *RedisClient) Ping() error {
	conn := this.client.Get()
	return conn.Err()
}

func (this *RedisClient) Stats() {
	ticker := time.NewTicker(this.stateTicker)
	var stats redis.PoolStats

	for {
		select {
		case <-ticker.C:

			stats = this.client.Stats()

			activeCountLab := prometheus.Labels{"stats": "active_count"}
			redisStats.With(activeCountLab).Set(float64(stats.ActiveCount))

			idleCountLab := prometheus.Labels{"stats": "idle_count"}
			redisStats.With(idleCountLab).Set(float64(stats.IdleCount))

		case <-this.closeChan:
			this.logger.Infof("stop stats")
			goto Stop
		}
	}
Stop:
	ticker.Stop()
}
