package mongodb

import (
	"context"
	"os"
	"testing"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/ory/dockertest/v3"
	dc "github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type User struct {
	ID   int
	Name string
}

var client *Client
var logger log.Logger

func TestMain(m *testing.M) {
	logger = log.NewLogger()

	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	opt := &dockertest.RunOptions{
		Repository: "mongo",
		Tag:        "latest",
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opt, func(hostConfig *dc.HostConfig) {
		hostConfig.PortBindings = map[dc.Port][]dc.PortBinding{
			"27017/tcp": {{HostIP: "", HostPort: "27017"}},
		}
	})

	if err != nil {
		logger.Fatalf("Could not start resource: %s", err)
	}

	resource.Expire(10)

	if err := pool.Retry(func() error {
		mgoClientOptions := ClientOptions{}
		client = NewClient(
			mgoClientOptions.WithDbConfig([]MgoConfig{
				{
					"test",
					"mongodb://0.0.0.0:27017/admin?connect=direct",
				},
			}))
		if len(client.Ping()) > 0 {
			return client.Ping()[0]
		}

		return nil
	}); err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	code := m.Run()

	client.Close()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		logger.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestGetColl(t *testing.T) {

	conf := config.NewMemConfig()

	mgoClientOptions := ClientOptions{}
	mongoClient := NewClient(
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
	assert.True(t, ok)

	mongoClient.Close()
}

func TestWithMonitorEvent(t *testing.T) {

	conf := config.NewMemConfig()
	conf.Set("debug", true)

	mgoClientOptions := ClientOptions{}
	mongoClient := NewClient(
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

func TestMulEvent(t *testing.T) {

	conf := config.NewMemConfig()
	conf.Set("debug", true)

	mgoClientOptions := ClientOptions{}
	mongoClient := NewClient(
		mgoClientOptions.WithLogger(logger),
		mgoClientOptions.WithConf(conf),
		mgoClientOptions.WithMonitorEvent(
			func() MgoEvent {
				monitorEventOptions := MonitorEventOptions{}
				return NewMonitorEvent(
					monitorEventOptions.WithConf(conf),
					monitorEventOptions.WithLogger(logger),
				)
			},
			func() MonitorEvent {
				return newSpyEvent(logger)
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
