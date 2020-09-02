package prometheus

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
)

func TestNewPrometheus(t *testing.T) {
	type args struct {
		httpAddr string
		logger   log.Logger
	}
	tests := []struct {
		name string
		args args
		want *Prometheus
	}{
		{"new prometheus with 9001", args{"9001", log.NewNullLogger()}, &Prometheus{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewPrometheus(tt.args.httpAddr, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewPrometheus() = %v, want %v", got, tt.want)
			}
			resp, err := http.Get("http://0.0.0.0:9001/metrics")
			assert.Nil(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

func TestNewNullProme(t *testing.T) {
	tests := []struct {
		name string
		want *Prometheus
	}{
		{"new", &Prometheus{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNullProme(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNullProme() = %v, want %v", got, tt.want)
			}
		})
	}
}
