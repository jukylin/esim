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
)

var mysqlOnce sync.Once

var onceClient *mysqlClient

type mysqlClient struct {
	dbs map[string]*gorm.DB

	mysqlLock *sync.RWMutex

	proxy []func () interface{}

	conf config.Config

	log log.Logger

	proxyOptions []interface{}

	dbConfigs []DbConfig
}

type Option func(c *mysqlClient)

type MysqlClientOptions struct{}

type DbConfig struct {
	Db          string `json:"db",yaml:"db"`
	Dsn         string `json:"dns",yaml:"dsn"`
	MaxIdle     int    `json:"max_idle",yaml:"maxidle"`
	MaxOpen     int    `json:"max_open",yaml:"maxopen"`
	MaxLifetime int    `json:"max_lifetime",yaml:"maxlifetime"`
}

func NewMysqlClient(options ...Option) *mysqlClient {
	mysqlOnce.Do(func() {

		onceClient = &mysqlClient{
			dbs:       make(map[string]*gorm.DB),
			mysqlLock: new(sync.RWMutex),
			proxy: make([]func () interface{}, 0),
		}

		for _, option := range options {
			option(onceClient)
		}

		if onceClient.conf == nil {
			onceClient.conf = config.NewNullConfig()
		}

		if onceClient.log == nil {
			onceClient.log = log.NewLogger()
		}

		onceClient.init()
	})

	return onceClient
}

func (MysqlClientOptions) WithConf(conf config.Config) Option {
	return func(m *mysqlClient) {
		m.conf = conf
	}
}

func (MysqlClientOptions) WithLogger(log log.Logger) Option {
	return func(m *mysqlClient) {
		m.log = log
	}
}

func (MysqlClientOptions) WithDbConfig(dbConfigs []DbConfig) Option {
	return func(m *mysqlClient) {
		m.dbConfigs = dbConfigs
	}
}

func (MysqlClientOptions) WithProxy(proxy ...func() interface{}) Option {
	return func(m *mysqlClient) {
		m.proxy = append(m.proxy, proxy...)
	}
}

// initializes mysqlClient.
func (m *mysqlClient) init() {

	dbConfigs := []DbConfig{}
	err := m.conf.UnmarshalKey("dbs", &dbConfigs)
	if err != nil {
		m.log.Panicf("Fatal error config file: %s \n", err.Error())
	}

	if len(m.dbConfigs) > 0 {
		dbConfigs = append(dbConfigs, m.dbConfigs...)
	}

	for _, dbConfig := range dbConfigs {
		var DB *gorm.DB
		if len(m.proxy) == 0 {
			DB, err = gorm.Open("mysql", dbConfig.Dsn)
			if err != nil {
				m.log.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
			}

			DB.DB().SetMaxIdleConns(dbConfig.MaxIdle)
			DB.DB().SetMaxOpenConns(dbConfig.MaxOpen)
			DB.DB().SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime))
		} else {

			dbSQL, err := sql.Open("mysql", dbConfig.Dsn)
			if err != nil {
				m.log.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
			}

			firstProxy := proxy.NewProxyFactory().GetFirstInstance("db_" + dbConfig.Db, dbSQL, m.proxy...)

			DB, err = gorm.Open("mysql", firstProxy)
			if err != nil {
				m.log.Panicf("[db] %s ping error : %s", dbConfig.Db, err.Error())
			}

			err = dbSQL.Ping()
			if err != nil {
				m.log.Panicf("[db] %s ping error : %s", dbConfig.Db, err.Error())
			}

			dbSQL.SetMaxIdleConns(dbConfig.MaxIdle)
			dbSQL.SetMaxOpenConns(dbConfig.MaxOpen)
			dbSQL.SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime))
		}

		m.setDb(dbConfig.Db, DB)

		if m.conf.GetBool("debug") == true {
			DB.LogMode(true)
		}

		//DB.SetLogger(log.L)
		m.log.Infof("[mysql] %s init success", dbConfig.Db)
	}
}


func (m *mysqlClient) setDb(db_name string, gdb *gorm.DB) bool {
	db_name = strings.ToLower(db_name)

	//m.mysqlLock.Lock()
	m.dbs[db_name] = gdb
	//m.mysqlLock.Unlock()
	return true
}

func (m *mysqlClient) GetDb(db_name string) *gorm.DB {
	return m.getDb(nil, db_name)
}

func (m *mysqlClient) getDb(ctx context.Context, db_name string) *gorm.DB {
	db_name = strings.ToLower(db_name)

	//m.mysqlLock.RLock()
	if db, ok := m.dbs[db_name]; ok {
		//m.mysqlLock.RUnlock()
		return db
	} else {
		//m.mysqlLock.RUnlock()
		m.log.Errorf("[db] %s not found", db_name)
		return nil
	}
}

func (m *mysqlClient) GetCtxDb(ctx context.Context, db_name string) *gorm.DB {
	return m.getDb(ctx, db_name)
}

//ping first db
func (m *mysqlClient) Ping() error {
	for _, db := range m.dbs {
		return db.DB().Ping()
	}
	return nil
}
