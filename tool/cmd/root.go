package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/jukylin/esim/log"
)

var cfgFile string
var v = viper.New()
var logger = log.NewLogger()


// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "esim",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Errorf(err.Error())
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.esim.yaml)")

	rootCmd.PersistentFlags().BoolP("inject", "i", true, "Automatic inject instance to infra")

	rootCmd.PersistentFlags().StringP("infra_dir", "", "internal/infra/", "Infra dir")

	rootCmd.PersistentFlags().StringP("infra_file", "", "infra.go", "Infra file name")

	rootCmd.PersistentFlags().BoolP("star", "", false, "With star")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	err := v.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		logger.Errorf(err.Error())
	}

	err = v.BindPFlags(rootCmd.Flags())
	if err != nil {
		logger.Errorf(err.Error())
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			logger.Errorf(err.Error())
			os.Exit(1)
		}

		// Search config in home directory with name ".esim" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".esim")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		logger.Errorf("Using config file: %s", viper.ConfigFileUsed())
	}
}
