package cmd

import (
	"github.com/jukylin/esim/tool/protoc"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Automatic execution go test",
	Long: `Watching for files change, If there is a change,
run the go test in the current directory
`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	err := v.BindPFlags(testCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}
