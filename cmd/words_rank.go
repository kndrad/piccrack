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
	"strconv"
	"time"

	"github.com/kndrad/wcrack/internal/textproc"
	"github.com/kndrad/wcrack/internal/textproc/database"
	"github.com/kndrad/wcrack/pkg/retry"
	"github.com/spf13/cobra"
)

var rankCmd = &cobra.Command{
	Use:   "rank",
	Short: "Displays ranking of words from a database.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := DefaultLogger(Verbose)

		config, err := textproc.LoadDatabaseConfig(DefaultEnvFilePath)
		if err != nil {
			logger.Error("Loading database config", "err", err.Error())

			return fmt.Errorf("loading config: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		pool, err := textproc.DatabasePool(ctx, *config)
		if err != nil {
			logger.Error("Loading database pool", "err", err.Error())

			return fmt.Errorf("database pool: %w", err)
		}
		defer pool.Close()

		if err := retry.Ping(ctx, pool, retry.MaxRetries); err != nil {
			logger.Error("Pinging database", "err", err.Error())

			return fmt.Errorf("database ping: %w", err)
		}

		conn, err := textproc.DatabaseConnection(ctx, pool)
		if err != nil {
			logger.Error("Connecting to database", "err", err.Error())

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
				logger.Error("Failed to strconv", "err", err.Error())
			}
			params.Limit = int32(limit)
		}

		rows, err := q.ListWordRankings(ctx, params)
		if err != nil {
			logger.Error("Failed to get words rank", "err", err.Error())

			return fmt.Errorf("words rank err: %w", err)
		}

		for _, row := range rows {
			fmt.Printf("WORD: %s | RANK: %d\n", row.Value, row.Ranking)
		}

		logger.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	wordsCmd.AddCommand(rankCmd)
}
