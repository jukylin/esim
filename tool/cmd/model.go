package cmd

import (
	"github.com/spf13/cobra"
	//"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jukylin/esim/tool/model"
	"github.com/jukylin/esim/log"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "对model进行优化",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		err := model.HandleModel(v)
		if err != nil {
			log.Log.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(modelCmd)

	modelCmd.Flags().BoolP("sort", "s", true, "按照内存对齐排序")

	modelCmd.Flags().BoolP("new", "n", false, "生成New方法")

	modelCmd.Flags().BoolP("star", "", false, "返回指针变量")

	modelCmd.Flags().BoolP("option", "o", false, "New with option")

	modelCmd.Flags().BoolP("gen_logger_option", "", false, "generate logger option")

	modelCmd.Flags().BoolP("gen_conf_option", "", false, "generate conf option")

	modelCmd.Flags().BoolP("pool", "p", false, "生成临时对象池")

	//modelCmd.Flags().BoolP("print", "", false, "扩展print方法")

	//modelCmd.Flags().BoolP("Print", "", false, "打印到终端")

	modelCmd.Flags().StringP("modelname", "m", "", "模型的名称")

	modelCmd.Flags().BoolP("plural", "", false, "支持复数")

	v.BindPFlags(modelCmd.Flags())
}
