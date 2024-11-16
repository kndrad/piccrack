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
	"path/filepath"

	api "github.com/kndrad/itcrack/internal/api/v1"
	"github.com/spf13/cobra"
)

var apiCheckHealthCmd = &cobra.Command{
	Use:   "checkhealth",
	Short: "Checks health of api http server",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := api.LoadConfig(".env")
		if err != nil {
			Logger.Error("Failed to load config", "err", err)

			return fmt.Errorf("loading config err: %w", err)
		}
		client := api.NewClient(config, Logger, api.WithTimeout(DefaultTimeout))
		defer client.Close()

		url := config.BaseURL() + string(filepath.Separator) + "health"

		if err := client.Get(context.Background(), url); err != nil {
			Logger.Error("Failed to check http server health", "err", err)

			return fmt.Errorf("checking health err: %w", err)
		}

		Logger.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	apiCmd.AddCommand(apiCheckHealthCmd)
}
