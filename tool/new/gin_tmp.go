package new

func GinInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/transports/http/controllers",
		Content: `package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"{{PROPATH}}{{service_name}}/internal/app/service"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"{{PROPATH}}{{service_name}}/internal/app/dto"
)

func Demo(c *gin.Context) {

	username := c.GetString("username")
	svc := service.NewUserSvc(infra.NewInfra())
	info := svc.GetUserInfo(c.Request.Context(), username)

	c.JSON(http.StatusOK, gin.H{
		"data": dto.NewUser(info),
	})
}


func Esim(c *gin.Context)  {
	inf := infra.NewInfra()

	c.JSON(http.StatusOK, gin.H{
		"Esim" : inf.String(),
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


func RegisterGinServer(en *gin.Engine)  {
	en.GET("/", controllers.Esim)

	en.GET("/demo", controllers.Demo)
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
	"github.com/jukylin/esim/container"
	"{{PROPATH}}{{service_name}}/internal/transports/http/routers"
	"github.com/jukylin/esim/middle-ware"
	"gitlab.etcchebao.cn/go_service/esim/pkg/log"
)

type GinServer struct{
	en *gin.Engine

	addr string

	logger log.Logger

	server *http.Server
}

func NewGinServer(esim *container.Esim) *GinServer {

	serviceName := esim.Conf.GetString("appname")
	httpport := esim.Conf.GetString("httpport")

	in := strings.Index(httpport, ":")
	if in < 0 {
		httpport = ":"+httpport
	}

	en := gin.Default()

	if esim.Conf.GetString("runmode") != "pro"{
		gin.SetMode(gin.DebugMode)
	}else{
		gin.SetMode(gin.ReleaseMode)
	}

	if esim.Conf.GetBool("http_tracer") == true{
		en.Use(middle_ware.GinTracer(serviceName, esim.Logger))
	}

	if esim.Conf.GetBool("http_metrics") == true {
		en.Use(middle_ware.GinMonitor())
	}

	server := &GinServer{
		en : en,
		addr : httpport,
	}

	return server
}


func (this *GinServer) Start(){
	routers.RegisterGinServer(this.en)

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

	Files = append(Files, fc1, fc2, fc3)
}
