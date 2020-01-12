package mongodb

import (
	"testing"
	"context"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/config"
	"go.mongodb.org/mongo-driver/bson"
)

type User struct{
	Id int
	Name string
}


func TestGetColl(t *testing.T)  {

	logger := log.NewLogger()
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
}


func TestWithMonitorEvent(t *testing.T)  {
	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
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
}


func TestMulEvent(t *testing.T)  {
	loggerOptions := log.LoggerOptions{}
	logger := log.NewLogger(loggerOptions.WithDebug(true))
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
}