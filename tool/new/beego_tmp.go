package new

func BeegoInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/transports/http/controllers",
		Content: `package controllers

import (
	"{{PROPATH}}{{service_name}}/internal/application"
	"github.com/astaxie/beego"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"{{PROPATH}}{{service_name}}/internal/transports/http/dto"
)

// Operations about Users
type UserController struct {
	beego.Controller
}


func (this *UserController) GetUserInfo() {
	username := this.GetString("username")

	svc := application.NewUserSvc(infra.NewInfra())

	user := svc.GetUserInfo(this.Ctx.Request.Context(), username)

	this.Data["json"] = dto.NewUser(user)
	this.ServeJSON()
}
`,
	}

	fc2 := &FileContent{
		FileName: "routers.go",
		Dir:      "internal/transports/http/routers",
		Content: `package routers

import (
	"{{PROPATH}}{{service_name}}/internal/transports/http/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.IndexController{})

	beego.Router("/ping", &controllers.PingController{})

	ns := beego.NewNamespace("/v1",
		beego.NSAutoRouter(&controllers.UserController{}),
	)
	beego.AddNamespace(ns)
}
`,
	}

	fc4 := &FileContent{
		FileName: "beego.go",
		Dir:      "internal/transports/http",
		Content: `package http

import (
	"net/http"
	"strings"
	"time"
	middle_ware "github.com/jukylin/esim/middle-ware"
	"github.com/astaxie/beego"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/jukylin/esim/container"
	_ "{{PROPATH}}{{service_name}}/internal/transports/http/routers"
)

type BeegoServer struct{
	httpport string

	esim *container.Esim
}


func NewBeegoServer(esim *container.Esim) *BeegoServer  {

	beegoServer := &BeegoServer{}

	httpport := esim.Conf.GetString("httpport")

	in := strings.Index(httpport, ":")
	if in < 0 {
		httpport = ":"+httpport
	}

	beegoServer.httpport = httpport
	beegoServer.esim = esim

	return beegoServer
}


func (this *BeegoServer) Start()  {
	go func() {
		beego.RunWithMiddleWares(this.httpport, getMwd(this.esim)...)
	}()
	time.Sleep(100 * time.Millisecond)
}

func (this *BeegoServer) GracefulShutDown()  {
	//beego do this itself
}

func getMwd(esim *container.Esim) []beego.MiddleWare {

	var mws []beego.MiddleWare

	if esim.Conf.GetBool("http_tracer") == true{
		mws = append(mws, func(handler http.Handler) http.Handler {
			return nethttp.Middleware(esim.Tracer, handler)
		})
	}

	if esim.Conf.GetBool("http_metrics") == true {
		mws = append(mws, func(handler http.Handler) http.Handler {
			return middle_ware.Monitor(handler)
		})
	}

	return mws
}
`,
	}

	fc5 := &FileContent{
		FileName: "index.go",
		Dir:      "internal/transports/http/controllers",
		Content: `package controllers

import (
	"github.com/astaxie/beego"
	"{{PROPATH}}{{service_name}}/internal/infra"
)

// Operations about Index
type IndexController struct {
	beego.Controller
}

// @router / [get]
func (this *IndexController) Get() {
	this.Data["json"] = map[string]interface{}{"Esim" : infra.NewInfra().String()}
	this.ServeJSON()
}

type PingController struct {
	beego.Controller
}

// @router / [get]
func (this *PingController) Get() {
	errs := infra.NewInfra().HealthCheck()
	if len(errs) > 0{
		for _, err := range errs {
			infra.NewInfra().Logger.Errorf(err.Error())
		}
		this.Abort("500")
	}else{
		this.Data["json"] = map[string]interface{}{"msg" : "success"}
		this.ServeJSON()
	}
}
`,
	}

	fc6 := &FileContent{
		FileName: "component_test.go",
		Dir:      "internal/transports/http/component-test",
		Content: `// +build component_test

package component_test

import (
	"os"
	"testing"
	"context"
	"io/ioutil"
	{{service_name}} "{{PROPATH}}{{service_name}}/internal"
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
	appOptions := {{package_name}}.AppOptions{}
	app := {{package_name}}.NewApp(appOptions.WithConfPath("../../../../conf/"))

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

	app.Trans = append(app.Trans, http.NewBeegoServer(app.Esim))

	app.Start()

	errs := app.Infra.HealthCheck()
	if len(errs) > 0{
		for _, err := range errs {
			app.Logger.Errorf(err.Error())
		}
	}
}


func tearDown(app *{{package_name}}.App)  {
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

	fc7 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/transports/http/dto",
		Content: `package dto

import "{{PROPATH}}{{service_name}}/internal/domain/user/entity"

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

	Files = append(Files, fc1, fc2, fc4, fc5, fc6, fc7)
}
