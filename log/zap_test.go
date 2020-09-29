package log

import (
	"context"
	"reflect"
	"testing"
)

func Test_esimZap_getGormArgs(t *testing.T) {
	type args struct {
		ctx context.Context
	}

	ctx := context.Background()

	tests := []struct {
		name string
		args args
		want []interface{}
	}{
		{"解析路径", args{ctx}, []interface{}{"caller", "log/zap_test.go:26"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ez := NewEsimZap()
			if got := ez.getGormArgs(tt.args.ctx); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("esimZap.getGormArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}
