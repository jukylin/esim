package log

import (
	"context"
	"time"

	"go.uber.org/zap"
	glogger "gorm.io/gorm/logger"
)

type gormLogger struct {
	sugar *zap.SugaredLogger

	logLevel glogger.LogLevel

	ez *EsimZap
}

type GormLoggerOptions struct{}

type GormLoggerOption func(c *gormLogger)

func NewGormLogger(options ...GormLoggerOption) glogger.Interface {
	glog := &gormLogger{}

	for _, option := range options {
		option(glog)
	}

	glog.logLevel = glogger.Warn
	glog.sugar = glog.ez.Logger.Sugar()

	return glog
}

func WithGLogEsimZap(ez *EsimZap) GormLoggerOption {
	return func(gl *gormLogger) {
		gl.ez = ez
	}
}

func (gl *gormLogger) LogMode(logLevel glogger.LogLevel) glogger.Interface {
	gl.logLevel = logLevel
	return gl
}

func (gl *gormLogger) Info(ctx context.Context, template string, args ...interface{}) {
	gl.sugar.With(gl.ez.getGormArgs(ctx)...).Infof(template, args...)
}

func (gl *gormLogger) Warn(ctx context.Context, template string, args ...interface{}) {
	gl.sugar.With(gl.ez.getGormArgs(ctx)...).Warnf(template, args...)
}

func (gl *gormLogger) Error(ctx context.Context, template string, args ...interface{}) {
	gl.sugar.With(gl.ez.getGormArgs(ctx)...).Errorf(template, args...)
}

func (gl *gormLogger) Trace(ctx context.Context, begin time.Time,
	fc func() (string, int64), err error) {
	if gl.logLevel > 0 {
		elapsed := time.Since(begin)
		sql, rows := fc()
		switch {
		case err != nil && gl.logLevel >= glogger.Error:
			gl.sugar.With(gl.ez.getGormArgs(ctx)...).Errorf("%.2fms [%d] %s : %s",
				float64(elapsed.Nanoseconds())/1e6, rows, sql, err.Error())
		case gl.logLevel >= glogger.Info:
			gl.sugar.With(gl.ez.getGormArgs(ctx)...).Debugf("%.2fms [%d] %s",
				float64(elapsed.Nanoseconds())/1e6, rows, sql)
		}
	}
}
