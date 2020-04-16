package new

func init()  {
	Files = append(Files, internalfc1)
}

var (
	internalfc1 = &FileContent{
		FileName: "app.go",
		Dir:      "internal",
		Content: `package {{package_name}}

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

	confPath string
}

type Option func(c *App)

type AppOptions struct{}

func NewApp(options ...Option) *App {

	app := &App{}

	for _, option := range options {
		option(app)
	}

	if app.confPath == ""{
		app.confPath = "conf/"
	}

	container.SetConfFunc(func() config.Config {
		options := config.ViperConfOptions{}

		monitFile := app.confPath + "monitoring.yaml"
		confFile := app.confPath + "conf.yaml"

		file := []string{monitFile, confFile}
		conf := config.NewViperConfig(options.WithConfigType("yaml"),
			options.WithConfFile(file))

		env := os.Getenv("ENV")
		if env == "" {
			conf.Set("runmode", "dev")
		}

		if conf.GetString("runmode") != "pro" {
			conf.Set("debug", true)
		}

		return conf
	})

	app.Esim = container.NewEsim()

	return app
}

func (AppOptions) WithConfPath(confPath string) Option {
	return func(app *App) {
		app.confPath = confPath
	}
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


)
