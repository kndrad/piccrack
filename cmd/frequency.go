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
	"os"
	"path/filepath"

	"github.com/kndrad/itcrack/internal/screenshot"
	"github.com/spf13/cobra"
)

// frequencyCmd represents the frequency command.
var frequencyCmd = &cobra.Command{
	Use:   "frequency",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			textFilePath = filepath.Clean(TextFilePath)
			outPath      = filepath.Clean(OutPath)
		)
		if verbose {
			logger.Info("frequencyCmd", "filename", textFilePath)
		}

		shutdown := OnShutdown()
		defer shutdown()

		content, err := os.ReadFile(textFilePath)
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd err: %w", err)
		}
		scanner := bufio.NewScanner(bytes.NewReader(content))
		scanner.Split(bufio.ScanWords)

		words := make([]string, 0)
		words = append(words, "test")

		for scanner.Scan() {
			word := scanner.Text()
			words = append(words, word)
		}
		if err := scanner.Err(); err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}

		analysis, err := screenshot.AnalyzeFrequency(words)
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}

		// Join to create new out file path with an extension.
		name, err := analysis.Name()
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}
		jsonPath := JoinPaths(outPath, name, "json")
		logger.Info("frequencyCmd opening file", "jsonPath", jsonPath)
		jsonFile, err := OpenCleanFile(jsonPath, os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}
		defer jsonFile.Close()

		jsonAnalysis, err := json.MarshalIndent(analysis, "", " ")
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}
		logger.Info("frequencyCmd writing analysisJson")
		if _, err := jsonFile.Write(jsonAnalysis); err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(frequencyCmd)

	frequencyCmd.PersistentFlags().StringVarP(
		&TextFilePath, "file", "f", "", "File to analyze words output frequency from",
	)
	if err := frequencyCmd.MarkPersistentFlagRequired("file"); err != nil {
		logger.Error("frequencyCmd", "err", err.Error())
	}

	frequencyCmd.Flags().StringVarP(&OutPath, "out", "o", ".", "Output path")
	frequencyCmd.Flags().BoolVarP(&verbose, "verbose", "v", true, "Verbose")
}
