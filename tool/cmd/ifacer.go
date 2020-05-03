package cmd

import (
	"github.com/spf13/cobra"
	//"log"
	"github.com/jukylin/esim/log"
	file_dir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/jukylin/esim/tool/ifacer"
)

var ifacerCmd = &cobra.Command{
	Use:   "ifacer",
	Short: "根据接口生成空实例",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		writer := &file_dir.EsimWriter{}
		ifacer := ifacer.NewIfacer(
			ifacer.WithIfacerLogger(log.NewLogger()),
			ifacer.WithIfacerTpl(templates.NewTextTpl()),
			ifacer.WithIfacerWriter(writer),
		)
		err := ifacer.Run(v)
		if err != nil {
			log.Log.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(ifacerCmd)

	ifacerCmd.Flags().StringP("iname", "", "", "接口名称")

	ifacerCmd.Flags().StringP("out", "o", "", "输出文件: abc.go")

	ifacerCmd.Flags().StringP("ipath", "i", "./*", "接口路径")

	ifacerCmd.Flags().BoolP("istar", "s", false, "带星")

	ifacerCmd.Flags().StringP("stname", "", "", "struct 名称：type struct_name struct{}")

	v.BindPFlags(ifacerCmd.Flags())
}
