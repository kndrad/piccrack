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
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/kndrad/itcrack/internal/screenshot"
	"github.com/spf13/cobra"
)

var logger *slog.Logger

var (
	filename string
	save     bool
	outPath  string
	verbose  bool
)

// wordsCmd represents the words command.
var wordsCmd = &cobra.Command{
	Use:   "words",
	Short: "",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running 'words' command.")
		filename = filepath.Clean(filename)

		exit := Exit()
		defer exit()

		var files []string

		// Check if filename is a directory
		stat, err := os.Stat(filename)
		if err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("cmd: %w", err)
		}
		if stat.IsDir() {
			// File represents a directory so append each screenshot file to files (with non image removal).
			logger.Info("wordsCmd: processing directory", "filename", filename)

			entries, err := os.ReadDir(filename)
			if err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("cmd: %w", err)
			}
			// Append image files only
			for _, e := range entries {
				if !e.IsDir() && isImageFile(e.Name()) {
					files = append(files, filepath.Join(filename+string(filepath.Separator)+e.Name()))
				}
			}
			logger.Info("wordsCmd: number of image files in a directory", "len(filename)", len(files))
		} else {
			files = append(files, filename)
		}
		outFile, err := os.OpenFile(filepath.Clean(outPath), os.O_APPEND|os.O_CREATE|os.O_RDWR, 0o600)
		if err != nil {
			fmt.Println(err)

			return fmt.Errorf("cmd: %w", err)
		}
		defer outFile.Close()

		// Clear outFile and reset to beginning
		if err := outFile.Truncate(0); err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("cmd: %w", err)
		}
		if _, err := outFile.Seek(0, 0); err != nil {
			logger.Error("wordsCmd", "err", err)

			return fmt.Errorf("cmd: %w", err)
		}

		// Process each screenshot file (header write + words recognition)
		for _, file := range files {
			if err := processScreenshot(file, outFile); err != nil {
				logger.Error("wordsCmd", "err", err)

				return fmt.Errorf("cmd: %w", err)
			}
		}

		return nil
	},
}

func init() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))

	rootCmd.AddCommand(wordsCmd)

	wordsCmd.PersistentFlags().StringVarP(&filename, "file", "f", "", "File to read words from")
	if err := wordsCmd.MarkPersistentFlagRequired("file"); err != nil {
		logger.Error("wordsCmd", "err", err.Error())
	}

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// wordsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	wordsCmd.Flags().BoolVarP(&save, "save", "s", false, "Save the output")
	wordsCmd.Flags().StringVarP(&outPath, "out", "o", "", "Output path")
	wordsCmd.MarkFlagsRequiredTogether("save", "out")

	wordsCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose")
}

func processScreenshot(filePath string, outFile *os.File) error {
	filePath = filepath.Clean(filePath)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	words, err := screenshot.RecognizeWords(content)
	if err != nil {
		return fmt.Errorf("failed to recognize words: %w", err)
	}

	if verbose {
		logger.Info("wordsCmd: recognized words", "file", filePath, "words", string(words))
	}

	if save {
		// Write '#filename + words'.
		// Header is a combination of # +'filename'.
		header := "#" + filepath.Base(filePath) + "\n"
		if _, err := outFile.WriteString(header); err != nil {
			return fmt.Errorf("failed to write header: %w", err)
		}
		if _, err := outFile.Write(words); err != nil {
			return fmt.Errorf("failed to write words: %w", err)
		}
		if _, err := outFile.WriteString("\n\n"); err != nil {
			return fmt.Errorf("failed to write newlines: %w", err)
		}
	}

	return nil
}

func isImageFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))

	return ext == ".png" || ext == ".jpg" || ext == ".jpeg"
}

func Exit(funcs ...func() error) func() error {
	return func() error {
		for _, f := range funcs {
			if err := f(); err != nil {
				return fmt.Errorf("onExit: %w", err)
			}
		}

		fmt.Println("Program is done.")
		os.Exit(1)

		return nil
	}
}
