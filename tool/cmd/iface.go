package cmd

import (
	"github.com/spf13/cobra"
	//"log"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/tool/factory"
)

var ifaceCmd = &cobra.Command{
	Use:   "factory",
	Short: "对model进行优化",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		err := factory.HandleModel(v)
		if err != nil {
			log.Log.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(ifaceCmd)

	ifaceCmd.Flags().StringP("name", "n", "", "接口名称")

	ifaceCmd.Flags().StringP("out", "o", "", "输出文件")


	v.BindPFlags(ifaceCmd.Flags())
}
