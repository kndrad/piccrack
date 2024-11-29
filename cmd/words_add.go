/*
Copyright © 2024 Konrad Nowara

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/kndrad/wordcrack/internal/textproc"
	"github.com/kndrad/wordcrack/internal/textproc/database"
	"github.com/kndrad/wordcrack/pkg/retry"
	"github.com/spf13/cobra"
)

// addWordCmd represents the add command.
var addWordCmd = &cobra.Command{
	Use:     "add",
	Short:   "Add word to a database.",
	Example: "wordcrack words add [WORD]",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := textproc.LoadDatabaseConfig(DefaultEnvFilePath)
		if err != nil {
			Logger.Error("Loading database config", "err", err.Error())

			return fmt.Errorf("loading database config failed: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		pool, err := textproc.DatabasePool(ctx, *config)
		if err != nil {
			Logger.Error("Loading database pool", "err", err.Error())

			return fmt.Errorf("database pool: %w", err)
		}
		defer pool.Close()

		if err := retry.Ping(ctx, pool, retry.MaxRetries); err != nil {
			Logger.Error("Pinging database", "err", err.Error())

			return fmt.Errorf("database ping: %w", err)
		}

		conn, err := textproc.DatabaseConnection(ctx, pool)
		if err != nil {
			Logger.Error("Connecting to database", "err", err.Error())

			return fmt.Errorf("database connection: %w", err)
		}
		defer conn.Close(ctx)

		q := database.New(conn)

		value := args[0]
		word, err := q.CreateWord(ctx, value)
		if err != nil {
			Logger.Error("Inserting word failed", "err", err.Error())

			return fmt.Errorf("word insert: %w", err)
		}
		Logger.Info("Inserted word",
			slog.Int64("word", word.ID),
			slog.String("value", word.Value),
			slog.Time("created_at_time", word.CreatedAt.Time),
		)

		Logger.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	wordsCmd.AddCommand(addWordCmd)
}
