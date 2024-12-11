package api

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "api",
	Short: "API root command.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var Verbose bool

var cfgFile string

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "./config/development.yaml", "config file path")

	rootCmd.PersistentFlags().String("host", "localhost", "http server host")
	viper.BindPFlag("host", rootCmd.Flags().Lookup("host"))

	rootCmd.PersistentFlags().String("port", "8080", "http server port")
	viper.BindPFlag("port", rootCmd.Flags().Lookup("port"))

	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "print verbose actions")
}

func RootCmd() *cobra.Command {
	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".wcrack")

		// Search .env file in project dir
		wd, err := os.Getwd()
		cobra.CheckErr(err)
		viper.AddConfigPath(wd)
		viper.SetConfigType("env")
		viper.SetConfigName("")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
