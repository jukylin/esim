package example1

import (
	"context"
	"github.com/jukylin/esim/pkg/file-dir"
)

type Test interface {
	Iface1(func(string) string) bool

	Iface2(ctx context.Context, found *bool) (bool, error)

	Iface3() func(string) string
}
