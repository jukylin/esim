package cmd

import (
	"github.com/spf13/cobra"
	//"log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jukylin/esim/tool/protoc"
)

var protocCmd = &cobra.Command{
	Use:   "protoc",
	Short: "将数据库表结构生成struct",
	Long: `1：在执行前需要注意，先把proto 文件 复制到项目下，
2：需要在项目根目录下执行
生成的protobuf文件会被放到项目的 internal/infra/third_party/package/*.pb.go,
`,
	Run: func(cmd *cobra.Command, args []string) {
		v.Set("debug", false)
		protoc.Gen(v)
	},
}

func init() {
	rootCmd.AddCommand(protocCmd)

	protocCmd.Flags().StringP("from_proto", "f", "", "proto 文件")

	protocCmd.Flags().StringP("target", "t", "internal/infra/third_party/protobuf", "生成的源码保存路径")

	protocCmd.Flags().BoolP("mock", "m", false, "生成 mock 文件")

	protocCmd.Flags().StringP("package", "p", "", "package名称")

	v.BindPFlags(protocCmd.Flags())
}
