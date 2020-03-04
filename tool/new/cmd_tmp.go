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
	{{service_name}} "{{PROPATH}}{{service_name}}/internal"
	"{{PROPATH}}{{service_name}}/internal/infra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {

		app := {{service_name}}.NewApp()

		app.Infra = infra.NewInfra()

{{RUN_SERVER}}

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