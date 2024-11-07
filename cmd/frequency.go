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

var wordsFreqCmd = &cobra.Command{
	Use:   "frequency",
	Short: "Analyze words frequency",
	Long: `itcrack frequency - Analyze word frequency
	-f, --file     Input text file to analyze (optional)
	-o, --out      Output directory for analysis results
	-v, --verbose  Enable verbose logging (default: false)`,
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
			logger.Error("Scanning failed", "err", err)

			return fmt.Errorf("scanner: %w", err)
		}

		analysis, err := textproc.AnalyzeFrequency(words)
		if err != nil {
			logger.Error("Analyzing words frequency failed", "err", err)

			return fmt.Errorf("frequency analysis: %w", err)
		}

		name, err := analysis.Name()
		if err != nil {
			logger.Error("Failed to get analysis name", "err", err)

			return fmt.Errorf("analysis name: %w", err)
		}
		// Join outPath, name and json extension to create new out file path with an extension.
		jsonPath := openf.Join(outPath, name, "json")
		logger.Info("Opening file",
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
			logger.Error("Failed to marshal json analysis", "err", err)

			return fmt.Errorf("json marshal: %w", err)
		}
		logger.Info("Writing analysis to json file",
			slog.String("json_path", jsonPath),
		)
		if _, err := jsonFile.Write(data); err != nil {
			logger.Error("Failed to write json analysis", "err", err)

			return fmt.Errorf("json write: %w", err)
		}

		logger.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	wordsCmd.AddCommand(wordsFreqCmd)

	wordsFreqCmd.Flags().StringVarP(
		&InputPath, "file", "f", "", ".txt file path to analyze words frequency.",
	)
	if err := wordsFreqCmd.MarkFlagRequired("file"); err != nil {
		logger.Error("Marking flag required failed", "err", err.Error())
	}

	wordsFreqCmd.Flags().StringVarP(&OutputPath, "out", "o", DefaultOutputPath, "JSON file output path")
}
