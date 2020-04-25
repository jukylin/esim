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
var onceRedisClient *Client

type Client struct {
	client *redis.Pool

	proxyConn []func() interface{}

	conf config.Config

	logger elog.Logger

	proxyNum int

	proxyInses []interface{}

	stateTicker time.Duration

	closeChan chan bool
}

type Option func(c *Client)

type ClientOptions struct{}

func NewClient(options ...Option) *Client {
	return newPoolRedis(options...)
}

func newPoolRedis(options ...Option) *Client {
	poolRedisOnce.Do(func() {

		onceRedisClient = &Client{
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

		redisEtc1Host := onceRedisClient.conf.GetString("redis_host")
		if redisEtc1Host == "" {
			redisEtc1Host = "0.0.0.0"
		}

		redisEtc1Port := onceRedisClient.conf.GetString("redis_post")
		if redisEtc1Port == "" {
			redisEtc1Port = "6379"
		}

		redisEtc1Password := onceRedisClient.conf.GetString("redis_password")

		redisReadTimeOut := onceRedisClient.conf.GetInt64("redis_read_time_out")
		if redisReadTimeOut == 0 {
			redisReadTimeOut = 300
		}

		redisWriteTimeOut := onceRedisClient.conf.GetInt64("redis_write_time_out")
		if redisWriteTimeOut == 0 {
			redisWriteTimeOut = 300
		}

		redisConnTimeOut := onceRedisClient.conf.GetInt64("redis_conn_time_out")
		if redisConnTimeOut == 0 {
			redisConnTimeOut = 300
		}

		onceRedisClient.client = &redis.Pool{
			MaxIdle:     redisMaxIdle,
			MaxActive:   redisMaxActive,
			IdleTimeout: time.Duration(redisIdleTimeout) * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", redisEtc1Host+":"+redisEtc1Port,
					redis.DialReadTimeout(time.Duration(redisReadTimeOut)*time.Millisecond),
					redis.DialWriteTimeout(time.Duration(redisWriteTimeOut)*time.Millisecond),
					redis.DialConnectTimeout(time.Duration(redisConnTimeOut)*time.Millisecond))

				if err != nil {
					onceRedisClient.logger.Panicf("redis.Dial err: %s", err.Error())
					return nil, err
				}
				if redisEtc1Password != "" {
					if _, err := c.Do("AUTH", redisEtc1Password); err != nil {
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

		onceRedisClient.logger.Infof("[redis] init success %s : %s", redisEtc1Host, redisEtc1Port)
	})

	return onceRedisClient
}

func (ClientOptions) WithConf(conf config.Config) Option {
	return func(r *Client) {
		r.conf = conf
	}
}

func (ClientOptions) WithLogger(logger elog.Logger) Option {
	return func(r *Client) {
		r.logger = logger
	}
}

func (ClientOptions) WithProxy(proxyConn ...func() interface{}) Option {
	return func(r *Client) {
		r.proxyConn = append(r.proxyConn, proxyConn...)
	}
}

func (ClientOptions) WithStateTicker(stateTicker time.Duration) Option {
	return func(r *Client) {
		r.stateTicker = stateTicker
	}
}

//使用原生redisgo
func (c *Client) GetRedisConn() redis.Conn {

	rc := c.client.Get()

	return rc
}

//Recommended
func (c *Client) GetCtxRedisConn() ContextConn {

	rc := c.client.Get()

	facadeProxy := NewFacadeProxy()
	facadeProxy.NextProxy(rc)

	var firstProxy ContextConn
	if c.proxyNum > 0 && rc.Err() == nil {
		firstProxy = c.proxyInses[len(c.proxyInses)-1].(ContextConn)
		firstProxy.(proxy.Proxy).NextProxy(facadeProxy)
	} else {
		firstProxy = facadeProxy
	}

	return firstProxy
}

func (c *Client) Close() {
	c.client.Close()
	c.closeChan <- true
}

func (c *Client) Ping() error {
	conn := c.client.Get()
	return conn.Err()
}

func (c *Client) Stats() {
	ticker := time.NewTicker(c.stateTicker)
	var stats redis.PoolStats

	for {
		select {
		case <-ticker.C:

			stats = c.client.Stats()

			activeCountLab := prometheus.Labels{"stats": "active_count"}
			redisStats.With(activeCountLab).Set(float64(stats.ActiveCount))

			idleCountLab := prometheus.Labels{"stats": "idle_count"}
			redisStats.With(idleCountLab).Set(float64(stats.IdleCount))

		case <-c.closeChan:
			c.logger.Infof("stop stats")
			goto Stop
		}
	}
Stop:
	ticker.Stop()
}
