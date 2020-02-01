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

	logger log.Logger

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

		if onceClient.logger == nil {
			onceClient.logger = log.NewLogger()
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

func (MysqlClientOptions) WithLogger(logger log.Logger) Option {
	return func(m *mysqlClient) {
		m.logger = logger
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
func (this *mysqlClient) init() {

	dbConfigs := []DbConfig{}
	err := this.conf.UnmarshalKey("dbs", &dbConfigs)
	if err != nil {
		this.logger.Panicf("Fatal error config file: %s \n", err.Error())
	}

	if len(this.dbConfigs) > 0 {
		dbConfigs = append(dbConfigs, this.dbConfigs...)
	}

	for _, dbConfig := range dbConfigs {
		var DB *gorm.DB
		if len(this.proxy) == 0 {
			DB, err = gorm.Open("mysql", dbConfig.Dsn)
			if err != nil {
				this.logger.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
			}

			DB.DB().SetMaxIdleConns(dbConfig.MaxIdle)
			DB.DB().SetMaxOpenConns(dbConfig.MaxOpen)
			DB.DB().SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime))
		} else {

			dbSQL, err := sql.Open("mysql", dbConfig.Dsn)
			if err != nil {
				this.logger.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
			}

			firstProxy := proxy.NewProxyFactory().GetFirstInstance("db_" + dbConfig.Db, dbSQL, this.proxy...)

			DB, err = gorm.Open("mysql", firstProxy)
			if err != nil {
				this.logger.Panicf("[db] %s ping error : %s", dbConfig.Db, err.Error())
			}

			err = dbSQL.Ping()
			if err != nil {
				this.logger.Panicf("[db] %s ping error : %s", dbConfig.Db, err.Error())
			}

			dbSQL.SetMaxIdleConns(dbConfig.MaxIdle)
			dbSQL.SetMaxOpenConns(dbConfig.MaxOpen)
			dbSQL.SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime))
		}

		this.setDb(dbConfig.Db, DB)

		if this.conf.GetBool("debug") == true {
			DB.LogMode(true)
		}

		//DB.SetLogger(log.L)
		this.logger.Infof("[mysql] %s init success", dbConfig.Db)
	}
}


func (this *mysqlClient) setDb(db_name string, gdb *gorm.DB) bool {
	db_name = strings.ToLower(db_name)

	//m.mysqlLock.Lock()
	this.dbs[db_name] = gdb
	//m.mysqlLock.Unlock()
	return true
}

func (this *mysqlClient) GetDb(db_name string) *gorm.DB {
	return this.getDb(nil, db_name)
}

func (this *mysqlClient) getDb(ctx context.Context, db_name string) *gorm.DB {
	db_name = strings.ToLower(db_name)

	//m.mysqlLock.RLock()
	if db, ok := this.dbs[db_name]; ok {
		//m.mysqlLock.RUnlock()
		return db
	} else {
		//m.mysqlLock.RUnlock()
		this.logger.Errorf("[db] %s not found", db_name)
		return nil
	}
}

func (this *mysqlClient) GetCtxDb(ctx context.Context, db_name string) *gorm.DB {
	return this.getDb(ctx, db_name)
}


func (this *mysqlClient) Ping() []error {
	var errs []error
	var err error
	for _, db := range this.dbs {
		err = db.DB().Ping()
		if err != nil{
			errs = append(errs, err)
		}
	}
	return errs
}


func (this *mysqlClient) Close() {
	var err error
	for _, db := range this.dbs {
		err = db.DB().Close()
		if err != nil{
			this.logger.Errorf(err.Error())
		}
	}

	return
}