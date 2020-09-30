package new

var (
	internalfc1 = &FileContent{
		FileName: "app.go",
		Dir:      "internal",
		Content: `package {{.PackageName}}

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/jukylin/esim/config"
	"github.com/jukylin/esim/transports"
	"{{.ProPath}}{{.ServerName}}/internal/infra"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/config"
	eot "github.com/jukylin/esim/opentracing"
	"github.com/jukylin/esim/prometheus"
	"github.com/jukylin/esim/log"
)

const defaultAppname = "esim"
const defaultPrometheusHTTPArrd = "9002"

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

	if app.confPath == "" {
		app.confPath = "conf/"
	}

	monitFile := app.confPath + "monitoring.yaml"
	confFile := app.confPath + "conf.yaml"
	file := []string{monitFile, confFile}

	confOps := config.ViperConfOptions{}
	conf := config.NewViperConfig(
		confOps.WithConfigType("yaml"),
		confOps.WithConfFile(file))

	env := os.Getenv("ENV")
	if env == "" {
		conf.Set("runmode", "dev")
	}

	ez := log.NewEsimZap(
		log.WithEsimZapDebug(conf.GetBool("debug")),
		log.WithEsimZapJSON(conf.GetString("runmode") == "pro"),
	)

	logger := log.NewLogger(
		log.WithEsimZap(ez),
	)

	appname := defaultAppname
	if conf.GetString("appname") != "" {
		appname = conf.GetString("appname")
	}
	tracer := eot.NewTracer(appname, logger)

	httpAddr := defaultPrometheusHTTPArrd
	if conf.GetString("prometheus_http_addr") != "" {
		httpAddr = conf.GetString("prometheus_http_addr")
	}
	promer := prometheus.NewPrometheus(httpAddr, logger)

	app.Esim = container.NewEsim(
		container.WithEsimZap(ez),
		container.WithLogger(logger),
		container.WithConf(conf),
		container.WithTracer(tracer),
		container.WithPromer(promer),
	)

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

	s := <-c
	app.Esim.Logger.Infof("receive a signal %s", s.String())
	app.stop()

	close(c)
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

func initInternalFiles() {
	Files = append(Files, internalfc1)
}
