package mongodb

import (
	"os"
	"testing"
	"context"
	"sync"

	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/config"
	"go.mongodb.org/mongo-driver/bson"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
)

type User struct{
	Id int
	Name string
}


var client *MgoClient
var logger log.Logger

func TestMain(m *testing.M) {
	logger = log.NewLogger()

	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	opt := &dockertest.RunOptions{
		Repository: "mongo",
		Tag: "latest",
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opt, func(hostConfig *dc.HostConfig) {
		hostConfig.PortBindings = map[dc.Port][]dc.PortBinding{
			"27017/tcp": {{HostIP: "", HostPort: "27017"}},
		}
	})

	if err := pool.Retry(func() error {
		mgoClientOptions := MgoClientOptions{}
		client = NewMongo(
			mgoClientOptions.WithDbConfig([]MgoConfig{
				{
					"test",
					"mongodb://0.0.0.0:27017/admin?connect=direct",
				},
			}))
		if len(client.Ping()) > 0{
			return client.Ping()[0]
		}else{
			return nil
		}
	}); err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	client.Close()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		logger.Fatalf("Could not purge resource: %s", err)
	}
	resource.Expire(60)

	os.Exit(code)
}

func TestGetColl(t *testing.T)  {
	mgoOnce = sync.Once{}

	conf := config.NewMemConfig()

	mgoClientOptions := MgoClientOptions{}
	mongoClient := NewMongo(
		mgoClientOptions.WithLogger(logger),
		mgoClientOptions.WithConf(conf),
		mgoClientOptions.WithDbConfig([]MgoConfig{
			{
				"test",
				"mongodb://127.0.0.1:27017",
			},
		}),
	)

	ctx := mongoClient.GetCtx(context.Background())
	coll := mongoClient.GetColl("test", "coll")
	mongoClient.GetCtx(ctx)

	filter := bson.M{"phone": "123456"}
	coll.Find(ctx, filter)
	_, ok := ctx.Value("command").(*string)
	if ok == false{
		t.Error("command not exists")
	}
	mongoClient.Close()
}


func TestWithMonitorEvent(t *testing.T)  {
	mgoOnce = sync.Once{}

	conf := config.NewMemConfig()
	conf.Set("debug", true)

	mgoClientOptions := MgoClientOptions{}
	mongoClient := NewMongo(
		mgoClientOptions.WithLogger(logger),
		mgoClientOptions.WithConf(conf),
		mgoClientOptions.WithMonitorEvent(
			func() MonitorEvent {
				monitorEventOptions := MonitorEventOptions{}
				return NewMonitorEvent(
					monitorEventOptions.WithConf(conf),
					monitorEventOptions.WithLogger(logger),
				)
			},
		),
		mgoClientOptions.WithDbConfig([]MgoConfig{
			{
				"test",
				"mongodb://127.0.0.1:27017",
			},
		}),
	)

	ctx := mongoClient.GetCtx(context.Background())
	coll := mongoClient.GetColl("test", "coll")
	mongoClient.GetCtx(ctx)

	u := User{}
	filter := bson.M{"phone": "123456"}
	coll.FindOne(ctx, filter).Decode(u)
	mongoClient.Close()
}


func TestMulEvent(t *testing.T)  {
	mgoOnce = sync.Once{}

	conf := config.NewMemConfig()
	conf.Set("debug", true)

	mgoClientOptions := MgoClientOptions{}
	mongoClient := NewMongo(
		mgoClientOptions.WithLogger(logger),
		mgoClientOptions.WithConf(conf),
		mgoClientOptions.WithMonitorEvent(
			func() MonitorEvent {
				monitorEventOptions := MonitorEventOptions{}
				return NewMonitorEvent(
					monitorEventOptions.WithConf(conf),
					monitorEventOptions.WithLogger(logger),
				)
			},
			func() MonitorEvent {
				return NewSpyEvent(logger)
			},
		),
		mgoClientOptions.WithDbConfig([]MgoConfig{
			{
				"test",
				"mongodb://127.0.0.1:27017",
			},
		}),
	)

	ctx := mongoClient.GetCtx(context.Background())
	coll := mongoClient.GetColl("test", "coll")
	mongoClient.GetCtx(ctx)

	u := User{}
	filter := bson.M{"phone": "123456"}
	coll.FindOne(ctx, filter).Decode(u)
	mongoClient.Close()
}