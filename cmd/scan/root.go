package scan

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use: "scan",
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Help()
		return nil
	},
}

func RootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
