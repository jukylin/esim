package new

func GinInit() {
	fc1 := &FileContent{
		FileName: "user.go",
		Dir:      "internal/server/gin/controllers",
		Content: `package controllers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"{{PROPATH}}{{service_name}}/internal/app"
	"{{PROPATH}}{{service_name}}/internal/infra"
	"{{PROPATH}}{{service_name}}/internal/domain/dto"
)

func Demo(c *gin.Context) {

	username := c.GetString("username")
	svc := app.NewUserSvc(infra.NewInfra())
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
		Dir:      "internal/server/gin/routers",
		Content: `package routers

import (
	"{{PROPATH}}{{service_name}}/internal/server/gin/controllers"
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
	"github.com/jukylin/esim/container"
	"{{PROPATH}}{{service_name}}/internal/server/gin"
)

// grpcCmd represents the grpc command
var ginCmd = &cobra.Command{
	Use:   "gin",
	Short: "提供外部调用的Restful api 接口",
	Long: {{!}}提供外部调用的Restful api 接口{{!}},
	Run: func(cmd *cobra.Command, args []string) {
		esim := container.NewEsim()
		gin.NewGinServer(esim)
	},
}

func init() {
	rootCmd.AddCommand(ginCmd)

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
		FileName: "gin.go",
		Dir:      "internal/server/gin",
		Content: `package gin

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jukylin/esim/container"
	"{{PROPATH}}{{service_name}}/internal/server/gin/routers"
	"github.com/jukylin/esim/server"
)

func NewGinServer(esim *container.Esim)  {

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
		en.Use(server.GinTracer(serviceName, esim.Log))
	}

	if esim.Conf.GetBool("http_metrics") == true {
		en.Use(server.GinMonitor())
	}

	routers.RegisterGinServer(en)

	en.Run(httpport)
}
`,
	}

	Files = append(Files, fc1, fc2, fc3, fc4)
}
