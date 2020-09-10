package cmd

import (
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/tool/factory"
	"github.com/spf13/cobra"
)

var factoryCmd = &cobra.Command{
	Use:   "factory",
	Short: "初始化结构体 NewXXX()",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		esimFactory := factory.NewEsimFactory(
			factory.WithEsimFactoryLogger(logger),
			factory.WithEsimFactoryWriter(filedir.NewEsimWriter()),
		)
		err := esimFactory.Run(v)
		if err != nil {
			logger.Errorf(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(factoryCmd)

	factoryCmd.Flags().BoolP("sort", "s", true, "字段按大小升序")

	factoryCmd.Flags().BoolP("new", "n", false, "生成NewXXX()")

	factoryCmd.Flags().BoolP("option", "o", false, "生成NewXXX(options ...xxxOption)")

	// factoryCmd.Flags().BoolP("pool", "p", false, "with pool")

	// factoryCmd.Flags().BoolP("ol", "", false, "generate logger option")

	// factoryCmd.Flags().BoolP("oc", "", false, "generate conf option")

	factoryCmd.Flags().BoolP("print", "", false, "在终端打印")

	factoryCmd.Flags().StringP("sname", "", "", "指定的结构体")

	factoryCmd.Flags().StringP("sdir", "", "", "结构体路径")

	// factoryCmd.Flags().BoolP("plural", "", false, "with plural")

	// factoryCmd.Flags().StringP("imp_iface", "", "", "implement the interface")

	err := v.BindPFlags(factoryCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}
