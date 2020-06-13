package log

import (
	"context"
	"runtime"
	"time"

	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log Logger

type logger struct {
	debug bool

	json bool

	logger *zap.Logger

	sugar *zap.SugaredLogger
}

type LoggerOptions struct{}

type Option func(c *logger)

func NewLogger(options ...Option) Logger {
	logger := &logger{}

	for _, option := range options {
		option(logger)
	}

	var level zap.AtomicLevel
	if logger.debug {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	encoderConfig := zap.NewProductionEncoderConfig()

	zapConfig := zap.Config{
		Level:       level,
		Development: logger.debug,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		DisableCaller:    true,
	}
	zapConfig.EncoderConfig.EncodeTime = logger.standardTimeEncoder

	if !logger.json {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig.Encoding = "json"
	}
	zapLogger, _ := zapConfig.Build()
	logger.logger = zapLogger
	logger.sugar = zapLogger.Sugar()

	Log = logger

	return logger
}

// Deprecated: use WithDebug instead.
func (LoggerOptions) WithDebug(debug bool) Option {
	return func(l *logger) {
		l.debug = debug
	}
}

// Deprecated: use WithJSON instead.
func (LoggerOptions) WithJSON(json bool) Option {
	return func(l *logger) {
		l.json = json
	}
}

func WithDebug(debug bool) Option {
	return func(l *logger) {
		l.debug = debug
	}
}

func WithJSON(json bool) Option {
	return func(l *logger) {
		l.json = json
	}
}

func (log *logger) Error(msg string) {
	log.logger.Error(msg)
}

func (log *logger) Debugf(template string, args ...interface{}) {
	log.sugar.With(log.getArgs(context.TODO())...).Debugf(template, args...)
}

func (log *logger) Infof(template string, args ...interface{}) {
	log.sugar.With(log.getArgs(context.TODO())...).Infof(template, args...)
}

func (log *logger) Warnf(template string, args ...interface{}) {
	log.sugar.With(log.getArgs(context.TODO())...).Warnf(template, args...)
}

func (log *logger) Errorf(template string, args ...interface{}) {
	log.sugar.With(log.getArgs(context.TODO())...).Errorf(template, args...)
}

func (log *logger) DPanicf(template string, args ...interface{}) {
	log.sugar.With(log.getArgs(context.TODO())...).DPanicf(template, args...)
}

func (log *logger) Panicf(template string, args ...interface{}) {
	log.sugar.With(log.getArgs(context.TODO())...).Panicf(template, args...)
}

func (log *logger) Fatalf(template string, args ...interface{}) {
	log.sugar.With(log.getArgs(context.TODO())...).Fatalf(template, args...)
}

func (log *logger) Debugc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.getArgs(ctx)...).Debugf(template, args...)
}

func (log *logger) Infoc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.getArgs(ctx)...).Infof(template, args...)
}

func (log *logger) Warnc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.getArgs(ctx)...).Warnf(template, args...)
}

func (log *logger) Errorc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.getArgs(ctx)...).Errorf(template, args...)
}

func (log *logger) DPanicc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.getArgs(ctx)...).DPanicf(template, args...)
}

func (log *logger) Panicc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.getArgs(ctx)...).Panicf(template, args...)
}

func (log *logger) Fatalc(ctx context.Context, template string, args ...interface{}) {
	log.sugar.With(log.getArgs(ctx)...).Fatalf(template, args...)
}

func (log *logger) standardTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func (log *logger) getArgs(ctx context.Context) []interface{} {
	args := make([]interface{}, 0)

	args = append(args, "caller", log.getCaller(runtime.Caller(2)))
	tracerID := log.getTracerID(ctx)
	if tracerID != "" {
		args = append(args, "tracer_id", tracerID)
	}

	return args
}

func (log *logger) getCaller(pc uintptr, file string, line int, ok bool) string {
	return zapcore.NewEntryCaller(pc, file, line, ok).TrimmedPath()
}

// getTracerID get tracer_id from context.
func (log *logger) getTracerID(ctx context.Context) string {
	sp := opentracing.SpanFromContext(ctx)
	if sp != nil {
		if jaegerSpanContext, ok := sp.Context().(jaeger.SpanContext); ok {
			return jaegerSpanContext.TraceID().String()
		}
	}

	val := ctx.Value(tracerid.ActiveEsimKey)
	if tracerID, ok := val.(string); ok {
		return tracerID
	}

	return ""
}
