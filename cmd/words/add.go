package words

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kndrad/wcrack/cmd/logger"
	"github.com/kndrad/wcrack/config"
	"github.com/kndrad/wcrack/internal/database"
	"github.com/kndrad/wcrack/pkg/retry"
	"github.com/spf13/cobra"
)

// addCmd represents the add command.
var addCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add word to a database.",
	Example: "wcrack words add [WORD]",
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

		q := database.New(conn)

		value := args[0]
		word, err := q.CreateWord(ctx, value)
		if err != nil {
			l.Error("Inserting word failed", "err", err.Error())

			return fmt.Errorf("word insert: %w", err)
		}
		l.Info("Inserted word",
			slog.Int64("word", word.ID),
			slog.String("value", word.Value),
			slog.Time("created_at_time", word.CreatedAt.Time),
		)

		l.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
