package example

import (
	"sync"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
)

var (
	testPool = sync.Pool{
		New: func() interface{} {
			return &Test{}
		},
	}

	testsPool = sync.Pool{
		New: func() interface{} {
			return &Tests{}
		},
	}
)

type Test struct {
	g byte

	c int8

	i bool

	d int16

	f float32

	a int32

	m map[string]interface{}

	b int64

	e string

	logger log.Logger

	conf config.Config

	h []int

	u [3]string
}

type Tests []Test

type TestOption func(*Test)

type TestOptions struct{}

func (TestOptions) WithConf(conf config.Config) TestOption {
	return func(T *Test) {
		T.conf = conf
	}
}

func (TestOptions) WithLogger(logger log.Logger) TestOption {
	return func(T *Test) {
		T.logger = logger
	}
}

func (this *Test) Release() {
	this.b = 0
	this.c = 0
	this.i = false
	this.f = 0.00
	this.a = 0
	this.h = this.h[:0]
	for k, _ := range this.m {
		delete(this.m, k)
	}
	this.e = ""
	this.g = 0
	for k, _ := range this.u {
		this.u[k] = ""
	}
	this.d = 0
	this.logger = nil
	this.conf = nil
	testPool.Put(this)
}

func NewTests() *Tests {
	tests := testsPool.Get().(*Tests)
	return tests
}

func (this *Tests) Release() {
	*this = (*this)[:0]
	testsPool.Put(this)
}

type empty struct{}
