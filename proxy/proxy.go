package proxy

import (
	"sync"

	"github.com/jukylin/esim/log"
)

type Proxy interface {
	//set next proxy
	NextProxy(proxy interface{})

	//get proxy name
	ProxyName() string
}

var factoryOnce sync.Once

var proxyFactory *Factory

type Factory struct {
	logger log.Logger
}

type FactoryOption func(c *Factory)

type FactoryOptions struct{}

func NewProxyFactory(options ...FactoryOption) *Factory {

	factoryOnce.Do(func() {

		proxyFactory = &Factory{}

		for _, option := range options {
			option(proxyFactory)
		}

		if proxyFactory.logger == nil {
			proxyFactory.logger = log.NewLogger()
		}
	})

	return proxyFactory
}

func (FactoryOptions) WithLogger(logger log.Logger) FactoryOption {
	return func(p *Factory) {
		p.logger = logger
	}
}

// GetFirstInstance implement init mul level proxy,
// RealInstance and proxys make sure both implement the same interface
// return firstProxy | realInstance
func (pf *Factory) GetFirstInstance(realName string, realInstance interface{}, proxys ...func() interface{}) interface{} {

	var firstProxy interface{}

	proxyInses := pf.GetInstances(realName, proxys...)

	proxyNum := len(proxyInses)
	if proxyNum > 0 {
		firstProxy = proxyInses[0]

		if realInstance != nil {
			proxyInses[len(proxyInses)-1].(Proxy).NextProxy(realInstance)
		}
	} else {
		firstProxy = realInstance
	}

	return firstProxy
}

func (pf *Factory) GetInstances(realName string, proxys ...func() interface{}) []interface{} {

	proxyNum := len(proxys)
	var proxyInses []interface{}
	if proxyNum > 0 {
		proxyInses = make([]interface{}, proxyNum)
		for k, proxyFunc := range proxys {
			if _, ok := proxyFunc().(Proxy); !ok {
				pf.logger.Panicf("[%s] not implement the Proxy interface", realName)
			} else {
				proxyInses[k] = proxyFunc()
			}
		}

		for k, proxyIns := range proxyInses {
			if proxyNum == 1 {
				proxyIns.(Proxy).NextProxy(proxyInses[k])
				pf.logger.Infof("[%s] %s init [%p]", realName, proxyIns.(Proxy).ProxyName(), proxyIns)
			} else if k+1 < proxyNum {
				proxyIns.(Proxy).NextProxy(proxyInses[k+1])
				pf.logger.Infof("[%s] %s init [%p]", realName, proxyIns.(Proxy).ProxyName(), proxyIns)
			} else {
				continue
			}
		}
	}

	return proxyInses
}
