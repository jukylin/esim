package log

import (
	"context"
	"runtime"
	"strings"
	"time"
	"strconv"

	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type EsimZap struct {
	debug bool

	json bool

	*zap.Logger

	// 日志等级，参考zapcore.Level.
	logLevel zapcore.Level
}

type ZapOption func(c *EsimZap)

func NewEsimZap(options ...ZapOption) *EsimZap {
	ez := &EsimZap{}

	ez.logLevel = zap.WarnLevel
	for _, option := range options {
		option(ez)
	}

	var level zap.AtomicLevel
	if ez.debug {
		ez.logLevel = zap.DebugLevel
	}
	level = zap.NewAtomicLevelAt(ez.logLevel)

	encoderConfig := zap.NewProductionEncoderConfig()
	zapConfig := zap.Config{
		Level:       level,
		Development: ez.debug,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
		DisableCaller:    true,
	}
	zapConfig.EncoderConfig.EncodeTime = ez.standardTimeEncoder

	if !ez.json {
		zapConfig.Encoding = "console"
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig.Encoding = "json"
	}

	zapLogger, _ := zapConfig.Build()
	ez.Logger = zapLogger
	return ez
}

func WithEsimZapDebug(debug bool) ZapOption {
	return func(ez *EsimZap) {
		ez.debug = debug
	}
}

func WithEsimZapJSON(json bool) ZapOption {
	return func(ez *EsimZap) {
		ez.json = json
	}
}

func WithLogLevel(level zapcore.Level) ZapOption {
	return func(ez *EsimZap) {
		ez.logLevel = level
	}
}

func (ez *EsimZap) standardTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func (ez *EsimZap) getArgs(ctx context.Context) []interface{} {
	args := make([]interface{}, 0)

	args = append(args, "caller", ez.getCaller(runtime.Caller(2)))
	tracerID := tracerid.ExtractTracerID(ctx)
	if tracerID != "" {
		args = append(args, "tracer_id", tracerID)
	}

	return args
}

func (ez *EsimZap) getGormArgs(ctx context.Context) []interface{} {
	args := make([]interface{}, 0)
	var fullPath string
	var offidx int

	for i := 0; i < 15; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok && (strings.Index(file, "esim") == -1) {
			fullPath = file + ":" + strconv.FormatInt(int64(line), 10)
			break
		}
	}

	offidx = len(fullPath)
	idx := strings.LastIndexByte(fullPath, '/')
	if idx != -1 {
		offidx = idx
	}
	// Find the penultimate separator.
	idx = strings.LastIndexByte(fullPath[:idx], '/')
	if idx != -1 {
		offidx = idx
	}

	args = append(args, "caller", fullPath[offidx+1:])
	tracerID := tracerid.ExtractTracerID(ctx)
	if tracerID != "" {
		args = append(args, "tracer_id", tracerID)
	}

	return args
}

func (ez *EsimZap) getCaller(pc uintptr, file string, line int, ok bool) string {
	return zapcore.NewEntryCaller(pc, file, line, ok).TrimmedPath()
}
