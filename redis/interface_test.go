package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_execInfo_Release(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"1"},
		{"2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ei := newExecInfo()
			assert.Len(t, (ei.args), 0)
			ei.Release()
		})
	}
}
