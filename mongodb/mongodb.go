package mongodb

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/jukylin/esim/config"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"github.com/jukylin/esim/log"
	"github.com/prometheus/client_golang/prometheus"
)

var mgoOnce sync.Once
var onceMgoClient *MgoClient

type MgoClient struct {
	Mgos map[string]*mongo.Client

	conf config.Config

	logger log.Logger

	monitorEvents []func() MonitorEvent

	mgoConfig []MgoConfig

	eventOptions []EventOption
}

type mongoBackEvent struct {
	succEvent *event.CommandSucceededEvent

	failedEvent *event.CommandFailedEvent
}

type Option func(c *MgoClient)

type MgoClientOptions struct{}

func NewMongo(options ...Option) *MgoClient {
	mgoOnce.Do(func() {
		onceMgoClient = &MgoClient{
			Mgos:    make(map[string]*mongo.Client),
		}

		for _, option := range options {
			option(onceMgoClient)
		}

		if onceMgoClient.conf == nil {
			onceMgoClient.conf = config.NewNullConfig()
		}

		if onceMgoClient.logger == nil {
			onceMgoClient.logger = log.NewLogger()
		}

		onceMgoClient.init()
	})

	return onceMgoClient
}

func (MgoClientOptions) WithConf(conf config.Config) Option {
	return func(m *MgoClient) {
		m.conf = conf
	}
}

func (MgoClientOptions) WithLogger(logger log.Logger) Option {
	return func(m *MgoClient) {
		m.logger = logger
	}
}

func (MgoClientOptions) WithDbConfig(dbConfigs []MgoConfig) Option {
	return func(m *MgoClient) {
		m.mgoConfig = dbConfigs
	}
}

func (MgoClientOptions) WithMonitorEvent(mongoEvent ...func() MonitorEvent) Option {
	return func(m *MgoClient) {
		m.monitorEvents = mongoEvent
	}
}


type MgoConfig struct {
	Db  string `json:"db",yaml:"db"`
	Uri string `json:"uri",yaml:"uri"`
}

func (this *MgoClient) init() {

	mgoConfigs := []MgoConfig{}
	err := this.conf.UnmarshalKey("mgos", &mgoConfigs)
	if err != nil {
		this.logger.Panicf("Fatal error config file: %s \n", err.Error())
	}

	if len(this.mgoConfig) > 0 {
		mgoConfigs = append(mgoConfigs, this.mgoConfig...)
	}

	for _, mgo := range mgoConfigs {

		clientOptions := options.Client()
		clientOptions.ApplyURI(mgo.Uri)

		if this.monitorEvents != nil {
			firstEvent := this.initMonitorMulLevelEvent(mgo.Db)
			//事件监控
			eventComMon := &event.CommandMonitor{
				Started: func(ctx context.Context, startEvent *event.CommandStartedEvent) {
					exec_command, ok := ctx.Value("command").(*string)
					if ok == true {
						*exec_command = startEvent.Command.String()
					}
					firstEvent.Start(ctx, startEvent)
				},
				Succeeded: func(ctx context.Context, succEvent *event.CommandSucceededEvent) {
					if succEvent.CommandName != "ping" {
						firstEvent.SucceededEvent(ctx, succEvent)
					}
				},
				Failed: func(ctx context.Context, failedEvent *event.CommandFailedEvent) {
					if failedEvent.CommandName != "ping" {
						firstEvent.FailedEvent(ctx, failedEvent)
					}
				},
			}
			clientOptions.SetMonitor(eventComMon)
		}

		//池子监控
		poolMon := &event.PoolMonitor{
			Event: func(pev *event.PoolEvent) {
				this.poolEvent(pev)
			},
		}
		clientOptions.SetPoolMonitor(poolMon)

		mgo_connect_timeout := this.conf.GetInt64("mgo_connect_timeout")
		if mgo_connect_timeout != 0 {
			clientOptions.SetConnectTimeout(time.Duration(mgo_connect_timeout) * time.Millisecond)
			clientOptions.SetServerSelectionTimeout(time.Duration(mgo_connect_timeout) * time.Millisecond)
		}

		mgo_max_conn_idle_time := this.conf.GetInt64("mgo_max_conn_idle_time")
		if mgo_max_conn_idle_time != 0 {
			clientOptions.SetMaxConnIdleTime(time.Duration(mgo_max_conn_idle_time) * time.Minute)
		}

		mgo_max_pool_size := this.conf.GetUint64("mgo_max_pool_size")
		if mgo_max_pool_size != 0 {
			clientOptions.SetMaxPoolSize(mgo_max_pool_size)
		}

		mgo_min_pool_size := this.conf.GetUint64("mgo_min_pool_size")
		if mgo_min_pool_size != 0 {
			clientOptions.SetMinPoolSize(mgo_min_pool_size)
		}

		client, err := mongo.NewClient(clientOptions)
		if err != nil {
			this.logger.Panicf("new mongo client error: %s , uri: %s \n", err.Error(), mgo.Uri)
		}

		ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)

		err = client.Connect(ctx)
		if err != nil {
			this.logger.Panicf("conn mongo error: %s , uri: %s \n", err.Error(), mgo.Uri)
		}

		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			this.logger.Panicf("ping mongo error: %s , uri: %s \n", err.Error(), mgo.Uri)
		}

		this.setMgo(mgo.Db, client)
		this.logger.Infof("[mongodb] %s init success", mgo.Db)
	}
}

