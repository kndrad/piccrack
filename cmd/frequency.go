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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/kndrad/itcrack/internal/textproc"
	"github.com/kndrad/itcrack/pkg/openf"
	"github.com/spf13/cobra"
)

// frequencyCmd represents the frequency command.
var frequencyCmd = &cobra.Command{
	Use:   "frequency",
	Short: "Analyze word frequency in a text file",
	Long: `itcrack frequency - Analyze word frequency in a text file
	-f, --file     Input text file to analyze (required)
	-o, --out      Output directory for analysis results (default: current directory)
	-v, --verbose  Enable verbose logging (default: true)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			txtPath = filepath.Clean(InputPath)
			outPath = filepath.Clean(OutputPath)
		)

		shutdown := OnShutdown()
		defer shutdown()

		content, err := os.ReadFile(txtPath)
		if err != nil {
			logger.Error("Failed to read txt file", "err", err)

			return fmt.Errorf("read file: %w", err)
		}
		scanner := bufio.NewScanner(bytes.NewReader(content))
		scanner.Split(bufio.ScanWords)

		words := make([]string, 0)
		for scanner.Scan() {
			word := scanner.Text()
			words = append(words, word)
		}
		if err := scanner.Err(); err != nil {
			logger.Error("scanning failed", "err", err)

			return fmt.Errorf("scanner: %w", err)
		}

		analysis, err := textproc.AnalyzeFrequency(words)
		if err != nil {
			logger.Error("Analyzing words frequency failed", "err", err)

			return fmt.Errorf("frequency analysis: %w", err)
		}

		// Join to create new out file path with an extension.
		name, err := analysis.Name()
		if err != nil {
			logger.Error("Failed to get analysis name", "err", err)

			return fmt.Errorf("analysis name: %w", err)
		}
		jsonPath := openf.Join(outPath, name, "json")
		logger.Info("opening file",
			slog.String("json_path", jsonPath),
		)
		flags := os.O_APPEND | openf.DefaultFlags
		jsonFile, err := openf.Cleaned(jsonPath, flags, 0o600)
		if err != nil {
			logger.Error("Failed to open cleaned json file", "err", err)

			return fmt.Errorf("open cleaned: %w", err)
		}
		defer jsonFile.Close()

		data, err := json.MarshalIndent(analysis, "", " ")
		if err != nil {
			logger.Error("marshalling json analysis", "err", err)

			return fmt.Errorf("json marshal: %w", err)
		}
		logger.Info("Writing analysis to json file",
			slog.String("json_path", jsonPath),
		)
		if _, err := jsonFile.Write(data); err != nil {
			logger.Error("failed to write json analysis", "err", err)

			return fmt.Errorf("json write: %w", err)
		}

		logger.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(frequencyCmd)

	frequencyCmd.Flags().StringVarP(
		&InputPath, "file", "f", "", ".txt file path to analyze words frequency.",
	)
	if err := frequencyCmd.MarkFlagRequired("file"); err != nil {
		logger.Error("Marking flag required failed", "err", err.Error())
	}

	frequencyCmd.Flags().StringVarP(&OutputPath, "out", "o", DefaultOutputPath, "JSON file output path")
}
