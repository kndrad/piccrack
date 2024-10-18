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

var (
	ScreenshotPath string
	OutFilePath    string
	Save           bool
	verbose        bool
)

// wordsCmd represents the words command.
var wordsCmd = &cobra.Command{
	Use:   "words",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("words called.")

		var (
			screenshotPath = filepath.Clean(ScreenshotPath)
			outFilePath    = filepath.Clean(OutFilePath)
		)

		shutdown := Shutdown()
		defer shutdown()

		var screenshotFiles []string

		stat, err := os.Stat(screenshotPath)
		if err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("wordsCmd: %w", err)
		}
		if stat.IsDir() {
			logger.Info("wordsCmd: processing directory", "file", screenshotPath)

			entries, err := os.ReadDir(filepath.Clean(screenshotPath))
			if err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
			// Append image files only
			for _, e := range entries {
				if !e.IsDir() && screenshot.IsImageFile(e.Name()) {
					screenshotFiles = append(screenshotFiles, filepath.Join(screenshotPath, "/", e.Name()))
				}
			}
			logger.Info("wordsCmd: number of image files in a directory", "len(files)", len(screenshotFiles))
		} else {
			screenshotFiles = append(screenshotFiles, screenshotPath)
		}

		out, err := OpenCleanFile(outFilePath, os.O_APPEND|DefaultFlag, DefaultPerm)
		if err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("wordsCmd: %w", err)
		}

		// Process each screenshot and write an out file
		for _, name := range screenshotFiles {
			content, err := os.ReadFile(name)
			if err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
			words, err := screenshot.RecognizeWords(content)
			if err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
			if err := screenshot.WriteWords(words, screenshot.NewWordsTextFileWriter(out)); err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
		}

		return nil
	},
}

// frequencyCmd represents the frequency command.
var frequencyCmd = &cobra.Command{
	Use:   "frequency",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("frequency called")
		ScreenshotPath = filepath.Clean(ScreenshotPath)
		if verbose {
			logger.Info("frequencyCmd", "filename", ScreenshotPath)
		}

		shutdown := Shutdown()
		defer shutdown()

		content, err := os.ReadFile(ScreenshotPath)
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd err: %w", err)
		}

		// Scan each word
		scanner := bufio.NewScanner(bytes.NewReader(content))
		scanner.Split(bufio.ScanWords)

		// Filter out non existing words?
		// Or try to adjust the word to the nearest possible?
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

		analysis, err := screenshot.AnalyzeWordFrequency(words)
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
		outPath := Join(OutFilePath, name, "json")
		logger.Info("frequencyCmd opening file", "jsonPath", outPath)

		// Open text_analysis JSON file and write analysis
		outFile, err := OpenCleanFile(outPath, os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}
		defer outFile.Close()

		// Write JSON analysis
		jsonAnalysis, err := json.MarshalIndent(analysis, "", " ")
		if err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}
		logger.Info("frequencyCmd writing analysisJson")
		if _, err := outFile.Write(jsonAnalysis); err != nil {
			logger.Error("frequencyCmd", "err", err)

			return fmt.Errorf("frequencyCmd: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(wordsCmd)

	wordsCmd.PersistentFlags().StringVarP(&ScreenshotPath, "file", "f", "", "Screenshot file to recognize words from")
	if err := wordsCmd.MarkPersistentFlagRequired("file"); err != nil {
		logger.Error("rootcmd", "err", err.Error())
	}
	wordsCmd.Flags().BoolVarP(&Save, "save", "s", false, "Save the output")
	wordsCmd.Flags().StringVarP(&OutFilePath, "out", "o", "", "Output path")
	wordsCmd.MarkFlagsRequiredTogether("save", "out")
	wordsCmd.Flags().BoolVarP(&verbose, "verbose", "v", true, "Verbose")

	rootCmd.AddCommand(frequencyCmd)
	frequencyCmd.PersistentFlags().StringVarP(
		&ScreenshotPath, "file", "f", "", "File to analyze words output frequency from",
	)
	if err := frequencyCmd.MarkPersistentFlagRequired("file"); err != nil {
		logger.Error("frequencyCmd", "err", err.Error())
	}

	frequencyCmd.Flags().StringVarP(&OutFilePath, "out", "o", ".", "Output path")
	frequencyCmd.Flags().BoolVarP(&verbose, "verbose", "v", true, "Verbose")
}
