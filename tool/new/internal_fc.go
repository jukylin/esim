package new

func init() {
	Files = append(Files, internalfc1)
}

var (
	internalfc1 = &FileContent{
		FileName: "app.go",
		Dir:      "internal",
		Content: `package {{.PackageName}}

import (
	"os"
	"os/signal"
	"syscall"
	"github.com/jukylin/esim/transports"
	"{{.ProPath}}{{.ServerName}}/internal/infra"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/config"
)

type App struct{

	*container.Esim

	trans []transports.Transports

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

func (app *App) Start()  {
	for _, tran := range app.trans {
		tran.Start()
	}
}

func (app *App) RegisterTran(tran transports.Transports) {
	app.trans = append(app.trans, tran)
}

func (app *App) AwaitSignal() {
	c := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)
	select {
	case s := <-c:
		app.Esim.Logger.Infof("receive a signal %s", s.String())
		app.stop()
		os.Exit(0)
	}
}


func (app *App) stop()  {
	for _, tran := range app.trans {
		tran.GracefulShutDown()
	}

	app.Infra.Close()
}
`,
	}
)
