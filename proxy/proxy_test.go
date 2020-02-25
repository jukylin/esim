package proxy

import (
	"testing"
	"github.com/jukylin/esim/log"
	"github.com/stretchr/testify/assert"
)


type DbRepo interface {
	Get(str string) string
}

type filterProxy struct {
	nextProxy DbRepo

	logger log.Logger
}

func NewFilterProxy() *filterProxy {
	filterProxy := &filterProxy{}

	filterProxy.logger = log.NewLogger()

	return filterProxy
}

func (this *filterProxy) Get(str string) string {
	this.logger.Infof("filterProxy")
	result := this.nextProxy.Get(str)
	return result
}

func (this *filterProxy) NextProxy(proxy interface{}) {
	this.nextProxy = proxy.(DbRepo)
}

func (this *filterProxy) ProxyName() string {
	return "filter_proxy"
}

type cacheProxy struct {
	nextProxy DbRepo

	logger log.Logger
}

func NewCacheProxy() *cacheProxy {
	cacheProxy := &cacheProxy{}

	cacheProxy.logger = log.NewLogger()

	return cacheProxy
}

func (this *cacheProxy) Get(str string) string {
	this.logger.Infof("cacheProxy")
	result := this.nextProxy.Get(str)
	return result
}

func (this *cacheProxy) NextProxy(proxy interface{}) {
	this.nextProxy = proxy.(DbRepo)
}

func (this *cacheProxy) ProxyName() string {
	return "cache_proxy"
}

type realDb struct {
}

func NewRealDb() *realDb {
	return &realDb{}
}

func (this *realDb) Get(str string) string {
	return "1.0.0"
}

func TestProxyFactory(t *testing.T)  {
	firstProxy := NewProxyFactory().GetFirstInstance("db_repo", NewRealDb(),
		func() interface{} {
			return NewFilterProxy()
		},
		func() interface{} {
			return NewCacheProxy()
		}).(DbRepo)

	assert.Equal(t, firstProxy.(Proxy).ProxyName(), "filter_proxy")

	result := firstProxy.Get("version")
	assert.Equal(t, result, "1.0.0")
}


func TestProxyFactoryNotProxy(t *testing.T)  {
	firstProxy := NewProxyFactory().GetFirstInstance("db_repo", NewRealDb())
	assert.IsType(t, firstProxy, &realDb{})

	result := firstProxy.(DbRepo).Get("version")
	assert.Equal(t, result, "1.0.0")
}


func TestProxyFactoryNotRealInstanceNotProxy(t *testing.T)  {
	firstProxy := NewProxyFactory().GetFirstInstance("db_repo", nil)
	assert.Nil(t, firstProxy)
}


func TestProxyFactoryNotRealInstanceButProxy(t *testing.T)  {
	firstProxy := NewProxyFactory().GetFirstInstance("db_repo", nil,
		func() interface{} {
			return NewFilterProxy()
		},
		func() interface{} {
			return NewCacheProxy()
		})
	assert.IsType(t, &filterProxy{}, firstProxy)
}


func TestProxyFactoryGetInstances(t *testing.T)  {
	firstProxy := NewProxyFactory().GetInstances("db_repo",
		func() interface{} {
			return NewFilterProxy()
		},
		func() interface{} {
			return NewCacheProxy()
		})
	assert.IsType(t, &filterProxy{}, firstProxy[0])
	assert.IsType(t, &cacheProxy{}, firstProxy[1])
}