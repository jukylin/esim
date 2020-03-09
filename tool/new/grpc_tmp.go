package new

func GrpcInit() {
	fc1 := &FileContent{
		FileName: "demo_controller.go",
		Dir:      "internal/transports/grpc/controllers",
		Content: `package controllers

import (
	"context"
	gp "{{PROPATH}}{{service_name}}/internal/infra/third_party/protobuf/passport"
	"{{PROPATH}}{{service_name}}/internal/application"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"{{PROPATH}}{{service_name}}/internal/transports/grpc/dto"
)

type DemoController struct {
	infra *infra.Infra

	userSvc *application.UserService
}


func (this *DemoController) GetUserByUserName(ctx context.Context,
	request *gp.GetUserByUserNameRequest) (*gp.GrpcUserReply, error) {
	grpcReply := &gp.GrpcUserReply{}
	userName := request.GetUsername()

	userInfo := this.userSvc.GetUserInfo(ctx, userName)

	grpcReply.Code = 0;

	grpcReply.Data = dto.NewUserInfo(userInfo)

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
	{{service_name}} "{{PROPATH}}{{service_name}}/internal"
	"{{PROPATH}}{{service_name}}/internal/transports/grpc/routers"
	"{{PROPATH}}{{service_name}}/internal/transports/grpc/controllers"
)

func NewGrpcServer(app *{{package_name}}.App) *grpc.GrpcServer {

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
		serverOptions.WithTracer(app.Tracer),
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
	{{service_name}} "{{PROPATH}}{{service_name}}/internal"
	"{{PROPATH}}{{service_name}}/internal/transports/grpc"
	"{{PROPATH}}{{service_name}}/internal/infra"
	gp "{{PROPATH}}{{service_name}}/internal/infra/third_party/protobuf/passport"
	"github.com/stretchr/testify/assert"
	_grpc "google.golang.org/grpc"
	"github.com/jukylin/esim/container"
	egrpc "github.com/jukylin/esim/grpc"
	"github.com/jukylin/esim/log"
)


func TestMain(m *testing.M) {
	appOptions := {{package_name}}.AppOptions{}
	app := {{package_name}}.NewApp(appOptions.WithConfPath("../../../../conf/"))

	setUp(app)

	code := m.Run()

	tearDown(app)

	os.Exit(code)
}


func provideStubsGrpcClient(esim *container.Esim) *egrpc.GrpcClient {
	clientOptional := egrpc.ClientOptionals{}
	clientOptions := egrpc.NewClientOptions(
		clientOptional.WithLogger(esim.Logger),
		clientOptional.WithConf(esim.Conf),
		clientOptional.WithDialOptions(_grpc.WithUnaryInterceptor(
			egrpc.ClientStubs(func(ctx context.Context, method string, req, reply interface{}, cc *_grpc.ClientConn, invoker _grpc.UnaryInvoker, opts ..._grpc.CallOption) error {
				esim.Logger.Infof(method)
				err := invoker(ctx, method, req, reply, cc, opts...)
				return err
			}),
		),),
	)

	grpcClient := egrpc.NewClient(clientOptions)

	return grpcClient
}


func setUp(app *{{package_name}}.App)  {

	app.Infra = infra.NewStubsInfra(provideStubsGrpcClient(app.Esim))

	app.Trans = append(app.Trans, grpc.NewGrpcServer(app))

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
func TestUserService_GetUserByUserName(t *testing.T)  {
	logger := log.NewLogger()

	ctx := context.Background()

	grpcClient := egrpc.NewClient(egrpc.NewClientOptions())
	conn := grpcClient.DialContext(ctx, ":50055")
	defer conn.Close()

	client := gp.NewUserInfoClient(conn)

	req := &gp.GetUserByUserNameRequest{}
	req.Username = "demo"
	reply, err := client.GetUserByUserName(ctx, req)
	if err != nil{
		logger.Errorf(err.Error())
	}else {
		assert.Equal(t, "demo", reply.Data.UserName)
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
	{{service_name}} "{{PROPATH}}{{service_name}}/internal"
	"github.com/google/wire"
	"{{PROPATH}}{{service_name}}/internal/application"
)


type Controllers struct {

	App *{{package_name}}.App

	Demo *DemoController
}


var controllersSet = wire.NewSet(
	wire.Struct(new(Controllers), "*"),
	provideDemoController,
)


func NewControllers(app *{{package_name}}.App) *Controllers {
	controllers := initControllers(app)
	return controllers
}


func provideDemoController(app *{{package_name}}.App) *DemoController {

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
	{{service_name}} "{{PROPATH}}{{service_name}}/internal"
)



func initControllers(app *{{package_name}}.App) *Controllers {
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
	{{service_name}} "{{PROPATH}}{{service_name}}/internal"
)

// Injectors from wire.go:

func initControllers(app *{{package_name}}.App) *Controllers {
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
		FileName: "user_dto.go",
		Dir:      "internal/transports/grpc/dto",
		Content: `package dto

import (
	"{{PROPATH}}{{service_name}}/internal/domain/user/entity"
	"{{PROPATH}}{{service_name}}/internal/infra/third_party/protobuf/passport"
)

type User struct {

	//用户名称
	UserName string {{!}}json:"user_name"{{!}}

	//密码
	PassWord string {{!}}json:"pass_word"{{!}}
}

func NewUserInfo(user entity.User) *passport.Info {
	info := &passport.Info{}
	info.UserName = user.UserName
	info.PassWord = user.PassWord
	return info
}`,
	}

	Files = append(Files, fc1, fc2, fc3, fc4, fc5, fc6, fc7, fc8)
}
