package cmd

import (
	"github.com/spf13/cobra"
	//"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jukylin/esim/tool/db2entity"
	"os"
)

var db2entityCmd = &cobra.Command{
	Use:   "db2entity",
	Short: "将数据库表结构生成实体",
	Long: `1：需要在项目根目录下执行
生成的实体文件会被放到项目的 domain/entity/table.go,
CRUD文件被放到 infra/dao/table.go下
`,
	Run: func(cmd *cobra.Command, args []string) {
		v.Set("debug", true)
		db2entity.GenEntity(v)
	},
}

func init() {
	rootCmd.AddCommand(db2entityCmd)

	db2entityCmd.Flags().StringP("host", "H", os.Getenv("ESIM_HOST"), "Host to check mariadb status of")

	db2entityCmd.Flags().StringP("port", "P", os.Getenv("ESIM_PORT"), "Specify a port to connect to")

	db2entityCmd.Flags().StringP("table", "t", "", "Table to build struct from")

	db2entityCmd.Flags().StringP("database", "d", "", "Database to for connection")

	db2entityCmd.Flags().StringP("user", "u", os.Getenv("ESIM_USER"), "user to connect to database")

	db2entityCmd.Flags().StringP("password", "p", os.Getenv("ESIM_PASSWORD"), "password to connect to database")

	db2entityCmd.Flags().StringP("package", "", "", "name to set for package")

	db2entityCmd.Flags().StringP("struct", "s", "", "name to set for struct")

	db2entityCmd.Flags().BoolP("json", "j", false, "Add json annotations (default) Disable json annotations")

	db2entityCmd.Flags().BoolP("gorm", "g", true, "Add gorm annotations (tags)")

	db2entityCmd.Flags().StringP("target", "", "", "Save file path")

	db2entityCmd.Flags().BoolP("guregu", "", false, "Add guregu null types")

	db2entityCmd.Flags().BoolP("valid", "", false, "Add valid annotations")

	db2entityCmd.Flags().BoolP("mod", "", false, "Add mod annotations")

	db2entityCmd.Flags().BoolP("mar", "", true, "Marshaler to json")

	db2entityCmd.Flags().StringP("etar", "", "internal/domain/entity", "Save entity file path")

	db2entityCmd.Flags().BoolP("disetar", "", false, "disable Save model")

	db2entityCmd.Flags().StringP("daotar", "", "internal/infra/dao", "Save dao file path")

	db2entityCmd.Flags().BoolP("disdaotar", "", false, "disable Save dao")

	db2entityCmd.Flags().StringP("repotar", "", "internal/infra/repo", "Save dao file path")

	db2entityCmd.Flags().BoolP("disrepotar", "", false, "disable Save repo")

	db2entityCmd.Flags().BoolP("inject", "i", true, "automatic inject")

	db2entityCmd.Flags().StringP("injtar", "", "infra", "inject target, target must in wire dir")

	db2entityCmd.Flags().BoolP("hasdata", "", true, "check hasdata after find")

	v.BindPFlags(db2entityCmd.Flags())
}
