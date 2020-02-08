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
		en.Use(middle_ware.GinTracer(esim.Tracer))
	}

	if esim.Conf.GetBool("http_metrics") == true {
		en.Use(middle_ware.GinMonitor())
	}

	server := &GinServer{
		en : en,
		addr : httpport,
		logger: esim.Logger,
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

	fc4 := &FileContent{
		FileName: "component_test.go",
		Dir:      "internal/transports/http/component-test",
		Content: `// +build compnent_test

package compnent_test

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
)

var app *{{service_name}}.App

func TestMain(m *testing.M) {

	app = {{service_name}}.NewApp()

	app.Trans = append(app.Trans, http.NewGinServer(app.Esim))

	app.Infra = infra.NewInfra()

	app.Start()
	errs := app.Infra.HealthCheck()
	if len(errs) > 0{
		for _, err := range errs {
			app.Logger.Errorf(err.Error())
		}
	}

	code := m.Run()

	app.Infra.Close()

	os.Exit(code)
}


func TestControllers_Esim(t *testing.T)  {

	client := http_client.NewHttpClient()
	ctx := context.Background()
	resp, err := client.Get(ctx, "http://localhost:"+ app.Conf.GetString("httpport"))

	if err != nil{
		app.Logger.Errorf(err.Error())
	}

	defer resp.Body.Close()


	body, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		app.Logger.Errorf(err.Error())
	}
	println(string(body))
	assert.Equal(t, 200, resp.StatusCode)
}
`,
	}

	Files = append(Files, fc1, fc2, fc3, fc4)
}
