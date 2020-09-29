package proxy

import (
	"testing"

	"fmt"

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

type notProxy struct{}

func newNotProxy() *notProxy {
	return &notProxy{}
}

func (np *notProxy) Get(str string) string {
	return "1.0.0"
}

func TestProxyFactory(t *testing.T) {
	realDb := newRealDb()
	firstProxy := NewProxyFactory().GetFirstInstance("db_repo", realDb,
		func() interface{} {
			return newFilterProxy()
		},
		func() interface{} {
			return newCacheProxy()
		}).(DbRepo)
	assert.Equal(t, firstProxy.(Proxy).ProxyName(), "filter_proxy")

	result := firstProxy.Get("version")
	assert.Equal(t, result, "1.0.0")
	assert.Equal(t, fmt.Sprintf("%p", firstProxy.(*filterProxy).nextProxy.(*cacheProxy).nextProxy),
		fmt.Sprintf("%p", realDb))
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
	filterProxyObj := newFilterProxy()
	cacheProxyObj := newCacheProxy()
	firstProxy := NewProxyFactory().GetInstances("db_repo",
		func() interface{} {
			return filterProxyObj
		},
		func() interface{} {
			return cacheProxyObj
		})
	assert.IsType(t, &filterProxy{}, firstProxy[0])
	assert.IsType(t, &cacheProxy{}, firstProxy[1])

	assert.Equal(t, fmt.Sprintf("%p", filterProxyObj.nextProxy),
		fmt.Sprintf("%p", cacheProxyObj))
	assert.Nil(t, cacheProxyObj.nextProxy)
}

func TestProxyFactoryGetOneInstances(t *testing.T) {
	firstProxy := NewProxyFactory(WithLogger(log.NewNullLogger())).GetInstances("db_repo",
		func() interface{} {
			return newFilterProxy()
		})
	assert.IsType(t, &filterProxy{}, firstProxy[0])
	assert.Nil(t, firstProxy[0].(*filterProxy).nextProxy)
}

func TestProxyFactoryGetZoreInstances(t *testing.T) {
	firstProxy := NewProxyFactory().GetInstances("zore_proxy")
	assert.Nil(t, firstProxy)
}

func TestProxyFactory_GetInstancesNotImplementTheProxyInterface(t *testing.T) {
	assert.Panics(t, func() {
		NewProxyFactory().GetInstances("not_implement_proxy",
			func() interface{} {
				return newNotProxy()
			})
	})
}
