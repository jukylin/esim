package new

func InternalInit() {
	fc1 := &FileContent{
		FileName: "app.go",
		Dir:      "internal",
		Content: `package {{service_name}}

import (
	"os"
	"os/signal"
	"syscall"
	"github.com/jukylin/esim/transports"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/config"
)

type App struct{

	*container.Esim

	Trans []transports.Transports

	Infra *infra.Infra
}

func NewApp() *App {

	container.SetConfFunc(func() config.Config {
		options := config.ViperConfOptions{}

		env := os.Getenv("ENV")
		if env == "" {
			env = "dev"
		}

		gopath := os.Getenv("GOPATH")

		monitFile := gopath + "/src/{{PROPATH}}{{service_name}}/" + "conf/monitoring.yaml"
		confFile := gopath + "/src/{{PROPATH}}{{service_name}}/" + "conf/" + env + ".yaml"

		file := []string{monitFile, confFile}
		return config.NewViperConfig(options.WithConfigType("yaml"),
			options.WithConfFile(file))
	})

	esim := container.NewEsim()

	app := &App{
		Esim: esim,
	}

	return app
}


func (this *App) Start()  {
	for _, tran := range this.Trans {
		tran.Start()
	}
}

func (this *App) AwaitSignal() {
	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	select {
	case s := <-c:
		this.Esim.Logger.Infof("receive a signal %s", s.String())
		this.stop()
		os.Exit(0)
	}
}


func (this *App) stop()  {
	for _, tran := range this.Trans {
		tran.GracefulShutDown()
	}

	this.Infra.Close()
}
`,
	}


	Files = append(Files, fc1)
}
