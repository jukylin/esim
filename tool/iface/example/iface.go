package example1

import (
	"context"
)

type TestStruct struct{

}

type Test interface {
	Iface1(func(string) string) (result bool)

	Iface2(ctx context.Context, found *bool) (result bool, err error)

	Iface3() (f func(string) string)

	Iface4() map[string]string
}
