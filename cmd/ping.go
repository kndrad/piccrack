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
	"time"

	"github.com/kndrad/wcrack/config"
	"github.com/kndrad/wcrack/internal/textproc"
	"github.com/kndrad/wcrack/pkg/retry"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command.
var pingCmd = &cobra.Command{
	Use:     "ping",
	Short:   "Pings a database",
	Example: "wcrack ping",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := DefaultLogger(Verbose)

		logger.Info("Loading database config.")

		cfg, err := config.Load("config/development.yaml")
		if err != nil {
			logger.Error("Failed to load db config", "err", err.Error())

			return fmt.Errorf("db config: %w", err)
		}

		logger.Info("Establishing connection to a database.")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		pool, err := textproc.DatabasePool(ctx, cfg.Database)
		if err != nil {
			logger.Error("Failed to get db pool", "err", err.Error())

			return fmt.Errorf("db pool: %w", err)
		}
		defer pool.Close()

		logger.Info("Pinging database...")
		if err := retry.Ping(ctx, pool, retry.MaxRetries); err != nil {
			logger.Error("Pinging db pool failed", "err", err.Error())

			return fmt.Errorf("db pool: %w", err)
		}
		logger.Info("Pinging db success.")

		conn, err := textproc.DatabaseConnection(ctx, pool)
		if err != nil {
			logger.Error("Failed to connect to a database", "err", err.Error())

			return fmt.Errorf("db connection: %w", err)
		}
		defer conn.Close(ctx)

		logger.Info("Database OK.")

		logger.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(pingCmd)
}
