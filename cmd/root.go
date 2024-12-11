package cmd

import (
	"fmt"
	"os"

	"github.com/kndrad/wcrack/cmd/api"
	"github.com/kndrad/wcrack/cmd/scan"
	"github.com/kndrad/wcrack/cmd/words"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	DefaultEnvFilePath = ".env"
	DefaultOutputPath  = "./output"
)

var (
	OutPath string
	Verbose bool
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "wcrack",
	Short: "Analyze text from screenshots and performing word frequency analysis.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cmd.Help(); err != nil {
			return fmt.Errorf("help display: %w", err)
		}

		return nil
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wcrack.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "print verbose actions")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	rootCmd.AddCommand(api.RootCmd())
	rootCmd.AddCommand(scan.RootCmd())
	rootCmd.AddCommand(words.RootCmd())
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
