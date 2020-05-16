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

func newFilterProxy() *filterProxy {
	filterProxy := &filterProxy{}

	filterProxy.logger = log.NewLogger()

	return filterProxy
}

func (fp *filterProxy) Get(str string) string {
	fp.logger.Infof("filterProxy")
	result := fp.nextProxy.Get(str)
	return result
}

func (fp *filterProxy) NextProxy(proxy interface{}) {
	fp.nextProxy = proxy.(DbRepo)
}

func (fp *filterProxy) ProxyName() string {
	return "filter_proxy"
}

type cacheProxy struct {
	nextProxy DbRepo

	logger log.Logger
}

func newCacheProxy() *cacheProxy {
	cacheProxy := &cacheProxy{}

	cacheProxy.logger = log.NewLogger()

	return cacheProxy
}

func (fp *cacheProxy) Get(str string) string {
	fp.logger.Infof("cacheProxy")
	result := fp.nextProxy.Get(str)
	return result
}

func (fp *cacheProxy) NextProxy(proxy interface{}) {
	fp.nextProxy = proxy.(DbRepo)
}

func (fp *cacheProxy) ProxyName() string {
	return "cache_proxy"
}

type realDb struct{}

func newRealDb() *realDb {
	return &realDb{}
}

func (fp *realDb) Get(str string) string {
	return "1.0.0"
}

func TestProxyFactory(t *testing.T) {
	firstProxy := NewProxyFactory().GetFirstInstance("db_repo", newRealDb(),
		func() interface{} {
			return newFilterProxy()
		},
		func() interface{} {
			return newCacheProxy()
		}).(DbRepo)

	assert.Equal(t, firstProxy.(Proxy).ProxyName(), "filter_proxy")

	result := firstProxy.Get("version")
	assert.Equal(t, result, "1.0.0")
}

func TestProxyFactoryNotProxy(t *testing.T) {
	firstProxy := NewProxyFactory().GetFirstInstance("db_repo", newRealDb())
	assert.IsType(t, firstProxy, &realDb{})

	result := firstProxy.(DbRepo).Get("version")
	assert.Equal(t, result, "1.0.0")
}

func TestProxyFactoryNotRealInstanceNotProxy(t *testing.T) {
	firstProxy := NewProxyFactory().GetFirstInstance("db_repo", nil)
	assert.Nil(t, firstProxy)
}

func TestProxyFactoryNotRealInstanceButProxy(t *testing.T) {
	firstProxy := NewProxyFactory().GetFirstInstance("db_repo", nil,
		func() interface{} {
			return newFilterProxy()
		},
		func() interface{} {
			return newCacheProxy()
		})
	assert.IsType(t, &filterProxy{}, firstProxy)
}

func TestProxyFactoryGetInstances(t *testing.T) {
	firstProxy := NewProxyFactory().GetInstances("db_repo",
		func() interface{} {
			return newFilterProxy()
		},
		func() interface{} {
			return newCacheProxy()
		})
	assert.IsType(t, &filterProxy{}, firstProxy[0])
	assert.IsType(t, &cacheProxy{}, firstProxy[1])
}
