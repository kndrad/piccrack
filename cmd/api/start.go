package api

import (
	"context"
	"fmt"
	"os"

	"github.com/kndrad/wcrack/config"
	apiv1 "github.com/kndrad/wcrack/internal/api/v1"
	"github.com/kndrad/wcrack/internal/database"
	"github.com/kndrad/wcrack/pkg/retry"
	"github.com/spf13/cobra"

	"github.com/kndrad/wcrack/cmd/logger"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts http API server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(Verbose)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
		if err != nil {
			l.Error("Loading database config", "err", err.Error())

			return fmt.Errorf("config load: %w", err)
		}

		fmt.Printf("CONFIG: %#v\n", cfg)

		pool, err := database.Pool(ctx, cfg.Database)
		if err != nil {
			l.Error("Loading database pool", "err", err.Error())

			return fmt.Errorf("database pool: %w", err)
		}
		defer pool.Close()

		if err := retry.Ping(ctx, pool, retry.MaxRetries); err != nil {
			l.Error("Pinging database", "err", err.Error())

			return fmt.Errorf("database ping: %w", err)
		}

		db, err := database.Connect(ctx, pool)
		if err != nil {
			l.Error("Connecting to database", "err", err.Error())

			return fmt.Errorf("database connection: %w", err)
		}
		defer db.Close(ctx)

		q := database.New(db)
		wordsService := apiv1.NewWordService(q, l)

		// Create server instance
		srv, err := apiv1.NewServer(
			cfg.HTTP,
			wordsService,
			l,
		)
		if err != nil {
			l.Error("Failed to init new http server", "err", err)

			return fmt.Errorf("new http server err: %w", err)
		}

		if err := srv.Start(ctx); err != nil {
			l.Error("Failed to listen and serve", "err", err)

			return fmt.Errorf("listen and serve err: %w", err)
		}
		l.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
