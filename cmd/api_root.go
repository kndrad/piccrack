package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "API root command.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(apiCmd)

	apiCmd.PersistentFlags().String("host", "localhost", "http server host")
	viper.BindPFlag("host", apiCmd.Flags().Lookup("host"))

	apiCmd.PersistentFlags().String("port", "8080", "http server port")
	viper.BindPFlag("port", apiCmd.Flags().Lookup("port"))
}
