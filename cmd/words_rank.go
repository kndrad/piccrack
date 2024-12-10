package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/kndrad/wcrack/cmd/logger"
	"github.com/kndrad/wcrack/config"
	"github.com/kndrad/wcrack/internal/database"
	"github.com/kndrad/wcrack/pkg/retry"
	"github.com/spf13/cobra"
)

var rankCmd = &cobra.Command{
	Use:   "rank",
	Short: "Displays ranking of words from a database.",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(Verbose)

		cfg, err := config.Load("config/development.yaml")
		if err != nil {
			l.Error("Loading database config", "err", err.Error())

			return fmt.Errorf("loading config: %w", err)
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

		q := database.New(conn)

		var limit int32 = 30
		params := database.ListWordRankingsParams{
			Limit: limit,
		}

		if len(args) > 0 {
			limit, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil {
				l.Error("Failed to strconv", "err", err.Error())
			}
			params.Limit = int32(limit)
		}

		rows, err := q.ListWordRankings(ctx, params)
		if err != nil {
			l.Error("Failed to get words rank", "err", err.Error())

			return fmt.Errorf("words rank err: %w", err)
		}

		for _, row := range rows {
			fmt.Printf("WORD: %s | RANK: %d\n", row.Value, row.Ranking)
		}

		l.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	wordsCmd.AddCommand(rankCmd)
}
