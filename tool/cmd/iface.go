package cmd

import (
	"github.com/spf13/cobra"
	//"log"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/tool/iface"
)

var ifaceCmd = &cobra.Command{
	Use:   "iface",
	Short: "根据接口生成空实例",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		ifa := &iface.Iface{}
		err := ifa.Run(v)
		if err != nil {
			log.Log.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(ifaceCmd)

	ifaceCmd.Flags().StringP("name", "n", "", "接口名称")

	ifaceCmd.Flags().StringP("out", "o", "", "输出文件: abc.go")

	ifaceCmd.Flags().StringP("iface_path", "i", ".", "接口路径")

	ifaceCmd.Flags().BoolP("star", "s", false, "带星")

	ifaceCmd.Flags().StringP("struct_name", "", "", "struct 名称：type struct_name struct{}")

	v.BindPFlags(ifaceCmd.Flags())
}
