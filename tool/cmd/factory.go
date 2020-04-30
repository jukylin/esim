package cmd

import (
	"github.com/spf13/cobra"
	//"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/tool/factory"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
)

var factoryCmd = &cobra.Command{
	Use:   "factory",
	Short: "初始化结构体",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger()
		esimFactory := factory.NewEsimFactory(
			factory.WithEsimFactoryLogger(logger),
			factory.WithEsimFactoryWriter(file_dir.NewEsimWriter()),
			factory.WithEsimFactoryTpl(templates.NewTextTpl()),
		)
		err := esimFactory.Run(v)
		if err != nil {
			log.Log.Error(err.Error())
		}
		esimFactory.Close()
	},
}

func init() {
	rootCmd.AddCommand(factoryCmd)

	factoryCmd.Flags().BoolP("sort", "s", true, "sort the field")

	factoryCmd.Flags().BoolP("new", "n", false, "with new")

	factoryCmd.Flags().BoolP("option", "o", false, "New with option")

	factoryCmd.Flags().BoolP("pool", "p", false, "with pool")

	factoryCmd.Flags().BoolP("ol", "", false, "generate logger option")

	factoryCmd.Flags().BoolP("oc", "", false, "generate conf option")

	factoryCmd.Flags().BoolP("star", "", false, "with star")

	factoryCmd.Flags().BoolP("print", "", false, "print to terminal")

	factoryCmd.Flags().StringP("sname", "", "", "struct name")

	factoryCmd.Flags().StringP("sdir", "", "", "struct path")

	factoryCmd.Flags().BoolP("plural", "", false, "with plural")

	factoryCmd.Flags().StringP("imp_iface", "", "", "implement the interface")

	v.BindPFlags(factoryCmd.Flags())
}
