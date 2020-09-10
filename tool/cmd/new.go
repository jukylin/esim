package cmd

import (
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/jukylin/esim/tool/new"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "创建一个新项目",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		new.InitProject(
			new.WithProjectLogger(logger),
			new.WithProjectWriter(filedir.NewEsimWriter()),
			new.WithProjectTpl(templates.NewTextTpl()),
		).Run(v)
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().BoolP("beego", "", false, "使用Beego")

	newCmd.Flags().BoolP("gin", "", true, "使用Gin")

	newCmd.Flags().BoolP("grpc", "", false, "使用GRPC")

	newCmd.Flags().BoolP("monitoring", "m", true, "监控配置设置为true")

	newCmd.Flags().StringP("server_name", "s", "", "服务名称")

	err := v.BindPFlags(newCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}
