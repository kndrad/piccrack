package api

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/kndrad/piccrack/cmd/logger"
	"github.com/kndrad/piccrack/config"
	"github.com/kndrad/piccrack/internal/database"
	"github.com/kndrad/piccrack/pkg/retry"
	"github.com/spf13/cobra"
)

var healthzCmd = &cobra.Command{
	Use:   "healthz",
	Short: "Checks health of http API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(Verbose)

		cfg, err := config.Load(cfgFile)
		if err != nil {
			l.Error("Failed to load config", "err", err)

			return fmt.Errorf("loading config err: %w", err)
		}
		url := "http://" + net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port) + "/api/v1/healthz"
		buf := new(bytes.Buffer)
		req, err := http.NewRequestWithContext(
			context.TODO(),
			http.MethodGet,
			url,
			buf,
		)
		if err != nil {
			l.Error("Failed to create request", "err", err)

			return fmt.Errorf("new request err: %w", err)
		}
		l.Info("Sending request",
			slog.String("url", url),
		)

		c := &http.Client{}
		defer c.CloseIdleConnections()

		// Ping http server
		resp, err := c.Do(req)
		if err != nil {
			l.Error("Failed to do request with a client", "err", err)

			return fmt.Errorf("client do request err: %w", err)
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:
			l.Info("Received response and server OK", "statusCode", resp.StatusCode)

			return nil
		case http.StatusNotFound:
			l.Info("Received not found", "statusCode", resp.StatusCode)
		default:
			l.Info("Received response", "statusCode", resp.StatusCode)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Ping db as well
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

		l.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthzCmd)
}
