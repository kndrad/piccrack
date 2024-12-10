package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/kndrad/wcrack/config"
	"github.com/kndrad/wcrack/internal/database"
	"github.com/kndrad/wcrack/pkg/retry"
	"github.com/spf13/cobra"

	"github.com/kndrad/wcrack/cmd/logger"
)

// pingCmd represents the ping command.
var pingCmd = &cobra.Command{
	Use:     "ping",
	Short:   "Pings a database",
	Example: "wcrack ping",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(Verbose)

		l.Info("Loading database config.")

		cfg, err := config.Load("config/development.yaml")
		if err != nil {
			l.Error("Failed to load db config", "err", err.Error())

			return fmt.Errorf("db config: %w", err)
		}

		l.Info("Establishing connection to a database.")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		pool, err := database.Pool(ctx, cfg.Database)
		if err != nil {
			l.Error("Failed to get db pool", "err", err.Error())

			return fmt.Errorf("db pool: %w", err)
		}
		defer pool.Close()

		l.Info("Pinging database...")
		if err := retry.Ping(ctx, pool, retry.MaxRetries); err != nil {
			l.Error("Pinging db pool failed", "err", err.Error())

			return fmt.Errorf("db pool: %w", err)
		}
		l.Info("Pinging db success.")

		conn, err := database.Connect(ctx, pool)
		if err != nil {
			l.Error("Failed to connect to a database", "err", err.Error())

			return fmt.Errorf("db connection: %w", err)
		}
		defer conn.Close(ctx)

		l.Info("Database OK.")

		l.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
}
