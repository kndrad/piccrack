package cmd

import (
	"context"
	"fmt"

	"github.com/kndrad/itcrack/internal/api"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var serveHTTPCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts an HTTP server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := api.LoadConfig(".env")
		if err != nil {
			logger.Error("Failed to load config", "err", err.Error())

			return fmt.Errorf("loading config err: %w", err)
		}
		srv := api.NewHTTPServer(config, logger)

		ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
		defer cancel()

		if err := api.StartServer(ctx, srv, logger); err != nil {
			logger.Error("Failed to listen and serve", "err", err.Error())

			return fmt.Errorf("listen and serve err: %w", err)
		}

		logger.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveHTTPCmd)

	serveHTTPCmd.Flags().String("host", "localhost", "http server host")
	serveHTTPCmd.MarkFlagRequired("host")
	viper.BindPFlag("host", serveHTTPCmd.Flags().Lookup("host"))

	serveHTTPCmd.Flags().String("port", "8080", "http server port")
	serveHTTPCmd.MarkFlagRequired("port")
	viper.BindPFlag("port", serveHTTPCmd.Flags().Lookup("port"))
}