func (this *MgoClient) initMonitorMulLevelEvent(db_name string) MonitorEvent {
	eventNum := len(this.monitorEvents)
	var firstProxy MonitorEvent
	proxyInses := make([]MonitorEvent, eventNum)
	for k, proxyFunc := range this.monitorEvents {
		if _, ok := proxyFunc().(MonitorEvent); ok == false {
			this.logger.Panicf("[mongodb] not implement MonitorEvent interface")
		} else {
			proxyInses[k] = proxyFunc()
		}
	}

	for k, proxyIns := range proxyInses {
		//first proxy
		if k == 0 {
			firstProxy = proxyIns.(MonitorEvent)
		}

		if k+1 != eventNum {
			proxyIns.(MonitorEvent).NextEvent(proxyInses[k+1])
		}

		this.logger.Infof("[mongodb] %s init %s [%p]", db_name, proxyIns.(MonitorEvent).EventName(), proxyIns)
	}

	return firstProxy
}

func (this *MgoClient) setMgo(mgo_name string, gdb *mongo.Client) bool {
	mgo_name = strings.ToLower(mgo_name)
	this.Mgos[mgo_name] = gdb
	return true
}

func (this *MgoClient) GetColl(dataBase, coll string) *mongo.Collection {
	dataBase = strings.ToLower(dataBase)
	if mgo, ok := this.Mgos[dataBase]; ok {
		return mgo.Database(dataBase).Collection(coll)
	} else {
		this.logger.Errorf("[db] %s not found", dataBase)
		return nil
	}
}

func (this *MgoClient) poolEvent(pev *event.PoolEvent) {
	lab := prometheus.Labels{"type": pev.Type}
	mongodbPoolTypes.With(lab).Inc()
}

func (this *MgoClient) Ping() []error {
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)

	var errs []error
	var err error
	for _, db := range this.Mgos {
		err = db.Ping(ctx, readpref.Primary())
		if err != nil{
			errs = append(errs, err)
		}
	}

	return errs
}

func (this *MgoClient) Close()  {
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)

	for _, db := range this.Mgos {
		err = db.Disconnect(ctx)
		if err != nil {
			this.logger.Errorf(err.Error())
		}
	}
}

//mongodb 的上下文
func (this *MgoClient) GetCtx(ctx context.Context) context.Context {
	var command string
	return context.WithValue(ctx, "command", &command)
}
