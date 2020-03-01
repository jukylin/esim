package example

import (
	"context"
	"gitlab.etcchebao.cn/go_service/esim/db2entity"
)

type Test interface {
	Iface1( func(string) string) bool

	Iface2(ctx context.Context, found *bool) (bool, error)

	Iface3(user db2entity.Field) func(string) string
}
