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

var poolOnce sync.Once
var onceClient *Client

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
	poolOnce.Do(func() {

		onceClient = &Client{
			proxyConn:   make([]func() interface{}, 0),
			stateTicker: 10 * time.Second,
			closeChan:   make(chan bool, 1),
		}

		for _, option := range options {
			option(onceClient)
		}

		if onceClient.conf == nil {
			onceClient.conf = config.NewNullConfig()
		}

		if onceClient.logger == nil {
			onceClient.logger = elog.NewLogger()
		}

		onceClient.proxyNum = len(onceClient.proxyConn)
		if onceClient.proxyNum > 0 {
			onceClient.proxyInses = proxy.NewProxyFactory().
				GetInstances("redis", onceClient.proxyConn...)
		}

		redisMaxActive := onceClient.conf.GetInt("redis_max_active")
		if redisMaxActive == 0 {
			redisMaxActive = 500
		}

		redisMaxIdle := onceClient.conf.GetInt("redis_max_idle")
		if redisMaxActive == 0 {
			redisMaxIdle = 100
		}

		redisIdleTimeout := onceClient.conf.GetInt("redis_idle_time_out")
		if redisIdleTimeout == 0 {
			redisIdleTimeout = 600
		}

		redisEtc1Host := onceClient.conf.GetString("redis_host")
		if redisEtc1Host == "" {
			redisEtc1Host = "0.0.0.0"
		}

		redisEtc1Port := onceClient.conf.GetString("redis_post")
		if redisEtc1Port == "" {
			redisEtc1Port = "6379"
		}

		redisEtc1Password := onceClient.conf.GetString("redis_password")

		redisReadTimeOut := onceClient.conf.GetInt64("redis_read_time_out")
		if redisReadTimeOut == 0 {
			redisReadTimeOut = 300
		}

		redisWriteTimeOut := onceClient.conf.GetInt64("redis_write_time_out")
		if redisWriteTimeOut == 0 {
			redisWriteTimeOut = 300
		}

		redisConnTimeOut := onceClient.conf.GetInt64("redis_conn_time_out")
		if redisConnTimeOut == 0 {
			redisConnTimeOut = 300
		}

		onceClient.client = &redis.Pool{
			MaxIdle:     redisMaxIdle,
			MaxActive:   redisMaxActive,
			IdleTimeout: time.Duration(redisIdleTimeout) * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", redisEtc1Host+":"+redisEtc1Port,
					redis.DialReadTimeout(time.Duration(redisReadTimeOut)*time.Millisecond),
					redis.DialWriteTimeout(time.Duration(redisWriteTimeOut)*time.Millisecond),
					redis.DialConnectTimeout(time.Duration(redisConnTimeOut)*time.Millisecond))

				if err != nil {
					onceClient.logger.Panicf("redis.Dial err: %s", err.Error())
					return nil, err
				}
				if redisEtc1Password != "" {
					if _, err := c.Do("AUTH", redisEtc1Password); err != nil {
						err = c.Close()
						onceClient.logger.Panicf(err.Error())
						onceClient.logger.Panicf("redis.AUTH err: %s", err.Error())
						return nil, err
					}
				}
				// 选择db
				c.Do("SELECT", 0)

				if onceClient.conf.GetBool("debug") {
					c = redis.NewLoggingConn(
						c, log.New(os.Stdout, "",
							log.Ldate|log.Ltime|log.Lshortfile), "")
				}
				return c, nil
			},
		}

		if onceClient.conf.GetString("runmode") == "pro" {
			//conn success ？
			rc := onceClient.client.Get()
			if rc.Err() != nil {
				onceClient.logger.Panicf(rc.Err().Error())
			}
			rc.Close()
		}

		go onceClient.Stats()

		onceClient.logger.Infof("[redis] init success %s : %s", redisEtc1Host, redisEtc1Port)
	})

	return onceClient
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

func (c *Client) Close() error {
	err := c.client.Close()
	c.closeChan <- true
	return err
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
