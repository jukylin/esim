package new

func GrpcInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/server/grpc/controllers",
		Content: `package controllers

import (
	"context"
	"encoding/json"
	gp "{{PROPATH}}{{service_name}}/internal/infra/third_party/protobuf/passport"
	"{{PROPATH}}{{service_name}}/internal/app"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"github.com/tidwall/gjson"
)

type UserService struct {
}


func (u *UserService) GetUserByUserName(ctx context.Context,
	request *gp.GetUserByUserNameRequest) (*gp.GrpcReplyMap, error) {
	grpcReply := &gp.GrpcReplyMap{}
	userName := request.GetUsername()

	svc := app.NewUserSvc(infra.NewInfra())
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
		Dir:      "internal/server/grpc/routers",
		Content: `package routers

import (
	"{{PROPATH}}{{service_name}}/internal/server/grpc/controllers"
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
		Dir:      "cmd",
		Content: `/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"{{PROPATH}}{{service_name}}/internal/server/grpc"
	"github.com/jukylin/esim/container"
)

// grpcCmd represents the grpc command
var grpcCmd = &cobra.Command{
	Use:   "grpc",
	Short: "提供内部调用的服务化接口",
	Long: {{!}}提供内部调用的服务化接口{{!}},
	Run: func(cmd *cobra.Command, args []string) {
		esim := container.NewEsim()
		grpc.NewGrpcServer(esim)
	},
}

func init() {
	rootCmd.AddCommand(grpcCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// grpcCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// grpcCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
`,
	}

	fc4 := &FileContent{
		FileName: "grpc.go",
		Dir:      "internal/server/grpc",
		Content: `package grpc

import (
	"net"
	"strings"

	"google.golang.org/grpc/reflection"
	"github.com/jukylin/esim/grpc"
	"{{PROPATH}}{{service_name}}/internal/server/grpc/routers"
	"github.com/jukylin/esim/container"
)

func NewGrpcServer(esim *container.Esim)  {

	serviceName := esim.Conf.GetString("appname")
	target := esim.Conf.GetString("grpc_server_tcp")

	in := strings.Index(target, ":")
	if in < 0 {
		target = ":"+target
	}

	stop := make(chan bool , 1)
	go func() {
		lis, err := net.Listen("tcp", target)
		if err != nil {
			esim.Log.Panicf("failed to listen: %s", err.Error())
		}

		serverOptions := grpc.ServerOptions{}

		//grpc服务初始化
		grpcServer :=  grpc.NewGrpcServer(serviceName,
			serverOptions.WithServerConf(esim.Conf),
			serverOptions.WithServerLogger(esim.Log),
			serverOptions.WithUnarySrvItcp(),
			serverOptions.WithGrpcServerOption(),
		)

		//注册grpc路由
		routers.RegisterGrpcServer(grpcServer.Server)

		// Register reflection service on gRPC server.
		reflection.Register(grpcServer.Server)
		if err := grpcServer.Server.Serve(lis); err != nil {
			esim.Log.Panicf("failed to serve: %s", err.Error())
		}

		stop <- true
	}()
	esim.Log.Infof("%s init success : %s", serviceName, target)

	<-stop
	esim.Log.Infof("%s stop success : %s", serviceName, target)
}

`,
	}

	Files = append(Files, fc1, fc2, fc3, fc4)
}
