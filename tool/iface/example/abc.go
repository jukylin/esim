package example1

import (
	"context"
)

//@ Interface Test
type testStub struct{}

type testStubOption func(testStub)

type testStubOptions struct{}

func NewTestStub(options ...testStubOption) testStub {

	t := testStub{}

	for _, option := range options {
		option(t)
	}

	return t
}

func (this testStub) Iface1(arg0 func(string) string) bool {

}

func (this testStub) Iface2(ctx context.Context, found *bool) (bool, error) {

}

func (this testStub) Iface3() func(string) string {

}
