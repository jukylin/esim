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

var proxyFactoryOnce sync.Once

var proxyFactory *ProxyFactory

type ProxyFactory struct{
	logger log.Logger
}

type ProxyFactoryOption func(c *ProxyFactory)

type ProxyFactoryOptions struct{}


func NewProxyFactory(options ...ProxyFactoryOption) *ProxyFactory {

	proxyFactoryOnce.Do(func() {

		proxyFactory = &ProxyFactory{}

		for _, option := range options {
			option(proxyFactory)
		}

		if proxyFactory.logger == nil {
			proxyFactory.logger = log.NewLogger()
		}
	})

	return proxyFactory
}

func (ProxyFactoryOptions) WithLogger(log log.Logger) ProxyFactoryOption {
	return func(p *ProxyFactory) {
		p.logger = log
	}
}

// GetFirstInstance implement init mul level proxy,
// RealInstance and proxys make sure both implement the same interface
// return firstProxy | realInstance
func (this *ProxyFactory) GetFirstInstance(realName string, realInstance interface{}, proxys ...func() interface{}) interface{} {

	var firstProxy interface{}
	var proxyInses []interface{}

	proxyInses = this.GetInstances(realName, proxys...)

	proxyNum := len(proxyInses)
	if proxyNum > 0 {
		firstProxy = proxyInses[0]

		if realInstance != nil {
			proxyInses[len(proxyInses)-1].(Proxy).NextProxy(realInstance)
		}
	}else{
		firstProxy = realInstance
	}

	return firstProxy
}


func (this *ProxyFactory) GetInstances(realName string, proxys ...func() interface{}) []interface{} {

	proxyNum := len(proxys)
	var proxyInses []interface{}
	if proxyNum > 0 {
		proxyInses = make([]interface{}, proxyNum)
		for k, proxyFunc := range proxys {
			if _, ok := proxyFunc().(Proxy); ok == false {
				this.logger.Panicf("[%s] not implement the Proxy interface", realName)
			} else {
				proxyInses[k] = proxyFunc()
			}
		}

		for k, proxyIns := range proxyInses {
			if proxyNum == 1 {
				proxyIns.(Proxy).NextProxy(proxyInses[k])
				this.logger.Infof("[%s] %s init [%p]", realName, proxyIns.(Proxy).ProxyName(), proxyIns)
			} else if k+1 < proxyNum{
				proxyIns.(Proxy).NextProxy(proxyInses[k+1])
				this.logger.Infof("[%s] %s init [%p]", realName, proxyIns.(Proxy).ProxyName(), proxyIns)
			}else{
				continue
			}
		}
	}

	return proxyInses
}