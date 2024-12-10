package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/kndrad/wcrack/cmd/logger"
	"github.com/kndrad/wcrack/config"
	"github.com/spf13/cobra"
)

var apiHealthzCmd = &cobra.Command{
	Use:   "healthz",
	Short: "Checks health http API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(Verbose)

		cfg, err := config.Load(os.Getenv("CONFIG_PATH"))
		if err != nil {
			l.Error("Failed to load config", "err", err)

			return fmt.Errorf("loading config err: %w", err)
		}
		url := net.JoinHostPort(cfg.HTTP.Host, cfg.HTTP.Port) + "/api/v1/healthz"
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

		l.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	apiCmd.AddCommand(apiHealthzCmd)
}
