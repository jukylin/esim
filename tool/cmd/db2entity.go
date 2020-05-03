package cmd

import (
	"github.com/spf13/cobra"
	//"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jukylin/esim/tool/db2entity"
	"os"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/tool/db2entity/domain-file"
	"github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/jukylin/esim/infra"
)

var db2entityCmd = &cobra.Command{
	Use:   "db2entity",
	Short: "table's fields to entity",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger()

		dbConf := domain_file.NewDbConfig()
		dbConf.ParseConfig(v, logger)

		//select table's columns from db
		columnsInter := domain_file.NewDBColumnsInter(logger)

		tpl := templates.NewTextTpl()

		daoDomainFile := domain_file.NewDaoDomainFile(
			domain_file.WithDaoDomainFileLogger(logger),
			domain_file.WithDaoDomainFileTpl(tpl),
		)

		entityDomainFile := domain_file.NewEntityDomainFile(
			domain_file.WithEntityDomainFileLogger(logger),
			domain_file.WithEntityDomainFileTpl(tpl),
		)

		repoDomainFile := domain_file.NewRepoDomainFile(
			domain_file.WithRepoDomainFileLogger(logger),
			domain_file.WithRepoDomainFileTpl(tpl),
		)

		writer := file_dir.NewEsimWriter()

		db2EntityOptions := db2entity.Db2EnOptions{}
		db2entity.NewDb2Entity(
			db2EntityOptions.WithLogger(logger),
			db2EntityOptions.WithDbConf(dbConf),
			db2EntityOptions.WithColumnsInter(columnsInter),
			db2EntityOptions.WithWriter(writer),
			db2EntityOptions.WithExecer(pkg.NewCmdExec()),
			db2EntityOptions.WithDomainFile(daoDomainFile, entityDomainFile, repoDomainFile),
			db2EntityOptions.WithInfraer(infra.NewInfraer(
				infra.WithIfacerInfraInfo(infra.NewInfraInfo()),
				infra.WithIfacerLogger(logger),
				infra.WithIfacerWriter(writer),
				infra.WithIfacerExecer(pkg.NewCmdExec()),
			)),
		).Run(v)
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

	db2entityCmd.Flags().StringP("entity_target", "", "", "Save entity file path")

	db2entityCmd.Flags().BoolP("disable_entity", "", false, "Disable Save entity")

	db2entityCmd.Flags().StringP("dao_target", "", "internal/infra/dao", "Save dao file path")

	db2entityCmd.Flags().BoolP("disable_dao", "", false, "Disable Save dao")

	db2entityCmd.Flags().StringP("repo_target", "", "internal/infra/repo", "Save dao file path")

	db2entityCmd.Flags().BoolP("disable_repo", "", false, "Disable Save repo")

	v.BindPFlags(db2entityCmd.Flags())
}
