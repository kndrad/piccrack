package words

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/kndrad/wcrack/cmd/logger"
	"github.com/kndrad/wcrack/config"
	"github.com/kndrad/wcrack/internal/database"
	"github.com/kndrad/wcrack/pkg/retry"
	"github.com/spf13/cobra"
)

var wordsFrequencyCmd = &cobra.Command{
	Use:     "frequency",
	Short:   "Outputs words frequency from a database",
	Example: "wcrack words frequency",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(Verbose)

		cfg, err := config.Load("config/development.yaml")
		if err != nil {
			l.Error("Loading database config", "err", err.Error())

			return fmt.Errorf("config load: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

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

		conn, err := database.Connect(ctx, pool)
		if err != nil {
			l.Error("Connecting to database", "err", err.Error())

			return fmt.Errorf("database connection: %w", err)
		}
		defer conn.Close(ctx)

		// Query db to get word frequency count.
		q := database.New(conn)

		var limit int32 = 30
		params := database.ListWordFrequenciesParams{Limit: limit}

		if len(args) > 0 {
			limit, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				l.Error("Failed to strconv", "err", err.Error())
			}
			params.Limit = int32(limit)
		}
		rows, err := q.ListWordFrequencies(ctx, params)
		if err != nil {
			l.Error("Failed to analyze word frequency count", "err", err.Error())

			return fmt.Errorf("getting word frequency count: %w", err)
		}
		l.Info("Got word frequency count rows",
			slog.Int("len", len(rows)),
		)

		if Verbose {
			for i, row := range rows {
				fmt.Printf("%v: ROW: [%v, %v] \n", i, row.Value, row.Total)
			}
		}

		l.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(wordsFrequencyCmd)
}
