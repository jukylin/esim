package new

func BeegoInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/transports/http/controllers",
		Content: `package controllers

import (
	"{{PROPATH}}{{service_name}}/internal/app"
	"github.com/astaxie/beego"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"{{PROPATH}}{{service_name}}/internal/domain/dto"
)

// Operations about Users
type UserController struct {
	beego.Controller
}


func (this *UserController) GetUserInfo() {
	username := this.GetString("username")

	svc := app.NewUserSvc(infra.NewInfra())

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

	"github.com/jukylin/esim/middle-ware"
	"github.com/astaxie/beego"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/jukylin/esim/opentracing"
	"github.com/jukylin/esim/container"
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
	beego.RunWithMiddleWares(this.httpport, getMwd(this.esim)...)

}

func (this *BeegoServer) GracefulShutDown()  {
	//beego do this itself
}

func getMwd(esim *container.Esim) []beego.MiddleWare {

	var mws []beego.MiddleWare

	serviceName := esim.Conf.GetString("appname")

	if esim.Conf.GetBool("http_tracer") == true{
		mws = append(mws, func(handler http.Handler) http.Handler {
			tracer := opentracing.NewTracer(serviceName, esim.Logger)
			return nethttp.Middleware(tracer, handler)
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
`,
	}

	Files = append(Files, fc1, fc2, fc4, fc5)
}
