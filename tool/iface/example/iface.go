package example1

import (
	"context"
)

type Test interface {
	Iface1(func(string) string) bool

	Iface2(ctx context.Context, found *bool) (bool, error)

	Iface3() func(string) string
}
