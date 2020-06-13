package redis

import (
	"context"
	"reflect"
	"testing"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	opentracing2 "github.com/opentracing/opentracing-go"
	"github.com/stretchr/testify/assert"
)

var monitorProxy *MonitorProxy

func TestNewMonitorProxy(t *testing.T) {
	conf := config.NewMemConfig()
	conf.Set("redis_trace", true)
	conf.Set("redis_check_slow", true)
	conf.Set("redis_metrics", true)

	monitorProxyOptions := MonitorProxyOptions{}
	monitorProxy = NewMonitorProxy(
		monitorProxyOptions.WithConf(conf),
	)
	assert.IsType(t, &MonitorProxy{}, monitorProxy)
	assert.True(t, len(monitorProxy.afterEvents) > 0)
}

func TestMonitorProxy_Do(t *testing.T) {
	type fields struct {
		name        string
		nextConn    ContextConn
		tracer      opentracing2.Tracer
		conf        config.Config
		logger      log.Logger
		afterEvents []afterEvents
	}
	type args struct {
		ctx         context.Context
		commandName string
		args        []interface{}
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantReply interface{}
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MonitorProxy{
				name:        tt.fields.name,
				nextConn:    tt.fields.nextConn,
				tracer:      tt.fields.tracer,
				conf:        tt.fields.conf,
				logger:      tt.fields.logger,
				afterEvents: tt.fields.afterEvents,
			}
			gotReply, err := mp.Do(tt.args.ctx, tt.args.commandName, tt.args.args...)
			if (err != nil) != tt.wantErr {
				t.Errorf("MonitorProxy.Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotReply, tt.wantReply) {
				t.Errorf("MonitorProxy.Do() = %v, want %v", gotReply, tt.wantReply)
			}
		})
	}
}

func TestMonitorProxy_Send(t *testing.T) {
	type fields struct {
		name        string
		nextConn    ContextConn
		tracer      opentracing2.Tracer
		conf        config.Config
		logger      log.Logger
		afterEvents []afterEvents
	}
	type args struct {
		ctx         context.Context
		commandName string
		args        []interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MonitorProxy{
				name:        tt.fields.name,
				nextConn:    tt.fields.nextConn,
				tracer:      tt.fields.tracer,
				conf:        tt.fields.conf,
				logger:      tt.fields.logger,
				afterEvents: tt.fields.afterEvents,
			}
			if err := mp.Send(tt.args.ctx, tt.args.commandName, tt.args.args...); (err != nil) != tt.wantErr {
				t.Errorf("MonitorProxy.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMonitorProxy_Flush(t *testing.T) {
	type fields struct {
		name        string
		nextConn    ContextConn
		tracer      opentracing2.Tracer
		conf        config.Config
		logger      log.Logger
		afterEvents []afterEvents
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MonitorProxy{
				name:        tt.fields.name,
				nextConn:    tt.fields.nextConn,
				tracer:      tt.fields.tracer,
				conf:        tt.fields.conf,
				logger:      tt.fields.logger,
				afterEvents: tt.fields.afterEvents,
			}
			if err := mp.Flush(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("MonitorProxy.Flush() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMonitorProxy_Receive(t *testing.T) {
	type fields struct {
		name        string
		nextConn    ContextConn
		tracer      opentracing2.Tracer
		conf        config.Config
		logger      log.Logger
		afterEvents []afterEvents
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		wantReply interface{}
		wantErr   bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MonitorProxy{
				name:        tt.fields.name,
				nextConn:    tt.fields.nextConn,
				tracer:      tt.fields.tracer,
				conf:        tt.fields.conf,
				logger:      tt.fields.logger,
				afterEvents: tt.fields.afterEvents,
			}
			gotReply, err := mp.Receive(tt.args.ctx)
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
	type fields struct {
		name        string
		nextConn    ContextConn
		tracer      opentracing2.Tracer
		conf        config.Config
		logger      log.Logger
		afterEvents []afterEvents
	}
	type args struct {
		ctx  context.Context
		info *execInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MonitorProxy{
				name:        tt.fields.name,
				nextConn:    tt.fields.nextConn,
				tracer:      tt.fields.tracer,
				conf:        tt.fields.conf,
				logger:      tt.fields.logger,
				afterEvents: tt.fields.afterEvents,
			}
			mp.redisTracer(tt.args.ctx, tt.args.info)
		})
	}
}

func TestMonitorProxy_redisSlowCommand(t *testing.T) {
	type fields struct {
		name        string
		nextConn    ContextConn
		tracer      opentracing2.Tracer
		conf        config.Config
		logger      log.Logger
		afterEvents []afterEvents
	}
	type args struct {
		ctx  context.Context
		info *execInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MonitorProxy{
				name:        tt.fields.name,
				nextConn:    tt.fields.nextConn,
				tracer:      tt.fields.tracer,
				conf:        tt.fields.conf,
				logger:      tt.fields.logger,
				afterEvents: tt.fields.afterEvents,
			}
			mp.redisSlowCommand(tt.args.ctx, tt.args.info)
		})
	}
}

func TestMonitorProxy_redisMetrics(t *testing.T) {
	type fields struct {
		name        string
		nextConn    ContextConn
		tracer      opentracing2.Tracer
		conf        config.Config
		logger      log.Logger
		afterEvents []afterEvents
	}
	type args struct {
		ctx  context.Context
		info *execInfo
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mp := &MonitorProxy{
				name:        tt.fields.name,
				nextConn:    tt.fields.nextConn,
				tracer:      tt.fields.tracer,
				conf:        tt.fields.conf,
				logger:      tt.fields.logger,
				afterEvents: tt.fields.afterEvents,
			}
			mp.redisMetrics(tt.args.ctx, tt.args.info)
		})
	}
}
