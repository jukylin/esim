package new

func GrpcInit() {
	fc1 := &FileContent{
		FileName: "demo.go",
		Dir:      "internal/transports/grpc/controllers",
		Content: `package controllers

import (
	"context"
	"encoding/json"
	gp "{{PROPATH}}{{service_name}}/internal/infra/third_party/protobuf/passport"
	"{{PROPATH}}{{service_name}}/internal/application"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"github.com/tidwall/gjson"
)

type DemoController struct {
	infra *infra.Infra

	userSvc *application.UserService
}


func (this *DemoController) GetUserByUserName(ctx context.Context,
	request *gp.GetUserByUserNameRequest) (*gp.GrpcReplyMap, error) {
	grpcReply := &gp.GrpcReplyMap{}
	userName := request.GetUsername()

	userInfo := this.userSvc.GetUserInfo(ctx, userName)

	grpcReply.Code = 0;
	userInfoJson, err := json.Marshal(userInfo)
	if err != nil {
		grpcReply.Code = -1;
		grpcReply.Msg = err.Error();
		return grpcReply, nil
	}

	grpcReply.Data = make(map[string]string)

	gjson.Parse(string(userInfoJson)).ForEach(func(key, value gjson.Result) bool {
		grpcReply.Data[key.String()] = value.String()
		return true
	})

	return grpcReply, nil
}
`,
	}

	fc2 := &FileContent{
		FileName: "routers.go",
		Dir:      "internal/transports/grpc/routers",
		Content: `package routers

import (
	"{{PROPATH}}{{service_name}}/internal/transports/grpc/controllers"
	"{{PROPATH}}{{service_name}}/internal/infra/third_party/protobuf/passport"
	"google.golang.org/grpc"
)


func RegisterGrpcServer(s *grpc.Server, controllers *controllers.Controllers)  {
	passport.RegisterUserInfoServer(s, controllers.Demo)
}
`,
	}

	fc3 := &FileContent{
		FileName: "grpc.go",
		Dir:      "internal/transports/grpc",
		Content: `package grpc

import (
	"strings"

	"github.com/jukylin/esim/grpc"
	"{{PROPATH}}{{service_name}}/internal"
	"{{PROPATH}}{{service_name}}/internal/transports/grpc/routers"
	"{{PROPATH}}{{service_name}}/internal/transports/grpc/controllers"
)

func NewGrpcServer(app *{{service_name}}.App) *grpc.GrpcServer {

	target := app.Conf.GetString("grpc_server_tcp")

	in := strings.Index(target, ":")
	if in < 0 {
		target = ":"+target
	}

	serverOptions := grpc.ServerOptions{}

	//grpc服务初始化
	grpcServer :=  grpc.NewGrpcServer(target,
		serverOptions.WithServerConf(app.Conf),
		serverOptions.WithServerLogger(app.Logger),
		serverOptions.WithUnarySrvItcp(),
		serverOptions.WithGrpcServerOption(),
	)

	//注册grpc路由
	routers.RegisterGrpcServer(grpcServer.Server, controllers.NewControllers(app))

	return grpcServer
}
`,
	}

	fc4 := &FileContent{
		FileName: "component_test.go",
		Dir:      "internal/transports/grpc/component-test",
		Content: `// +build component_test

package component_test

import (
	"os"
	"testing"
	"context"
	"{{PROPATH}}{{service_name}}/internal"
	"{{PROPATH}}{{service_name}}/internal/transports/grpc"
	"{{PROPATH}}{{service_name}}/internal/infra"
	gp "{{PROPATH}}{{service_name}}/internal/infra/third_party/protobuf/passport"
	"github.com/stretchr/testify/assert"
)

var app *{{service_name}}.App

func TestMain(m *testing.M) {

	app = {{service_name}}.NewApp()

	app.Infra = infra.NewInfra()

	app.Trans = append(app.Trans, grpc.NewGrpcServer(app))

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

//go test -v -tags="component_test"
func TestUserService_GetUserByUserName(t *testing.T)  {
	ctx := context.Background()

	conn := app.Infra.GrpcClient.DialContext(ctx, ":" + app.Conf.GetString("grpc_server_tcp"))

	client := gp.NewUserInfoClient(conn)

	req := &gp.GetUserByUserNameRequest{}
	req.Username = "demo"
	reply, err := client.GetUserByUserName(ctx, req)
	if err != nil{
		app.Logger.Errorf(err.Error())
	}else {
		assert.Equal(t, "demo", reply.Data["UserName"])
		assert.Equal(t, int32(0), reply.Code)
	}
}
`,
	}


	fc5 := &FileContent{
		FileName: "controllers.go",
		Dir:      "internal/transports/grpc/controllers",
		Content: `package controllers

import (
	"{{PROPATH}}{{service_name}}/internal"
	"github.com/google/wire"
	"{{PROPATH}}{{service_name}}/internal/application"
)


type Controllers struct {

	App *{{service_name}}.App

	Demo *DemoController
}


var controllersSet = wire.NewSet(
	wire.Struct(new(Controllers), "*"),
	provideDemoController,
)


func NewControllers(app *{{service_name}}.App) *Controllers {
	controllers := initControllers(app)
	return controllers
}


func provideDemoController(app *{{service_name}}.App) *DemoController {

	userSvc := application.NewUserSvc(app.Infra)

	demoController := &DemoController{}
	demoController.infra = app.Infra
	demoController.userSvc = userSvc

	return demoController
}
`,
	}


	fc6 := &FileContent{
		FileName: "wire.go",
		Dir:      "internal/transports/grpc/controllers",
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
		Dir:      "internal/transports/grpc/controllers",
		Content: `// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package controllers

import (
	"{{PROPATH}}{{service_name}}/internal"
)

// Injectors from wire.go:

func initControllers(app *{{service_name}}.App) *Controllers {
	demoController := provideDemoController(app)
	controllers := &Controllers{
		App:  app,
		Demo: demoController,
	}
	return controllers
}
`,
	}

	fc8 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/transports/grpc/dto",
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
