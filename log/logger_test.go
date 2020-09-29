package log

import (
	"context"
	"testing"

	tracerid "github.com/jukylin/esim/pkg/tracer-id"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {

	ez := NewEsimZap(
		WithEsimZapDebug(true),
		WithEsimZapJSON(true),
	)

	logger := NewLogger(
		WithEsimZap(ez))

	ctx := context.WithValue(context.Background(), tracerid.ActiveEsimKey, tracerid.TracerID()())

	logger.Debugf("debug")
	logger.Debugc(ctx, "debug")

	logger.Infof("info")
	logger.Infoc(ctx, "info")

	logger.Warnf("warn")
	logger.Warnc(ctx, "warn")

	logger.Error("Error")
	logger.Errorf("Errorf")
	logger.Errorc(ctx, "Errorf")

	assert.Panics(t, func() {
		logger.Panicf("Panicf")
		logger.DPanicf("DPanicf")
	})

	assert.Panics(t, func() {
		logger.Panicc(ctx, "Panicc")
		logger.DPanicc(ctx, "DPanicf")
	})

	//logger.Fatalf("Fatalf")
	//logger.Fatalc(ctx, "Fatalc")

}

//
//func TestNewLogger(t *testing.T) {
//	type args struct {
//		options []Option
//	}
//
//	loggerOptions := LoggerOptions{}
//	tests := []struct {
//		name string
//		args args
//	}{
//		{"debug with json", args{[]Option{WithDebug(true), WithJSON(true)}}},
//		{"debug not with json", args{[]Option{WithDebug(true), WithJSON(true)}}},
//		{"json", args{[]Option{WithJSON(true)}}},
//		{"no options", args{}},
//		{"object-oriented options", args{[]Option{loggerOptions.WithJSON(true),
//			loggerOptions.WithDebug(true)}}},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			NewLogger(tt.args.options...)
//		})
//	}
//}
