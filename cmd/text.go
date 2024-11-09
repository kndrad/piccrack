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
	"time"

	"github.com/kndrad/itcrack/internal/screenshot"
	"github.com/kndrad/itcrack/internal/textproc"
	"github.com/kndrad/itcrack/pkg/openf"
	"github.com/spf13/cobra"
)

var InputPath string

// textCmd represents the words command.
var textCmd = &cobra.Command{
	Use:   "text",
	Short: "Extract text from image (screenshot) files (PNG/JPG/JPEG) using OCR",
	Long: `Extract text from image (screenshot) files (PNG/JPG/JPEG) using OCR
  -f, --file     Screenshot file or directory path to process (required)
  -o, --out      Output text file path (default: current directory)
  -v, --verbose  Enable verbose logging (default: true)`,
	SuggestFor: []string{
		"txt",
	},
	Example: "go run main.go text <path/to/file.png> -o <path/to/out/dir>",
	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			inputPath = filepath.Clean(args[0])
			outPath   = filepath.Clean(OutputPath)
		)

		var filePaths []string

		stat, err := os.Stat(inputPath)
		if err != nil {
			logger.Error("getting stat of screenshot", "err", err)

			return fmt.Errorf("stat: %w", err)
		}

		// Switched to true if inputPath points to a directory.
		addDirSuffix := false

		if stat.IsDir() {
			addDirSuffix = true

			logger.Info("Processing directory",
				slog.String("input_path", inputPath),
			)

			entries, err := os.ReadDir(inputPath)
			if err != nil {
				logger.Error("reading dir", "err", err)

				return fmt.Errorf("reading dir: %w", err)
			}
			// Append image files only
			for _, e := range entries {
				if !e.IsDir() && screenshot.IsImageFile(e.Name()) {
					filePaths = append(filePaths, filepath.Join(inputPath, "/", e.Name()))
				}
			}
			logger.Info(
				"Number of image files in a directory",
				slog.String("input_path", inputPath),
				slog.Int("files_total", len(filePaths)),
			)
		} else {
			// Only add input path if path was not a directory.
			filePaths = append(filePaths, inputPath)
		}

		// Add the suffix if addDirSuffix was changed to true.
		if addDirSuffix {
			suffix := "dir"
			id, err := textproc.NewAnalysisIDWithSuffix(suffix)
			if err != nil {
				logger.Error("Failed to add suffix to an out path",
					slog.String("suffix", suffix),
					slog.String("id", id),
				)
			}
		}

		ppath, err := openf.PreparePath(outPath, time.Now())
		if err != nil {
			logger.Error("Failed to prepare out path",
				slog.String("outPath", outPath),
				slog.String("err", err.Error()),
			)

			return fmt.Errorf("open file cleaned: %w", err)
		}

		txtFile, err := openf.Open(
			ppath.String(),
			os.O_APPEND|openf.DefaultFlags,
			openf.DefaultFileMode,
		)
		if err != nil {
			logger.Error("Failed to open cleaned file", "err", err)

			return fmt.Errorf("open file cleaned: %w", err)
		}

		// Process each screenshot and write output to .txt file.
		for _, path := range filePaths {
			content, err := os.ReadFile(path)
			if err != nil {
				logger.Error("reading file", "err", err)

				return fmt.Errorf("reading file: %w", err)
			}
			words, err := screenshot.RecognizeWords(content)
			if err != nil {
				logger.Error("Failed to recognize words in a screenshot content", "err", err)

				return fmt.Errorf("screenshot words recognition: %w", err)
			}
			w := textproc.NewWordsTextFileWriter(txtFile)
			if err := textproc.WriteWords(words, w); err != nil {
				logger.Error("Failed to write words to a txt file", "err", err)

				return fmt.Errorf("writing words: %w", err)
			}
		}

		logger.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(textCmd)

	textCmd.Flags().StringVarP(&OutputPath, "out", "o", "", "output path")
	textCmd.MarkFlagRequired("out")
}
