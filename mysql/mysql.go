package mysql

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/proxy"
	"github.com/prometheus/client_golang/prometheus"
)

var clientOnce sync.Once

var onceClient *Client

type Client struct {
	gdbs map[string]*gorm.DB

	sqlDbs map[string]*sql.DB

	proxy []func() interface{}

	conf config.Config

	logger log.Logger

	proxyOptions []interface{}

	dbConfigs []DbConfig

	closeChan chan bool

	stateTicker time.Duration

	//for integration tests
	db *sql.DB
}

type Option func(c *Client)

type ClientOptions struct{}

type DbConfig struct {
	Db          string `json:"db",yaml:"db"`
	Dsn         string `json:"dns",yaml:"dsn"`
	MaxIdle     int    `json:"max_idle",yaml:"maxidle"`
	MaxOpen     int    `json:"max_open",yaml:"maxopen"`
	MaxLifetime int    `json:"max_lifetime",yaml:"maxlifetime"`
}

func NewClient(options ...Option) *Client {
	clientOnce.Do(func() {

		onceClient = &Client{
			gdbs:        make(map[string]*gorm.DB),
			sqlDbs:      make(map[string]*sql.DB),
			proxy:       make([]func() interface{}, 0),
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
			onceClient.logger = log.NewLogger()
		}

		onceClient.init()
	})

	return onceClient
}

func (ClientOptions) WithConf(conf config.Config) Option {
	return func(m *Client) {
		m.conf = conf
	}
}

func (ClientOptions) WithLogger(logger log.Logger) Option {
	return func(m *Client) {
		m.logger = logger
	}
}

func (ClientOptions) WithDbConfig(dbConfigs []DbConfig) Option {
	return func(m *Client) {
		m.dbConfigs = dbConfigs
	}
}

func (ClientOptions) WithProxy(proxy ...func() interface{}) Option {
	return func(m *Client) {
		m.proxy = append(m.proxy, proxy...)
	}
}

func (ClientOptions) WithStateTicker(stateTicker time.Duration) Option {
	return func(m *Client) {
		m.stateTicker = stateTicker
	}
}

func (ClientOptions) WithDB(db *sql.DB) Option {
	return func(m *Client) {
		m.db = db
	}
}

// initializes Client.
func (c *Client) init() {

	dbConfigs := make([]DbConfig, 0)
	err := c.conf.UnmarshalKey("dbs", &dbConfigs)
	if err != nil {
		c.logger.Panicf("Fatal error config file: %s \n", err.Error())
	}

	if len(c.dbConfigs) > 0 {
		dbConfigs = append(dbConfigs, c.dbConfigs...)
	}

	for _, dbConfig := range dbConfigs {
		if len(c.proxy) == 0 {
			var DB *gorm.DB

			if c.db != nil {
				DB, err = gorm.Open("mysql", c.db)
			} else {
				DB, err = gorm.Open("mysql", dbConfig.Dsn)
			}

			if err != nil {
				c.logger.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
			}

			DB.DB().SetMaxIdleConns(dbConfig.MaxIdle)
			DB.DB().SetMaxOpenConns(dbConfig.MaxOpen)
			DB.DB().SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime))

			c.setDb(dbConfig.Db, DB, DB.DB())

			if c.conf.GetBool("debug") {
				DB.LogMode(true)
			}
		} else {
			var DB *gorm.DB
			var dbSQL *sql.DB

			if c.db == nil {
				dbSQL, err = sql.Open("mysql", dbConfig.Dsn)
				if err != nil {
					c.logger.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
				}
			} else {
				dbSQL = c.db
			}

			firstProxy := proxy.NewProxyFactory().GetFirstInstance("db_"+dbConfig.Db, dbSQL, c.proxy...)

			DB, err = gorm.Open("mysql", firstProxy)
			if err != nil {
				c.logger.Panicf("[db] %s ping error : %s", dbConfig.Db, err.Error())
			}

			err = dbSQL.Ping()
			if err != nil {
				c.logger.Panicf("[db] %s ping error : %s", dbConfig.Db, err.Error())
			}

			dbSQL.SetMaxIdleConns(dbConfig.MaxIdle)
			dbSQL.SetMaxOpenConns(dbConfig.MaxOpen)
			dbSQL.SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime))

			c.setDb(dbConfig.Db, DB, dbSQL)

			if c.conf.GetBool("debug") {
				DB.LogMode(true)
			}
		}

		go c.Stats()
		//DB.SetLogger(log.L)
		c.logger.Infof("[mysql] %s init success", dbConfig.Db)
	}
}

func (c *Client) setDb(dbName string, gdb *gorm.DB, db *sql.DB) bool {
	dbName = strings.ToLower(dbName)

	//m.mysqlLock.Lock()
	c.gdbs[dbName] = gdb
	c.sqlDbs[dbName] = db

	//m.mysqlLock.Unlock()
	return true
}

func (c *Client) GetDb(dbName string) *gorm.DB {
	return c.getDb(context.Background(), dbName)
}

func (c *Client) getDb(ctx context.Context, dbName string) *gorm.DB {
	dbName = strings.ToLower(dbName)

	//m.mysqlLock.RLock()
	if db, ok := c.gdbs[dbName]; ok {
		//m.mysqlLock.RUnlock()
		return db
	} else {
		//m.mysqlLock.RUnlock()
		c.logger.Errorf("[db] %s not found", dbName)
		return nil
	}
}

func (c *Client) GetCtxDb(ctx context.Context, dbName string) *gorm.DB {
	return c.getDb(ctx, dbName)
}

func (c *Client) Ping() []error {
	var errs []error
	var err error
	for _, db := range c.sqlDbs {
		err = db.Ping()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (c *Client) Close() {
	var err error
	for _, db := range c.gdbs {
		err = db.Close()
		if err != nil {
			c.logger.Errorf(err.Error())
		}
	}

	//c.closeChan <- true
}

func (c *Client) Stats() {

	defer func() {
		if err := recover(); err != nil {
			c.logger.Infof(err.(error).Error())
		}
	}()

	ticker := time.NewTicker(c.stateTicker)
	var stats sql.DBStats

	for {
		select {
		case <-ticker.C:
			for dbName, db := range c.sqlDbs {

				stats = db.Stats()

				maxOpenConnLab := prometheus.Labels{"db": dbName, "stats": "max_open_conn"}
				mysqlStats.With(maxOpenConnLab).Set(float64(stats.MaxOpenConnections))

				openConnLab := prometheus.Labels{"db": dbName, "stats": "open_conn"}
				mysqlStats.With(openConnLab).Set(float64(stats.OpenConnections))

				inUseLab := prometheus.Labels{"db": dbName, "stats": "in_use"}
				mysqlStats.With(inUseLab).Set(float64(stats.InUse))

				idleLab := prometheus.Labels{"db": dbName, "stats": "idle"}
				mysqlStats.With(idleLab).Set(float64(stats.Idle))

				waitCountLab := prometheus.Labels{"db": dbName, "stats": "wait_count"}
				mysqlStats.With(waitCountLab).Set(float64(stats.WaitCount))
			}
		case <-c.closeChan:
			c.logger.Infof("stop stats")
			goto Stop
		}
	}

Stop:
	ticker.Stop()
}
