package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/kndrad/wcrack/cmd/logger"
	"github.com/kndrad/wcrack/pkg/ocr"
	"github.com/kndrad/wcrack/pkg/openf"
	"github.com/kndrad/wcrack/pkg/textproc"
	"github.com/spf13/cobra"
)

// textCmd represents the words command.
var textCmd = &cobra.Command{
	Use:   "text",
	Short: "Extract text from image (screenshot) files (PNG/JPG/JPEG) using OCR",
	SuggestFor: []string{
		"txt",
	},
	Example: "itcrack text --path <path/to/file.png> -o <path/to/out/dir>",
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.New(Verbose)

		path, err := cmd.Flags().GetString("path")
		if err != nil {
			l.Error("Failed to get path string flag value", "err", err)

			return fmt.Errorf("get string: %w", err)
		}
		path = filepath.Clean(path)

		stat, err := os.Stat(path)
		if err != nil {
			l.Error("getting stat of screenshot", "err", err)

			return fmt.Errorf("stat: %w", err)
		}

		// Switched to true if inputPath points to a directory.
		addDirSuffix := false

		var filePaths []string
		if stat.IsDir() {
			addDirSuffix = true

			l.Info("Processing directory",
				slog.String("input_path", path),
			)

			entries, err := os.ReadDir(path)
			if err != nil {
				l.Error("reading dir", "err", err)

				return fmt.Errorf("reading dir: %w", err)
			}
			// Append image files only
			for _, e := range entries {
				if !e.IsDir() {
					filePaths = append(filePaths, filepath.Join(path, "/", e.Name()))
				}
			}
			l.Info(
				"Number of image files in a directory",
				slog.String("input_path", path),
				slog.Int("files_total", len(filePaths)),
			)
		} else {
			// Only add input path if path was not a directory.
			filePaths = append(filePaths, path)
		}

		// Add the suffix if addDirSuffix was changed to true.
		if addDirSuffix {
			suffix := "dir"
			id, err := textproc.NewAnalysisIDWithSuffix(suffix)
			if err != nil {
				l.Error("Failed to add suffix to an out path",
					slog.String("suffix", suffix),
					slog.String("id", id),
				)
			}
		}

		outPath := filepath.Clean(OutPath)
		ppath, err := openf.PreparePath(outPath, time.Now())
		if err != nil {
			l.Error("Failed to prepare out path",
				slog.String("outPath", outPath),
				slog.String("err", err.Error()),
			)

			return fmt.Errorf("open file cleaned: %w", err)
		}

		txtF, err := openf.Open(
			ppath.String(),
			os.O_APPEND|openf.DefaultFlags,
			openf.DefaultFileMode,
		)
		if err != nil {
			l.Error("Failed to open cleaned file", "err", err)

			return fmt.Errorf("open file cleaned: %w", err)
		}

		c := ocr.NewClient()
		defer c.Close()

		if err != nil {
			l.Error("Failed to init tesseract client", "err", err)

			return fmt.Errorf("new client: %w", err)
		}

		// Process each screenshot and write output to .txt file.
		for _, path := range filePaths {
			result, err := ocr.Do(c, path)
			if err != nil {
				l.Error("Failed to recognize words in a screenshot content", "err", err)

				return fmt.Errorf("screenshot words recognition: %w", err)
			}
			w := textproc.NewFileWriter(txtF)
			if err := textproc.Write(w, []byte(result.String())); err != nil {
				l.Error("Failed to write words to a txt file", "err", err)

				return fmt.Errorf("writing words: %w", err)
			}
		}

		l.Info("Program completed successfully.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(textCmd)

	textCmd.Flags().String("path", "", "Path to image")
	textCmd.MarkFlagRequired("path")

	textCmd.Flags().StringVarP(&OutPath, "out", "o", ".", "output path")
	textCmd.MarkFlagRequired("out")
}
