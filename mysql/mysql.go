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

var mysqlOnce sync.Once

var onceClient *MysqlClient

type MysqlClient struct {
	dbs map[string]*gorm.DB

	proxy []func () interface{}

	conf config.Config

	logger log.Logger

	proxyOptions []interface{}

	dbConfigs []DbConfig

	closeChan chan bool

	stateTicker time.Duration

	//for integration tests
	db *sql.DB
}

type Option func(c *MysqlClient)

type MysqlClientOptions struct{}

type DbConfig struct {
	Db          string `json:"db",yaml:"db"`
	Dsn         string `json:"dns",yaml:"dsn"`
	MaxIdle     int    `json:"max_idle",yaml:"maxidle"`
	MaxOpen     int    `json:"max_open",yaml:"maxopen"`
	MaxLifetime int    `json:"max_lifetime",yaml:"maxlifetime"`
}

func NewMysqlClient(options ...Option) *MysqlClient {
	mysqlOnce.Do(func() {

		onceClient = &MysqlClient{
			dbs:       make(map[string]*gorm.DB),
			proxy: make([]func () interface{}, 0),
			stateTicker : 10 * time.Second,
			closeChan: make(chan bool, 1),
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
	return func(m *MysqlClient) {
		m.conf = conf
	}
}

func (MysqlClientOptions) WithLogger(logger log.Logger) Option {
	return func(m *MysqlClient) {
		m.logger = logger
	}
}

func (MysqlClientOptions) WithDbConfig(dbConfigs []DbConfig) Option {
	return func(m *MysqlClient) {
		m.dbConfigs = dbConfigs
	}
}

func (MysqlClientOptions) WithProxy(proxy ...func() interface{}) Option {
	return func(m *MysqlClient) {
		m.proxy = append(m.proxy, proxy...)
	}
}

func (MysqlClientOptions) WithStateTicker(stateTicker time.Duration) Option {
	return func(m *MysqlClient) {
		m.stateTicker = stateTicker
	}
}


func (MysqlClientOptions) WithDB(db *sql.DB) Option {
	return func(m *MysqlClient) {
		m.db = db
	}
}

// initializes mysqlClient.
func (this *MysqlClient) init() {

	dbConfigs := []DbConfig{}
	err := this.conf.UnmarshalKey("dbs", &dbConfigs)
	if err != nil {
		this.logger.Panicf("Fatal error config file: %s \n", err.Error())
	}

	if len(this.dbConfigs) > 0 {
		dbConfigs = append(dbConfigs, this.dbConfigs...)
	}

	for _, dbConfig := range dbConfigs {
		if len(this.proxy) == 0 {
			var DB *gorm.DB

			if this.db != nil{
				DB, err = gorm.Open("mysql", this.db)
			}else{
				DB, err = gorm.Open("mysql", dbConfig.Dsn)
			}

			if err != nil {
				this.logger.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
			}

			DB.DB().SetMaxIdleConns(dbConfig.MaxIdle)
			DB.DB().SetMaxOpenConns(dbConfig.MaxOpen)
			DB.DB().SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime))

			this.setDb(dbConfig.Db, DB)

			if this.conf.GetBool("debug") == true {
				DB.LogMode(true)
			}
		} else {
			var DB *gorm.DB
			var dbSQL *sql.DB

			if this.db == nil{
				dbSQL, err = sql.Open("mysql", dbConfig.Dsn)
				if err != nil {
					this.logger.Panicf("[db] %s init error : %s", dbConfig.Db, err.Error())
				}
			}else{
				dbSQL = this.db
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

			this.setDb(dbConfig.Db, DB)

			if this.conf.GetBool("debug") == true {
				DB.LogMode(true)
			}
		}

		go this.Stats()
		//DB.SetLogger(log.L)
		this.logger.Infof("[mysql] %s init success", dbConfig.Db)
	}
}


func (this *MysqlClient) setDb(db_name string, gdb *gorm.DB) bool {
	db_name = strings.ToLower(db_name)

	//m.mysqlLock.Lock()
	this.dbs[db_name] = gdb
	//m.mysqlLock.Unlock()
	return true
}

func (this *MysqlClient) GetDb(db_name string) *gorm.DB {
	return this.getDb(nil, db_name)
}

func (this *MysqlClient) getDb(ctx context.Context, db_name string) *gorm.DB {
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

func (this *MysqlClient) GetCtxDb(ctx context.Context, db_name string) *gorm.DB {
	return this.getDb(ctx, db_name)
}


func (this *MysqlClient) Ping() []error {
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


func (this *MysqlClient) Close() {
	var err error
	for _, db := range this.dbs {
		err = db.Close()
		if err != nil{
			this.logger.Errorf(err.Error())
		}
	}

	//this.closeChan <- true
	return
}


func (this *MysqlClient) Stats() {

	ticker := time.NewTicker(this.stateTicker)
	var stats sql.DBStats

	for {
		select {
		case <- ticker.C:
			for db_name, db := range this.dbs {

				stats = db.DB().Stats()
				maxOpenConnLab := prometheus.Labels{"db": db_name, "stats" : "max_open_conn"}
				mysqlStats.With(maxOpenConnLab).Set(float64(stats.MaxOpenConnections))

				openConnLab := prometheus.Labels{"db": db_name, "stats" : "open_conn"}
				mysqlStats.With(openConnLab).Set(float64(stats.OpenConnections))

				inUseLab := prometheus.Labels{"db": db_name, "stats" : "in_use"}
				mysqlStats.With(inUseLab).Set(float64(stats.InUse))

				idleLab := prometheus.Labels{"db": db_name, "stats" : "idle"}
				mysqlStats.With(idleLab).Set(float64(stats.Idle))

				waitCountLab := prometheus.Labels{"db": db_name, "stats" : "wait_count"}
				mysqlStats.With(waitCountLab).Set(float64(stats.WaitCount))
			}
		case <- this.closeChan:
			this.logger.Infof("stop stats")
			goto Stop
		}
	}

Stop:
	ticker.Stop()

	return
}