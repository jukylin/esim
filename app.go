package esim

import (
	"os"
	"os/signal"
	"syscall"
	"gitlab.etcchebao.cn/go_service/esim/pkg/log"
	"github.com/jukylin/esim/transports"
)

type APP struct{

	logger log.Logger

	Trans []transports.Transports
}

func NewApp(logger log.Logger) *APP {

	app := &APP{
		logger: logger,
	}

	return app
}


func (this *APP) Start()  {
	for _, tran := range this.Trans {
		tran.Start()
	}
}

func (this *APP) AwaitSignal() {
	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	select {
	case s := <-c:
		this.logger.Infof("receive a signal %s", s.String())
		this.stop()
		os.Exit(0)
	}
}


func (this *APP) stop()  {
	for _, tran := range this.Trans {
		tran.GracefulShutDown()
	}
}