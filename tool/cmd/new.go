package cmd

import (
	"github.com/spf13/cobra"
	"github.com/jukylin/esim/tool/new"
	"github.com/jukylin/esim/log"
)

// grpcCmd represents the grpc command
var newCmd = &cobra.Command{
	Use:   "new",
	Short: "create a new project",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		loggerOptions := log.LoggerOptions{}
		log := log.NewLogger(loggerOptions.WithDebug(true))
		err := new.Build(v, log)
		if err != nil {
			log.Fatalf(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(newCmd)

	newCmd.Flags().BoolP("beego", "", false, "init beego server")

	newCmd.Flags().BoolP("gin", "", true, "init gin server")

	newCmd.Flags().BoolP("grpc", "", false, "init grpc server")

	newCmd.Flags().BoolP("monitoring", "m", true, "enable monitoring")

	newCmd.Flags().StringP("service_name", "s", "", "service name")

	v.BindPFlags(newCmd.Flags())
}
