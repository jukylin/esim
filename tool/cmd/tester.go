package cmd

import (
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/tool/tester"
	"github.com/spf13/cobra"
)

var testerCmd = &cobra.Command{
	Use:   "test",
	Short: "自动运行go test 命令",
	Long: `
监听文件修改，捕获事件并在当前项目下执行go test命令
`,
	Run: func(cmd *cobra.Command, args []string) {
		watcher := tester.NewFsnotifyWatcher(
			tester.WithFwLogger(logger),
		)

		execer := pkg.NewCmdExec(
			pkg.WithCmdExecLogger(logger),
		)

		ter := tester.NewTester(
			tester.WithTesterLogger(logger),
			tester.WithTesterWatcher(watcher),
			tester.WithTesterExec(execer),
		)

		ter.Run(v)
	},
}

func init() {
	rootCmd.AddCommand(testerCmd)

	testerCmd.Flags().BoolP(pkg.WireCmd, "w", true, "运行wire 命令")

	testerCmd.Flags().BoolP(pkg.MockeryCmd, "", true, "运行mockery 命令")

	testerCmd.Flags().BoolP(pkg.LintCmd, "", false, "运行 golangci-lint 命令")

	err := v.BindPFlags(testerCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}
