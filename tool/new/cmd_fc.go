package new

func init()  {
	Files = append(Files, cmdfc)
}



var (
	cmdfc = &FileContent{
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
	"os"
	"github.com/spf13/cobra"
{{ range $Import := .ImportServer}}"{{$Import}}"
{{end}}
	"{{.ProPath}}{{.ServerName}}/internal"
	"{{.ProPath}}{{.ServerName}}/internal/infra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "",
	Short: "",
	Long: "",
	Run: func(cmd *cobra.Command, args []string) {

		app := {{.ServerName}}.NewApp()

		app.Infra = infra.NewInfra()

{{ range $Server := .RunServer}}{{$Server}}
{{end}}
		app.Start()
		app.AwaitSignal()

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
  if err := rootCmd.Execute(); err != nil {
    os.Exit(1)
  }
}

func init() {
}
`,
}
)