package esim

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"google.golang.org/grpc"
	"gitlab.etcchebao.cn/go_service/esim/pkg/log"
	"context"
	"time"
)

type APP struct{

	logger log.Logger

	Http *http.Server

	Grpc *grpc.Server
}

func NewApp(logger log.Logger) *APP {

	app := &APP{
		logger: logger,
	}

	return app
}

func (this *APP) Stop()  {
	
}

func (this *APP) AwaitSignal() {
	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	select {
	case s := <-c:
		this.logger.Infof("receive a signal %s", s.String())
		if this.Http != nil {
			ctx, cannel := context.WithTimeout(context.Background(), 3 * time.Second)
			defer cannel()
			if err := this.Http.Shutdown(ctx); err != nil {
				this.logger.Errorf("stop http server error %s", err.Error())
			}
		}

		if this.Grpc != nil {
			this.Grpc.GracefulStop();
		}

		os.Exit(0)
	}
}