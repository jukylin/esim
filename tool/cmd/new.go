package cmd

import (
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/jukylin/esim/tool/new"
	"github.com/spf13/cobra"
)

// grpcCmd represents the grpc command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "create a new project",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		new.NewProject(
			new.WithProjectLogger(logger),
			new.WithProjectWriter(file_dir.NewEsimWriter()),
			new.WithProjectTpl(templates.NewTextTpl()),
		).Run(v)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().BoolP("beego", "", false, "init beego server")

	newCmd.Flags().BoolP("gin", "", true, "init gin server")

	newCmd.Flags().BoolP("grpc", "", false, "init grpc server")

	newCmd.Flags().BoolP("monitoring", "m", true, "enable monitoring")

	newCmd.Flags().StringP("server_name", "s", "", "server name")

	err := v.BindPFlags(newCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}
