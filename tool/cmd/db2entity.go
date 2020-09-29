package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/jukylin/esim/infra"
	"github.com/jukylin/esim/pkg"
	filedir "github.com/jukylin/esim/pkg/file-dir"
	"github.com/jukylin/esim/pkg/templates"
	"github.com/jukylin/esim/tool/db2entity"
	domainfile "github.com/jukylin/esim/tool/db2entity/domain-file"
)

var db2entityCmd = &cobra.Command{
	Use:   "db2entity",
	Short: "将对应的表生成实体",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		dbConf := domainfile.NewDbConfig()
		dbConf.ParseConfig(v, logger)

		// select table's columns from db
		columnsInter := domainfile.NewDBColumnsInter(logger)

		tpl := templates.NewTextTpl()

		daoDomainFile := domainfile.NewDaoDomainFile(
			domainfile.WithDaoDomainFileLogger(logger),
			domainfile.WithDaoDomainFileTpl(tpl),
		)

		entityDomainFile := domainfile.NewEntityDomainFile(
			domainfile.WithEntityDomainFileLogger(logger),
			domainfile.WithEntityDomainFileTpl(tpl),
		)

		repoDomainFile := domainfile.NewRepoDomainFile(
			domainfile.WithRepoDomainFileLogger(logger),
			domainfile.WithRepoDomainFileTpl(tpl),
		)

		writer := filedir.NewEsimWriter()

		shareInfo := domainfile.NewShareInfo()
		shareInfo.DbConf = dbConf

		db2EntityOptions := db2entity.Db2EnOptions{}
		db2Entity := db2entity.NewDb2Entity(
			db2EntityOptions.WithLogger(logger),
			db2EntityOptions.WithDbConf(dbConf),
			db2EntityOptions.WithColumnsInter(columnsInter),
			db2EntityOptions.WithWriter(writer),
			db2EntityOptions.WithShareInfo(shareInfo),

			db2EntityOptions.WithExecer(pkg.NewCmdExec()),
			db2EntityOptions.WithDomainFile(entityDomainFile, daoDomainFile, repoDomainFile),
			db2EntityOptions.WithInfraer(infra.NewInfraer(
				infra.WithIfacerLogger(logger),
				infra.WithIfacerWriter(writer),
				infra.WithIfacerExecer(pkg.NewCmdExec()),
			)),
		)

		err := db2Entity.Run(v)
		if err != nil {
			logger.Errorf(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(db2entityCmd)

	db2entityCmd.Flags().StringP("host", "H", os.Getenv("ESIM_DB_HOST"), "数据库 host")

	db2entityCmd.Flags().StringP("port", "P", os.Getenv("ESIM_DB_PORT"), "数据库 port")

	db2entityCmd.Flags().StringP("table", "t", "", "数据表表")

	db2entityCmd.Flags().StringP("database", "d", "", "数据库名")

	db2entityCmd.Flags().StringP("user", "u", os.Getenv("ESIM_DB_USER"), "链接数据库的用户")

	db2entityCmd.Flags().StringP("password", "p", os.Getenv("ESIM_DB_PASSWORD"), "链接数据库的密码")

	db2entityCmd.Flags().StringP("boubctx", "b", "", "用于界限上下文")

	db2entityCmd.Flags().StringP("package", "", "", "包名，空使用数据表名")

	db2entityCmd.Flags().StringP("struct", "s", "", "实体的名称，空使用数据表名")

	db2entityCmd.Flags().BoolP("gorm", "g", true, "增加grom注解")

	db2entityCmd.Flags().StringP("entity_target", "", "", "实体保存的路径")

	db2entityCmd.Flags().BoolP("disable_entity", "", false, "不生成实体")

	db2entityCmd.Flags().StringP("dao_target", "", "internal/infra/dao", "DAO存放路径")

	db2entityCmd.Flags().BoolP("disable_dao", "", false, "关闭DAO的生成")

	db2entityCmd.Flags().StringP("repo_target", "", "internal/infra/repo", "repo的存放路径")

	db2entityCmd.Flags().BoolP("disable_repo", "", false, "关闭repo的生成")

	err := v.BindPFlags(db2entityCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}
