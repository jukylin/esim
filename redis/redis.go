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

	stateTicker time.Duration

	closeChan chan bool

	redisMaxActive int

	redisMaxIdle int

	redisIdleTimeout int

	redisHost string

	redisPort string

	redisPassword string

	redisReadTimeOut int64

	redisWriteTimeOut int64

	redisConnTimeOut int64
}

type Option func(c *Client)

type ClientOptions struct{}

func NewClient(options ...Option) *Client {
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
		//if onceClient.proxyNum > 0 {
		//	onceClient.proxyInses = proxy.NewProxyFactory().
		//		GetInstances("redis", onceClient.proxyConn...)
		//}

		onceClient.redisMaxActive = onceClient.conf.GetInt("redis_max_active")
		if onceClient.redisMaxActive == 0 {
			onceClient.redisMaxActive = 500
		}

		onceClient.redisMaxIdle = onceClient.conf.GetInt("redis_max_idle")
		if onceClient.redisMaxIdle == 0 {
			onceClient.redisMaxIdle = 100
		}

		onceClient.redisIdleTimeout = onceClient.conf.GetInt("redis_idle_time_out")
		if onceClient.redisIdleTimeout == 0 {
			onceClient.redisIdleTimeout = 600
		}

		onceClient.redisHost = onceClient.conf.GetString("redis_host")
		if onceClient.redisHost == "" {
			onceClient.redisHost = "0.0.0.0"
		}

		onceClient.redisPort = onceClient.conf.GetString("redis_post")
		if onceClient.redisPort == "" {
			onceClient.redisPort = "6379"
		}

		onceClient.redisPassword = onceClient.conf.GetString("redis_password")

		onceClient.redisReadTimeOut = onceClient.conf.GetInt64("redis_read_time_out")
		if onceClient.redisReadTimeOut == 0 {
			onceClient.redisReadTimeOut = 300
		}

		onceClient.redisWriteTimeOut = onceClient.conf.GetInt64("redis_write_time_out")
		if onceClient.redisWriteTimeOut == 0 {
			onceClient.redisWriteTimeOut = 300
		}

		onceClient.redisConnTimeOut = onceClient.conf.GetInt64("redis_conn_time_out")
		if onceClient.redisConnTimeOut == 0 {
			onceClient.redisConnTimeOut = 300
		}

		onceClient.initPool()

		if onceClient.conf.GetString("runmode") == "pro" {
			// conn success ？
			rc := onceClient.client.Get()
			if rc.Err() != nil {
				onceClient.logger.Panicf(rc.Err().Error())
			}
			rc.Close()
		}

		go onceClient.Stats()

		onceClient.logger.Infof("[redis] init success %s : %s",
			onceClient.redisHost, onceClient.redisPort)
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

// initClient Initialize the pool of connections.
func (c *Client) initPool() {
	c.client = &redis.Pool{
		MaxIdle:     c.redisMaxIdle,
		MaxActive:   c.redisMaxActive,
		IdleTimeout: time.Duration(c.redisIdleTimeout) * time.Second,
		Wait:true,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", c.redisHost+":"+c.redisPort,
				redis.DialReadTimeout(time.Duration(c.redisReadTimeOut)*time.Millisecond),
				redis.DialWriteTimeout(time.Duration(c.redisWriteTimeOut)*time.Millisecond),
				redis.DialConnectTimeout(time.Duration(c.redisConnTimeOut)*time.Millisecond))

			if err != nil {
				c.logger.Errorf("redis.Dial err: %s", err.Error())
				return nil, err
			}

			if c.redisPassword != "" {
				if _, err = conn.Do("AUTH", c.redisPassword); err != nil {
					c.logger.Errorf("redis.AUTH err: %s", err.Error())
					err = conn.Close()
					c.logger.Errorf(err.Error())
					return nil, err
				}
			}

			// select db
			_, err = conn.Do("SELECT", 0)
			if err != nil {
				c.logger.Errorf("Select err: %s", err.Error())
				return nil, err
			}

			if c.conf.GetBool("debug") {
				conn = redis.NewLoggingConn(
					conn, log.New(os.Stdout, "",
						log.Ldate|log.Ltime|log.Lshortfile), "")
			}
			return conn, nil
		},
	}
}

func (c *Client) GetRedisConn() redis.Conn {
	rc := c.client.Get()

	return rc
}

// Recommended.
func (c *Client) GetCtxRedisConn() ContextConn {
	rc := c.client.Get()

	facadeProxy := NewFacadeProxy()
	facadeProxy.NextProxy(rc)

	var firstProxy ContextConn
	if c.proxyNum > 0 && rc.Err() == nil {
		firstProxy = proxy.NewProxyFactory().
			GetFirstInstance("redis", facadeProxy, c.proxyConn...).(ContextConn)
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
