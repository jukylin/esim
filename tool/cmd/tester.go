package cmd

import (
	"github.com/spf13/cobra"
	"github.com/jukylin/esim/tool/tester"
	"github.com/jukylin/esim/log"
	"github.com/jukylin/esim/pkg"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Automatic execution go test",
	Long: `Watching for files change, If there is a change,
run the go test in the current directory
`,
	Run: func(cmd *cobra.Command, args []string) {
		logger := log.NewLogger()

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
	rootCmd.AddCommand(testCmd)

	err := v.BindPFlags(testCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}
