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

	"github.com/kndrad/wcrack/internal/textproc"
	"github.com/kndrad/wcrack/pkg/openf"
	"github.com/spf13/cobra"
)

var wordsFrequencyAnalyzeCmd = &cobra.Command{
	Use:     "analyze",
	Short:   "Analyze words frequency in .txt and write output to .json",
	Example: "wcrack words frequency analyze --path=./testdata/words.txt --out=./output",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := DefaultLogger(Verbose)

		path, err := cmd.Flags().GetString("path")
		if err != nil {
			logger.Error("Failed to read path string flag value", "err", err)

			return fmt.Errorf("get string: %w", err)
		}

		content, err := os.ReadFile(filepath.Clean(path))
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

		outPath := filepath.Clean(OutPath)
		// Join outPath, id and json extension to create new out file path with an extension.
		jsonPath := openf.Join(outPath, analysis.ID, "json")
		logger.Info("Opening file",
			slog.String("json_path", jsonPath),
		)
		flags := os.O_APPEND | openf.DefaultFlags

		jsonFile, err := openf.Open(jsonPath, flags, 0o600)
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
	wordsFrequencyCmd.AddCommand(wordsFrequencyAnalyzeCmd)

	wordsFrequencyAnalyzeCmd.Flags().String("path", "", "Path of txt input file")
	wordsFrequencyAnalyzeCmd.MarkFlagRequired("path")
	wordsFrequencyAnalyzeCmd.Flags().StringVarP(&OutPath, "out", "o", ".", "JSON file output path")
}
