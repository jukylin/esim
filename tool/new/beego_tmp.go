package new

func BeegoInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/server/controllers",
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
		Dir:      "internal/server/routers",
		Content: `package routers

import (
	"{{PROPATH}}{{service_name}}/internal/server/controllers"
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

	fc3 := &FileContent{
		FileName: "beego.go",
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
	"{{PROPATH}}{{service_name}}/internal/server"
	"github.com/jukylin/esim/container"
	_ "{{PROPATH}}{{service_name}}/internal/server/routers"
)

// grpcCmd represents the grpc command
var beegoCmd = &cobra.Command{
	Use:   "beego",
	Short: "提供外部调用的Restful api 接口",
	Long: {{!}}提供外部调用的Restful api 接口{{!}},
	Run: func(cmd *cobra.Command, args []string) {
		esim := container.NewEsim()
		server.NewBeegoServer(esim)
	},
}

func init() {
	rootCmd.AddCommand(beegoCmd)

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
		FileName: "beego.go",
		Dir:      "internal/server",
		Content: `package server

import (
	"net/http"
	"strings"

	"github.com/jukylin/esim/server"
	"github.com/astaxie/beego"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/jukylin/esim/opentracing"
	"github.com/jukylin/esim/container"
)

func NewBeegoServer(esim *container.Esim)  {

	httpport := esim.Conf.GetString("httpport")

	in := strings.Index(httpport, ":")
	if in < 0 {
		httpport = ":"+httpport
	}

	beego.RunWithMiddleWares(httpport, getMwd(esim)...)
}

func getMwd(esim *container.Esim) []beego.MiddleWare {

	var mws []beego.MiddleWare

	serviceName := esim.Conf.GetString("appname")

	if esim.Conf.GetBool("http_tracer") == true{
		mws = append(mws, func(handler http.Handler) http.Handler {
			tracer := opentracing.NewTracer(serviceName, esim.Log)
			return nethttp.Middleware(tracer, handler)
		})
	}

	if esim.Conf.GetBool("http_metrics") == true {
		mws = append(mws, func(handler http.Handler) http.Handler {
			return server.Monitor(handler)
		})
	}

	return mws
}
`,
	}

	fc5 := &FileContent{
		FileName: "index.go",
		Dir:      "internal/server/controllers",
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

	Files = append(Files, fc1, fc2, fc3, fc4, fc5)
}
