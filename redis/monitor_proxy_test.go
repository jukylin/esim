package redis

import (
	"context"
	"reflect"
	"testing"

	"errors"
	"github.com/jukylin/esim/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"time"
)

var monitorProxy *MonitorProxy

func TestNewMonitorProxy(t *testing.T) {
	conf := config.NewMemConfig()
	conf.Set("redis_trace", true)
	conf.Set("redis_check_slow", true)
	conf.Set("redis_metrics", true)
	conf.Set("redis_slow_time", 100)

	monitorProxyOptions := MonitorProxyOptions{}
	monitorProxy = NewMonitorProxy(
		monitorProxyOptions.WithConf(conf),
	)
	monitorProxy.NextProxy(&DummyContextConn{})
	assert.IsType(t, &MonitorProxy{}, monitorProxy)
	assert.True(t, len(monitorProxy.afterEvents) > 0)

	assert.NotNil(t, monitorProxy.tracer)
}

func TestMonitorProxy_Do(t *testing.T) {
	type args struct {
		ctx         context.Context
		commandName string
		args        []interface{}
	}

	ctx := context.Background()
	tests := []struct {
		name      string
		args      args
		wantReply interface{}
		wantErr   bool
	}{
		{"one", args{ctx, "get", []interface{}{"one"}}, nil, false},
		{"two", args{ctx, "get", []interface{}{"tow"}}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReply, err := monitorProxy.Do(tt.args.ctx, tt.args.commandName, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("MonitorProxy.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(gotReply, tt.wantReply) {
				t.Errorf("MonitorProxy.Do() = %v, want %v", gotReply, tt.wantReply)
			}
			monitorProxy.Close()
		})
	}
}

func TestMonitorProxy_Send(t *testing.T) {
	type args struct {
		ctx         context.Context
		commandName string
		args        []interface{}
	}
	ctx := context.Background()
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"one", args{ctx, "get", []interface{}{"one"}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := monitorProxy.Send(tt.args.ctx, tt.args.commandName, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("MonitorProxy.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMonitorProxy_Flush(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	ctx := context.Background()
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"one", args{ctx}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := monitorProxy.Flush(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("MonitorProxy.Flush() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMonitorProxy_Receive(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	ctx := context.Background()
	tests := []struct {
		name      string
		args      args
		wantReply interface{}
		wantErr   bool
	}{
		{"one", args{ctx}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotReply, err := monitorProxy.Receive(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("MonitorProxy.Receive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotReply, tt.wantReply) {
				t.Errorf("MonitorProxy.Receive() = %v, want %v", gotReply, tt.wantReply)
			}
		})
	}
}

func TestMonitorProxy_redisTracer(t *testing.T) {
	type args struct {
		ctx  context.Context
		info *execInfo
	}

	ctx := context.Background()

	tests := []struct {
		name string
		args args
	}{
		{"span_record_error", args{ctx, &execInfo{
			err: errors.New("This is an error"), commandName: "",
			startTime: time.Now(), endTime: time.Now()}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitorProxy.redisTracer(tt.args.ctx, tt.args.info)
		})
	}
}

func TestMonitorProxy_redisMetrics(t *testing.T) {
	type args struct {
		ctx  context.Context
		info *execInfo
	}

	redisTotal.Reset()
	ctx := context.Background()

	tests := []struct {
		name string
		args args
	}{
		{"metrics", args{ctx,
			&execInfo{
				commandName: "get",
				startTime:   time.Now().Add(-2 * time.Second),
				endTime:     time.Now()}}}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitorProxy.redisMetrics(tt.args.ctx, tt.args.info)
			lab := prometheus.Labels{"cmd": "get"}
			c, _ := redisTotal.GetMetricWith(lab)
			metric := &io_prometheus_client.Metric{}
			err := c.Write(metric)
			assert.Nil(t, err)
			assert.True(t, metric.Counter.GetValue() == 1)
		})
	}
}

func TestMonitorProxy_redisSlowCommand(t *testing.T) {
	type args struct {
		ctx  context.Context
		info *execInfo
	}

	ctx := context.Background()
	tests := []struct {
		name string
		args args
	}{
		{"show_command", args{ctx,
			&execInfo{
				commandName: "get",
				startTime:   time.Now().Add(-2 * time.Second),
				endTime:     time.Now()}}},
		{"quick_command", args{ctx,
			&execInfo{
				commandName: "get",
				startTime:   time.Now().Add(-10 * time.Millisecond),
				endTime:     time.Now()}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitorProxy.redisSlowCommand(tt.args.ctx, tt.args.info)
		})
	}
}
