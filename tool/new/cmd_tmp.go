package new

func CmdInit() {

	fc1 := &FileContent{
		FileName: "root.go",
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
	"fmt"
	"os"
	"github.com/spf13/cobra"
{{IMPORT_SERVER}}
	"github.com/jukylin/esim"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/config"
	"{{service_name}}/internal/infra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {

		container.SetConfFunc(func() config.Config {
			options := config.ViperConfOptions{}

			env := os.Getenv("ENV")
			if env == "" {
				env = "dev"
			}

			gopath := os.Getenv("GOPATH")

			monitFile := gopath + "/src/{{service_name}}/" + "conf/monitoring.yaml"
			confFile := gopath + "/src/{{service_name}}/" + "conf/" + env + ".yaml"

			file := []string{monitFile, confFile}
			return config.NewViperConfig(options.WithConfigType("yaml"),
				options.WithConfFile(file))
		})

		em := container.NewEsim()
		app := esim.NewApp(em.Logger)

{{RUN_SERVER}}

		app.Infra = infra.NewInfra()

		app.Start()
		app.AwaitSignal()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func init() {
}
`,
	}

	Files = append(Files, fc1)
}