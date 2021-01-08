package log

import (
	"context"
	"testing"
	"time"

	glogger "gorm.io/gorm/logger"
	"errors"
	"gorm.io/gorm"
)

func Test_gormLogger_Trace(t *testing.T) {
	type fields struct {
		logLevel glogger.LogLevel
		ez       *EsimZap
	}
	type args struct {
		ctx   context.Context
		begin time.Time
		fc    func() (string, int64)
		err   error
	}

	z := NewEsimZap(
		WithEsimZapJSON(true),
		WithEsimZapDebug(true),
	)
	ctx := context.Background()

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"Error",
		fields{ez:z, logLevel : glogger.Error},
		args{ctx, time.Now(), func() (string, int64) {
			return "select * from Error;", 10
		}, errors.New("error")}},
		{"Error ErrRecordNotFound",
			fields{ez:z, logLevel : glogger.Error},
			args{ctx, time.Now(), func() (string, int64) {
				return "select * from ErrRecordNotFound;", 0
			}, gorm.ErrRecordNotFound}},
		{"Warn",
			fields{ez:z, logLevel : glogger.Warn},
			args{ctx, time.Now(), func() (string, int64) {
				return "select * from Warn;", 0
			}, nil}},

		{"Info",
			fields{ez:z, logLevel : glogger.Info},
			args{ctx, time.Now(), func() (string, int64) {
				return "select * from Info;", 0
			}, nil}},
		{"fc 为 nil",
			fields{ez:z, logLevel : glogger.Info},
			args{ctx, time.Now(), nil, nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gl := NewGormLogger(
				WithGLogEsimZap(tt.fields.ez),
			)
			gl.LogMode(tt.fields.logLevel)
			gl.Trace(tt.args.ctx, tt.args.begin, tt.args.fc, tt.args.err)
		})
	}
}
