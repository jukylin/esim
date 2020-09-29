package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/jukylin/esim/config"
	"go.mongodb.org/mongo-driver/event"
)

func TestMonitorEvent_withSlowCommand(t *testing.T) {
	type args struct {
		ctx       context.Context
		backEvent *mongoBackEvent
		beginTime time.Time
		endTime   time.Time
	}

	var command string
	command = "test"
	ctx := context.Background()
	ctx = context.WithValue(ctx, keyCtx, &command)

	tests := []struct {
		name string
		args args
	}{
		{"慢查询-成功", args{ctx, &mongoBackEvent{
			&event.CommandSucceededEvent{
				event.CommandFinishedEvent{
					1000000000,
					"test",
					100,
					"",
				},
				nil,
			},
			nil,
		}, time.Now(), time.Now()}},
		{"慢查询-失败", args{ctx, &mongoBackEvent{
			failedEvent: &event.CommandFailedEvent{
				event.CommandFinishedEvent{
					1000000000,
					"test",
					100,
					"",
				},
				"",
			},
		}, time.Now(), time.Now()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := config.NewMemConfig()
			conf.Set("mgo_check_slow", true)
			conf.Set("mgo_slow_time", 100)
			monitorEventOptions := MonitorEventOptions{}
			m := NewMonitorEvent(
				monitorEventOptions.WithLogger(logger),
				monitorEventOptions.WithConf(conf),
			)
			m.(*MonitorEvent).withSlowCommand(tt.args.ctx, tt.args.backEvent, tt.args.beginTime, tt.args.endTime)
		})
	}
}
