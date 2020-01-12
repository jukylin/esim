package log

import (
	"context"
	"github.com/jukylin/esim/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"runtime"
	"time"
	"github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
)

var l Logger

type logger struct {
	debug bool

	logger *zap.Logger

	conf config.Config
}

type LoggerOptions struct{}

type Option func(c *logger)

func NewLogger(options ...Option) Logger {

	log := &logger{}

	for _, option := range options {
		option(log)
	}

	if log.conf == nil {
		log.conf = config.NewNullConfig()
	}

	var level zap.AtomicLevel
	if log.debug == true {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	encoderConfig := zap.NewProductionEncoderConfig()

	zapConfig := zap.Config{
		Level:       level,
		Development: log.debug,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		DisableCaller:    true,
	}
	zapConfig.EncoderConfig.EncodeTime = log.standardTimeEncoder

	if log.debug == true {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig.Encoding = "json"
		zapConfig.InitialFields = map[string]interface{}{"service_name": log.conf.GetString("appname")}
	}
	logger, _ := zapConfig.Build()
	log.logger = logger

	l = log

	return log
}

func (LoggerOptions) WithConf(conf config.Config) Option {
	return func(l *logger) {
		l.conf = conf
	}
}

func (LoggerOptions) WithDebug(debug bool) Option {
	return func(l *logger) {
		l.debug = debug
	}
}

func (log *logger) getLogger() *zap.Logger {
	return log.logger
}

func (log *logger) getSugarLogger() *zap.SugaredLogger {
	sugar := log.logger.Sugar()
	return sugar
}

func Error(msg string) { l.Error(msg) }
func (log *logger) Error(msg string) {
	log.getLogger().Error(msg)
}

func Debugf(template string, args ...interface{}) { l.Debugf(template, args...) }
func (log *logger) Debugf(template string, args ...interface{}) {
	log.getSugarLogger().With("caller", log.getCaller(runtime.Caller(1))).Debugf(template, args...)
}

func Infof(template string, args ...interface{}) { l.Infof(template, args...) }
func (log *logger) Infof(template string, args ...interface{}) {
	log.getSugarLogger().With("caller", log.getCaller(runtime.Caller(1))).Infof(template, args...)
}

func Warnf(template string, args ...interface{}) { l.Warnf(template, args...) }
func (log *logger) Warnf(template string, args ...interface{}) {
	log.getSugarLogger().With("caller", log.getCaller(runtime.Caller(1))).Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) { l.Errorf(template, args...) }
func (log *logger) Errorf(template string, args ...interface{}) {
	log.getSugarLogger().With("caller", log.getCaller(runtime.Caller(1))).Errorf(template, args...)
}

func DPanicf(template string, args ...interface{}) { l.DPanicf(template, args...) }
func (log *logger) DPanicf(template string, args ...interface{}) {
	log.getSugarLogger().With("caller", log.getCaller(runtime.Caller(1))).DPanicf(template, args...)
}

func Panicf(template string, args ...interface{}) { l.Panicf(template, args...) }
func (log *logger) Panicf(template string, args ...interface{}) {
	log.getSugarLogger().With("caller", log.getCaller(runtime.Caller(1))).Panicf(template, args...)
}

func Fatalf(template string, args ...interface{}) { l.Fatalf(template, args...) }
func (log *logger) Fatalf(template string, args ...interface{}) {
	log.getSugarLogger().With("caller", log.getCaller(runtime.Caller(1))).
		Fatalf(template, args...)
}

func Debugc(ctx context.Context, template string, args ...interface{}) {
	l.Debugc(ctx, template, args...)
}
func (log *logger) Debugc(ctx context.Context, template string, args ...interface{}) {
	sugar := log.getSugarLogger()
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		sugar.With("tracer_id", tracerId)
	}
	sugar.With("caller", log.getCaller(runtime.Caller(1))).Debugf(template, args...)
}

func Infoc(ctx context.Context, template string, args ...interface{}) { l.Infoc(ctx, template, args...) }
func (log *logger) Infoc(ctx context.Context, template string, args ...interface{}) {
	sugar := log.getSugarLogger()
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		sugar.With("tracer_id", tracerId)
	}
	sugar.With("caller", log.getCaller(runtime.Caller(1))).Infof(template, args...)
}

func Warnc(ctx context.Context, template string, args ...interface{}) { l.Warnc(ctx, template, args...) }
func (log *logger) Warnc(ctx context.Context, template string, args ...interface{}) {
	sugar := log.getSugarLogger()
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		sugar.With("tracer_id", tracerId)
	}
	sugar.With("caller", log.getCaller(runtime.Caller(1))).Warnf(template, args...)
}

func Errorc(ctx context.Context, template string, args ...interface{}) {
	l.Errorc(ctx, template, args...)
}
func (log *logger) Errorc(ctx context.Context, template string, args ...interface{}) {
	sugar := log.getSugarLogger()
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		sugar.With("tracer_id", tracerId)
	}
	sugar.With("caller", log.getCaller(runtime.Caller(1))).Errorf(template, args...)
}

func DPanicc(ctx context.Context, template string, args ...interface{}) {
	l.DPanicc(ctx, template, args...)
}
func (log *logger) DPanicc(ctx context.Context, template string, args ...interface{}) {
	sugar := log.getSugarLogger()
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		sugar.With("tracer_id", tracerId)
	}
	sugar.With("caller", log.getCaller(runtime.Caller(1))).DPanicf(template, args...)
}

func Panicc(ctx context.Context, template string, args ...interface{}) {
	l.Panicc(ctx, template, args...)
}
func (log *logger) Panicc(ctx context.Context, template string, args ...interface{}) {
	sugar := log.getSugarLogger()
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		sugar.With("tracer_id", tracerId)
	}
	sugar.With("caller", log.getCaller(runtime.Caller(1))).Panicf(template, args...)
}

func Fatalc(ctx context.Context, template string, args ...interface{}) {
	l.Fatalc(ctx, template, args...)
}
func (log *logger) Fatalc(ctx context.Context, template string, args ...interface{}) {
	sugar := log.getSugarLogger()
	if tracerId := log.getTracerId(ctx); tracerId != "" {
		sugar.With("tracer_id", tracerId)
	}
	sugar.With("caller", log.getCaller(runtime.Caller(1))).Fatalf(template, args...)
}

func (log *logger) getCaller(pc uintptr, file string, line int, ok bool) string {
	return zapcore.NewEntryCaller(pc, file, line, ok).TrimmedPath()
}

func (log *logger) standardTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

//get opentracing global tracer_id
func (log *logger) getTracerId(ctx context.Context) string {
	sp := opentracing.SpanFromContext(ctx)
	if sp != nil {
		if jaegerSpanContext, ok := sp.Context().(jaeger.SpanContext); ok{
			return jaegerSpanContext.TraceID().String()
		}else{
			return ""
		}
	} else {
		return ""
	}
}
