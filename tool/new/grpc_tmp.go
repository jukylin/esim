package new

func GrpcInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/transports/grpc/controllers",
		Content: `package controllers

import (
	"context"
	"encoding/json"
	gp "{{PROPATH}}{{service_name}}/internal/infra/third_party/protobuf/passport"
	"{{PROPATH}}{{service_name}}/internal/app/service"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"github.com/tidwall/gjson"
)

type UserService struct {
}


func (u *UserService) GetUserByUserName(ctx context.Context,
	request *gp.GetUserByUserNameRequest) (*gp.GrpcReplyMap, error) {
	grpcReply := &gp.GrpcReplyMap{}
	userName := request.GetUsername()

	svc := service.NewUserSvc(infra.NewInfra())
	userInfo := svc.GetUserInfo(ctx, userName)

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


func RegisterGrpcServer(s *grpc.Server)  {
	passport.RegisterUserInfoServer(s, &controllers.UserService{})
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
	"{{PROPATH}}{{service_name}}/internal/transports/grpc/routers"
	"github.com/jukylin/esim/container"
)

func NewGrpcServer(esim *container.Esim) *grpc.GrpcServer {

	target := esim.Conf.GetString("grpc_server_tcp")

	in := strings.Index(target, ":")
	if in < 0 {
		target = ":"+target
	}

	serverOptions := grpc.ServerOptions{}

	//grpc服务初始化
	grpcServer :=  grpc.NewGrpcServer(target,
		serverOptions.WithServerConf(esim.Conf),
		serverOptions.WithServerLogger(esim.Logger),
		serverOptions.WithUnarySrvItcp(),
		serverOptions.WithGrpcServerOption(),
	)

	//注册grpc路由
	routers.RegisterGrpcServer(grpcServer.Server)

	return grpcServer
}
`,
	}

	fc4 := &FileContent{
		FileName: "component_test.go",
		Dir:      "internal/transports/grpc/component-test",
		Content: `// +build compnent_test

package compnent_test

import (
	"os"
	"testing"
	"{{PROPATH}}{{service_name}}/internal"
	"{{PROPATH}}{{service_name}}/internal/transports/grpc"
	"{{PROPATH}}{{service_name}}/internal/infra"
)

var app *{{service_name}}.App

func TestMain(m *testing.M) {

	app = {{service_name}}.NewApp()

	app.Infra = infra.NewInfra()

	app.Trans = append(app.Trans, grpc.NewGrpcServer(app.Esim))

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
`,
	}

	Files = append(Files, fc1, fc2, fc3, fc4)
}
