package cmd

import (
	"github.com/spf13/cobra"
	//"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/tool/factory"
)

var factoryCmd = &cobra.Command{
	Use:   "factory",
	Short: "初始化结构体",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		err := factory.NewEsimFactory().Run(v)
		if err != nil {
			log.Log.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(factoryCmd)

	factoryCmd.Flags().BoolP("sort", "s", true, "按照内存对齐排序")

	factoryCmd.Flags().BoolP("new", "n", false, "生成New方法")

	factoryCmd.Flags().BoolP("option", "o", false, "New with option")

	factoryCmd.Flags().BoolP("pool", "p", false, "生成临时对象池")

	factoryCmd.Flags().BoolP("gen_logger_option", "", false, "generate logger option")

	factoryCmd.Flags().BoolP("gen_conf_option", "", false, "generate conf option")

	factoryCmd.Flags().BoolP("star", "", false, "返回指针变量")

	factoryCmd.Flags().BoolP("print", "", false, "print the result")

	//modelCmd.Flags().BoolP("Print", "", false, "打印到终端")

	factoryCmd.Flags().StringP("sname", "", "", "结构体名称")

	factoryCmd.Flags().StringP("sdir", "", "", "结构体路径")

	factoryCmd.Flags().BoolP("plural", "", false, "支持复数")

	factoryCmd.Flags().StringP("imp_iface", "", "", "实现了接口")

	v.BindPFlags(factoryCmd.Flags())
}
