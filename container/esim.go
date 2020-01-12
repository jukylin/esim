package container

import (
	"sync"
	"github.com/google/wire"
	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/prome"
)

var esimOnce sync.Once
var onceEsim *Esim

//esim init start
type Esim struct {
	prome *prome.Prome

	Log log.Logger

	Conf config.Config
}

var esimSet = wire.NewSet(
	wire.Struct(new(Esim), "*"),
	provideConf,
	provideLogger,
	provideProme,
)

var confFunc func() config.Config

func SetConfFunc(conf func() config.Config) {
	confFunc = conf
}
func provideConf() config.Config {
	return confFunc()
}

var promeFunc = func(conf config.Config, log log.Logger) *prome.Prome {
	return prome.NewProme(conf, log)
}

func SetPromeFunc(prome func(conf config.Config, log log.Logger) *prome.Prome) {
	promeFunc = prome
}
func provideProme(conf config.Config, log log.Logger) *prome.Prome {
	return promeFunc(conf, log)
}

var loggerFunc = func(conf config.Config) log.Logger {
	var loggerOptions log.LoggerOptions

	logger := log.NewLogger(
		loggerOptions.WithConf(conf),
		loggerOptions.WithDebug(conf.GetBool("debug")),
	)
	return logger
}

func SetLogger(log func(config.Config) log.Logger) {
	loggerFunc = log
}
func provideLogger(conf config.Config) log.Logger {
	return loggerFunc(conf)
}

//esim init end

//使用单例模式，避免多次初始化
func NewEsim() *Esim {
	esimOnce.Do(func() {
		onceEsim = initEsim()
	})

	return onceEsim
}

func (this *Esim) String() string {
	return "相信，相信自己！！！"
}
