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

	v1 "github.com/kndrad/itcrack/internal/api/v1"
	"github.com/spf13/cobra"
)

var apiServeHTTPCmd = &cobra.Command{
	Use:   "servehttp",
	Short: "Start http server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := v1.LoadConfig(".env")
		if err != nil {
			Logger.Error("Failed to load config", "err", err.Error())

			return fmt.Errorf("loading config err: %w", err)
		}

		srv, err := v1.NewHTTPServer(config, Logger)
		if err != nil {
			Logger.Error("Failed to init new http server", "err", err)

			return fmt.Errorf("new http server err: %w", err)
		}

		if err := srv.Start(context.TODO()); err != nil {
			Logger.Error("Failed to listen and serve", "err", err)

			return fmt.Errorf("listen and serve err: %w", err)
		}
		Logger.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	apiCmd.AddCommand(apiServeHTTPCmd)
}
