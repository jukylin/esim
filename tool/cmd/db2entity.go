package cmd

import (
	"github.com/spf13/cobra"
	//"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jukylin/esim/tool/db2entity"
	"os"
	"github.com/jukylin/esim/log"
)

var db2entityCmd = &cobra.Command{
	Use:   "db2entity",
	Short: "将数据库表结构生成实体",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger()
		db2EntityOptions := db2entity.Db2EntityOptions{}
		db2entity.NewDb2Entity(db2EntityOptions.WithLogger(logger)).Run(v)
	},
}

func init() {
	rootCmd.AddCommand(db2entityCmd)

	db2entityCmd.Flags().StringP("host", "H", os.Getenv("ESIM_DB_HOST"), "Specify a host to connect to")

	db2entityCmd.Flags().StringP("port", "P", os.Getenv("ESIM_DB_PORT"), "Specify a port to connect to")

	db2entityCmd.Flags().StringP("table", "t", "", "Database's table")

	db2entityCmd.Flags().StringP("database", "d", "", "Database to for connection")

	db2entityCmd.Flags().StringP("user", "u", os.Getenv("ESIM_DB_USER"), "User to connect to database")

	db2entityCmd.Flags().StringP("password", "p", os.Getenv("ESIM_DB_PASSWORD"), "Password to connect to database")

	db2entityCmd.Flags().StringP("boubctx", "b", "", "Name to set for bounded context")

	db2entityCmd.Flags().StringP("package", "", "", "Name to set for package")

	db2entityCmd.Flags().StringP("struct", "s", "", "Name to set for struct")

	db2entityCmd.Flags().BoolP("gorm", "g", true, "Add gorm annotations (tags)")

	db2entityCmd.Flags().StringP("target", "", "", "Save file path")

	db2entityCmd.Flags().StringP("entity_target", "", "", "Save entity file path")

	db2entityCmd.Flags().BoolP("disable_entity", "", false, "Disabled Save model")

	db2entityCmd.Flags().StringP("dao_target", "", "internal/infra/dao", "Save dao file path")

	db2entityCmd.Flags().BoolP("disable_dao", "", false, "Disabled Save dao")

	db2entityCmd.Flags().StringP("repo_target", "", "internal/infra/repo", "Save dao file path")

	db2entityCmd.Flags().BoolP("disable_repo", "", false, "Disabled Save repo")

	db2entityCmd.Flags().BoolP("inject", "i", true, "Automatic inject")

	db2entityCmd.Flags().StringP("infra_dir", "", "internal/infra/", "Infra dir")

	db2entityCmd.Flags().StringP("infra_file", "", "infra.go", "Infra file name")

	v.BindPFlags(db2entityCmd.Flags())
}
