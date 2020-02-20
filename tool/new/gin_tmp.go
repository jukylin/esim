package new

func GinInit() {
	fc1 := &FileContent{
		FileName: "demo.go",
		Dir:      "internal/transports/http/controllers",
		Content: `package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"{{PROPATH}}{{service_name}}/internal/application"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"{{PROPATH}}{{service_name}}/internal/transports/http/dto"
)

type DemoController struct {
	infra *infra.Infra

	UserSvc *application.UserService
}

func (this *DemoController) Demo(c *gin.Context) {

	username := c.GetString("username")
	info := this.UserSvc.GetUserInfo(c.Request.Context(), username)

	c.JSON(http.StatusOK, gin.H{
		"data": dto.NewUser(info),
	})
}


type PingController struct {
	infra *infra.Infra
}

func (this *PingController) Ping(c *gin.Context)  {
	errs := this.infra.HealthCheck()

	if len(errs) > 0{
		for _, err := range errs{
			infra.NewInfra().Logger.Errorf(err.Error())
		}
		c.JSON(http.StatusInternalServerError, gin.H{})
	}else{
		c.JSON(http.StatusOK, gin.H{})
	}
}

type EsimController struct {
	infra *infra.Infra
}

func (this *EsimController) Esim(c *gin.Context)  {
	c.JSON(http.StatusOK, gin.H{
		"Esim" : this.infra.String(),
	})
}
`,
	}

	fc2 := &FileContent{
		FileName: "routers.go",
		Dir:      "internal/transports/http/routers",
		Content: `package routers

import (
	"{{PROPATH}}{{service_name}}/internal/transports/http/controllers"
	"github.com/gin-gonic/gin"

)


func RegisterGinServer(en *gin.Engine, controllers *controllers.Controllers)  {
	en.GET("/", controllers.Esim.Esim)

	en.GET("/demo", controllers.Demo.Demo)

	en.GET("/ping", controllers.Ping.Ping)
}
`,
	}

	fc3 := &FileContent{
		FileName: "gin.go",
		Dir:      "internal/transports/http",
		Content: `package http

import (
	"strings"
	"net/http"
	"context"
	"time"
	"github.com/gin-gonic/gin"
	"{{PROPATH}}{{service_name}}/internal/transports/http/routers"
	"github.com/jukylin/esim/middle-ware"
	"github.com/jukylin/esim/log"
	"{{PROPATH}}{{service_name}}/internal/transports/http/controllers"
	"{{PROPATH}}{{service_name}}/internal"
)

type GinServer struct{
	en *gin.Engine

	addr string

	logger log.Logger

	server *http.Server

	app *{{service_name}}.App
}

func NewGinServer(app *{{service_name}}.App) *GinServer {

	httpport := app.Conf.GetString("httpport")

	in := strings.Index(httpport, ":")
	if in < 0 {
		httpport = ":"+httpport
	}

	en := gin.Default()

	if app.Conf.GetString("runmode") != "pro"{
		gin.SetMode(gin.DebugMode)
	}else{
		gin.SetMode(gin.ReleaseMode)
	}

	if app.Conf.GetBool("http_tracer") == true{
		en.Use(middle_ware.GinTracer(app.Tracer))
	}

	if app.Conf.GetBool("http_metrics") == true {
		en.Use(middle_ware.GinMonitor())
	}

	server := &GinServer{
		en : en,
		addr : httpport,
		logger: app.Logger,
		app: app,
	}

	return server
}


func (this *GinServer) Start(){
	routers.RegisterGinServer(this.en, controllers.NewControllers(this.app))

	server := &http.Server{Addr: this.addr, Handler: this.en}
	this.server = server
	go func() {
		if err := server.ListenAndServe(); err != nil{
			this.logger.Fatalf("start http server err %s", err.Error())
			return
		}
	}()
}

func (this *GinServer) GracefulShutDown()  {
	ctx, cannel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cannel()
	if err := this.server.Shutdown(ctx); err != nil {
		this.logger.Errorf("stop http server error %s", err.Error())
	}
}
`,
	}

	fc4 := &FileContent{
		FileName: "component_test.go",
		Dir:      "internal/transports/http/component-test",
		Content: `// +build component_test

package component_test

import (
	"os"
	"testing"
	"context"
	"io/ioutil"
	"{{PROPATH}}{{service_name}}/internal"
	"{{PROPATH}}{{service_name}}/internal/transports/http"
	"{{PROPATH}}{{service_name}}/internal/infra"
	http_client "github.com/jukylin/esim/http"
	"github.com/stretchr/testify/assert"
	"github.com/jukylin/esim/log"
	_grpc "google.golang.org/grpc"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/grpc"
)


func TestMain(m *testing.M) {
	appOptions := {{service_name}}.AppOptions{}
	app := {{service_name}}.NewApp(appOptions.WithConfPath("../../../../conf/"))

	setUp(app)

	code := m.Run()

	tearDown(app)

	os.Exit(code)
}


func provideStubsGrpcClient(esim *container.Esim) *grpc.GrpcClient {
	clientOptional := grpc.ClientOptionals{}
	clientOptions := grpc.NewClientOptions(
		clientOptional.WithLogger(esim.Logger),
		clientOptional.WithConf(esim.Conf),
		clientOptional.WithDialOptions(_grpc.WithUnaryInterceptor(
			grpc.ClientStubs(func(ctx context.Context, method string, req, reply interface{}, cc *_grpc.ClientConn, invoker _grpc.UnaryInvoker, opts ..._grpc.CallOption) error {
				esim.Logger.Infof(method)
				err := invoker(ctx, method, req, reply, cc, opts...)
				return err
			}),
		),),
	)

	grpcClient := grpc.NewClient(clientOptions)

	return grpcClient
}

func setUp(app *{{service_name}}.App)  {

	app.Infra = infra.NewStubsInfra(provideStubsGrpcClient(app.Esim))

	app.Trans = append(app.Trans, http.NewGinServer(app))

	app.Start()

	errs := app.Infra.HealthCheck()
	if len(errs) > 0{
		for _, err := range errs {
			app.Logger.Errorf(err.Error())
		}
	}
}


func tearDown(app *{{service_name}}.App)  {
	app.Infra.Close()
}

//go test -v -tags="component_test"
func TestControllers_Esim(t *testing.T)  {
	logger := log.NewLogger()

	client := http_client.NewHttpClient()
	ctx := context.Background()
	resp, err := client.Get(ctx, "http://localhost:8080")

	if err != nil{
		logger.Errorf(err.Error())
	}

	defer resp.Body.Close()


	body, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		logger.Errorf(err.Error())
	}
	logger.Debugf(string(body))
	assert.Equal(t, 200, resp.StatusCode)
}
`,
	}

	fc5 := &FileContent{
		FileName: "controllers.go",
		Dir:      "internal/transports/http/controllers",
		Content: `package controllers

import (
	"{{PROPATH}}{{service_name}}/internal"
	"github.com/google/wire"
	"{{PROPATH}}{{service_name}}/internal/application"
)


type Controllers struct {

	App *{{service_name}}.App

	Ping *PingController

	Esim *EsimController

	Demo *DemoController

}


var controllersSet = wire.NewSet(
	wire.Struct(new(Controllers), "*"),
	providePingController,
	provideEsimController,
	provideDemoController,
)


func NewControllers(app *{{service_name}}.App) *Controllers {
	controllers := initControllers(app)
	return controllers
}


func providePingController(app *{{service_name}}.App) *PingController {
	pingController := &PingController{}
	pingController.infra = app.Infra
	return pingController
}


func provideEsimController(app *{{service_name}}.App) *EsimController {
	esimController := &EsimController{}
	esimController.infra = app.Infra
	return esimController
}


func provideDemoController(app *{{service_name}}.App) *DemoController {

	userSvc := application.NewUserSvc(app.Infra)

	demoController := &DemoController{}
	demoController.infra = app.Infra
	demoController.UserSvc = userSvc

	return demoController
}
`,
	}


	fc6 := &FileContent{
		FileName: "wire.go",
		Dir:      "internal/transports/http/controllers",
		Content: `//+build wireinject

package controllers

import (
	"github.com/google/wire"
	"{{PROPATH}}{{service_name}}/internal"
)



func initControllers(app *{{service_name}}.App) *Controllers {
	wire.Build(controllersSet)
	return nil
}
`,
	}


	fc7 := &FileContent{
		FileName: "wire_gen.go",
		Dir:      "internal/transports/http/controllers",
		Content: `// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package controllers

import (
	"{{PROPATH}}{{service_name}}/internal"
)

// Injectors from wire.go:

func initControllers(app *{{service_name}}.App) *Controllers {
	pingController := providePingController(app)
	esimController := provideEsimController(app)
	demoController := provideDemoController(app)
	controllers := &Controllers{
		App:  app,
		Ping: pingController,
		Esim: esimController,
		Demo: demoController,
	}
	return controllers
}

`,
	}

	fc8 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/transports/http/dto",
		Content: `package dto

import "{{service_name}}/internal/domain/user/entity"

type User struct {

	//用户名称
	UserName string {{!}}json:"user_name"{{!}}

	//密码
	PassWord string {{!}}json:"pass_word"{{!}}
}

func NewUser(user entity.User) User {
	dto := User{}

	dto.UserName = user.UserName
	dto.PassWord = user.PassWord
	return dto
}`,
	}

	Files = append(Files, fc1, fc2, fc3, fc4, fc5, fc6, fc7, fc8)
}
