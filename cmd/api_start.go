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

	v1 "github.com/kndrad/wordcrack/internal/api/v1"
	"github.com/kndrad/wordcrack/internal/textproc"
	"github.com/kndrad/wordcrack/internal/textproc/database"
	"github.com/kndrad/wordcrack/pkg/retry"
	"github.com/spf13/cobra"
)

var apiStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts http API server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		dbconf, err := textproc.LoadDatabaseConfig(".env")
		if err != nil {
			Logger.Error("Failed to load database config", "err", err.Error())

			return fmt.Errorf("loading config err: %w", err)
		}

		pool, err := textproc.DatabasePool(ctx, *dbconf)
		if err != nil {
			Logger.Error("Loading database pool", "err", err.Error())

			return fmt.Errorf("database pool: %w", err)
		}
		defer pool.Close()

		if err := retry.Ping(ctx, pool, retry.MaxRetries); err != nil {
			Logger.Error("Pinging database", "err", err.Error())

			return fmt.Errorf("database ping: %w", err)
		}

		db, err := textproc.DatabaseConnection(ctx, pool)
		if err != nil {
			Logger.Error("Connecting to database", "err", err.Error())

			return fmt.Errorf("database connection: %w", err)
		}
		defer db.Close(ctx)

		config, err := v1.LoadConfig(".env")
		if err != nil {
			Logger.Error("Failed to load api config", "err", err.Error())

			return fmt.Errorf("loading config err: %w", err)
		}

		q := database.New(db)
		wordsService := v1.NewWordService(q, Logger)
		srv, err := v1.NewServer(
			config,
			wordsService,
			Logger,
		)
		if err != nil {
			Logger.Error("Failed to init new http server", "err", err)

			return fmt.Errorf("new http server err: %w", err)
		}

		if err := srv.Start(ctx); err != nil {
			Logger.Error("Failed to listen and serve", "err", err)

			return fmt.Errorf("listen and serve err: %w", err)
		}
		Logger.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	apiCmd.AddCommand(apiStartCmd)
}
