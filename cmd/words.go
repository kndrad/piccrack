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
	"fmt"
	"os"
	"path/filepath"

	"github.com/kndrad/itcrack/internal/screenshot"
	"github.com/spf13/cobra"
)

var (
	ScreenshotPath string
	TextFilePath   string
	OutPath        string
	Save           bool
	verbose        bool
)

// wordsCmd represents the words command.
var wordsCmd = &cobra.Command{
	Use:   "words",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			screenshotPath = filepath.Clean(ScreenshotPath)
			textFilePath   = filepath.Clean(TextFilePath)
		)

		shutdown := OnShutdown()
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

		textFile, err := OpenCleanFile(textFilePath, os.O_APPEND|DefaultFlag, DefaultPerm)
		if err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("wordsCmd: %w", err)
		}

		// Process each screenshot and write an out file
		for _, filename := range screenshotFiles {
			content, err := os.ReadFile(filename)
			if err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
			words, err := screenshot.RecognizeWords(content)
			if err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
			if err := screenshot.WriteWords(words, screenshot.NewWordsTextFileWriter(textFile)); err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("wordsCmd: %w", err)
			}
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
	wordsCmd.Flags().StringVarP(&TextFilePath, "out", "o", "", "Output path")
	wordsCmd.MarkFlagsRequiredTogether("save", "out")
	wordsCmd.Flags().BoolVarP(&verbose, "verbose", "v", true, "Verbose")
}
