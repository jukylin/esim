package cmd

import (
	"github.com/jukylin/esim/pkg"
	"github.com/jukylin/esim/tool/tester"
	"github.com/spf13/cobra"
)

var testerCmd = &cobra.Command{
	Use:   "test",
	Short: "Automatic execution go test",
	Long: `Watching for files change, If there is a change,
run the go test in the current directory
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

	testerCmd.Flags().BoolP("wire", "w", true, "run with wire")

	err := v.BindPFlags(testerCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}
