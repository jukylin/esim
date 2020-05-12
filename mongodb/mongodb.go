package mongodb

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/prometheus/client_golang/prometheus"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var clientOnce sync.Once
var onceClient *Client

type Client struct {
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

type Option func(c *Client)

type ClientOptions struct{}

func NewClient(options ...Option) *Client {
	clientOnce.Do(func() {
		onceClient = &Client{
			Mgos: make(map[string]*mongo.Client),
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

func (ClientOptions) WithDbConfig(dbConfigs []MgoConfig) Option {
	return func(m *Client) {
		m.mgoConfig = dbConfigs
	}
}

func (ClientOptions) WithMonitorEvent(mongoEvent ...func() MonitorEvent) Option {
	return func(m *Client) {
		m.monitorEvents = mongoEvent
	}
}

type MgoConfig struct {
	Db  string `json:"db",yaml:"db"`
	URI string `json:"uri",yaml:"uri"`
}

func (c *Client) init() {

	mgoConfigs := make([]MgoConfig, 0)
	err := c.conf.UnmarshalKey("mgos", &mgoConfigs)
	if err != nil {
		c.logger.Panicf("Fatal error config file: %s \n", err.Error())
	}

	if len(c.mgoConfig) > 0 {
		mgoConfigs = append(mgoConfigs, c.mgoConfig...)
	}

	for _, mgo := range mgoConfigs {

		clientOptions := options.Client()
		clientOptions.ApplyURI(mgo.URI)

		if c.monitorEvents != nil {
			firstEvent := c.initMonitorMulLevelEvent(mgo.Db)
			//事件监控
			eventComMon := &event.CommandMonitor{
				Started: func(ctx context.Context, startEvent *event.CommandStartedEvent) {
					execCommand, ok := ctx.Value("command").(*string)
					if ok {
						*execCommand = startEvent.Command.String()
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
				c.poolEvent(pev)
			},
		}
		clientOptions.SetPoolMonitor(poolMon)

		mgoConnectTimeout := c.conf.GetInt64("mgo_connect_timeout")
		if mgoConnectTimeout != 0 {
			clientOptions.SetConnectTimeout(time.Duration(mgoConnectTimeout) * time.Millisecond)
			clientOptions.SetServerSelectionTimeout(time.Duration(mgoConnectTimeout) * time.Millisecond)
		}

		mgoMaxConnIdleTime := c.conf.GetInt64("mgo_max_conn_idle_time")
		if mgoMaxConnIdleTime != 0 {
			clientOptions.SetMaxConnIdleTime(time.Duration(mgoMaxConnIdleTime) * time.Minute)
		}

		mgoMaxPoolSize := c.conf.GetUint64("mgo_max_pool_size")
		if mgoMaxPoolSize != 0 {
			clientOptions.SetMaxPoolSize(mgoMaxPoolSize)
		}

		mgoMinPoolSize := c.conf.GetUint64("mgo_min_pool_size")
		if mgoMinPoolSize != 0 {
			clientOptions.SetMinPoolSize(mgoMinPoolSize)
		}

		client, err := mongo.NewClient(clientOptions)
		if err != nil {
			c.logger.Panicf("new mongo client error: %s , uri: %s \n", err.Error(), mgo.URI)
		}

		ctx, _ := context.WithTimeout(context.Background(), 2 * time.Second)

		err = client.Connect(ctx)
		if err != nil {
			c.logger.Panicf("conn mongo error: %s , uri: %s \n", err.Error(), mgo.URI)
		}

		err = client.Ping(ctx, readpref.Primary())
		if err != nil {
			c.logger.Panicf("ping mongo error: %s , uri: %s \n", err.Error(), mgo.URI)
		}

		c.setMgo(mgo.Db, client)
		c.logger.Infof("[mongodb] %s init success", mgo.Db)
	}
}

func (c *Client) initMonitorMulLevelEvent(dbName string) MonitorEvent {
	eventNum := len(c.monitorEvents)
	var firstProxy MonitorEvent
	proxyInses := make([]MonitorEvent, eventNum)
	for k, proxyFunc := range c.monitorEvents {
		if _, ok := proxyFunc().(MonitorEvent); !ok {
			c.logger.Panicf("[mongodb] not implement MonitorEvent interface")
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

		c.logger.Infof("[mongodb] %s init %s [%p]", dbName, proxyIns.(MonitorEvent).EventName(), proxyIns)
	}

	return firstProxy
}

func (c *Client) setMgo(mgoName string, gdb *mongo.Client) bool {
	mgoName = strings.ToLower(mgoName)
	c.Mgos[mgoName] = gdb
	return true
}

func (c *Client) GetColl(dataBase, coll string) *mongo.Collection {
	dataBase = strings.ToLower(dataBase)
	if mgo, ok := c.Mgos[dataBase]; ok {
		return mgo.Database(dataBase).Collection(coll)
	}

	c.logger.Errorf("[db] %s not found", dataBase)
	return nil
}

func (c *Client) poolEvent(pev *event.PoolEvent) {
	lab := prometheus.Labels{"type": pev.Type}
	mongodbPoolTypes.With(lab).Inc()
}

func (c *Client) Ping() []error {
	ctx, _ := context.WithTimeout(context.Background(), 2 * time.Second)

	var errs []error
	var err error
	for _, db := range c.Mgos {
		err = db.Ping(ctx, readpref.Primary())
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (c *Client) Close() {
	var err error
	ctx, _ := context.WithTimeout(context.Background(), 2 * time.Second)

	for _, db := range c.Mgos {
		err = db.Disconnect(ctx)
		if err != nil {
			c.logger.Errorf(err.Error())
		}
	}
}

//mongodb 的上下文
func (c *Client) GetCtx(ctx context.Context) context.Context {
	var command string
	return context.WithValue(ctx, "command", &command)
}
